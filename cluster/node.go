package cluster

import (
	"encoding/json"
	"fmt"
	"log"

	etcd "github.com/coreos/etcd/client"
)

// Represents a node available for running jobs in a Kubernotes cluster.
type Node struct {
	// Arbitrary unit to represent available IO shares for scheduling
	BlockIOShares int

	// Arbitrary unit to represent available CPU shares for scheduling
	CPUShares int

	// Megabytes available for job scheduling
	MemoryMegabytes int

	Endpoint          string
	Name              string
	Namespace         string
	JobIDs            []string
	LastModifiedIndex uint64 `json:"-"`
}

type Resources struct {
	BlockIOShares   int
	CPUShares       int
	MemoryMegabytes int
}

// Get the name of the node.
func (n *Node) GetName() string {
	return n.Name
}

// Get the name of the cluster this node belongs to.
func (n *Node) GetNamespace() string {
	return n.Namespace
}

// Deserialize a Node from JSON string.
func (n *Node) Deserialize(jsonBlob string) error {
	err := json.Unmarshal([]byte(jsonBlob), n)
	return err
}

// Serialize a Node to JSON string.
func (n *Node) Serialize() (string, error) {
	jsonBlob, err := json.Marshal(n)
	return string(jsonBlob), err
}

// Join as a node of an existing cluster.
func (n *Node) JoinCluster(backend Backend) error {
	// see if node exists
	exists, err := backend.CheckIfKeyExists(getNodePath(n.Namespace, n.Name))
	if err != nil {
		return fmt.Errorf("problem determining cluster membership %v", err)
	}

	// create it, if it doesn't
	if !exists {
		return n.SaveIfNotModified(backend, etcd.PrevNoExist)
	} else {
		return n.Load(backend)
	}

	return err
}

// Save information to provided backend, if not modified externally, or to be newly created.
func (n *Node) SaveIfNotModified(backend Backend, exists etcd.PrevExistType) error {
	json, err := n.Serialize()
	if err != nil {
		return fmt.Errorf("problem serializing node %v", err)
	}
	err = backend.WriteKey(getNodePath(n.Namespace, n.Name), json, false, exists, n.LastModifiedIndex)
	if err != nil {
		return fmt.Errorf("problem joining cluster %v", err)
	}
	return nil
}

// Loads an existing Node from it's information stored in the backend.
func (n *Node) Load(backend Backend) error {
	var json string
	var err error
	json, n.LastModifiedIndex, err = backend.ReadKey(getNodePath(n.Namespace, n.Name))
	if err != nil {
		return fmt.Errorf("could not get cluster node %v", err)
	}
	err = n.Deserialize(json)
	return err
}

// Remove this Node from the current cluster.
func (n *Node) LeaveCluster(backend Backend) error {
	// delete cluster membership
	err := backend.DeleteKey(getNodePath(n.Namespace, n.Name), true)
	if err != nil {
		return fmt.Errorf("problem leaving cluster %v", err)
	}

	return nil
}

// Get list of jobs currently assigned to this node.
func (n *Node) GetJobs(backend Backend) ([]Job, error) {
	ret := make([]Job, len(n.JobIDs))
	namespace := NewNamespace(n.Namespace)
	for i, jobID := range n.JobIDs {
		job, err := namespace.GetJob(backend, jobID)
		if err != nil {
			return nil, err
		}
		ret[i] = *job
	}
	return ret, nil
}

func (n *Node) GetFreeResources(backend Backend) (*Resources, error) {
	resources := &Resources{
		CPUShares:       n.CPUShares,
		BlockIOShares:   n.BlockIOShares,
		MemoryMegabytes: n.MemoryMegabytes,
	}

	// look at assigned jobs to determine current resource utilization
	jobs, err := n.GetJobs(backend)
	if err != nil {
		return nil, err
	}
	for _, job := range jobs {
		resources.CPUShares -= job.CPUShares
		resources.BlockIOShares -= job.BlockIOWeight
		resources.MemoryMegabytes -= job.MemoryLimitMegabytes
	}

	return resources, nil
}

// Assign a job for a node to run.
func (n *Node) AssignJob(backend Backend, jobID string) error {
	for _, v := range n.JobIDs {
		if v == jobID {
			return fmt.Errorf("cannot run duplicate job %s on node %s", jobID, n.Name)
		}
	}
	n.JobIDs = append(n.JobIDs, jobID)
	return n.SaveIfNotModified(backend, etcd.PrevExist)
}

// Unassign a job from a node.
func (n *Node) UnassignJob(backend Backend, jobID string) error {
	for i, v := range n.JobIDs {
		if v == jobID {
			// remove the specified job id
			n.JobIDs = append(n.JobIDs[:i], n.JobIDs[i+1:]...)
		}
	}
	return n.SaveIfNotModified(backend, etcd.PrevExist)
}

// Block on an endpoint waiting to be notified of scheduling changes.
func (n *Node) WatchForChanges(backend Backend, since uint64) (uint64, error) {
	path := getNodeChangesPath(n.Namespace, n.Name)
	log.Println("watching for changes to", path)
	return backend.WatchForChanges(path, since)
}
