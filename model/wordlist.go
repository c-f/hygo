package model

import "sync"

// Wordlist stores Credentials
type Wordlist struct {
	mux   sync.RWMutex
	Items []Credential
}

// NewWordlist creates a new Wordlist
func NewWordlist() *Wordlist {
	return &Wordlist{
		Items: make([]Credential, 0),
	}
}

// Add add a new basic Credential (user:pass) to wordlist
func (w *Wordlist) Add(user, passwd string) {
	w.mux.Lock()
	defer w.mux.Unlock()

	w.Items = append(w.Items, *NewCredential(user, passwd))
}

// AddCred adds a new Credential to the wordlist
func (w *Wordlist) AddCred(c Credential) {
	w.mux.Lock()
	defer w.mux.Unlock()

	w.Items = append(w.Items, c)
}

// Length returns the lenght of the wordlist
func (w *Wordlist) Length() int {
	w.mux.RLock()
	defer w.mux.RUnlock()

	return len(w.Items)
}

// Get returns the Credential from the wordlist
func (w *Wordlist) Get(idx int) Credential {
	w.mux.RLock()
	defer w.mux.RUnlock()
	return w.Items[idx]
}
