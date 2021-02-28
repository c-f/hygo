package model

import "sync"

// AtomicBool provides a synced boolean variable.
type AtomicBool struct {
	sync.RWMutex
	value bool
}

// Set stores the new value of AtomicBool.
func (ab *AtomicBool) Set(value bool) {
	ab.Lock()
	defer ab.Unlock()
	ab.value = value
}

// Get retrieves the current value of AtomicBool.
func (ab *AtomicBool) Get() bool {
	ab.RLock()
	defer ab.RUnlock()
	return ab.value
}
