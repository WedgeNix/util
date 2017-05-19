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
)

// HTTPLogin allows basic HTTP authorization for getting simple responses.
type HTTPLogin struct {
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
