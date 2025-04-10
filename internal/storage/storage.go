package storage

import "time"

type StorageInterface interface {
	Increment(key string, window time.Duration) (int64, error)
	IsBlocked(key string) (bool, error)
	Block(key string, duration time.Duration) error
}
