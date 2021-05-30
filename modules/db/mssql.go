package db

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c-f/hygo/model"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	MssqlName = "mssql"
)

// Mssql is the bruter
type Mssql struct {
	//stuff
	Host string
	Port int

	// Options
	StopIfSuccess     bool
	StopIfNetErr      bool
	LogFailedAttempts bool

	// timeout
	timeout time.Duration
	sleep   time.Duration

	// internals for threading
	queue  chan model.Credential
	active model.AtomicBool
	wg     sync.WaitGroup
}

// NewMssql creates a new Mssql Bruter
func NewMssql(host string, port int, sleep time.Duration, timeout time.Duration, logAttempts bool) *Mssql {
	return &Mssql{
		Host: host,
		Port: port,
		// Time Stuff
		timeout:           timeout,
		sleep:             sleep,
		StopIfSuccess:     true, // default
		StopIfNetErr:      true, // default
		LogFailedAttempts: logAttempts,

		queue: make(chan model.Credential),
	}
}

// GoStart defines the amout of handlers per target and sets the out and err channels
func (bruter *Mssql) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *Mssql) Add(cred model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- cred
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *Mssql) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()
}

// handle handles the parallel auth request againts a single host:port combination
func (bruter *Mssql) handle(outChan chan model.Result, errChan chan model.Err) {
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {
		if !bruter.active.Get() { // possibility to get a recheck or counter
			continue
		}

		time.Sleep(bruter.sleep)
		ok, err := ConnectMssql(addr, c, bruter.timeout)
		if err != nil {
			errStr := err.Error()
			shouldContinue := true

			// login error: mssql: Login failed for user 'sa'.
			if strings.Contains(errStr, "login error: mssql: Login failed for user ") {
				if bruter.LogFailedAttempts {
					log.Println("Failed for", addr, string(c.ToJson()))
				}
				continue
			}

			// unable to open tcp connection with host '127.0.0.1:16002': dial tcp 127.0.0.1:16002: connect: connection refused
			if strings.Contains(errStr, "unable to open tcp connection with host") {
				shouldContinue = false
			}

			if !shouldContinue {
				log.Println("Disable Bruteforce (conn reset) for ", MssqlName, bruter.Host)
				bruter.active.Set(false)
			}
			errChan <- model.Err{
				Error: err.Error(),
				Host:  bruter.Host,
				Port:  strconv.Itoa(bruter.Port),
				Addr:  addr,
			}
		}

		if ok {
			if bruter.StopIfSuccess {
				bruter.active.Set(false)
			}

			outChan <- model.Result{
				Service:    MssqlName,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
	}
}

// ConnectMssql connects to maria or mysql srvs
func ConnectMssql(addr string, c model.Credential, timeout time.Duration) (bool, error) {

	var u url.URL
	u.Scheme = "sqlserver"
	u.User = url.UserPassword(c.User, c.Password)
	u.Host = addr
	q := url.Values{}
	q.Set("connection timeout", fmt.Sprintf("%.0f", timeout.Seconds()))
	q.Set("TrustServerCertificate", "true")
	u.RawQuery = q.Encode()
	dataSourceName := u.String()

	return Connect("sqlserver", dataSourceName, timeout)
}
