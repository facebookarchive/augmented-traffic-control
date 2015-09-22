package api

import (
	"testing"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

var (
	FakeShaping *atc_thrift.Setting = &atc_thrift.Setting{
		Up:   &atc_thrift.Shaping{},
		Down: nil,
	}
)

func TestDBCreatesSchema(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.db.Query("SELECT * FROM Profiles")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBInsertsProfile(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	profile, err := db.updateProfile(Profile{
		Name:     "sample-profile",
		Settings: FakeShaping,
	})
	if err != nil {
		t.Fatal(err)
	}
	if profile.Id < 0 {
		t.Fatalf("Id should be >= 0: %d", profile.Id)
	}
	if profile.Name != "sample-profile" {
		t.Fatalf(`Mismatched name settings: %q != %q`, "sample-profile", profile.Name)
	}
}

func TestDBGetsProfiles(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.updateProfile(Profile{Name: "sample-profile", Settings: FakeShaping}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.updateProfile(Profile{Name: "fake-profile", Settings: FakeShaping}); err != nil {
		t.Fatal(err)
	}

	profiles, err := db.getProfiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(profiles) != 2 {
		t.Fatalf("Wrong number of profiles: 2 != %d", len(profiles))
	}

	presentProfiles := make(map[string]int64)
	for _, p := range profiles {
		presentProfiles[p.Name] = p.Id
	}
	if id, ok := presentProfiles["sample-profile"]; !ok {
		t.Error("sample-profile wasn't returned")
	} else if id <= 0 {
		t.Errorf("sample-profile's id wasn't set: %d", id)
	}

	if id, ok := presentProfiles["fake-profile"]; !ok {
		t.Error("fake-profile wasn't returned")
	} else if id <= 0 {
		t.Errorf("fake-profile's id wasn't set: %d", id)
	}
}

func TestDBDeletesProfiles(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		profile *Profile
	)
	if _, err := db.updateProfile(Profile{Name: "sample-profile", Settings: FakeShaping}); err != nil {
		t.Fatal(err)
	}
	if profile, err = db.updateProfile(Profile{Name: "fake-profile", Settings: FakeShaping}); err != nil {
		t.Fatal(err)
	}

	if err := db.deleteProfile(profile.Id); err != nil {
		t.Fatal(err)
	}

	profiles, err := db.getProfiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(profiles) != 1 {
		t.Errorf("Wrong number of profiles: 1 != %d", len(profiles))
		if len(profiles) < 1 {
			// Checks below assume there's at least 1 profile remaining
			return
		}
	}

	if profiles[0].Name != "sample-profile" {
		t.Errorf("Wrong profile was deleted; remaining profile: %q", profiles[0].Name)
	}
}
