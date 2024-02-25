package main

import (
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

type Store interface {
	GetAll(entityname string, size int) (error, [][]byte)
	Get(entityname string, key []byte) (error, []byte)
	Set(entityname string, key []byte, value []byte) error
	Delete(entityname string, key []byte) error
}

type DB struct {
	db *badger.DB
}

func NewDB(isInMemory bool) *DB {
	opt := badger.DefaultOptions("")
	if isInMemory {
		opt = opt.WithInMemory(true)
	}
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal("Couldn't Open Database")
	}
	return &DB{db}
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) GetAll(entityname string, size int) (error, [][]byte) {
	result := make([][]byte, 0)
	var err error
	prefix := []byte(entityname)
	db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err = item.Value(func(v []byte) error {
				result = append(result, v)
				if len(result) == size {
					return nil
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err, nil
	}
	return nil, result
}

func (db *DB) Get(entityname string, key []byte) (error, []byte) {
	var result []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append([]byte(entityname+"-"), key...))
		if err != nil {
			return err
		}
		err = item.Value(func(v []byte) error {
			result = v
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err, nil
	}
	return nil, result
}

func (db *DB) Set(entityname string, key []byte, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Delete(entityname string, key []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(append([]byte(entityname+"-"),
			key...))
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Put(entityname string, key []byte, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Patch(entityname string, key []byte, valye []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(append([]byte(entityname+"-"), key...))
		if err != nil {
			return err
		}
		var value []byte
		err = item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return err
		}
		value = append(value, valye...)
		err = txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
