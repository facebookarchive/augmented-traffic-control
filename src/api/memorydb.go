package api

import (
	"log"
	"sync"
)

type memoryRunner struct {
	mutex         sync.RWMutex
	lastProfileId int64
	profiles      map[int64]*Profile
}

func NewMemoryRunner() (DbRunner, error) {
	runner := &memoryRunner{
		profiles: make(map[int64]*Profile),
	}

	return runner, nil
}

func (runner *memoryRunner) Close() {
}

/**
*** Porcelain (public)
**/

func (runner *memoryRunner) UpdateProfile(profile Profile) chan *Profile {
	result := make(chan *Profile)
	go func() {
		defer close(result)
		profile, err := runner.updateProfile(profile)
		if err == nil {
			result <- profile
		}
		runner.log(err)
	}()
	return result
}

func (runner *memoryRunner) GetProfiles() chan []Profile {
	result := make(chan []Profile)
	go func() {
		defer close(result)
		profiles, err := runner.getProfiles()
		if err == nil {
			result <- profiles
		}
		runner.log(err)
	}()
	return result
}

func (runner *memoryRunner) DeleteProfile(id int64) {
	go func() {
		runner.log(runner.deleteProfile(id))
	}()
}

/**
*** Plumbing (private...ish)
**/

func (runner *memoryRunner) log(err error) {
	if err != nil {
		log.Printf("DB: error: %v\n", err)
	}
}

func (runner *memoryRunner) nextProfileId() (int64, error) {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	runner.lastProfileId++
	return runner.lastProfileId, nil
}

func (runner *memoryRunner) updateProfile(profile Profile) (*Profile, error) {
	var err error

	_, ok := runner.profiles[profile.Id]

	if profile.Id <= 0 || !ok {
		profile.Id, err = runner.nextProfileId()
		if err != nil {
			return nil, err
		}
	}

	runner.mutex.Lock()
	defer runner.mutex.Unlock()

	runner.profiles[profile.Id] = &profile
	return &profile, nil
}

func (runner *memoryRunner) getProfiles() (px []Profile, err error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()

	for _, p := range runner.profiles {
		px = append(px, *p)
	}

	return
}

func (runner *memoryRunner) deleteProfile(id int64) error {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	delete(runner.profiles, id)
	return nil
}
