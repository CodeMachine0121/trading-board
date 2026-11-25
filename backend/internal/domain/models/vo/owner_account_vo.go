package vo

// OwnerAccountVo is the one account, as configured: its email and the bcrypt hash of its password.
// Empty means "not configured", which admits nobody.
type OwnerAccountVo struct {
	Email        string
	PasswordHash string
}
