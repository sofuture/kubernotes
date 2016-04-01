package cluster

import (
	"testing"

	"github.com/sofuture/kubernotes/testtools"
)

func TestNamespaceCreatesNamespaceAndNewNode(t *testing.T) {
	// create a namespace and node
	c := NewNamespace("test")
	tb := testtools.TestBackend{}

	err := c.CreateNode(tb, &Node{Name: "testnode", Namespace: "test"})
	if err != nil {
		t.Fatal("error creating node", err)
	}

	// make sure we can get the node
	_, err = c.GetNode(tb, "testnode")
	if err != nil {
		t.Fatal("error getting node", err)
	}

	if _, ok := tb["/kubernotes/clusters/test"]; !ok {
		t.Fatal("cluster not created")
	}

	if _, ok := tb["/kubernotes/clusters/test/nodes/testnode"]; !ok {
		t.Fatal("clusternode not created")
	}
}

func TestNamespaceCreatesNamespaceAndMultipleNodes(t *testing.T) {
	// create a namespace and multiple nodes
	c := NewNamespace("test")
	tb := testtools.TestBackend{}

	err := c.CreateNode(tb, &Node{Name: "testnode", Namespace: "test"})
	if err != nil {
		t.Fatal("error creating node", err)
	}

	err = c.CreateNode(tb, &Node{Name: "testnode1", Namespace: "test"})
	if err != nil {
		t.Fatal("error creating node", err)
	}

	err = c.CreateNode(tb, &Node{Name: "testnode2", Namespace: "test"})
	if err != nil {
		t.Fatal("error creating node", err)
	}

	// make sure we get all the nodes back
	nodes, err := c.GetNodes(tb)
	if err != nil || len(nodes) != 3 {
		t.Fatal("error getting nodes", err)
	}
}

func TestNamespaceCanFindRunningJob(t *testing.T) {
	// create a node and namespace
	c := NewNamespace("test")
	tb := testtools.TestBackend{}

	// assign a job to the node
	err := c.CreateNode(tb, &Node{Name: "testnode", Namespace: "test", JobIDs: []string{"foo"}})
	if err != nil {
		t.Fatal("error creating node", err)
	}

	// create a job
	job := &Job{
		ID:                   "foo",
		UnitFile:             "unit file",
		CPUShares:            10,
		BlockIOWeight:        10,
		MemoryLimitMegabytes: 10,
	}

	err = c.CreateJob(tb, job)
	if err != nil {
		t.Fatal("error creating node", err)
	}

	// get the job from the node
	node, err := c.GetNodeRunningJob(tb, "foo")
	if err != nil || node == nil {
		t.Fatal("couldn't find node running job", err)
	}

	// try to get a job that doesnt exist
	node, err = c.GetNodeRunningJob(tb, "notreal")
	if err != nil {
		t.Fatal("should not have gotten an error looking for job", err)
	}
	if node != nil {
		t.Fatal("should not have found a node running fake job", node)
	}
}

func TestNamespaceCreatesJobs(t *testing.T) {
	// create a namespace
	c := NewNamespace("test")
	tb := testtools.TestBackend{}

	job := &Job{
		ID:                   "foo",
		UnitFile:             "unit file",
		CPUShares:            10,
		BlockIOWeight:        10,
		MemoryLimitMegabytes: 10,
	}

	// create a job
	err := c.CreateJob(tb, job)
	if err != nil {
		t.Fatal("error creating node", err)
	}

	if _, ok := tb["/kubernotes/clusters/test"]; !ok {
		t.Fatal("cluster not created")
	}

	if _, ok := tb["/kubernotes/clusters/test/jobs/foo"]; !ok {
		t.Fatal("job not created")
	}

	// get nonexistent job
	job1, err := c.GetJob(tb, "foo1")
	if err == nil {
		t.Fatal("expected error getting nonexistent job")
	}

	// get existing job
	job1, err = c.GetJob(tb, "foo")
	if err != nil {
		t.Fatal("failed to get job", err)
	}

	if job1.ID != job.ID || job1.UnitFile != job.UnitFile || job1.CPUShares != job.CPUShares ||
		job1.BlockIOWeight != job.BlockIOWeight ||
		job1.MemoryLimitMegabytes != job.MemoryLimitMegabytes {
		t.Fatal("job we got back had different values than the one we stored", job, job1)
	}

}
