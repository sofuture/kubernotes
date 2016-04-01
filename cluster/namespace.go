package cluster

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
)

// Represents a Kubernotes namespace.
type Namespace struct {
	namespace     string
	lastSeenIndex uint64
}

// Create a Namespace representing a Kubernotes scheduling namespace.
func NewNamespace(namespace string) *Namespace {
	return &Namespace{
		namespace: namespace,
	}
}

// Get the name of this namespace.
func (n *Namespace) GetName() string {
	return n.namespace
}

// Get the Node for the specified name if it already exists.
func (n *Namespace) GetNode(backend Backend, nodeName string) (*Node, error) {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Name:      nodeName,
		Namespace: n.namespace,
	}

	err = node.Load(backend)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// Join the provided name to this namespace if it does not already exist.
func (n *Namespace) CreateNode(backend Backend, node *Node) error {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return err
	}

	return node.JoinCluster(backend)
}

// Get all Nodes in the specified namespace.
func (n *Namespace) GetNodes(backend Backend) ([]Node, error) {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return nil, err
	}

	// load all children of the namespaces node path
	nodes, _, err := backend.ReadKeyChildren(getNodesPath(n.namespace))
	if err != nil {
		return nil, err
	}

	// load each of the children as a node
	ret := make([]Node, len(nodes))
	for i, node := range nodes {
		n := &Node{}
		err := n.Deserialize(node)
		if err != nil {
			return nil, err
		}
		ret[i] = *n
	}

	return ret, nil
}

// Store a job definition in the namespace, overwriting it if it exists.
func (n *Namespace) CreateJob(backend Backend, job *Job) error {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return err
	}

	// get JSON for Job
	json, err := job.Serialize()
	if err != nil {
		return err
	}

	// create job, overwriting existing
	err = backend.WriteKey(getJobPath(n.namespace, job.ID), string(json), false, etcd.PrevNoExist, 0)
	if err != nil {
		return fmt.Errorf("problem creating job %v", err)
	}

	return nil
}

// Retrieve a job definition from the namespace.
func (n *Namespace) GetJob(backend Backend, jobID string) (*Job, error) {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return nil, err
	}

	// get the stored JSON
	json, _, err := backend.ReadKey(getJobPath(n.namespace, jobID))
	if err != nil {
		return nil, fmt.Errorf("problem retrieving job %v", err)
	}

	// load into a Job
	job := &Job{}
	err = job.Deserialize(json)
	return job, err
}

// Find Node that's running a job.
func (n *Namespace) GetNodeRunningJob(backend Backend, jobID string) (*Node, error) {

	// ensure the namespace exists
	err := n.checkOrCreateNamespace(backend)
	if err != nil {
		return nil, err
	}

	nodes, err := n.GetNodes(backend)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		for _, nodeJobID := range node.JobIDs {
			if jobID == nodeJobID {
				return &node, nil
			}
		}
	}

	return nil, nil
}

func (n *Namespace) checkOrCreateNamespace(backend Backend) error {

	// see if namespace exists
	namespaceExists, err := backend.CheckIfKeyExists(getNamespacePath(n.namespace))
	if err != nil {
		return fmt.Errorf("problem accessing namespace %v", err)
	}

	// create it, if it doesn't
	if !namespaceExists {
		err = backend.WriteKey(getNamespacePath(n.namespace), "", true, etcd.PrevNoExist, 0)
		if err != nil {
			return fmt.Errorf("problem creating namespace %v", err)
		}
	}

	return nil
}
