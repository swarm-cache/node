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

	// If key is already set i don't check for F_MAX_CACHED_KEYS.
	// Work only with F_MAX_USED_MEMORY
	//
	if oData, is := bag[key]; is {
		if dLen+(usedMem-int64(len(oData))) > glob.F_MAX_USED_MEMORY && glob.F_MAX_USED_MEMORY != 0 {
			return fmt.Errorf("F_MAX_USED_MEMORY exceeded!")
		}

		usedMem = (usedMem - int64(len(oData))) + dLen
	} else {
		if bagSize == glob.F_MAX_CACHED_KEYS && glob.F_MAX_CACHED_KEYS != 0 {
			return fmt.Errorf("F_MAX_CACHED_KEYS exceeded!")
		}
		if dLen+kLen+usedMem > glob.F_MAX_USED_MEMORY && glob.F_MAX_USED_MEMORY != 0 {
			return fmt.Errorf("F_MAX_USED_MEMORY exceeded!")
		}

		bagSize++
		usedMem += kLen + dLen
	}

	bag[key] = *data

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
