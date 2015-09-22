package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"runtime"
	"testing"

	"github.com/facebook/augmented-traffic-control/src/api/stub"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

// IP address for test server to bind to.
var ServerAddr = ""

// Client addresses to use to connect to atc.
// Need to provide two in order to simulate two different clients connecting
// to ATC.
// Loopback addresses are easy to use and can be hardcoded for this purpose
// since traffic originating at a particular loopback address has the same
// loopback address as the source.
var Addr1 = "::1"
var Addr2 = "127.0.0.1"

func TestGetsServerInfo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli := srv.client(Addr1)

	var info ServerInfo
	cli.GetJson(1, &info, "/info")

	if info.Atcd.Platform != atc_thrift.PlatformType_LINUX.String() {
		t.Error("Wrong platform type:", info.Atcd.Platform)
	}
	if info.Atcd.Version != "2.0-gostub" {
		t.Error("Wrong daemon version:", info.Atcd.Version)
	}
	if info.Api.Version != "2.0-go" {
		t.Error("Wrong daemon version:", info.Atcd.Version)
	}
}

func TestCreatesGroup(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli := srv.client(Addr1)

	var group atc_thrift.ShapingGroup
	cli.PostJson(1, nil, &group, "/group")

	if group.Id <= 0 {
		t.Error("Wrong group id:", group.Id)
	}
	if group.Shaping != nil {
		t.Error("New group is being shaped:", group.Shaping)
	}
	if len(group.Members) == 0 {
		t.Error("Group has no members:", len(group.Members))
	} else if group.Members[0] != cli.host() {
		t.Errorf("Group has wrong member: %q != %q", cli.host(), group.Members[0])
	}
}

func TestGetsToken(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli := srv.client(Addr1)

	var group atc_thrift.ShapingGroup
	cli.PostJson(1, nil, &group, url("/group"))

	var token Token
	cli.GetJson(1, &token, url("group", group.Id, "token"))

	if token.Token == "" {
		t.Errorf("Invalid token: %q", token.Token)
	}
}

func TestJoinsGroup(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli1 := srv.client(Addr1)
	cli2 := srv.client(Addr2)

	var group atc_thrift.ShapingGroup
	cli1.PostJson(1, nil, &group, "/group")
	var token Token
	cli1.GetJson(1, &token, url("group", group.Id, "token"))

	var resp MemberResponse
	cli2.PostJson(1, token, &resp, url("group", group.Id, "join"))

	if resp.Member != Addr2 {
		t.Errorf("Invalid member: %q != %q", Addr2, resp.Member)
	}
	if resp.Id != group.Id {
		t.Errorf("Invalid group ID: %d != %d", group.Id, resp.Id)
	}

	cli1.GetJson(1, &group, url("group", group.Id))

	checkSetContains(t, group.Members, Addr1, Addr2)
}

func TestLeavesGroup(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli1 := srv.client(Addr1)
	cli2 := srv.client(Addr2)

	var group atc_thrift.ShapingGroup
	cli1.PostJson(1, nil, &group, "/group")
	var token Token
	cli1.GetJson(1, &token, url("group", group.Id, "token"))

	var resp MemberResponse
	cli2.PostJson(1, token, &resp, url("group", group.Id, "join"))

	cli2.PostJson(1, token, &resp, url("group", group.Id, "leave"))

	if resp.Member != Addr2 {
		t.Errorf("Invalid member: %q != %q", Addr2, resp.Member)
	}
	if resp.Id != group.Id {
		t.Errorf("Invalid group ID: %d != %d", group.Id, resp.Id)
	}

	cli1.GetJson(1, &group, url("group", group.Id))

	checkSetContains(t, group.Members, Addr1)

}

func TestCreatesProfile(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli1 := srv.client(Addr1)

	req := ProfileRequest{
		Name: "asdf",
		Settings: &atc_thrift.Setting{
			Down: &atc_thrift.Shaping{
				Rate: 64,
			},
			Up: &atc_thrift.Shaping{
				Rate: 32,
			},
		},
	}
	var resp Profile
	cli1.PostJson(1, req, &resp, "/profile")

	if resp.Id <= 0 {
		t.Error("Wrong profile id:", resp.Id)
	}
	if resp.Settings.Down.Rate != 64 {
		t.Error("Wrong profile down rate:", resp.Settings.Down.Rate)
	}
}

func TestGetsProfiles(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli := srv.client(Addr1)

	cli.PostJson(1, ProfileRequest{Name: "foo", Settings: &atc_thrift.Setting{}}, nil, "/profile")
	cli.PostJson(1, ProfileRequest{Name: "bar", Settings: &atc_thrift.Setting{}}, nil, "/profile")

	var profiles Profiles
	cli.GetJson(1, &profiles, "/profile")

	if len(profiles.Profiles) != 2 {
		t.Error("Wrong number of profiles:", len(profiles.Profiles))
	}
	set := make(map[string]struct{})
	for _, prof := range profiles.Profiles {
		set[prof.Name] = struct{}{}
	}
	if _, ok := set["foo"]; !ok {
		t.Errorf("Profile wasn't returned: foo")
	}
	if _, ok := set["bar"]; !ok {
		t.Errorf("Profile wasn't returned: bar")
	}
}

func TestDeletesProfiles(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Cleanup()
	cli := srv.client(Addr1)

	cli.PostJson(1, ProfileRequest{Name: "foo", Settings: &atc_thrift.Setting{}}, nil, "/profile")
	var profile Profile
	cli.PostJson(1, ProfileRequest{Name: "bar", Settings: &atc_thrift.Setting{}}, &profile, "/profile")

	cli.Delete(1, url("profile", profile.Id))

	var profiles Profiles
	cli.GetJson(1, &profiles, "/profile")

	if len(profiles.Profiles) != 1 {
		t.Error("Wrong number of profiles:", len(profiles.Profiles))
	}
	set := make(map[string]struct{})
	for _, prof := range profiles.Profiles {
		set[prof.Name] = struct{}{}
	}
	if _, ok := set["foo"]; !ok {
		t.Errorf("Profile wasn't returned: foo")
	}
	if _, ok := set["bar"]; ok {
		t.Errorf("Profile wasn't deleted: bar")
	}
}

/**
*** Utilities
**/

type testServer struct {
	*Server
	t *testing.T
}

func newTestServer(t *testing.T) testServer {
	srv, err := ListenAndServe(net.JoinHostPort(ServerAddr, "0"), "", "", "sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	srv.Atcd = stub.NewFakeAtcd()

	// Yield to the server
	runtime.Gosched()

	return testServer{srv, t}
}

func (s testServer) Cleanup() {
	s.Atcd.Close()
	s.Atcd = nil
	s.Kill()
	return
}

func (s testServer) client(addr string) testClient {
	if addr == "" {
		addr = s.GetAddr()
	} else {
		_, port, _ := net.SplitHostPort(s.GetAddr())
		addr = net.JoinHostPort(addr, port)
	}
	return testClient{addr, s.t}
}

type testClient struct {
	addr string
	t    *testing.T
}

func (c testClient) Get(version int, url string) *http.Response {
	if url[0] != '/' {
		url = "/" + url
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/api/v%d%s", c.addr, version, url))
	if err != nil {
		c.t.Fatalf("Couldn't fetch url %q: %v", url, err)
	}
	return resp
}

func (c testClient) Delete(version int, url string) *http.Response {
	if url[0] != '/' {
		url = "/" + url
	}
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s/api/v%d%s", c.addr, version, url), &bytes.Buffer{})
	if err != nil {
		c.t.Fatalf("Couldn't fetch url %q: %v", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Couldn't fetch url %q: %v", url, err)
	}
	return resp
}

func (c testClient) Post(version int, request interface{}, url string) *http.Response {
	if url[0] != '/' {
		url = "/" + url
	}
	buf := &bytes.Buffer{}
	if request != nil {
		if err := json.NewEncoder(buf).Encode(request); err != nil {
			c.t.Fatalf("Could not encode json request for %q: %v", url, err)
		}
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/api/v%d%s", c.addr, version, url), "application/json", buf)
	if err != nil {
		c.t.Fatalf("Couldn't fetch url %q: %v", url, err)
	}
	return resp
}

func (c testClient) GetJson(version int, response interface{}, url string) {
	resp := c.Get(version, url)
	defer resp.Body.Close()
	if response != nil {
		err := json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			c.t.Fatalf("Could not decode json from %q: %v", c.url(resp), err)
		}
	}
}

func (c testClient) PostJson(version int, request, response interface{}, url string) {
	resp := c.Post(version, request, url)
	defer resp.Body.Close()
	if response != nil {
		err := json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			c.t.Fatalf("Could not decode json from %q: %v", c.url(resp), err)
		}
	}
}

func (c testClient) host() string {
	host, _, _ := net.SplitHostPort(c.addr)
	return host
}

func (c testClient) url(resp *http.Response) string {
	return resp.Request.URL.Path
}

func url(url_items ...interface{}) string {
	url := "/"
	for _, i := range url_items {
		url = path.Join(url, fmt.Sprint(i))
	}
	return path.Clean(url)
}

func checkSetContains(t *testing.T, actuals []string, expecteds ...string) {
	set := make(map[string]struct{})
	for _, item := range actuals {
		set[item] = struct{}{}
	}
	for _, item := range expecteds {
		if _, ok := set[item]; !ok {
			t.Errorf("Not a member of the set: %q", item)
		}
	}
}
