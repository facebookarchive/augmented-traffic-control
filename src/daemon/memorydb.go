package daemon

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/facebook/augmented-traffic-control/src/iptables"
)

type memoryRunner struct {
	mutex       sync.RWMutex
	members     map[string]*DbMember
	groups      map[int64]*DbGroup
	lastGroupID int64
}

// NewMemoryRunner builds a new database runner
// that stores everything in memory
func NewMemoryRunner() (DbRunner, error) {
	runner := &memoryRunner{
		members: make(map[string]*DbMember),
		groups:  make(map[int64]*DbGroup),
	}

	return runner, nil
}

func (runner *memoryRunner) Close() {
}

/**
*** Porcelain (public)
**/

func (runner *memoryRunner) GetGroup(id int64) (*DbGroup, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	group, ok := runner.groups[id]
	if !ok {
		return nil, nil
	}
	return group, nil
}

func (runner *memoryRunner) GetAllGroups() (chan *DbGroup, error) {
	result := make(chan *DbGroup)
	go func() {
		defer close(result)
		runner.log(runner.getAllGroups(result))
	}()
	return result, nil
}

func (runner *memoryRunner) DeleteGroup(id int64) error {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	delete(runner.groups, id)
	return nil
}

func (runner *memoryRunner) UpdateGroup(group DbGroup) (*DbGroup, error) {
	var err error
	if group.id == 0 {
		group.id, err = runner.nextGroupID()
		if err != nil {
			runner.log(err)
			return nil, err
		}
	}

	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	group.timeout = time.Now().Add(SHAPING_TIMEOUT_LENGTH)
	fmt.Printf("%T %T %T %T", group.id, group.secret, group.tc, group.timeout.Unix())
	runner.groups[group.id] = &group
	return &group, nil
}

func (runner *memoryRunner) GetMember(addr iptables.Target) (*DbMember, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	member, ok := runner.members[addr.String()]
	if !ok {
		return nil, nil
	}
	return member, nil
}

func (runner *memoryRunner) UpdateMember(member DbMember) (*DbMember, error) {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	runner.members[member.addr.String()] = &member
	return &member, nil
}

func (runner *memoryRunner) DeleteMember(addr iptables.Target) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	delete(runner.members, addr.String())
	return nil
}

func (runner *memoryRunner) GetMembersOf(id int64) (chan iptables.Target, error) {
	result := make(chan iptables.Target)
	go func() {
		defer close(result)
		runner.log(runner.getMembersOf(id, result))
	}()
	return result, nil
}

func (runner *memoryRunner) GetAllMembers() (chan *DbMember, error) {
	results := make(chan *DbMember)
	go func() {
		defer close(results)
		runner.log(runner.getAllMembers(results))
	}()
	return results, nil
}

func (runner *memoryRunner) Cleanup() error {
	go func() {
		n, err := runner.cleanupEmptyGroups()
		runner.log(err)
		if n > 0 {
			Log.Printf("DB: Cleaned %d empty groups\n", n)
		}
		n, err = runner.cleanupOldGroups()
		runner.log(err)
		if n > 0 {
			Log.Printf("DB: Cleaned %d expired groups\n", n)
		}
	}()
	return nil
}

func (runner *memoryRunner) getAllGroups(grps chan *DbGroup) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	for _, g := range runner.groups {
		grps <- g
	}
	return nil
}

func (runner *memoryRunner) nextGroupID() (int64, error) {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	runner.lastGroupID++
	return runner.lastGroupID, nil
}

func (runner *memoryRunner) getMembersOf(id int64, members chan iptables.Target) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	for _, m := range runner.members {
		if m.group_id == id {
			members <- m.addr
		}
	}
	return nil
}

func (runner *memoryRunner) getAllMembers(members chan *DbMember) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	for _, m := range runner.members {
		members <- m
	}
	return nil
}

func (runner *memoryRunner) cleanupOldGroups() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	toRemove := make([]int64, 0)
	membersToRemove := make([]string, 0)

	for _, g := range runner.groups {
		if g.timeout.Before(time.Now()) {
			toRemove = append(toRemove, g.id)
			for _, m := range runner.members {
				if m.group_id == g.id {
					membersToRemove = append(membersToRemove, m.addr.String())
				}
			}
		}
	}

	// delay the deletion of the groups until
	// we have a lock on the mutex
	go func(toRemove []int64, membersToRemove []string) {
		runner.mutex.Lock()
		defer runner.mutex.Unlock()

		for _, id := range toRemove {
			delete(runner.groups, id)
		}

		for _, id := range membersToRemove {
			delete(runner.members, id)
		}

	}(toRemove, membersToRemove)

	return int64(len(toRemove)), nil
}

func (runner *memoryRunner) cleanupEmptyGroups() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	knownGroups := make([]int, 0)
	for _, m := range runner.members {
		knownGroups = append(knownGroups, int(m.group_id))
	}

	sort.Ints(knownGroups)

	toRemove := make([]int64, 0)
	for _, g := range runner.groups {
		x := sort.SearchInts(knownGroups, int(g.id))
		if x >= len(knownGroups) || int64(knownGroups[x]) != g.id {
			toRemove = append(toRemove, g.id)
		}
	}

	// delay the deletion of the groups until
	// we have a lock on the mutex
	go func(toRemove []int64) {
		runner.mutex.Lock()
		defer runner.mutex.Unlock()

		for _, d := range toRemove {
			delete(runner.groups, d)
		}
	}(toRemove)

	return int64(len(toRemove)), nil
}

func (runner *memoryRunner) log(err error) {
	if err != nil {
		Log.Printf("DB: error: %v\n", err)
	}
}
