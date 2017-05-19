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
)

// HTTPLogin allows basic HTTP authorization for getting simple responses.
type HTTPLogin struct {
	User string
	Pass string
}

// Get receives an HTTP response from the given URL using authorization.
func (lgn *HTTPLogin) Get(url string) *http.Response {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	E(err)
	enc := base64.StdEncoding.EncodeToString([]byte(lgn.User + ":" + lgn.Pass))
	req.Header.Add("Authorization", "Basic "+enc)
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

// E reports the error if there is any.
func E(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
