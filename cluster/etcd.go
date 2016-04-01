package cluster

import (
	"fmt"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	etcd "github.com/coreos/etcd/client"
)

// Etcd backend for Kubernotes cluster
type Etcd struct {
	cfg    etcd.Config
	client etcd.Client
}

// Create a new Etcd backend, and connect to it.
func NewEtcd(servers []string) (*Etcd, error) {
	etcd := &Etcd{
		cfg: etcd.Config{
			Endpoints: servers,
			Transport: etcd.DefaultTransport,
		},
	}

	return etcd, etcd.connect()
}

// Block on an Etcd endpoint waiting to be notified of changes to it.
func (e *Etcd) WatchForChanges(key string, lastSeen uint64) (uint64, error) {

	// watch for changes only since lastSeen, if nonzero
	kapi := etcd.NewKeysAPI(e.client)
	watcher := kapi.Watcher(key, &etcd.WatcherOptions{
		Recursive:  true,
		AfterIndex: lastSeen,
	})

	// block, waiting for change
	resp, err := watcher.Next(context.Background())
	if err != nil {
		return 0, err
	}

	// return the new modified index
	return resp.Index, nil
}

// Store a value or directory in the backend. Fail if specified prevExist condition is not met, or if the key has changed since prevIndex, if specified.
func (e *Etcd) WriteKey(key string, value string, directory bool, prevExist etcd.PrevExistType, prevIndex uint64) (err error) {

	// set the value of a key
	kapi := etcd.NewKeysAPI(e.client)
	_, err = kapi.Set(context.Background(), key, value, &etcd.SetOptions{
		Dir:       directory,
		PrevExist: prevExist,
		PrevIndex: prevIndex,
	})

	if err != nil {
		// C+S locking error
		etcdErr, ok := err.(etcd.Error)
		if ok && etcdErr.Code == etcd.ErrorCodeTestFailed {
			return fmt.Errorf("optimistic lock of key failed %s %v", key, err)
		}
		return err
	}

	return err
}

// Delete a key from the backend. Recursively delete children of directories if directory is true.
func (e *Etcd) DeleteKey(key string, directory bool) (err error) {

	// delete the key
	kapi := etcd.NewKeysAPI(e.client)
	_, err = kapi.Delete(context.Background(), key, &etcd.DeleteOptions{
		Dir:       directory,
		Recursive: directory,
	})

	return err
}

// Retrieve the value of a key from the backend. Also return the last modified index of the key.
func (e *Etcd) ReadKey(key string) (string, uint64, error) {

	// get the value
	kapi := etcd.NewKeysAPI(e.client)
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		return "", 0, err
	}

	// return the value and last modified index
	if resp.Node != nil {
		return resp.Node.Value, resp.Node.ModifiedIndex, nil
	}

	return "", 0, nil
}

// Retrieve the values of all direct children of the specified key. Also return the last modified index of the specified key.
func (e *Etcd) ReadKeyChildren(key string) ([]string, uint64, error) {

	// get the value of the key
	kapi := etcd.NewKeysAPI(e.client)
	opts := &etcd.GetOptions{Recursive: true}
	resp, err := kapi.Get(context.Background(), key, opts)
	if err != nil {
		return nil, 0, err
	}

	// loop through all child nodes
	if resp.Node != nil {
		ret := make([]string, len(resp.Node.Nodes))
		for i, node := range resp.Node.Nodes {
			ret[i] = node.Value
		}
		return ret, resp.Node.ModifiedIndex, nil
	}

	return nil, 0, nil
}

// Determine if a key exists in the backend.
func (e *Etcd) CheckIfKeyExists(key string) (bool, error) {

	// check if a key exists
	kapi := etcd.NewKeysAPI(e.client)
	_, err := kapi.Get(context.Background(), key, nil)
	if err != nil {

		// specifically check for a key not found error
		etcdErr, ok := err.(etcd.Error)
		if ok && etcdErr.Code == etcd.ErrorCodeKeyNotFound {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

func (e *Etcd) connect() (err error) {
	e.client, err = etcd.New(e.cfg)
	return err
}
