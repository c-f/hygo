package model

// Item contains information about the target and its service
type Item struct {
	Target  string `json:"host"`
	Port    string `json:"port"`
	Service string `json:"service"`
}
