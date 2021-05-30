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

	_ "github.com/go-sql-driver/mysql"
)

var (
	MysqlName = "mysql"
)

// Mysql defines config for a specific target, which can be checked with credentials
type Mysql struct {
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

// NewMysql creates a new Mysql Bruter
func NewMysql(host string, port int, sleep time.Duration, timeout time.Duration, logAttempts bool) *Mysql {
	return &Mysql{
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
func (bruter *Mysql) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *Mysql) Add(cred model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- cred
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *Mysql) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()
}

// handle handles the parallel auth request againts a single host:port combination
func (bruter *Mysql) handle(outChan chan model.Result, errChan chan model.Err) {
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {
		if !bruter.active.Get() { // possibility to get a recheck or counter
			continue
		}

		time.Sleep(bruter.sleep)
		ok, err := ConnectMysql(addr, c, bruter.timeout)
		if err != nil {
			errStr := err.Error()
			shouldContinue := true

			if strings.Contains(errStr, "Access denied for user") {
				if bruter.LogFailedAttempts {
					log.Println("Failed for", addr, string(c.ToJson()))
				}
				continue
			}
			// dial tcp 127.0.0.1:14105: connect: connection refused
			if strings.Contains(errStr, "connect: connection refused") {
				shouldContinue = false
			}

			if !shouldContinue {
				log.Println("Disable Bruteforce (conn reset) for ", MysqlName, bruter.Host)
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
				Service:    MysqlName,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
	}
}

// ConnectMysql connects to maria or mysql srvs
func ConnectMysql(addr string, c model.Credential, timeout time.Duration) (bool, error) {
	dataSourceName := fmt.Sprintf("%s@tcp(%s)/?timeout=%s", url.UserPassword(c.User, c.Password).String(), addr, timeout)
	return Connect("mysql", dataSourceName, timeout)
}
