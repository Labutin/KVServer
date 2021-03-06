package kvstorage

import (
	"errors"
	"github.com/Labutin/concurrent-map"
	"strconv"
	"sync"
	"time"
)

const TTLTimeout = 60 * time.Second

type ttlValue struct {
	sync.RWMutex
	keys []string
}

type cmapValue struct {
	value interface{}
	ttl   int64
}

type Storage struct {
	cmap           concurrent_map.CMapInterface
	ttl            concurrent_map.CMapInterface
	ttlMutex       sync.Mutex
	done           chan interface{}
	wg             *sync.WaitGroup
	lastClearedTTL int64
	ttlTimeout     time.Duration
}

// NewKVStorage creates new key value storage
func NewKVStorage(chunks uint32, startTTLRemoval bool) *Storage {
	kvstorage := &Storage{}
	kvstorage.cmap = concurrent_map.NewCMap(chunks)
	kvstorage.ttl = concurrent_map.NewCMap(chunks)
	kvstorage.wg = &sync.WaitGroup{}
	kvstorage.lastClearedTTL = time.Now().Unix() - 1
	kvstorage.ttlTimeout = TTLTimeout
	if startTTLRemoval {
		go kvstorage.StartTTLProcessing()
	}

	return kvstorage
}

// ensureTTLKey creates empty record with given key in ttl map
func (t *Storage) ensureTTLKey(key string) *ttlValue {
	t.ttlMutex.Lock()
	defer t.ttlMutex.Unlock()
	if res, ok := t.ttl.Get(key); !ok {
		newValue := &ttlValue{}
		t.ttl.Put(key, newValue)
		return newValue
	} else {
		return res.(*ttlValue)
	}
}

// Set stores value for given key and TTL
func (t *Storage) Set(key string, value interface{}, TTL time.Duration) {
	storeValue := &cmapValue{value: value}
	whenToDelete := int64(0)
	if TTL > 0 {
		whenToDelete = time.Now().Add(TTL).Unix()
		storeValue.ttl = whenToDelete
	}
	t.cmap.Put(key, storeValue)
	if TTL > 0 {
		ttlKey := strconv.FormatInt(whenToDelete, 10)
		var ttlRecord *ttlValue
		if value, ok := t.ttl.Get(ttlKey); !ok {
			ttlRecord = t.ensureTTLKey(ttlKey)
		} else {
			ttlRecord = value.(*ttlValue)
		}
		ttlRecord.Lock()
		ttlRecord.keys = append(ttlRecord.keys, key)
		ttlRecord.Unlock()
	} else {
		if TTL < 0 {
			t.cmap.Remove(key)
		}
	}
}

// Update updates value for given key
func (t *Storage) Update(key string, value interface{}, TTL time.Duration) error {
	if !t.cmap.IsExist(key) {
		return errors.New("Key not found")
	}
	t.Set(key, value, TTL)
	return nil
}

// Remove deletes value for given key
func (t *Storage) Remove(key string) error {
	return t.cmap.Remove(key)
}

// getRaw returns data
func (t *Storage) getRaw(key string) (*cmapValue, bool) {
	value, ok := t.cmap.Get(key)
	if !ok {
		return nil, false
	}
	tValue := value.(*cmapValue)

	return tValue, true
}

// Get returns value for given key
func (t *Storage) Get(key string) (interface{}, bool) {
	cmapValue, ok := t.getRaw(key)
	if !ok {
		return nil, false
	}
	return cmapValue.value, true
}

// GetWithTTL returns value and TTL for given key
func (t *Storage) GetWithTTL(key string) (interface{}, int64, bool) {
	cmapValue, ok := t.getRaw(key)
	if !ok {
		return nil, int64(0), false
	}
	return cmapValue.value, cmapValue.ttl, true
}

// GetListElement returns i-th element from List value
func (t *Storage) GetListElement(key string, i int) (interface{}, error) {
	value, ok := t.Get(key)
	if !ok {
		return nil, errors.New("Key not found")
	}
	if vl, ok := value.([]interface{}); !ok {
		return nil, errors.New("Value not List")
	} else {
		if len(vl) <= i || i < 0 {
			return nil, errors.New("Out of bound")
		}
		return vl[i], nil
	}
}

// GetDictElement returns element with key 'dictKey' from dictionary with given key
func (t *Storage) GetDictElement(key, dictKey string) (interface{}, error) {
	value, ok := t.Get(key)
	if !ok {
		return nil, errors.New("Key not found")
	}
	if vl, ok := value.(map[string]interface{}); !ok {
		return nil, errors.New("Value not Dictionary")
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
			keysToRemove := value.(*ttlValue).keys
			for _, key := range keysToRemove {
				if v, ok := t.cmap.Get(key); ok {
					tv := v.(*cmapValue)
					if tv.ttl <= lastTime {
						t.Remove(key)
					}
				}
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
func (t *Storage) StopTTLProcessing() {
	close(t.done)
	t.wg.Wait()
}

// startTTLProcessing starts processing records TTL
func (t *Storage) StartTTLProcessing() {
	t.done = make(chan interface{})
	go t.ttlRemoval()
}
