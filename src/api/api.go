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
		"/":                 RedirectHandler(ServerData.ApiUrl + "shape"),
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
	atcd := GetAtcd(r)
	daemon_info, err := atcd.GetAtcdInfo()
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not communicate with ATC Daemon: %v", err)
	}
	info := ServerInfo{
		Api: APIInfo{Version: VERSION},
		Atcd: DaemonInfo{
			Platform: daemon_info.Platform.String(),
			Version:  daemon_info.Version,
		},
	}
	return info, nil
}

func GroupsHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST")
	atcd := GetAtcd(r)
	switch r.Method {
	case "POST":
		grp, err := atcd.CreateGroup(GetClientAddr(r))
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not create group: %v", err)
		}
		return grp, nil
	case "GET":
		addr := GetClientAddr(r)
		group, err := atcd.GetGroupWith(addr)
		if err != nil {
			// No group found.
			return nil, nil
		}
		return group, nil
	case "OPTIONS":
		return nil, nil
	default:
		return nil, InvalidMethod
	}
}

func GroupHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET")
	atcd := GetAtcd(r)
	if r.Method == "OPTIONS" {
		return nil, nil
	}
	if r.Method != "GET" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	group, err := atcd.GetGroup(id)
	if err != nil {
		if IsNoSuchItem(err) {
			return nil, HttpErrorf(http.StatusNotFound, "Invalid group")
		}
		return nil, HttpErrorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
	}
	return group, nil
}

func GroupJoinHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "POST")
	atcd := GetAtcd(r)
	if r.Method == "OPTIONS" {
		return nil, nil
	}
	if r.Method != "POST" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	req_info := &Token{}
	if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	member := GetClientAddr(r)
	err = atcd.JoinGroup(id, member, req_info.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
	}
	return MemberResponse{
		Id:     id,
		Member: member,
	}, nil
}

func GroupLeaveHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "POST")
	atcd := GetAtcd(r)
	if r.Method == "OPTIONS" {
		return nil, nil
	}
	if r.Method != "POST" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	req_info := &Token{}
	if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	member := GetClientAddr(r)
	err = atcd.LeaveGroup(id, member, req_info.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not join group: %v", err)
	}
	return MemberResponse{
		Id:     id,
		Member: member,
	}, nil
}

func GroupTokenHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET")
	atcd := GetAtcd(r)
	if r.Method != "GET" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	grp, err := atcd.GetGroupWith(GetClientAddr(r))
	if err != nil {
		if IsNoSuchItem(err) {
			return nil, HttpErrorf(http.StatusUnauthorized, "Invalid group")
		}
		return nil, HttpErrorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
	}
	if grp.ID != id {
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
}

func GroupShapeHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "GET", "POST", "DELETE")
	atcd := GetAtcd(r)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
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
	default:
		return nil, InvalidMethod
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
	default:
		return nil, InvalidMethod
	}
}

func getSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr := GetClientAddr(r)
	group, err := atcd.GetGroupWith(addr)
	if err != nil {
		// Not being shaped
		return nil, nil
	}
	return GroupShaping{
		Id:      group.ID,
		Shaping: group.Shaping,
	}, nil
}

func createSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr := GetClientAddr(r)
	group, err := atcd.GetGroupWith(addr)
	if err != nil {
		group, err = atcd.CreateGroup(addr)
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
		req_info.Token, err = atcd.GetGroupToken(group.ID)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	setting, err := atcd.ShapeGroup(group.ID, req_info.Shaping, req_info.Token)
	if err != nil {
		return nil, HttpErrorf(http.StatusBadGateway, "Could not shape: %v", err)
	}
	return GroupShaping{
		Id:      group.ID,
		Shaping: setting,
	}, nil
}

func deleteSimpleShaping(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr := GetClientAddr(r)
	group, err := atcd.GetGroupWith(addr)
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
		req_info.Token, err = atcd.GetGroupToken(group.ID)
		if err != nil {
			return nil, HttpErrorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
		}
	}
	err = atcd.UnshapeGroup(group.ID, req_info.Token)
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
	default:
		return nil, InvalidMethod
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	CORS(w, "DELETE")
	if r.Method != "DELETE" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, HttpErrorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	db := GetDB(r)
	db.DeleteProfile(id)
	return nil, nil
}
