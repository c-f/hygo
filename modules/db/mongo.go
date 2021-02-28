package db

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/c-f/hygo/model"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	MongoDBName   = "mongo"
	AuthMechanism = "auth_mechanism"
)

type MongoDB struct {
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

// NewMongoDB creates a new MongoDB Bruter
func NewMongoDB(host string, port int, sleep time.Duration, timeout time.Duration) *MongoDB {
	return &MongoDB{
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
func (bruter *MongoDB) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *MongoDB) Add(cred model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- cred
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *MongoDB) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()
}

// handle handles the parallel auth request againts a single host:port combination
func (bruter *MongoDB) handle(outChan chan model.Result, errChan chan model.Err) {
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {
		ok, err := ConnectMongoDB(addr, c, bruter.timeout)
		if err != nil {
			errStr := err.Error()
			shouldContinue := true

			// Error":"connection() error occured during connection handshake: auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": (AuthenticationFailed) Authentication failed.
			if strings.Contains(errStr, "(AuthenticationFailed) Authentication failed") {
				continue
			}

			// server selection error: context deadline exceeded, current topology: { Type: Unknown, Servers: [{ Addr: 127.0.0.1:1, Type: Unknown, Average RTT: 0, Last error: connection() error occured during connection handshake: dial tcp 127.0.0.1:1: connect: connection refused }, ] }
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
				Service:    MongoDBName,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
		time.Sleep(bruter.sleep)
	}
}

// ConnectMongoDB connects to mongodb srvs
// https://docs.mongodb.com/manual/reference/connection-string/
func ConnectMongoDB(addr string, c model.Credential, timeout time.Duration) (ok bool, err error) {

	// Connection Definition
	var u url.URL
	u.Scheme = "mongodb"
	u.User = url.UserPassword(c.User, c.Password)
	u.Host = addr
	u.Path = "/"
	q := url.Values{}

	// https://docs.mongodb.com/manual/reference/connection-string/#connections-connection-options
	q.Set("tlsAllowInvalidCertificates", "true")
	q.Set("tlsAllowInvalidHostnames", "true")
	q.Set("tlsInsecure", "true")
	q.Set("connectTimeoutMS", fmt.Sprintf("%.0f", timeout.Seconds()*1000))

	if authMechanism, ok := c.Data[AuthMechanism]; ok {
		// https://docs.mongodb.com/manual/reference/connection-string/#urioption.authMechanism
		q.Set("authMechanism", authMechanism)
	}

	u.RawQuery = q.Encode()
	dataSourceName := u.String()

	// Configure the figure
	client, err := mongo.NewClient(options.Client().ApplyURI(dataSourceName))
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Connect
	if err = client.Connect(ctx); err != nil {
		return
	}
	defer client.Disconnect(ctx)

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return
	}
	ok = true
	return
}
