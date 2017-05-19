package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	gomail "gopkg.in/gomail.v2"
)

// HTTPLogin allows basic HTTP authorization for getting simple responses.
type HTTPLogin struct {
	User string
	Pass string
}

// EmilLogin login for SMPT.
type EmailLogin struct {
	User string
	Pass string
}

// Get receives an HTTP response from the given URL using authorization.
func (lgn *HTTPLogin) Get(url string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	enc := base64.StdEncoding.EncodeToString([]byte(lgn.User + ":" + lgn.Pass))
	req.Header.Add("Authorization", "Basic "+enc)
	cl := http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

// Load assumes to read the
func Load(name string, v interface{}) {
	var r io.Reader
	if strings.Contains(name, "https://") || strings.Contains(name, "http://") {
		resp, err := http.Get(name)
		if err != nil {
			fmt.Println(err)
		}
		r = resp.Body
		defer resp.Body.Close()
	} else {
		f, err := os.Open(name + ".json")
		if err != nil {
			fmt.Println(err)
		}
		r = f
		defer f.Close()
	}
	json.NewDecoder(r).Decode(v)
}

// E reports the error if there is any.
func E(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

//Email to send basic emails from a particular gmail account.
func (lgn *EmailLogin) Email(to []string, subject string, body string, attachment string) {
	email := lgn.User + "@gmail.com"
	msg := gomail.NewMessage()
	msss := map[string][]string{"From": {email}, "To": to, "Subject": {subject}}
	msg.SetHeaders(msss)
	msg.SetBody("text/html", body)
	if len(attachment) > 0 {
		msg.Attach(attachment)
	}
	d := gomail.NewDialer("smtp.gmail.com", 587, email, lgn.Pass)
	err := d.DialAndSend(msg)
	if err != nil {
		fmt.Println(err)

	}
}
