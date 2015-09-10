package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/facebook/augmented-traffic-control/atc/atc_thrift"
	"github.com/gorilla/mux"
)

var (
	URL_MAP = map[string]map[string]HandlerFunc{
		"/": {
			"/": RedirectHandler("/api/v1/shape"),
		},
		"/api/v1": {
			"/info":             InfoHandler,
			"/group":            GroupsHandler,
			"/group/{id}":       GroupHandler,
			"/group/{id}/join":  GroupJoinHandler,
			"/group/{id}/leave": GroupLeaveHandler,
			"/group/{id}/token": GroupTokenHandler,
			"/group/{id}/shape": GroupShapeHandler,
			"/shape":            ShapeHandler,
		},
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
		return nil, Errorf(http.StatusBadGateway, "Could not communicate with ATC Daemon: %v", err)
	}
	info := ServerInfo{
		Api: GetApiInfo(),
		Atcd: DaemonInfo{
			Platform: daemon_info.Platform.String(),
			Version:  daemon_info.Version,
		},
	}
	return info, nil
}

func GroupsHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	if r.Method != "POST" {
		return nil, InvalidMethod
	}
	grp, err := atcd.CreateGroup(GetClientAddr(r))
	if err != nil {
		return nil, Errorf(http.StatusBadGateway, "Could not create group: %v", err)
	}
	return grp, nil
}

func GroupHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	if r.Method != "GET" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	group, err := atcd.GetGroup(id)
	if err != nil {
		if IsNoSuchItem(err) {
			return nil, Errorf(http.StatusNotFound, "Invalid group")
		}
		return nil, Errorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
	}
	return group, nil
}

func GroupJoinHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	if r.Method != "POST" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	req_info := &MemberTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	if req_info.Member == "" {
		req_info.Member = GetClientAddr(r)
	}
	// FIXME: maybe we need to check auth here?
	err = atcd.JoinGroup(id, req_info.Member, req_info.Token)
	if err != nil {
		return nil, Errorf(http.StatusBadGateway, "Could not join group: %v", err)
	}
	return MemberResponse{
		Id:     id,
		Member: req_info.Member,
	}, nil
}

func GroupLeaveHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	if r.Method != "POST" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	req_info := &MemberTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
	}
	if req_info.Member == "" {
		req_info.Member = GetClientAddr(r)
	}
	// FIXME: maybe we need to check auth here?
	err = atcd.LeaveGroup(id, req_info.Member, req_info.Token)
	if err != nil {
		return nil, Errorf(http.StatusBadGateway, "Could not join group: %v", err)
	}
	return MemberResponse{
		Id:     id,
		Member: req_info.Member,
	}, nil
}

func GroupTokenHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	if r.Method != "GET" {
		return nil, InvalidMethod
	}
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	grp, err := atcd.GetGroupWith(GetClientAddr(r))
	if err != nil {
		if IsNoSuchItem(err) {
			return nil, Errorf(http.StatusUnauthorized, "Invalid group")
		}
		return nil, Errorf(http.StatusBadGateway, "Could not get group from daemon: %v", err)
	}
	if grp.Id != id {
		return nil, Errorf(http.StatusUnauthorized, "Invalid group")
	}
	token, err := atcd.GetGroupToken(id)
	if err != nil {
		return nil, Errorf(http.StatusBadGateway, "Could not get token from daemon: %v", err)
	}

	return GroupToken{
		Token: token,
		Id:    id,
	}, nil
}

func GroupShapeHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, Errorf(http.StatusNotAcceptable, "Could not get ID from url: %v", err)
	}
	switch r.Method {
	case "GET":
		group, err := atcd.GetGroup(id)
		if err != nil {
			return nil, Errorf(http.StatusBadGateway, "Could not get shaping from atcd: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: group.Shaping,
		}, nil
	case "POST":
		req_info := &TokenShaping{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		setting, err := atcd.ShapeGroup(id, req_info.Shaping, req_info.Token)
		if err != nil {
			return nil, Errorf(http.StatusBadGateway, "Could not shape: %v", err)
		}
		return GroupShaping{
			Id:      id,
			Shaping: setting,
		}, nil
	case "DELETE":
		req_info := &Token{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		err = atcd.UnshapeGroup(id, req_info.Token)
		if err != nil {
			return nil, Errorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
		}
		return nil, nil
	default:
		return nil, InvalidMethod
	}
}

func ShapeHandler(atcd atc_thrift.Atcd, w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
	addr := GetClientAddr(r)
	group, err := atcd.GetGroupWith(addr)
	if err != nil {
		return nil, Errorf(http.StatusNotFound, "Address not being shaped")
	}
	switch r.Method {
	case "GET":
		return GroupShaping{
			Id:      group.Id,
			Shaping: group.Shaping,
		}, nil
	case "POST":
		req_info := &TokenShaping{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		setting, err := atcd.ShapeGroup(group.Id, req_info.Shaping, req_info.Token)
		if err != nil {
			return nil, Errorf(http.StatusBadGateway, "Could not shape: %v", err)
		}
		return GroupShaping{
			Id:      group.Id,
			Shaping: setting,
		}, nil
	case "DELETE":
		req_info := &Token{}
		if err := json.NewDecoder(r.Body).Decode(req_info); err != nil {
			return nil, Errorf(http.StatusNotAcceptable, "Could not parse json from request: %v", err)
		}
		err = atcd.UnshapeGroup(group.Id, req_info.Token)
		if err != nil {
			return nil, Errorf(http.StatusBadGateway, "Could not delete shaping from atcd: %v", err)
		}
		return nil, nil
	default:
		return nil, InvalidMethod
	}
}
