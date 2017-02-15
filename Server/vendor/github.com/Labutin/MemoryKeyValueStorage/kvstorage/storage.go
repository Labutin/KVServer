package kvstorage

import (
	"errors"
	"github.com/Labutin/concurrent-map"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const TTLTimeout = 60 * time.Second

type Storage struct {
	cmap           concurrent_map.CMapInterface
	ttl            concurrent_map.CMapInterface
	done           chan interface{}
	wg             *sync.WaitGroup
	lastClearedTTL int64
	ttlTimeout     time.Duration
}

func NewKVStorage(chunks uint32, startTTLRemoval bool) *Storage {
	kvstorage := &Storage{}
	kvstorage.cmap = concurrent_map.NewCMap(chunks)
	kvstorage.ttl = concurrent_map.NewCMap(chunks)
	kvstorage.wg = &sync.WaitGroup{}
	kvstorage.lastClearedTTL = time.Now().Unix() - 1
	kvstorage.ttlTimeout = TTLTimeout
	if startTTLRemoval {
		go kvstorage.startTTLProcessing()
	}

	return kvstorage
}

// Set stores value for given key and TTL
func (t *Storage) Set(key string, value interface{}, TTL time.Duration) {
	t.cmap.Put(key, value)
	if TTL > 0 {
		value, ok := t.ttl.Get(key)
		var keys []string
		if !ok {
			keys = []string{}
		} else {
			keys = value.([]string)
		}
		keys = append(keys, key)
		t.ttl.Put(strconv.FormatInt(time.Now().Add(TTL).Unix(), 10), keys)
	} else {
		if TTL < 0 {
			t.cmap.Remove(key)
		}
	}
}

// Update updates value for given key
func (t *Storage) Update(key string, value interface{}) error {
	if !t.cmap.IsExist(key) {
		return errors.New("Key not found")
	}
	t.cmap.Put(key, value)
	return nil
}

// Remove deletes value for given key
func (t *Storage) Remove(key string) error {
	return t.cmap.Remove(key)
}

// Get returns value for given key
func (t *Storage) Get(key string) (interface{}, bool) {
	value, ok := t.cmap.Get(key)
	if !ok {
		return nil, false
	}
	tValue := value

	return tValue, true
}

// GetListElement returns i-th element from List value
func (t *Storage) GetListElement(key string, i int) (interface{}, error) {
	value, ok := t.Get(key)
	if !ok {
		return nil, errors.New("Key not found")
	}
	if vl, ok := value.([]interface{}); !ok {
		return nil, errors.New("Value not List " + reflect.TypeOf(value).String())
	} else {
		if len(vl) <= i {
			return nil, errors.New("Out of bound")
		}
		return vl[i], nil
	}
}

func (t *Storage) GetDictElement(key, dictKey string) (interface{}, error) {
	value, ok := t.Get(key)
	if !ok {
		return nil, errors.New("Key not found")
	}
	if vl, ok := value.(map[string]interface{}); !ok {
		return nil, errors.New("Value not Dictionary " + reflect.TypeOf(value).String())
	} else {
		if value, ok := vl[dictKey]; ok {
			return value, nil
		} else {
			return nil, errors.New("Key in dictionary not found")
		}
	}
}

// Keys returns all keys in map
func (t *Storage) Keys() []string {
	return t.cmap.Keys()
}

// clearTTLExpiredRecords removes old records from map
func (t *Storage) clearTTLExpiredRecords() {
	lastTime := time.Now().Unix()
	for i := t.lastClearedTTL; i < lastTime; i++ {
		strI := strconv.FormatInt(i, 10)
		if value, ok := t.ttl.Get(strI); ok {
			keysToRemove := value.([]string)
			for _, key := range keysToRemove {
				t.Remove(key)
			}
			t.ttl.Remove(strI)
		}
	}
	t.lastClearedTTL = lastTime
}

// ttlRemoval starts removing TTL expired records
func (t *Storage) ttlRemoval() {
	t.wg.Add(1)
	defer t.wg.Done()
	for true {
		select {
		case <-t.done:
			return

		case <-time.After(t.ttlTimeout):
			t.clearTTLExpiredRecords()
		}
	}
}

// stopTTLProcessing stops processing records TTL
func (t *Storage) stopTTLProcessing() {
	close(t.done)
	t.wg.Wait()
}

// startTTLProcessing starts processing records TTL
func (t *Storage) startTTLProcessing() {
	t.done = make(chan interface{})
	go t.ttlRemoval()
}
