package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/facebook/augmented-traffic-control/src/api"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	atclog "github.com/facebook/augmented-traffic-control/src/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Log *log.Logger

	atcd *api.AtcdConn
)

func init() {
	Log = atclog.Stdlog()
}

func GetEnv(name, def string) string {
	if value := os.Getenv(name); value == "" {
		return def
	} else {
		return value
	}
}

func main() {
	thriftUrl := kingpin.Flag("thrift-addr", "thrift server url in the format 'proto://host:port' (env:ATCD_ADDR)").Short('r').
		Default("json://127.0.0.1:9090").Envar("ATCD_ADDR").URL()

	var (
		defMember = GetEnv("ATC_MEMBER", "")

		// Duplicate flags/args
		id      int64
		members []string
		member  string
	)

	kingpin.Command("info", "Prints info about the ATC shaping")

	kingpin.Command("list", "List groups")

	groupMembers := kingpin.Command("members", "List members of a group")
	groupMembers.Arg("id", "id of the group").Default("-1").Int64Var(&id)

	groupAdd := kingpin.Command("create", "create a group")
	if defMember != "" {
		groupAdd.Arg("member", "IP address of the member (env:ATC_MEMBER)").Default(defMember).StringVar(&member)
	} else {
		groupAdd.Arg("member", "IP address of the member (env:ATC_MEMBER)").Required().StringVar(&member)
	}

	groupFind := kingpin.Command("find", "find the group containing a member")
	if defMember != "" {
		groupFind.Arg("member", "IP address of the member (env:ATC_MEMBER)").Default(defMember).StringVar(&member)
	} else {
		groupFind.Arg("member", "IP address of the member (env:ATC_MEMBER)").Required().StringVar(&member)
	}

	groupShow := kingpin.Command("show", "show info about a group")
	groupShow.Arg("id", "id of the group").Required().Int64Var(&id)

	groupJoin := kingpin.Command("join", "leave a group")
	groupJoin.Arg("id", "id of the group").Required().Int64Var(&id)
	if defMember != "" {
		groupJoin.Arg("members", "IP address of the members (env:ATC_MEMBER)").Default(defMember).StringsVar(&members)
	} else {
		groupJoin.Arg("members", "IP address of the members (env:ATC_MEMBER)").Required().StringsVar(&members)
	}

	groupLeave := kingpin.Command("leave", "leave a group")
	groupLeave.Arg("id", "id of the group").Required().Int64Var(&id)
	if defMember != "" {
		groupLeave.Arg("members", "IP address of the members (env:ATC_MEMBER)").Default(defMember).StringsVar(&members)
	} else {
		groupLeave.Arg("members", "IP address of the members (env:ATC_MEMBER)").Required().StringsVar(&members)
	}

	groupToken := kingpin.Command("token", "get a token")
	groupToken.Arg("id", "id of the group").Required().Int64Var(&id)

	groupUnshape := kingpin.Command("unshape", "remove shaping from a group")
	groupUnshape.Arg("id", "id of the group").Required().Int64Var(&id)

	groupShape := kingpin.Command("shape", "apply shaping to a group")

	// Uplink shaping
	upRate := groupShape.Flag("up.rate", "uplink rate in Kb/s").Default("0").Int32()
	upDelay := groupShape.Flag("up.delay", "uplink delay in ms").Default("0").Int32()
	upLoss := groupShape.Flag("up.loss", "uplink loss percentage").Default("0").Float64()

	// Downlink shaping
	dnRate := groupShape.Flag("dn.rate", "downlink rate in Kb/s").Default("0").Int32()
	dnDelay := groupShape.Flag("dn.delay", "downlink delay in ms").Default("0").Int32()
	dnLoss := groupShape.Flag("dn.loss", "downlink loss percentage").Default("0").Float64()

	groupShape.Arg("id", "id of the group").Required().Int64Var(&id)

	cmd := kingpin.Parse()

	atcd = api.NewAtcdConn(*thriftUrl)
	if err := atcd.Open(); err != nil {
		Log.Fatalln("Could not open connection to atcd:", err)
	}
	defer atcd.Close()

	switch cmd {
	case "info":
		ServerInfo()
	case "create":
		GroupAdd(member)
	case "find":
		GroupFind(member)
	case "show":
		GroupShow(id)
	case "list":
		GroupList()
	case "members":
		GroupMembers(id)
	case "join":
		token, err := atcd.GetGroupToken(id)
		if err != nil {
			Log.Fatalln("Could not get group token:", err)
		}
		for _, member := range members {
			GroupJoin(id, member, token)
		}
	case "leave":
		token, err := atcd.GetGroupToken(id)
		if err != nil {
			Log.Fatalln("Could not get group token:", err)
		}
		for _, member := range members {
			GroupLeave(id, member, token)
		}
	case "token":
		GroupToken(id)
	case "unshape":
		token, err := atcd.GetGroupToken(id)
		if err != nil {
			Log.Fatalln("Could not get group token:", err)
		}
		GroupUnshape(id, token)
	case "shape":
		token, err := atcd.GetGroupToken(id)
		if err != nil {
			Log.Fatalln("Could not get group token:", err)
		}
		GroupShape(id, token, &atc_thrift.Shaping{
			Up: &atc_thrift.LinkShaping{
				Rate:  *upRate,
				Delay: &atc_thrift.Delay{Delay: *upDelay},
				Loss:  &atc_thrift.Loss{Percentage: *upLoss},
			},
			Down: &atc_thrift.LinkShaping{
				Rate:  *dnRate,
				Delay: &atc_thrift.Delay{Delay: *dnDelay},
				Loss:  &atc_thrift.Loss{Percentage: *dnLoss},
			},
		})
	default:
		Log.Fatalln("unknown command:", cmd)
	}
}

func ServerInfo() {
	info, err := atcd.GetAtcdInfo()
	if err != nil {
		Log.Fatalln(err)
	}
	fmt.Printf("atcd %v %v\n", info.Version, info.Platform)
}

func GroupAdd(member string) {
	grp, err := atcd.CreateGroup(member)
	if err != nil {
		Log.Fatalln(err)
	}
	printShortGroup(grp)
}

func GroupFind(member string) {
	grp, err := atcd.GetGroupWith(member)
	if err != nil {
		if api.IsNoSuchItem(err) {
			Log.Fatalf("Could not find group for member %q", member)
		} else {
			Log.Fatalln(err)
		}
	}
	printShortGroup(grp)
}

func GroupShow(id int64) {
	grp, err := atcd.GetGroup(id)
	if err != nil {
		Log.Fatalf("Could not find group (%d): %v", id, err)
	}
	printLongGroup(grp)
}

func GroupList() {
	grps, err := atcd.ListGroups()
	if err != nil {
		Log.Fatalln(err)
	}
	for _, grp := range grps {
		printShortGroup(grp)
	}
}

func GroupMembers(id int64) {
	if id != -1 {
		grp, err := atcd.GetGroup(id)
		if err != nil {
			Log.Fatalf("Could not find group (%d): %v", id, err)
		}
		listGroupMembers(grp)

	} else {
		grps, err := atcd.ListGroups()
		if err != nil {
			Log.Fatalln(err)
		}
		for _, grp := range grps {
			listGroupMembers(grp)
		}
	}
}

func GroupJoin(id int64, member string, token string) {
	if err := atcd.JoinGroup(id, member, token); err != nil {
		Log.Fatalf("Could not join group: %v", err)
	}
	grp, err := atcd.GetGroup(id)
	if err != nil {
		Log.Fatalln(err)
	}
	printShortGroup(grp)
}

func GroupLeave(id int64, member string, token string) {
	if err := atcd.LeaveGroup(id, member, token); err != nil {
		Log.Fatalf("Could not leave group: %v", err)
	}
}

func GroupToken(id int64) {
	token, err := atcd.GetGroupToken(id)
	if err != nil {
		Log.Fatalln(err)
	}
	fmt.Println(token)
}

func GroupShape(id int64, token string, shaping *atc_thrift.Shaping) {
	_, err := atcd.ShapeGroup(id, shaping, token)
	if err != nil {
		Log.Fatalln(err)
	}
	grp, err := atcd.GetGroup(id)
	if err != nil {
		Log.Fatalln(err)
	}
	printLongGroup(grp)
}

func GroupUnshape(id int64, token string) {
	err := atcd.UnshapeGroup(id, token)
	if err != nil {
		Log.Fatalln(err)
	}
}

func listGroupMembers(grp *atc_thrift.ShapingGroup) {
	for _, m := range grp.Members {
		fmt.Println(m)
	}
}

func printShortGroup(group *atc_thrift.ShapingGroup) {
	fmt.Printf("%d: %s\n", group.Id, strings.Join(group.Members, " "))
}

func printLongGroup(group *atc_thrift.ShapingGroup) {
	printShortGroup(group)
	if group.Shaping != nil {
		b, _ := json.Marshal(group.Shaping)
		var out bytes.Buffer
		json.Indent(&out, b, "    ", "  ")
		// json.Indent doesn't pad the first line for some reason.
		fmt.Print("    ")
		out.WriteTo(os.Stdout)
		fmt.Println()
	}
}
