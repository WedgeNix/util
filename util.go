package util

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load assumes to read the
func Load(name string, v interface{}) {
	f, err := os.Open(name + ".json")
	if err != nil {
		fmt.Println(err)
	}
	json.NewDecoder(f).Decode(v)
}
