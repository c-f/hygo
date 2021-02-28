package hygo

import "github.com/c-f/hygo/model"

// Bruter is the Interface to check a specific server
// The initialization is done by the factory GetBruter
type Bruter interface {

	// GoStart starts the handler and configure the outgoing communication channel
	GoStart(threads int, outChan chan model.Result, errChan chan model.Err)

	// Add adds a credential, which should be checked by the bruter (blocking)
	Add(model.Credential)

	// Closes the bruter and its connection greatfully - waiting for goroutines
	Close()
}

/*


bruter := GetBruter(<opts>)
bruter.GoStart(<opts>)

done := make(chan bool)
go func(){
	bruter.Add(stuff)
	done <- true
}
<-done
bruter.Close()

*/
