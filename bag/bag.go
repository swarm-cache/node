package bag

import "sync"

// Holds the actual key->pair data in a hashmap
var bag = make(map[string][]byte, 0)

// The required mutex to avoid data race
var mux = sync.Mutex{}

// Retrieves data
func Get(key string) (*[]byte, bool) {
	mux.Lock()
	defer mux.Unlock()

	v, i := bag[key]

	return &v, i
}

// Set data
func Set(key string, data *[]byte) {
	mux.Lock()
	defer mux.Unlock()

	bag[key] = *data
}

// Delete data
func Del(key string) {
	mux.Lock()
	defer mux.Unlock()

	delete(bag, key)
}
