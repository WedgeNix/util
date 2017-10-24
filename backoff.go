package util

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	BackoffTimeout = 48 * time.Second

	backoffLookupl sync.Mutex
	backoffLookup  = map[interface{}]*backoffData{}
)

type backoffData struct {
	sync.Mutex
	wait  int
	tries float64
}

func Backoff(f func() (interface{}, error)) {
	backoffLookupl.Lock()
	data, found := backoffLookup[f]
	if !found {
		backoffLookup[f] = &backoffData{}
	}
	backoffLookupl.Unlock()

	maxWait := int(BackoffTimeout.Seconds() * 1000)

	var res interface{}
	var err error

	data.Lock()
	for ; data.wait < maxWait && res == nil; data.tries++ {
		res, err = f()
		if err != nil {
			data.wait = int(math.Min(float64(maxWait), math.Pow(2, data.tries)+float64(rand.Intn(1000))+1))
			time.Sleep(time.Duration(data.wait) * time.Millisecond)
		}
	}
	data.Unlock()

	println("util.go: backoff timed out")
}
