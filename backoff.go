package util

import (
	"math"
	"math/rand"
	"net/http"
	"time"
)

// var (
// 	BackoffTimeout = 48 * time.Second

// 	backoffLookupl sync.Mutex
// 	backoffLookup  = map[interface{}]*backoffData{}
// )

// type backoffData struct {
// 	sync.Mutex
// 	wait  int
// 	tries float64
// }

type Backoff struct {
	Step    time.Duration
	Attempt int
	Timeout time.Duration
}

func (b *Backoff) Wait(resp *http.Response, err error) (*http.Response, error) {
	if b.Timeout == 0 {
		b.Timeout = 64000
	} else if b.Timeout < 32000 {
		panic("backoff timeout too low; 32s minimum")
	}
	if b.Step.Seconds() < 1 {
		b.Step = 1000 * time.Millisecond
	}
	maxWait := b.Timeout.Seconds() * 1000
	wait := int(math.Min(maxWait, math.Pow(2, float64(b.Attempt))+float64(rand.Intn(int(b.Step.Seconds()*1000)))+1))
	if err != nil || resp.StatusCode > 200 {
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}
	b.Attempt++
	return resp, err
}

// func Backoff(f func() (interface{}, error)) {
// 	backoffLookupl.Lock()
// 	data, found := backoffLookup[f]
// 	if !found {
// 		backoffLookup[f] = &backoffData{}
// 	}
// 	backoffLookupl.Unlock()

// 	maxWait := int(BackoffTimeout.Seconds() * 1000)

// 	var res interface{}
// 	var err error

// 	data.Lock()
// 	for ; data.wait < maxWait && res == nil; data.tries++ {
// 		res, err = f()
// 		if err != nil {
// 			data.wait = int(math.Min(float64(maxWait), math.Pow(2, data.tries)+float64(rand.Intn(1000))+1))
// 			time.Sleep(time.Duration(data.wait) * time.Millisecond)
// 		}
// 	}
// 	data.Unlock()

// 	println("util.go: backoff timed out")
// }
