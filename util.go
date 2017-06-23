package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"log"

	"sync"

	"io/ioutil"

	"gopkg.in/gomail.v2"
)

var (
	debug       bool
	debugSync   sync.RWMutex
	literal     bool
	literalSync sync.RWMutex
	gids        []uint64
	gidSync     sync.Mutex
)

func init() {
	// default settings
	Debug(true)
	Literal(false)

	gids = []uint64{}
}

// EmailLogin login for SMTP.
type EmailLogin struct {
	User string
	Pass string
}

//Email to send basic emails from a particular gmail account.
func (l *EmailLogin) Email(to []string, subject string, body string, attachment string) {
	email := l.User + "@gmail.com"
	msg := gomail.NewMessage()
	msss := map[string][]string{"From": {"WedgeNix<" + email + ">"}, "To": to, "Subject": {subject}}
	msg.SetHeaders(msss)
	msg.SetBody("text/html", body)
	if len(attachment) > 0 {
		msg.Attach(attachment)
	}
	d := gomail.NewDialer("smtp.gmail.com", 587, email, l.Pass)
	err := d.DialAndSend(msg)
	if err != nil {
		fmt.Println(err)
	}
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
func (l HTTPLogin) Get(url string) *http.Response {
	return l.req(http.MethodGet, url)
}

// Post receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) Post(url string) *http.Response {
	return l.req(http.MethodPost, url)
}

// Get receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) req(method string, url string) *http.Response {
	req, err := http.NewRequest(method, url, nil)
	E(err)
	req.Header.Add("Authorization", "Basic "+l.Base64())
	cl := http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	E(err)
	return resp
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
func Read(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	R(err)
	return string(b)
}

// Debug sets whether or not to print debug statements.
func Debug(b bool) {
	debugSync.Lock()
	debug = b
	debugSync.Unlock()
}

// Literal turns on/off channel printing markup.
func Literal(b bool) {
	literalSync.Lock()
	literal = b
	literalSync.Unlock()
}

func getLiteral() bool {
	literalSync.RLock()
	defer literalSync.RUnlock()
	return literal
}

func getDebug() bool {
	debugSync.RLock()
	defer debugSync.RUnlock()
	return debug
}

// Comment adds visual comments if debugging.
func Comment(args ...string) {
	if !getDebug() {
		return
	}
	P()
	for _, comment := range args {
		P("// ", comment)
	}
	P()
}

// E reports the error if there is any and exits.
func E(err error) {
	if err == nil {
		return
	}
	log.Fatalln(tabs() + err.Error())
}

// R reports the error if there is any.
func R(err error) {
	if err != nil {
		P(err)
	}
}

// P prints the arguments as they are.
func P(args ...interface{}) {
	fmt.Print(tabs())
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

func tabs() string {
	tabs := ""

	gidSync.Lock()
	defer gidSync.Unlock()

	if getLiteral() {
		return tabs
	}

	// find gid (CAUTION; anti-pattern)
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	gid, _ := strconv.ParseUint(string(b), 10, 64)

	index := -1
	for i, id := range gids {
		if id == gid {
			index = i
		}
	}
	if index >= 0 { // found
		tabs = strings.Repeat("\t", index)
	} else { // not found; new channel
		index = len(gids)
		gids = append(gids, gid)
		tabs = strings.Repeat("\t", index)
		p()
		p(tabs, "Ch| ", index+1)
		p(tabs, "|||")
	}

	return tabs
}
