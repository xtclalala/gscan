package id

import (
	"math/rand"
	"sync"
	"time"
)

func New(key string) (id uint16) {
	mx.Lock()
	defer mx.Unlock()
	id = gen()
	val, ok := m[key]
	if !ok {
		val = make(map[uint16]bool)
		m[key] = val
	}
	var o = true
	for o {
		if _, o = val[id]; !o {
			val[id] = true
			return
		}
		id = gen()
	}

	return
}

func Del(key string, k uint16) {
	mx.Lock()
	delete(m[key], k)
	mx.Unlock()
}

func Length(key string) int {
	mx.Lock()
	defer mx.Unlock()
	return len(m[key])

}

var m = make(map[string]map[uint16]bool)
var mx sync.RWMutex

func gen() uint16 {
	rand.Seed(time.Now().UnixNano())
	return uint16(rand.Intn(99999))
}
