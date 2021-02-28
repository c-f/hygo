package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c-f/hygo/model"

	"github.com/go-redis/redis/v8"
)

var (
	RedisName = "redis"
)

// Redis defines config for a specific target, which can be checked with credentials
type Redis struct {
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

// NewRedis creates a new Redis Bruter
func NewRedis(host string, port int, sleep time.Duration, timeout time.Duration) *Redis {
	return &Redis{
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
func (bruter *Redis) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *Redis) Add(cred model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- cred
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *Redis) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()
}

// handle handles the parallel auth request againts a single host:port combination
func (bruter *Redis) handle(outChan chan model.Result, errChan chan model.Err) {
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {
		ok, err := ConnectRedis(addr, c, bruter.timeout)
		if err != nil {
			errStr := err.Error()
			shouldContinue := true

			if strings.Contains(errStr, "WRONGPASS invalid username-password pair") {
				continue
			}

			if strings.Contains(errStr, "ERR invalid password") {
				continue
			}

			if strings.Contains(errStr, "connect: connection refused") {
				shouldContinue = false
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
				Service:    RedisName,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
		time.Sleep(bruter.sleep)
	}
}

// ConnectRedis connects to maria or Redis srvs
func ConnectRedis(addr string, c model.Credential, timeout time.Duration) (ok bool, err error) {

	// https://pkg.go.dev/github.com/go-redis/redis/v8#Options
	opts := &redis.Options{
		Addr:     addr,
		Password: c.Password, // no password set
		DB:       0,          // use default DB

		// Dial timeout for establishing new connections.
		// Default is 5 seconds.
		DialTimeout: timeout,
		// Timeout for socket reads. If reached, commands will fail
		// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
		// Default is 3 seconds.
		ReadTimeout: timeout,
		// Timeout for socket writes. If reached, commands will fail
		// with a timeout instead of blocking.
		// Default is ReadTimeout.
		WriteTimeout: timeout,
	}
	if c.User != "" {
		opts.Username = c.User
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	statusCmd := rdb.Ping(ctx)
	_, err = statusCmd.Result()

	if err != nil {
		return false, err
	}

	// Luls
	return true, nil
}
