package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	thriftAddr := kingpin.Flag("thrift-addr", "thrift server address (env:ATCD_ADDR)").Short('T').Default(GetEnv("ATCD_ADDR", "127.0.0.1:9090")).TCP()
	thriftProto := kingpin.Flag("thrift-proto", "thrift server protocol (env:ATCD_PROTO)").Short('P').Default(GetEnv("ATCD_PROTO", "json")).String()

	var (
		defMember = GetEnv("ATC_MEMBER", "")

		// Duplicate flags/args
		token   string
		id      int64
		members []net.IP
		member  net.IP
	)

	kingpin.Command("info", "Prints info about the ATC shaping")

	group := kingpin.Command("group", "Manage atc shaping groups")

	group.Command("list", "List groups")

	groupAdd := group.Command("create", "create a group")
	if defMember != "" {
		groupAdd.Arg("member", "IP address of the member (env:ATC_MEMBER)").Default(defMember).IPVar(&member)
	} else {
		groupAdd.Arg("member", "IP address of the member (env:ATC_MEMBER)").Required().IPVar(&member)
	}

	groupShow := group.Command("show", "show info about a group")
	groupShow.Arg("id", "id of the group").Required().Int64Var(&id)

	groupJoin := group.Command("join", "leave a group")
	groupJoin.Flag("token", "token").Short('t').Default("").StringVar(&token)
	groupJoin.Arg("id", "id of the group").Required().Int64Var(&id)
	if defMember != "" {
		groupJoin.Arg("members", "IP address of the members (env:ATC_MEMBER)").Default(defMember).IPListVar(&members)
	} else {
		groupJoin.Arg("members", "IP address of the members (env:ATC_MEMBER)").Required().IPListVar(&members)
	}

	groupLeave := group.Command("leave", "leave a group")
	groupLeave.Flag("token", "token").Short('t').Default("").StringVar(&token)
	groupLeave.Arg("id", "id of the group").Required().Int64Var(&id)
	if defMember != "" {
		groupLeave.Arg("members", "IP address of the members (env:ATC_MEMBER)").Default(defMember).IPListVar(&members)
	} else {
		groupLeave.Arg("members", "IP address of the members (env:ATC_MEMBER)").Required().IPListVar(&members)
	}

	groupToken := group.Command("token", "get a token")
	groupToken.Arg("id", "id of the group").Required().Int64Var(&id)

	groupUnshape := group.Command("unshape", "remove shaping from a group")
	groupUnshape.Flag("token", "token").Short('t').Default("").StringVar(&token)
	groupUnshape.Arg("id", "id of the group").Required().Int64Var(&id)

	groupShape := group.Command("shape", "apply shaping to a group")
	groupShape.Flag("token", "token").Short('t').Default("").StringVar(&token)

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

	atcd = api.NewAtcdConn(*thriftAddr, *thriftProto)
	if err := atcd.Open(); err != nil {
		Log.Fatalln("Could not open connection to atcd:", err)
	}
	defer atcd.Close()

	switch cmd {
	case "info":
		ServerInfo()
	case "group create":
		GroupJoin(id, member, token)
	case "group show":
		GroupShow(id)
	case "group list":
		GroupList()
	case "group join":
		for _, member := range members {
			GroupJoin(id, member, token)
		}
	case "group leave":
		for _, member := range members {
			GroupLeave(id, member, token)
		}
	case "group token":
		GroupToken(id)
	case "group unshape":
		GroupUnshape(id, token)
	case "group shape":
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

func GroupAdd(member net.IP) {
	grp, err := atcd.CreateGroup(member.String())
	if err != nil {
		Log.Fatalln(err)
	}
	printShortGroup(grp)
}

func GroupShow(id int64) {
	grp, err := atcd.GetGroup(id)
	if err != nil {
		Log.Fatalln(err)
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

func GroupJoin(id int64, member net.IP, token string) {
	if err := atcd.JoinGroup(id, member.String(), token); err != nil {
		Log.Fatalf("Could not join group: %v", err)
	}
	grp, err := atcd.GetGroup(id)
	if err != nil {
		Log.Fatalln(err)
	}
	printShortGroup(grp)
}

func GroupLeave(id int64, member net.IP, token string) {
	if err := atcd.LeaveGroup(id, member.String(), token); err != nil {
		Log.Fatalln(err)
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

func printShortGroup(group *atc_thrift.ShapingGroup) {
	fmt.Printf("%d: %s\n", group.ID, strings.Join(group.Members, " "))
}

func printLongGroup(group *atc_thrift.ShapingGroup) {
	printShortGroup(group)
	if group.Shaping != nil {
		b, _ := json.Marshal(group.Shaping)
		var out bytes.Buffer
		json.Indent(&out, b, "    ", "  ")
		out.WriteTo(os.Stdout)
	}
}
