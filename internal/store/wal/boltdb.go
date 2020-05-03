package wal

import (
	"encoding/binary"

	"go.etcd.io/bbolt"
)

var bucketName = []byte("data")

func BoltDB(db *bbolt.DB) (Log, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &boltdbLog{
		db: db,
	}, nil
}

func uint64ToBytes(id uint64) ([]byte, error) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, id)
	return data, nil
}

func bytesToUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

type boltdbLog struct {
	db *bbolt.DB
}

func (l *boltdbLog) Close() error {
	return l.db.Close()
}

func (l *boltdbLog) Last() (*Entry, error) {
	var key []byte
	var value []byte

	err := l.db.View(func(tx *bbolt.Tx) error {
		key, value = tx.Bucket(bucketName).Cursor().Last()
		return nil
	})

	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	return &Entry{
		Id:   bytesToUint64(key),
		Data: value,
	}, nil
}

func (l *boltdbLog) Append(obj *Entry) error {
	key, err := uint64ToBytes(obj.Id)
	if err != nil {
		return err
	}

	err = l.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketName).Put(key, obj.Data)
	})

	return err
}

func (l *boltdbLog) Since(id uint64) ([]*Entry, error) {
	key, err := uint64ToBytes(id)
	if err != nil {
		return nil, err
	}

	data := make([]*Entry, 0)
	err = l.db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket(bucketName).Cursor()
		c.Seek(key)

		for key, value := c.Next(); value != nil; key, value = c.Next() {
			if value == nil { // reached end of log
				break
			}

			data = append(data, &Entry{
				Id:   bytesToUint64(key),
				Data: value,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

var _ Log = &boltdbLog{}
