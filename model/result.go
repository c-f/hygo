package model

// Result contains all information about a sucessfull attempt
type Result struct {
	Service string `json:"service"`
	Host    string `json:"host"`
	Port    string `json:"port"`

	Credential Credential `json:"credential"`
}
