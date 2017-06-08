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

	"gopkg.in/gomail.v2"
)

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
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

// E reports the error if there is any and exits.
func E(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// R reports the error if there is any.
func R(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// S strings together the arguments as they are.
func S(args ...interface{}) string {
	str := ""
	for _, a := range args {
		str += fmt.Sprint(a)
	}
	return str
}

// P prints the arguments as they are.
func P(args ...interface{}) {
	for _, a := range args {
		fmt.Print(a)
	}
	fmt.Println()
}
