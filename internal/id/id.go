package id

import (
	"math/rand"
	"time"
)

func New(key string) (id uint16) {
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
	delete(m[key], k)
}

var m = make(map[string]map[uint16]bool)

func gen() uint16 {
	rand.Seed(time.Now().UnixNano())
	return uint16(rand.Intn(99999))
}
