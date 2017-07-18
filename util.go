package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"log"

	"io/ioutil"

	"gopkg.in/gomail.v2"
)

// EmailLogin login for SMTP.
type EmailLogin struct {
	User string
	Pass string
	SMTP string
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
	return gomail.NewDialer(l.SMTP, 587, l.User, l.Pass).DialAndSend(msg)
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
	return l.req(http.MethodGet, url, nil)
}

// Post receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) Post(url string, v interface{}) *http.Response {
	b, err := json.Marshal(v)
	R(err)
	return l.req(http.MethodPost, url, bytes.NewReader(b))
}

// Get receives an HTTP response from the given URL using authorization.
func (l HTTPLogin) req(method string, url string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, url, body)
	E(err)
	req.Header.Add("Authorization", "Basic "+l.Base64())
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
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

// E reports the error if there is any and exits.
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
