package testtools

import (
	"fmt"
	"strings"

	etcd "github.com/coreos/etcd/client"
)

// simple mock backend for testing
type TestBackend map[string]string

func (t TestBackend) WatchForChanges(key string, since uint64) (uint64, error) {
	return 0, nil
}

func (t TestBackend) WriteKey(key string, value string, directory bool, prevExist etcd.PrevExistType, prevIndex uint64) error {
	t[key] = value
	return nil
}

func (t TestBackend) ReadKey(key string) (string, uint64, error) {
	val, ok := t[key]
	if !ok {
		return "", 0, fmt.Errorf("key %s does not exist", key)
	}
	return val, 0, nil
}

func (t TestBackend) ReadKeyChildren(key string) ([]string, uint64, error) {
	ret := make([]string, 0)
	for k, v := range t {
		if strings.HasPrefix(k, key) {
			ret = append(ret, v)
		}
	}
	return ret, 0, nil
}

func (t TestBackend) CheckIfKeyExists(key string) (bool, error) {
	_, ok := t[key]
	if !ok {
		return false, nil
	}
	return true, nil
}

func (t TestBackend) DeleteKey(key string, directory bool) error {
	delete(t, key)
	return nil
}
