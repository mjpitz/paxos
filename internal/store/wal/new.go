package wal

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go.etcd.io/bbolt"
)

func New(filePath string) (Log, error) {
	parts := strings.Split(filePath, "://")

	switch parts[0] {
	case "memory":
		return Memory(), nil
	case "boltdb":
		home := path.Dir(parts[1])
		if _, err := os.Stat(home); os.IsNotExist(err) {
			if err := os.Mkdir(home, 0755); err != nil {
				return nil, err
			}
		}

		db, err := bbolt.Open(parts[1], 0644, bbolt.DefaultOptions)
		if err != nil {
			return nil, err
		}
		return BoltDB(db)
	}

	return nil, fmt.Errorf("unrecognized scheme: %s", parts[0])
}
