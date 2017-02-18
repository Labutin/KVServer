package persist

import (
	"github.com/Labutin/KVServer/Server/logs"
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type PersistStorage interface {
	SaveToDb(*kvstorage.Storage) error
	LoadFromDb(*kvstorage.Storage) error
}

type MongoStorage struct {
	connectionString string
	dbName           string
	collection       string
}

const (
	TYPE_GENERAL = "general"
	TYPE_LIST    = "list"
	TYPE_DICT    = "dict"
	GOROUTINE_ID = "persist"
)

func NewMongoStorage(connectionString, dbName, collection string) *MongoStorage {
	mongoStorage := &MongoStorage{
		connectionString: connectionString,
		dbName:           dbName,
		collection:       collection,
	}
	return mongoStorage
}

func (t MongoStorage) SaveToDb(storage *kvstorage.Storage) error {
	session, err := t.getConnection(t.connectionString)
	if err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(t.dbName).C(t.collection)
	if err := c.DropCollection(); err != nil && err.Error() != "ns not found" {
		return err
	}
	keys := storage.Keys()
	count := 100
	bulk := c.Bulk()
	for _, key := range keys {
		if value, ttl, ok := storage.GetWithTTL(key); ok {
			vType := TYPE_GENERAL
			switch value.(type) {
			case []interface{}:
				vType = TYPE_LIST
			case map[string]interface{}:
				vType = TYPE_DICT
			}
			bulk.Insert(map[string]interface{}{"key": key, "value": value, "type": vType, "ttl": ttl})
			count--
			if count == 0 {
				count = 100
				if _, err := bulk.Run(); err != nil {
					return err
				}
			}
		}
	}
	if count < 100 {
		if _, err := bulk.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (t MongoStorage) LoadFromDb(storage *kvstorage.Storage) error {
	session, err := t.getConnection(t.connectionString)
	if err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(t.dbName).C(t.collection)
	iter := c.Find(bson.M{}).Iter()
	item := struct {
		Key   string
		Value interface{}
		TTL   int64
		Type  string
	}{}
	currentTime := time.Now().Unix()
	for iter.Next(&item) {
		if item.Type == TYPE_DICT {
			if bM, ok := item.Value.(bson.M); ok {
				tmpValue := map[string]interface{}{}
				for bMKey := range bM {
					tmpValue[bMKey] = bM[bMKey]
				}
				item.Value = tmpValue
			}
		}
		if currentTime < item.TTL || item.TTL == 0 {
			var nsec time.Duration = 0
			if item.TTL > 0 {
				nsec = time.Duration(time.Unix(item.TTL, 0).Sub(time.Now()).Nanoseconds())
			}
			storage.Set(item.Key, item.Value, time.Nanosecond*nsec)
			log.Println(logs.MakeLogString(logs.DEBUG, GOROUTINE_ID, "Loaded key: "+item.Key, nil))
		} else {
			log.Println(logs.MakeLogString(logs.DEBUG, GOROUTINE_ID, "Skipped key: "+item.Key, nil))
		}
	}
	return nil
}

func (t MongoStorage) getConnection(connectionString string) (*mgo.Session, error) {
	session, err := mgo.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return session, nil
}
