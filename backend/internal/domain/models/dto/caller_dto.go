package dto

// CallerDto is who a request comes from: its role and the name its records are attributed to
// (the member name for an API key, "owner" for the signed-in owner).
type CallerDto struct {
	Role string
	Name string
}
