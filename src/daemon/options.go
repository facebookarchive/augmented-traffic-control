package daemon

type AtcdOptions struct {
	Secure     bool
	OtpTimeout uint8
}

var DefaultAtcdOptions = AtcdOptions{
	Secure:     true,
	OtpTimeout: 60,
}
