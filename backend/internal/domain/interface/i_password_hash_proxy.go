package _interface

//go:generate go tool mockgen -source=i_password_hash_proxy.go -destination=mocks/mock_i_password_hash_proxy.go -package=mocks

// IPasswordHashProxy checks a password against its stored one-way hash.
type IPasswordHashProxy interface {
	// Matches is false for an empty or malformed hash, never an error: a broken setting admits nobody.
	Matches(passwordHash string, password string) bool
}
