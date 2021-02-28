package hygo

import "time"

type Config struct {
	Sleep   time.Duration
	Timeout time.Duration

	Threads        int
	DefaultService string
}
