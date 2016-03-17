package daemon

import (
	"fmt"
	"net"
	"os"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/shaping"
)

type ShapingEngine struct {
	thriftAddr *net.TCPAddr
	conf       *Config
	shaper     shaping.Shaper
	hooks      [][]Hook
}

func NewShapingEngine(thriftAddr *net.TCPAddr, conf *Config) (*ShapingEngine, error) {
	shaper, err := shaping.GetShaper()
	if err != nil {
		return nil, err
	}
	if err := shaper.Initialize(); err != nil {
		return nil, err
	}
	return buildShapingEngine(thriftAddr, shaper, conf)
}

// Separate than NewShapingEngine so that tests have an entrypoint.
func buildShapingEngine(thriftAddr *net.TCPAddr, shaper shaping.Shaper, conf *Config) (*ShapingEngine, error) {
	hooks := make([][]Hook, len(all_hook_types))
	for i := range all_hook_types {
		hooks[i] = make([]Hook, 0, 5)
	}
	for _, hook := range conf.Hooks {
		types, err := hook.GetTypes()
		if err != nil {
			return nil, err
		}
		for _, t := range types {
			hooks[t] = append(hooks[t], hook)
		}
	}
	return &ShapingEngine{
		thriftAddr: thriftAddr,
		conf:       conf,
		shaper:     shaper,
		hooks:      hooks,
	}, nil
}

func (eng *ShapingEngine) GetPlatform() atc_thrift.PlatformType {
	return eng.shaper.GetPlatform()
}

func (eng *ShapingEngine) CreateGroup(id int64, target shaping.Target) error {
	if err := eng.shaper.CreateGroup(id, target); err != nil {
		return err
	}
	return eng.runHooks(GROUP_JOIN, id, target)
}

func (eng *ShapingEngine) JoinGroup(id int64, target shaping.Target) error {
	if err := eng.shaper.JoinGroup(id, target); err != nil {
		return err
	}
	return eng.runHooks(GROUP_JOIN, id, target)
}

func (eng *ShapingEngine) LeaveGroup(id int64, target shaping.Target) error {
	if err := eng.shaper.LeaveGroup(id, target); err != nil {
		return err
	}
	return eng.runHooks(GROUP_LEAVE, id, target)
}

func (eng *ShapingEngine) DeleteGroup(id int64) error {
	return eng.shaper.DeleteGroup(id)
}

func (eng *ShapingEngine) Shape(id int64, settings *atc_thrift.Shaping) error {
	return eng.shaper.Shape(id, settings)
}

func (eng *ShapingEngine) Unshape(id int64) error {
	return eng.shaper.Unshape(id)
}

func (eng *ShapingEngine) runHooks(t HookType, id int64, addr shaping.Target) error {
	// Make sure the hook type looks up correctly
	if len(eng.hooks) >= int(t) {
		env := eng.buildHookEnv(t, addr)
		for _, hook := range eng.hooks[t] {
			if err := hook.Run(env, fmt.Sprintf("%d", id), addr.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (eng *ShapingEngine) buildHookEnv(t HookType, addr shaping.Target) []string {
	return []string{
		fmt.Sprintf("ATCD_ADDR=json://%s", eng.thriftAddr),
		fmt.Sprintf("ATCD_PATH=%s", os.Args[0]),
		fmt.Sprintf("ATC_MEMBER=%s", addr),
		fmt.Sprintf("ATC_HOOK_TYPE=%s", t.String()),
	}
}
