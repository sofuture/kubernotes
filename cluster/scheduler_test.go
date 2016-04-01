package cluster

import (
	"testing"

	"github.com/sofuture/kubernotes/testtools"
)

func TestSchedulerAssignJobs(t *testing.T) {
	tb := testtools.TestBackend{}

	// create some nodes
	node := &Node{Namespace: "test", Name: "testnode", CPUShares: 1000}
	node1 := &Node{Namespace: "test", Name: "testnode1", CPUShares: 200}

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

	// create  jobs
	job1 := &Job{ID: "job1", CPUShares: 500}
	err = c.CreateJob(tb, job1)
	if err != nil {
		t.Fatal("unable to create job", err)
	}

	job2 := &Job{ID: "job2", CPUShares: 100}
	err = c.CreateJob(tb, job2)
	if err != nil {
		t.Fatal("unable to create job", err)
	}

	job3 := &Job{ID: "job3", CPUShares: 5000}
	err = c.CreateJob(tb, job3)
	if err != nil {
		t.Fatal("unable to create job", err)
	}

	// schedule job that only fits on one node
	status, err := c.Schedule(tb, job1)
	if err != nil {
		t.Fatal("got an error scheduling job1", err)
	}

	if !status.IsScheduled {
		t.Fatal("unable to schedule job1")
	}

	if status.Node != node.Name {
		t.Fatal("should have scheduled job on", node.Name, "but instead it's on", status.Node)
	}

	_ = node.Load(tb)
	jobs, _ := node.GetJobs(tb)
	if len(jobs) != 1 {
		t.Fatal("node assigned doesn't have scheduled job")
	}

	jobs, _ = node1.GetJobs(tb)
	if len(jobs) > 0 {
		t.Fatal("node should be empty, but has a scheduled job")
	}

	// schedule job that fits on either node
	status, err = c.Schedule(tb, job2)
	if err != nil {
		t.Fatal("got an error scheduling job2", err)
	}

	if !status.IsScheduled {
		t.Fatal("unable to schedule job2")
	}

	_ = node.Load(tb)
	_ = node1.Load(tb)

	found := false
	jobs, _ = node.GetJobs(tb)
	for _, job := range jobs {
		if job.ID == job2.ID {
			found = true
		}
	}

	jobs, _ = node1.GetJobs(tb)
	for _, job := range jobs {
		if job.ID == job2.ID {
			found = true
		}
	}

	if !found {
		t.Fatal("neither node had job2 which should have been scheduled")
	}

	// schedule job that fits nowhere
	status, err = c.Schedule(tb, job3)
	if err != nil {
		t.Fatal("got an error scheduling job3", err)
	}

	if status.IsScheduled {
		t.Fatal("should be unable to schedule job3")
	}

	if status.Node != "" {
		t.Fatal("should be unable to schedule job3")
	}

	// unschedule job2 from wherever it is
	err = c.Unschedule(tb, job2)
	if err != nil {
		t.Fatal("got an error unscheduling job2", err)
	}

	_ = node.Load(tb)
	_ = node1.Load(tb)

	jobs, _ = node.GetJobs(tb)
	if len(jobs) != 1 {
		t.Fatal("expected node to still have job1, and no other jobs")
	}

	jobs, _ = node1.GetJobs(tb)
	if len(jobs) != 0 {
		t.Fatal("expected node to have no assigned jobs")
	}
}
