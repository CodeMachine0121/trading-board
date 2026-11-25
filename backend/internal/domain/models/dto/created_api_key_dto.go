package dto

// CreatedApiKeyDto is the answer to creating a key: the listed key plus the full key, shown this once only.
type CreatedApiKeyDto struct {
	ApiKeyDto
	Key string `json:"key"`
}
