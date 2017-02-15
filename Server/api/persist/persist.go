package persist

import (
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TYPE_GENERAL = "general"
	TYPE_LIST    = "list"
	TYPE_DICT    = "dict"
)

func SaveToDb(storage *kvstorage.Storage, connectionString, dbName, collection string) error {
	session, err := getConnection(connectionString)
	if err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(dbName).C(collection)
	if err := c.DropCollection(); err != nil && err.Error() != "ns not found" {
		return err
	}
	keys := storage.Keys()
	count := 100
	bulk := c.Bulk()
	for _, key := range keys {
		if value, ok := storage.Get(key); ok {
			vType := TYPE_GENERAL
			switch value.(type) {
			case []interface{}:
				vType = TYPE_LIST
			case map[string]interface{}:
				vType = TYPE_DICT
			}
			bulk.Insert(map[string]interface{}{"key": key, "value": value, "type": vType})
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

func LoadFromDb(storage *kvstorage.Storage, connectionString, dbName, collection string) error {
	session, err := getConnection(connectionString)
	if err != nil {
		return err
	}
	defer session.Close()
	c := session.DB(dbName).C(collection)
	iter := c.Find(bson.M{}).Iter()
	item := struct {
		Key   string
		Value interface{}
		Type  string
	}{}
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
		storage.Set(item.Key, item.Value, 0)
	}
	return nil
}

func getConnection(connectionString string) (*mgo.Session, error) {
	session, err := mgo.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	return session, nil
}
