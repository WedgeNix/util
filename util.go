package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Load assumes to read the
func Load(name string, v interface{}) {
	var r io.Reader
	if strings.Contains(name, "https://") || strings.Contains(name, "http://") {
		res, err := http.Get(name)
		if err != nil {
			fmt.Println(err)
		}
		r = res.Body
		defer res.Body.Close()
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
