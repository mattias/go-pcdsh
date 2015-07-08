package main

import (
	"fmt"
	"time"
)

type Cache struct {
	Expires time.Time
	Data    interface{}
}

var supercache map[string]Cache

func init() {
	supercache = make(map[string]Cache)
}

func GetCache(key string) (interface{}, error) {
	val, ok := supercache[key]

	if !ok {
		return nil, fmt.Errorf("cache: '%s' did not exist", key)
	}

	if val.Expires.After(time.Now()) {
		return nil, fmt.Errorf("cache: '%s' has expired", key)
	}

	return val.Data, nil
}

func SetCache(key string, val interface{}) {
	supercache[key] = Cache{Expires: time.Now().Add(time.Hour * 24), Data: val}
}
