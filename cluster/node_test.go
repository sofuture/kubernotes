package cluster

import (
	"testing"

	"github.com/sofuture/kubernotes/testtools"
)

func TestNodeJoinCluster(t *testing.T) {
	// create a node and join a cluster
	node := &Node{Namespace: "test", Name: "testnode"}
	tb := testtools.TestBackend{}

	err := node.JoinCluster(tb)
	if err != nil {
		t.Fatal("failed to join cluster", err)
	}

	if _, ok := tb["/kubernotes/clusters/test/nodes/testnode"]; !ok {
		t.Fatal("clusternode not created")
	}

	// make sure we can leave clusters
	err = node.LeaveCluster(tb)
	if err != nil {
		t.Fatal("failed to leave cluster", err)
	}

	if _, ok := tb["/kubernotes/clusters/test/nodes/testnode"]; ok {
		t.Fatal("clusternode not destroyed")
	}

}

func TestNodeStoreResources(t *testing.T) {
	// create a node with allocated resources
	node := &Node{
		Namespace:       "test",
		Name:            "testnode",
		BlockIOShares:   10,
		CPUShares:       10,
		MemoryMegabytes: 10,
	}
	tb := testtools.TestBackend{}

	err := node.JoinCluster(tb)
	if err != nil {
		t.Fatal("failed to join cluster", err)
	}

	if _, ok := tb["/kubernotes/clusters/test/nodes/testnode"]; !ok {
		t.Fatal("clusternode not created")
	}

	c := NewNamespace("test")
	node1, err := c.GetNode(tb, "testnode")
	if err != nil {
		t.Fatal("failed to get node", err)
	}

	// make sure we store the correct resource information
	if node1.CPUShares != node.CPUShares ||
		node1.BlockIOShares != node.BlockIOShares ||
		node1.MemoryMegabytes != node.MemoryMegabytes {
		t.Fatal("node we got back had different values than the one we stored", node, node1)
	}

}

func TestNodeAssignJobs(t *testing.T) {
	tb := testtools.TestBackend{}

	// create some nodes
	node := &Node{Namespace: "test", Name: "testnode", CPUShares: 1000}
	node1 := &Node{Namespace: "test", Name: "testnode1"}

	err := node.JoinCluster(tb)
	if err != nil {
		t.Fatal("failed to join cluster", err)
	}
	err = node1.JoinCluster(tb)
	if err != nil {
		t.Fatal("failed to join cluster", err)
	}

	// and a namespace
	c := NewNamespace("test")
	c.CreateJob(tb, &Job{ID: "testjob", CPUShares: 200})

	// check node resources
	resources, err := node.GetFreeResources(tb)
	if err != nil {
		t.Fatal("unable to get resources for node", err)
	}

	if resources.CPUShares != 1000 {
		t.Fatal("got unexpected node resources", resources)
	}

	// initial assignment
	err = node.AssignJob(tb, "testjob")
	if err != nil {
		t.Fatal("failed to assign job to node", err)
	}

	// check resource use
	resources, err = node.GetFreeResources(tb)
	if err != nil {
		t.Fatal("unable to get resources for node", err)
	}

	if resources.CPUShares != 800 {
		t.Fatal("node didn't account for resources used by job", resources)
	}

	// get all jobs for node
	jobs, err := node.GetJobs(tb)
	if err != nil {
		t.Fatal("error getting node jobs")
	}

	if len(jobs) != 1 {
		t.Fatal("expected to get 1 job for node, got", len(jobs), "instead")
	}

	if jobs[0].ID != "testjob" {
		t.Fatal("got an unexpected job from node", jobs[0])
	}

	// another assignment of the same job should fail
	err = node.AssignJob(tb, "testjob")
	if err == nil {
		t.Fatal("allowed to assign duplicate job to node")
	}

	// unassignment should work
	err = node.UnassignJob(tb, "testjob")
	if err != nil {
		t.Fatal("failed to unassign job from node")
	}
}
