package api

type DbRunner interface {
	Close()

	UpdateProfile(profile Profile) chan *Profile
	GetProfiles() chan []Profile
	DeleteProfile(id int64)
}
