package daemon

import (
	"fmt"
	"os/exec"
	"strings"
)

type Config struct {
	Hooks []Hook `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

type HookType int

const (
	GROUP_JOIN HookType = iota
	GROUP_LEAVE
	NEVER
)

var (
	all_hook_types = []HookType{
		GROUP_JOIN,
		GROUP_LEAVE,
		// NEVER not included intentionally.
	}
)

type Hook struct {
	// Name of the hook
	Name string `json:"name" yaml:"name"`

	// String forms of the hook types.
	// This is represented internally as a bitmask. Don't use
	// this field directly. Use `hook.Type` instead.
	When []string `json:"when" yaml:"when"`

	// Shell command to be run. See hook documentation for details.
	Command string `json:"command" yaml:"command"`

	// Should the hook be run asynchronously
	Async bool `json:"async,omitempty" yaml:"aync,omitempty"`

	// Should the hook be required to succeed
	SuccessRequired bool `json:"success_required,omitempty" yaml:"success_required,omitempty"`
}

func (c Hook) GetTypes() ([]HookType, error) {
	if c.Async && c.SuccessRequired {
		return nil, fmt.Errorf("Hook (%v) cannot be both async and success_required", c.Name)
	}
	hookTypes := []HookType{}
	for _, p := range c.When {
		switch p {
		case "group.join":
			hookTypes = append(hookTypes, GROUP_JOIN)
		case "group.leave":
			hookTypes = append(hookTypes, GROUP_LEAVE)
		case "never":
			// Do nothing
		default:
			return nil, fmt.Errorf("Unknown hook position: %s", p)
		}
	}
	return hookTypes, nil
}

// Runs the hook with the provided arguments.
// Returns an error IFF SuccessRequired and the command fails.
// Forks a goroutine if hook.Async
func (hook Hook) Run(args []string) error {
	if hook.Async {
		go hook.doHook(args)
		return nil
	}
	err := hook.doHook(args)
	if hook.SuccessRequired && err != nil {
		return err
	}
	return nil
}

// Runs the hook's command with the given arguments in this goroutine.
// Logs and returns an error if the command fails.
func (hook Hook) doHook(args []string) error {
	cmd := exec.Command(hook.Command, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		Log.Printf("Could not run hook (%s) command: %v", hook.Name, err)
		Log.Printf("Failed hook command: '%s %s'\n%s", hook.Command, strings.Join(args, " "), string(out))
		return fmt.Errorf("Hook failed: %v", err)
	}
	return nil
}
