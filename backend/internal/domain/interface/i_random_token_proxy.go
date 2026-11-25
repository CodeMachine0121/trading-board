package _interface

//go:generate go tool mockgen -source=i_random_token_proxy.go -destination=mocks/mock_i_random_token_proxy.go -package=mocks

// IRandomTokenProxy draws unguessable secrets for codes, refresh tokens and API keys.
type IRandomTokenProxy interface {
	// NewToken returns at least 32 bytes of cryptographic randomness, URL-safe encoded. It cannot fail: the
	// operating system's random source never runs dry, and Go crashes rather than hand out weak randomness.
	NewToken() string
}
