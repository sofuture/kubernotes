package cluster

import (
	etcd "github.com/coreos/etcd/client"
)

// Interface to make our Cluster backend pluggable
type Backend interface {

	// Watch the specified key for changes since the provided index.
	WatchForChanges(key string, since uint64) (uint64, error)

	// Write the specified value to the provided key. If directory is true, create a directory. Only write the key if the current revision matches the prevIndex provided (i.e. has not been externally modified).
	WriteKey(key string, value string, directory bool, prevExist etcd.PrevExistType, prevIndex uint64) error

	// Read the value and last modified index of the specified key.
	ReadKey(key string) (string, uint64, error)

	// Read the immediate child values, and last modified index of the specified key.
	ReadKeyChildren(key string) ([]string, uint64, error)

	// Determine if the specified key exists.
	CheckIfKeyExists(key string) (bool, error)

	// Delete the specified key. If directory is true, delete all it's children as well.
	DeleteKey(key string, directory bool) error
}
