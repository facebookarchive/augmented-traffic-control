package daemon

// AtcdOptions contains the options for the ATC daemon
type AtcdOptions struct {
	// Run in secure mode?
	Secure bool

	// OTP token timeout (seconds)
	OtpTimeout uint8
}

// DefaultAtcdOptions are the default options for the ATC daemon
var DefaultAtcdOptions = AtcdOptions{
	Secure:     true,
	OtpTimeout: 60,
}
