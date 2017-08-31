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
	// APIURLMap contains the mapping between the URLS and the HTTP Handlers
	APIURLMap = map[string]HandlerFunc{
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

// RedirectHandler is a handler builder which returns a function that sends an HTTP
// redirect to the given URL.
func RedirectHandler(url string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
		http.Redirect(w, r, url, http.StatusFound)
		return nil, NoStatus
	}
}

// InfoHandler is the HTTP handler for information operations.
func InfoHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet)
	switch r.Method {
	case http.MethodGet:
		serv := GetServer(r)
		atcd := GetAtcd(r)
		addr, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		daemonInfo, err := atcd.GetAtcdInfo()
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
				Platform: daemonInfo.Platform.String(),
				Version:  daemonInfo.Version,
			},
			Client: ClientInfo{
				Addr:      addr.String(),
				Primary:   p,
				Secondary: s,
			},
		}
		return info, nil
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

// GroupsHandler is the HTTP endpoint for handling Group requests
func GroupsHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet, http.MethodPost)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodPost:
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
	case http.MethodGet:
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
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

// GroupHandler if the HTTP Handler for the Group requests.
func GroupHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodGet:
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
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

// GroupJoinHandler is the HTTP Hanlder for joining a member to a group
func GroupJoinHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodPost)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodPost:
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		reqInfo := &Token{}
		if err = json.NewDecoder(r.Body).Decode(reqInfo); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		member, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		err = atcd.JoinGroup(id, member.String(), reqInfo.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
		}
		return MemberResponse{
			Id:     id,
			Member: member.String(),
		}, nil
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

// GroupLeaveHandler is the HTTP Handler for removing a member from a group
func GroupLeaveHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodPost)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodPost:
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		reqInfo := &Token{}
		if err = json.NewDecoder(r.Body).Decode(reqInfo); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		member, cerr := GetClientAddr(r)
		if cerr != nil {
			return nil, cerr
		}
		err = atcd.LeaveGroup(id, member.String(), reqInfo.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
		}
		return MemberResponse{
			Id:     id,
			Member: member.String(),
		}, nil
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

// GroupTokenHandler is the HTTP Handler for reading group tokens
func GroupTokenHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodGet:
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
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupShapeHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	atcd := GetAtcd(r)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
	}
	switch r.Method {
	case http.MethodGet:
		group, err := atcd.GetGroup(id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get shaping from atcd: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: group.Shaping,
		}, nil
	case http.MethodPost:
		reqInfo := &TokenShaping{}
		if err := json.NewDecoder(r.Body).Decode(reqInfo); err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		setting, err := atcd.ShapeGroup(id, reqInfo.Shaping, reqInfo.Token)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not shape: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: setting,
		}, nil
	case http.MethodDelete:
		return GroupShapeDelete(w, r, atcd, id)
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func GroupShapeDelete(w http.ResponseWriter, r *http.Request, atcd atc_thrift.Atcd, id int64) (interface{}, HttpError) {
	reqInfo := &Token{}
	if err := json.NewDecoder(r.Body).Decode(reqInfo); err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	err := atcd.UnshapeGroup(id, reqInfo.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
	}
	return nil, nil
}

// ShapeHandler is the HTTP method for operating on shaping profiles
func ShapeHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	atcd := GetAtcd(r)
	switch r.Method {
	case http.MethodGet:
		return getSimpleShaping(atcd, w, r)
	case http.MethodPost:
		return createSimpleShaping(atcd, w, r)
	case http.MethodDelete:
		return deleteSimpleShaping(atcd, w, r)
	case http.MethodOptions:
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
	reqInfo := &TokenShaping{}
	if err = json.NewDecoder(r.Body).Decode(reqInfo); err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	// This is allowed since the requestor is shaping their own device!
	if reqInfo.Token == "" {
		reqInfo.Token, err = atcd.GetGroupToken(group.Id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	setting, err := atcd.ShapeGroup(group.Id, reqInfo.Shaping, reqInfo.Token)
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
	reqInfo := &Token{}
	err = json.NewDecoder(r.Body).Decode(reqInfo)
	if !(err == nil || err == io.EOF) {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	// This is allowed since the requestor is shaping their own device!
	if reqInfo.Token == "" {
		reqInfo.Token, err = atcd.GetGroupToken(group.Id)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	err = atcd.UnshapeGroup(group.Id, reqInfo.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
	}
	return nil, nil
}

func ProfilesHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodGet, http.MethodPost)
	db := GetDB(r)
	switch r.Method {
	case http.MethodGet:
		profiles := <-db.GetProfiles()
		if profiles == nil {
			return nil, HttpErrorf(http.StatusInternalServerError, "Couldn't load profiles from database")
		}
		return Profiles{profiles}, nil
	case http.MethodPost:
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
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, http.MethodDelete)
	switch r.Method {
	case http.MethodDelete:
		id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil {
			return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get Id from url: %v", err)
		}
		db := GetDB(r)
		db.DeleteProfile(id)
		return nil, nil
	case http.MethodOptions:
		return nil, nil
	default:
		return nil, InvalidMethod(r)
	}
}
