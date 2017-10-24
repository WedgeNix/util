package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"log"

	"io/ioutil"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

// MustGetenv must get the environment variable or it panics.
func MustGetenv(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		panic(key + " not found")
	}
	return value
}

// EmailLogin login for SMTP.
type EmailLogin struct {
	User   string
	Pass   string
	SMTP   string
	sender struct {
		sync.Mutex
		s *gomail.SendCloser
	}
}

//Email to send basic emails from a particular gmail account.
func (l *EmailLogin) Email(to []string, subject string, body string, attachment string) error {
	msg := gomail.NewMessage()
	msss := map[string][]string{"From": {"WedgeNix<" + l.User + ">"}, "To": to, "Subject": {subject}}
	msg.SetHeaders(msss)
	msg.SetBody("text/html", body)
	if len(attachment) > 0 {
		msg.Attach(attachment)
	}

	l.sender.Lock()
	defer l.sender.Unlock()

	if l.sender.s == nil {
		d := gomail.NewDialer(l.SMTP, 587, l.User, l.Pass)

		s, err := d.Dial()
		if err != nil {
			return err
		}
		l.sender.s = &s
	}
	return gomail.Send(*l.sender.s, msg)
}

// Stop stops and clears the dialed email instance sender.
func (l *EmailLogin) Stop() {
	l.sender.Lock()
	defer l.sender.Unlock()

	if l.sender.s == nil {
		return
	}

	(*l.sender.s).Close()
	l.sender.s = nil
}

// HTTPLogin allows basic HTTP authorization for getting simple responses.
type HTTPLogin struct {
	User string
	Pass string
}

// Base64 encodes an HTTP username and password.
func (l HTTPLogin) Base64() string {
	return base64.StdEncoding.EncodeToString([]byte(l.User + ":" + l.Pass))
}

// Get receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) Get(url string) (*http.Response, error) {
	return l.req(http.MethodGet, url, nil)
}

// Post receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) Post(url string, v interface{}) (*http.Response, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return l.req(http.MethodPost, url, bytes.NewReader(b))
}

// Get receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) req(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+l.Base64())
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	cl := http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// LANow grabs time from Los Angeles.
func LANow() time.Time {
	la, _ := time.LoadLocation("America/Los_Angeles")
	return time.Now().In(la)
}

// Load reads in a JSON from either an HTTP GET response or a local file.
func Load(name string, v interface{}) {
	var r io.Reader
	if strings.Contains(name, "https://") || strings.Contains(name, "http://") {
		resp, err := http.Get(name)
		E(err)
		r = resp.Body
		defer resp.Body.Close()
	} else {
		f, err := os.Open(name + ".json")
		E(err)
		r = f
		defer f.Close()
	}
	json.NewDecoder(r).Decode(v)
}

// Save writes a JSON to either an HTTP POST response or a local file.
func Save(name string, v interface{}) {
	if strings.Contains(name, "https://") || strings.Contains(name, "http://") {
		b, err := json.Marshal(v)
		E(err)
		r := bytes.NewReader(b)
		_, err = http.Post(name, "application/json", r)
		E(err)
	} else {
		f, err := os.Create(name + ".json")
		E(err)
		defer f.Close()
		json.NewEncoder(f).Encode(v)
	}
}

// Read reads a reader in full, returning the entire string.
func Read(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	return string(b), err
}

// HTTPStatus responds to the HTTP request with a JSON containing the status and message.
func HTTPStatus(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"Status":  http.StatusText(status),
		"Message": message,
	})
}

// MergeErr merges errors into one channel.
func MergeErr(errcs ...<-chan error) <-chan error {
	newerrc := make(chan error)
	var wg sync.WaitGroup

	for _, errc := range errcs {
		errc := errc
		wg.Add(1)
		go func() {
			defer wg.Done()
			for err := range errc {
				newerrc <- err
			}
		}()
	}

	go func() {
		defer close(newerrc)
		wg.Wait()
	}()

	return newerrc
}

type logError struct {
	err  error
	file string
	line int
}

func (le logError) Error() string {
	err := ""
	if le.err != nil {
		err = le.err.Error()
	}
	return le.file + ":" + strconv.Itoa(le.line) + ": " + err
}

// Err wraps the filename and line number around the error.
func Err(err error) error {
	if err == nil {
		return err
	}
	_, f, ln, _ := runtime.Caller(1)
	return &logError{err, f[strings.LastIndex(f, "/")+1:], ln}
}

// NewErr wraps the filename and line number around the error.
func NewErr(err string) error {
	return Err(errors.New(err))
}

var ready = make(chan bool, 1)

// NewLoader creates a console loading bar.
func NewLoader(msg string, stream ...bool) chan<- bool {
	<-ready

	done := make(chan bool)

	frames := []string{"░", "▒", "▓", "█", "▓", "▒"}
	frame := 0
	delay := 200

	r := "\r"
	if len(stream) > 0 && stream[0] {
		r = ""
	}

	go func() {
		for {
			select {
			case <-done:
				fmt.Println(r + msg + "  !")
				ready <- true
				return
			default:
			}

			fmt.Print(r + msg + "  " + frames[frame] + " ")
			frame = (frame + 1) % len(frames)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}()

	return done
}

// E reports the error if there is any and exits.
// Deprecated: handle errors by returning them as values
func E(err error, skip ...bool) {
	if err == nil {
		return
	}
	if len(skip) > 0 && skip[0] {
		log.Println(err)
	} else {
		log.Fatalln(err)
	}
}

// R reports the error if there is any.
// Deprecated: use util.E(err [, skip?])
func R(err error) {
	if err != nil {
		P(err)
	}
}

// P prints the arguments as they are.
func P(args ...interface{}) {
	for _, a := range args {
		fmt.Print(a)
	}
	fmt.Println()
}

// S strings together the arguments as they are.
func S(args ...interface{}) string {
	str := ""
	for _, a := range args {
		str += fmt.Sprint(a)
	}
	return str
}

func p(args ...interface{}) {
	for _, a := range args {
		fmt.Print(a)
	}
	fmt.Println()
}
