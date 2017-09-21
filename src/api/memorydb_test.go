package api

import (
	"testing"
	"time"
)

func getRunner(t *testing.T) *memoryRunner {
	runner, err := NewMemoryRunner()
	if err != nil {
		t.Fatalf("Error creating memory runner: %s", err)
		return nil
	}
	return runner.(*memoryRunner)
}

func getProfiles(t *testing.T, r *memoryRunner) (px []Profile) {
	ch := r.GetProfiles()

	select {
	case profiles := <-ch:
		px = profiles
	case <-time.After(200 * time.Millisecond):
		t.Errorf("Timeout reading from channel")
		return
	}

	select {
	case <-ch:
		// closed channel will always trigger here
	case <-time.After(10 * time.Millisecond):
		t.Errorf("Channel should have been closed")
	}

	return
}

func TestMemoryRunner(t *testing.T) {
	t.Run("emptyMemoryRunner", func(t *testing.T) {
		runner := getRunner(t)
		defer runner.Close()

		if runner.lastProfileId != 0 {
			t.Errorf("Expected runner lastProfileId to be 0, got %d", runner.lastProfileId)
		}

		if runner.profiles == nil {
			t.Errorf("Expected profiles to be non-nil")
		}

		if len(runner.profiles) != 0 {
			t.Errorf("Expected profile size to be 0, got %d", len(runner.profiles))
		}
	})

	t.Run("getProfiles", func(t *testing.T) {
		runner := getRunner(t)
		defer runner.Close()

		t.Run("emptyList", func(t *testing.T) {
			runner.profiles = make(map[int64]*Profile)
			px := getProfiles(t, runner)
			if len(px) != 0 {
				t.Errorf("Expected profiles to be empty, got %d items", len(px))
			}
		})

		t.Run("nonEmptyList", func(t *testing.T) {
			runner.profiles = make(map[int64]*Profile)
			runner.profiles[0] = &Profile{}
			runner.profiles[1] = &Profile{}
			runner.profiles[3] = &Profile{}
			px := getProfiles(t, runner)
			if len(px) != 3 {
				t.Errorf("Expected profiles to be of size 3, got %d items", len(px))
			}
		})
	})

	t.Run("deleteProfile", func(t *testing.T) {
		runner := getRunner(t)
		runner.profiles = make(map[int64]*Profile)
		runner.profiles[3] = &Profile{}

		defer runner.Close()

		t.Run("nonExistentProfile", func(t *testing.T) {
			if err := runner.deleteProfile(12); err != nil {
				t.Errorf("Unexpected error deleting non-existent profile: %s", err)
			}
		})

		t.Run("existingProfile", func(t *testing.T) {
			if err := runner.deleteProfile(3); err != nil {
				t.Errorf("Unexpected error deleting profile: %s", err)
			}

			if len(runner.profiles) != 0 {
				t.Errorf("Profile not deleted")
			}
		})
	})

	t.Run("updateProfile", func(t *testing.T) {
		t.Run("noID", func(t *testing.T) {

			runner := getRunner(t)
			defer runner.Close()

			pchan := runner.UpdateProfile(Profile{})
			select {
			case p := <-pchan:
				if p.Id != 1 {
					t.Errorf("Profile identifier should be 1, got %d", p.Id)
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for profile updating")
			}

			if runner.lastProfileId != 1 {
				t.Errorf("Last Profile Id should be 1, got %d", runner.lastProfileId)
			}

			select {
			case _, more := <-pchan:
				// closed pchan should trigger
				if more {
					t.Errorf("Expected pchan to be closed")
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for pchan to be closed")
			}
		})
		t.Run("withIDoutsideDb", func(t *testing.T) {
			runner := getRunner(t)
			defer runner.Close()

			// if the profile id does not exist, we overwrite it
			pchan := runner.UpdateProfile(Profile{Id: 12})
			select {
			case p := <-pchan:
				if p.Id != 1 {
					t.Errorf("Profile identifier should be 1, got %d", p.Id)
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for profile updating")
			}

			if runner.lastProfileId != 1 {
				t.Errorf("Last Profile Id should be 1, got %d", runner.lastProfileId)
			}

			select {
			case _, more := <-pchan:
				// closed pchan should trigger
				if more {
					t.Errorf("Expected pchan to be closed")
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for pchan to be closed")
			}
		})

		t.Run("withID", func(t *testing.T) {
			runner := getRunner(t)
			runner.profiles[12] = &Profile{Id: 12}
			runner.lastProfileId = 12
			defer runner.Close()

			pchan := runner.UpdateProfile(Profile{Id: 12})
			select {
			case p := <-pchan:
				if p.Id != 12 {
					t.Errorf("Profile identifier should be 12, got %d", p.Id)
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for profile updating")
			}

			if runner.lastProfileId != 12 {
				t.Errorf("Last Profile Id should be 12, got %d", runner.lastProfileId)
			}

			select {
			case _, more := <-pchan:
				// closed pchan should trigger
				if more {
					t.Errorf("Expected pchan to be closed")
				}
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Timeout waiting for pchan to be closed")
			}
		})
	})
}
