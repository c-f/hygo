package ssh

import (
	"fmt"
	"github.com/c-f/hygo/model"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	Name = "ssh"
)

// SSH defines configs for a specific target, which can be checked with credentials
type SSH struct {
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

// New creates a new SSH Bruter
func New(host string, port int, sleep time.Duration, timeout time.Duration) *SSH {
	bruter := &SSH{
		Host: host,
		Port: port,
		// Time Stuff
		timeout:       timeout,
		sleep:         sleep,
		StopIfSuccess: true, // default

		queue: make(chan model.Credential),
	}
	return bruter
}

// GoStart defines the amout of handlers per target and sets the out and err channels
func (bruter *SSH) GoStart(threads int, outChan chan model.Result, errChan chan model.Err) {
	bruter.active.Set(true)
	for i := 0; i < threads; i++ {
		bruter.wg.Add(1)
		go bruter.handle(outChan, errChan)
	}
}

// Add uses a credential pair for the connection blocking !
func (bruter *SSH) Add(c model.Credential) {
	if bruter.active.Get() {
		bruter.queue <- c
	}
	return
}

// Close signals the target to cease all activity.
func (bruter *SSH) Close() {
	bruter.active.Set(false)

	close(bruter.queue)
	bruter.wg.Wait()

}

// handle handles the parallel auth request againts a single host:port combination
func (bruter *SSH) handle(outChan chan model.Result, errChan chan model.Err) {
	// defer func() {
	// 	if x := recover(); x != nil {
	// 		fmt.Println("PAAAANIIIICCCC")
	// 		panic("!!!")
	// 	}
	// }()
	defer bruter.wg.Done()
	addr := fmt.Sprintf("%s:%d", bruter.Host, bruter.Port)

	for c := range bruter.queue {

		ok, err := ConnectWithCredential(addr, c, bruter.timeout)
		if err != nil {
			// TODO(Chris): err handling should be better
			errStr := err.Error()
			shouldContinue := true

			// ssh: handshake failed: read tcp 127.0.0.1:41946-\u003e127.0.0.1:14002: read: connection reset by peer
			// ssh: handshake failed: EOF

			// If password is wrong
			// ssh: unable to authenticate, attempted methods [none], no supported methods remain
			if strings.Contains(errStr, "no supported methods remain") {
				continue
			}

			// dial tcp 178.254.26.222:15681: connect: connection refused
			if bruter.StopIfNetErr && strings.Contains(errStr, "connect: connection refused") {
				shouldContinue = false
			}
			if !shouldContinue {
				bruter.active.Set(false)
			}

			// log err
			errChan <- model.Err{
				Error: errStr,
				Host:  bruter.Host,
				Port:  strconv.Itoa(bruter.Port),
				Addr:  addr,
			}
		}

		// check if pw is successfull
		if ok {
			if bruter.StopIfSuccess {
				bruter.active.Set(false)
			}

			// output
			outChan <- model.Result{
				Service:    Name,
				Host:       bruter.Host,
				Credential: c,
				Port:       strconv.Itoa(bruter.Port),
			}
		}
		time.Sleep(bruter.sleep)
	}
}

// GetAuthMethod returns the appropriate ssh.AuthMethod based on the provided Credentials
func GetAuthMethod(cred model.Credential) (ssh.AuthMethod, error) {
	if sshKey, ok := cred.Data["ssh_key"]; ok {
		var key ssh.Signer
		var err error
		if cred.Password == "" {
			key, err = ssh.ParsePrivateKey([]byte(sshKey))
			if err != nil {
				log.Println("Error while parsing key", cred.User)
				return nil, err
			}
		} else {
			key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(sshKey), []byte(cred.Password))
			if err != nil {
				log.Println("Error while parsing key", cred.User)
				return nil, err
			}
		}
		return ssh.PublicKeys(key), nil
	}
	return ssh.Password(cred.Password), nil
}

// ConnectWithCredential returns ok == true, if the credential work, else err is not nil
func ConnectWithCredential(addr string, cred model.Credential, timeout time.Duration) (ok bool, err error) {
	authMethod, err := GetAuthMethod(cred)
	if err != nil {
		return
	}
	config := &ssh.ClientConfig{
		User:            cred.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{authMethod},
		Timeout:         timeout,
		// TODO(chris): does not work :/
		BannerCallback: ssh.BannerDisplayStderr(),
	}
	return Connect(addr, config)
}

// ConnectWithCredential returns ok == true, if the credential work, else err is not nil
func ConnectWithPassword(addr, user, passwd string, timeout time.Duration) (ok bool, err error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
		Timeout:         timeout,
		// TODO(chris): does not work :/
		BannerCallback: ssh.BannerDisplayStderr(),
	}
	return Connect(addr, config)
}

// Connect connects to the server using the provided ClientConfig
func Connect(addr string, config *ssh.ClientConfig) (ok bool, err error) {
	cl, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return ok, err
	}

	// --[Connection successfull]--
	ok = true
	sess, err := cl.NewSession()
	if err != nil {
		return true, err
	}
	err = sess.Close()
	if err != nil {
		return ok, err
	}
	return true, nil
}
