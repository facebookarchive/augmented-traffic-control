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
	}
)

func RedirectHandler(url string) HandlerFunc {
	return func(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
		http.Redirect(w, r, url, http.StatusFound)
		return nil, NoStatus
	}
}

func InfoHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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

func GroupsHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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
			return nil, HttpErrorf(http.StatusNotFound, "No group found")
		}
		return group, nil
	default:
		return nil, InvalidMethod
	}

}

func GroupHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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

func GroupJoinHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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

func GroupLeaveHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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

func GroupTokenHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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
}

func GroupShapeHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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

func ShapeHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
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
		return nil, HttpErrorf(http.StatusNotFound, "Address not being shaped")
	}
	return GroupShaping{
		Id:      group.Id,
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
