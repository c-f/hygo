package db

import (
	"fmt"
	"github.com/c-f/hygo/model"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Postgres is the bruter
type Postgres struct {
	//stuff
	Host string
	Port int

	// Options
	StopIfSuccess bool
	StopIfNetErr  bool

	// timeout
	timeout time.Duration
	sleep   time.Duration

	// internals for threading
	queue  chan model.Credential
	active model.AtomicBool
	wg     sync.WaitGroup
}

// NewPostgres creates a new Postgres Bruter
func NewPostgres(host string, port int, sleep time.Duration, timeout time.Duration) *Postgres {
	return &Postgres{
		Host: host,
		Port: port,
		// Time Stuff
		timeout:       timeout,
		sleep:         sleep,
		StopIfSuccess: true, // default

		queue: make(chan model.Credential),
	}
}

// GoStart defines the amout of handlers per target and sets the out and err channels
func (bruter *Postgres) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *Postgres) Add(cred model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- cred
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *Postgres) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()
}

var (
	PostgresName = "postgres"
)

// handle handles the parallel auth request againts a single host:port combination
func (bruter *Postgres) handle(outChan chan model.Result, errChan chan model.Err) {
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {
		ok, err := ConnectPostgres(addr, c, bruter.timeout)
		if err != nil {
			errStr := err.Error()
			shouldContinue := true

			if strings.Contains(errStr, "password authentication failed for user") {
				continue
			}

			if !shouldContinue {
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
				Service:    PostgresName,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
		time.Sleep(bruter.sleep)
	}
}

// ConnectPostgres connects to a db and checks if the credentials are valid or not
func ConnectPostgres(addr string, c model.Credential, timeout time.Duration) (bool, error) {
	// https://pkg.go.dev/github.com/lib/pq#ParseURL
	var u url.URL
	u.Scheme = "postgres"
	u.User = url.UserPassword(c.User, c.Password)
	u.Host = addr
	q := url.Values{}

	q.Set("connect_timeout", fmt.Sprintf("%.0f", timeout.Seconds()))
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	dataSourceName := u.String()

	return Connect("postgres", dataSourceName, timeout)
}
