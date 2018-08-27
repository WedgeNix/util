package util

import (
	"math"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

type Backoff struct {
	initOnce sync.Once

	Step    time.Duration
	Max     int
	Timeout time.Duration
	Attempt int
}

func (b *Backoff) init() {
	b.initOnce.Do(func() {
		if b.Max == 0 {
			if b.Timeout == 0 {
				b.Timeout = 64000 * time.Millisecond
			} else if b.Timeout < 32000*time.Millisecond {
				panic("backoff timeout too low; 32 seconds minimum")
			}
		}
		if b.Step.Seconds() < 1 {
			b.Step = 1000 * time.Millisecond
		}
	})
}

type backoffNext struct {
	*Backoff
	i   interface{}
	err error
}

func (b *Backoff) Func(i interface{}, err error) backoffNext {
	b.init()
	b.Attempt++
	return backoffNext{b, i, err}
}

func (b backoffNext) Wait(v interface{}, err ...*error) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		panic("not a pointer")
	}

	ri := reflect.ValueOf(b.i)
	if ri.Kind() != reflect.Ptr {
		ri = ri.Elem()
	}
	if !ri.IsNil() && rv.Elem().CanSet() {
		rv.Elem().Set(ri)
	}
	if len(err) > 0 {
		*err[0] = b.err
	}

	if b.err == nil || b.Attempt >= b.Max {
		return false
	}

	wait := time.Duration(math.Min(
		b.Timeout.Seconds()*1000,
		math.Pow(2, float64(b.Attempt))+float64(rand.Intn(int(b.Step.Seconds()*1000)))+1,
	))
	time.Sleep(wait * time.Millisecond)
	return true
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
