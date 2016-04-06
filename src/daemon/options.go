package daemon

type AtcdOptions struct {
	// Run in secure mode?
	Secure bool

	// OTP token timeout (seconds)
	OtpTimeout uint8
}

var DefaultAtcdOptions = AtcdOptions{
	Secure:     true,
	OtpTimeout: 60,
}
