package bag

import (
	"fmt"
	"sync"

	"github.com/tgbv/swarm-cache/glob"
)

// Holds the actual key->pair data in a hashmap
var bag = make(map[string][]byte, 0)

// The required mutex to avoid data race
var mux = sync.Mutex{}

var bagSize int64 = 0
var usedMem int64 = 0 // bytes

// Retrieves data
func Get(key string) (*[]byte, bool) {
	mux.Lock()
	defer mux.Unlock()

	v, i := bag[key]

	return &v, i
}

// Set data
func Set(key string, data *[]byte) error {
	mux.Lock()
	defer mux.Unlock()

	kLen := int64(len(key))
	dLen := int64(len(*data))

	if bagSize == glob.F_MAX_CACHED_KEYS && glob.F_MAX_CACHED_KEYS != 0 {
		return fmt.Errorf("F_MAX_CACHED_KEYS exceeded!")
	}
	if dLen+kLen+usedMem > glob.F_MAX_USED_MEMORY/(1024*1024) && glob.F_MAX_USED_MEMORY != 0 {
		return fmt.Errorf("F_MAX_USED_MEMORY exceeded!")
	}

	bag[key] = *data
	bagSize++
	usedMem += kLen + dLen

	return nil
}

// Delete data
func Del(key string) {
	mux.Lock()
	defer mux.Unlock()

	usedMem -= int64((len(key) + len(bag[key])))
	bagSize--
	delete(bag, key)
}
