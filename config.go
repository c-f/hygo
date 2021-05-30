package hygo

import "time"

type Config struct {
	Sleep             time.Duration
	Timeout           time.Duration
	LogFailedAttempts bool
	Force             bool // should bruteforce continue even if connection is reset or similar events

	Threads        int
	DefaultService string
}
