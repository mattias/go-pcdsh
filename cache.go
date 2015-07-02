package main
import "fmt"

var supercache map[string]interface{}

func init() {
	supercache = make(map[string]interface{})
}

func GetCache(key string) (interface{}, error) {
	val, ok := supercache[key]

	if !ok {
		return nil, fmt.Errorf("cache: '%s' did not exist", key)
	}

	return val, nil
}

func SetCache(key string, val interface{}) {
	supercache[key]	= val
}