package concurrent_map

import (
	"errors"
	"hash/fnv"
	"sync"
)

// Concurrent map interface
type CMapInterface interface {
	Put(key string, value interface{})
	Get(key string) (interface{}, bool)
	Remove(key string) error
	IsExist(key string) bool
	Keys() []string
	Count() int
}

// A thread safe map
type CMap []*chunk

// A thread safe chunk of map
type chunk struct {
	data map[string]interface{}
	sync.RWMutex
}

// NewCMap creates a new concurrent map
func NewCMap(chunks uint32) CMapInterface {
	cmap := make(CMap, chunks)
	for i := uint32(0); i < chunks; i++ {
		cmap[i] = &chunk{data: map[string]interface{}{}}
	}
	return cmap
}

// getShard returns chunk for given key
func (t CMap) getShard(key string) *chunk {
	fnv := fnv.New32()
	fnv.Write([]byte(key))
	hash := fnv.Sum32() % uint32(len(t))
	return t[hash]
}

// Put sets given value for key
func (t CMap) Put(key string, value interface{}) {
	shard := t.getShard(key)
	shard.Lock()
	shard.data[key] = value
	shard.Unlock()
}

// Get returns value for given key
func (t CMap) Get(key string) (interface{}, bool) {
	shard := t.getShard(key)
	shard.RLock()
	value, ok := shard.data[key]
	shard.RUnlock()
	return value, ok

}

// Remove deletes record with given key in map
func (t CMap) Remove(key string) error {
	if !t.IsExist(key) {
		return errors.New("Key not found.")
	}
	shard := t.getShard(key)
	shard.Lock()
	delete(shard.data, key)
	shard.Unlock()
	return nil
}

// IsExist check for key present in map
func (t CMap) IsExist(key string) bool {
	shard := t.getShard(key)
	shard.RLock()
	_, ok := shard.data[key]
	shard.RUnlock()
	return ok
}

// Count returns total number elements in the map
func (t CMap) Count() int {
	count := 0
	for i := 0; i < len(t); i++ {
		chunk := t[i]
		chunk.RLock()
		count += len(chunk.data)
		chunk.RUnlock()
	}
	return count
}

// Keys returns all keys in map
func (t CMap) Keys() []string {
	keys := []string{}
	for i := 0; i < len(t); i++ {
		t[i].RLock()
		for k, _ := range t[i].data {
			keys = append(keys, k)
		}
		t[i].RUnlock()
	}
	return keys
}
