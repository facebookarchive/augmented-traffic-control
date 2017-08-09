package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/gorilla/mux"
)

var (
	API_URL_MAP = map[string]HandlerFunc{
		"/":                 RedirectHandler(ROOT_URL + "/shape"),
		"/info":             InfoHandler,
		"/group":            GroupsHandler,
		"/group/{id}":       GroupHandler,
		"/group/{id}/join":  GroupJoinHandler,
		"/group/{id}/leave": GroupLeaveHandler,
		"/group/{id}/token": GroupTokenHandler,
		"/group/{id}/shape": GroupShapeHandler,
		"/shape":            ShapeHandler,
		"/profile":          ProfilesHandler,
		"/profile/{id}":     ProfileHandler,
	}
)

func RedirectHandler(url string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
		http.Redirect(w, r, url, http.StatusFound)
		return nil, NoStatus
	}
}

func InfoHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET")
	switch r.Method {
	case "GET":
		serv := GetServer(r)
		atcd := GetAtcd(r)
		addr, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		daemon_info, err := atcd.GetAtcdInfo()
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not communicate with ATC Daemon: %v", err)
		}
		p, s, cerr := serv.bind_info.getPrimarySecondaryAddrs(r)
		if cerr != nil {
			return nil, cerr
		}
		info := ServerInfo{
			Api: serv.GetInfo(r),
			Atcd: DaemonInfo{
				Platform: daemon_info.Platform.String(),
				Version:  daemon_info.Version,
			},
			Client: ClientInfo{
				Addr:      addr.String(),
				Primary:   p,
				Secondary: s,
			},
		}
		return info, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupsHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST")
	atcd := GetAtcd(r)
	switch r.Method {
	case "POST":
		addr, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		grp, err := atcd.CreateGroup(addr.String())
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not create group: %v", err)
		}
		token, err := atcd.GetGroupToken(grp.Id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get group token: %v", err)
		}
		return CreatedGroup{grp, token}, nil
	case "GET":
		addr, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		group, err := atcd.GetGroupWith(addr.String())
		if err != nil {
			// No group found.
			return nil, nil
		}
		return group, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET")
	atcd := GetAtcd(r)
	switch r.Method {
	case "GET":
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		group, err := atcd.GetGroup(id)
		if err != nil {
			if IsNoSuchItem(err) {
				return nil, HttpErrorf(http.StatusNotFound, "Invalid group")
			}
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
		}
		return group, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupJoinHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "POST")
	atcd := GetAtcd(r)
	switch r.Method {
	case "POST":
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		req_info := &Token{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		member, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		err = atcd.JoinGroup(id, member.String(), req_info.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
		}
		return MemberResponse{
			Id:     id,
			Member: member.String(),
		}, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupLeaveHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "POST")
	atcd := GetAtcd(r)
	switch r.Method {
	case "POST":
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		req_info := &Token{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		member, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		err = atcd.LeaveGroup(id, member.String(), req_info.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
		}
		return MemberResponse{
			Id:     id,
			Member: member.String(),
		}, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupTokenHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET")
	atcd := GetAtcd(r)
	switch r.Method {
	case "GET":
		addr, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		grp, err := atcd.GetGroupWith(addr.String())
		if err != nil {
			if IsNoSuchItem(err) {
				return nil, HttpErrorf(http.StatusUnauthorized, "Invalid group")
			}
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
		}
		if grp.Id != id {
			return nil, HttpErrorf(http.StatusUnauthorized, "Invalid group")
		}
		token, err := atcd.GetGroupToken(id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}

		return GroupToken{
			Token: token,
			Id:    id,
		}, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupShapeHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST", "DELETE")
	atcd := GetAtcd(r)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
	}
	switch r.Method {
	case "GET":
		group, err := atcd.GetGroup(id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get shaping from atcd: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: group.Shaping,
		}, nil
	case "POST":
		req_info := &TokenShaping{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		setting, err := atcd.ShapeGroup(id, req_info.Shaping, req_info.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not shape: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: setting,
		}, nil
	case "DELETE":
		req_info := &Token{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		err = atcd.UnshapeGroup(id, req_info.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
		}
		return nil, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func ShapeHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST", "DELETE")
	atcd := GetAtcd(r)
	switch r.Method {
	case "GET":
		return getSimpleShaping(atcd, w, r)
	case "POST":
		return createSimpleShaping(atcd, w, r)
	case "DELETE":
		return deleteSimpleShaping(atcd, w, r)
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func getSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr, cerr := GetClientAddr(r)
	if cerr != nil {
		return nil, cerr
	}
	group, err := atcd.GetGroupWith(addr.String())
	if err != nil {
		// Not being shaped
		return nil, nil
	}
	return GroupShaping{
		Id:      group.Id,
		Shaping: group.Shaping,
	}, nil
}

func createSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr, cerr := GetClientAddr(r)
	if cerr != nil {
		return nil, cerr
	}
	group, err := atcd.GetGroupWith(addr.String())
	if err != nil {
		group, err = atcd.CreateGroup(addr.String())
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not create group: %v", err)
		}
	}
	req_info := &TokenShaping{}
	if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	// This is allowed since the requestor is shaping their own device!
	if req_info.Token == "" {
		req_info.Token, err = atcd.GetGroupToken(group.Id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	setting, err := atcd.ShapeGroup(group.Id, req_info.Shaping, req_info.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not shape: %v", err)
	}
	return GroupShaping{
		Id:      group.Id,
		Shaping: setting,
	}, nil
}

func deleteSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr, cerr := GetClientAddr(r)
	if cerr != nil {
		return nil, cerr
	}
	group, err := atcd.GetGroupWith(addr.String())
	if err != nil {
		return nil, HttpErrorf(http.StatusNotFound, "Address not being shaped")
	}
	req_info := &Token{}
	err = json.NewDecoder(r.Body).Decode(req_info)
	if !(err == nil || err == io.EOF) {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	// This is allowed since the requestor is shaping their own device!
	if req_info.Token == "" {
		req_info.Token, err = atcd.GetGroupToken(group.Id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	err = atcd.UnshapeGroup(group.Id, req_info.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
	}
	return nil, nil
}

func ProfilesHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST")
	db := GetDB(r)
	switch r.Method {
	case "GET":
		profiles := <-db.GetProfiles()
		if profiles == nil {
			return nil, HttpErrorf(http.StatusInternalServerError, "Couldn't load profiles from database")
		}
		return Profiles{profiles}, nil
	case "POST":
		var p ProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		if p.Name == "" {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Mandatory field 'name' not provided")
		} else if p.Shaping == nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Mandatory field 'settings' not provided")
		}
		prof := <-db.UpdateProfile(Profile{
			Id:      -1,
			Name:    p.Name,
			Shaping: p.Shaping,
		})
		if prof == nil {
			return nil, HttpErrorf(http.StatusInternalServerError, "Couldn't save profile to database")
		}
		return prof, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "DELETE")
	switch r.Method {
	case "DELETE":
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		db := GetDB(r)
		db.DeleteProfile(id)
		return nil, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}
