package model

// Credential contains information which secrets should be used
type Credential struct {
	User     string `json:"user"`
	Password string `json:"password"`

	// Data can contain service related data, such as ssh_keys or specific attributes
	// optional
	Data map[string]string `json:"data"`
}

// NewCredential returns a new Credential obj
func NewCredential(user, passwd string) *Credential {
	return &Credential{
		User:     user,
		Password: passwd,
		Data:     make(map[string]string),
	}
}
