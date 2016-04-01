package agent

import (
	"testing"

	"github.com/sofuture/kubernotes/cluster"
	"github.com/sofuture/kubernotes/testtools"
)

var unitFile = `[Unit]
Description=foobar service

[Service]
ExecStart=/bin/bash -c "while true; do echo 'foo'; sleep 1; done"
MemoryLimit=1M
CPUShares=10
BlockIOWeight=10

[Install]
WantedBy=multi-user.target
`

func getTestingAgent() (*Agent, TestLocal) {
	local := TestLocal{}
	agent := &Agent{
		Bind:            "127.0.0.1:23142",
		NodeName:        "testnode",
		CPUShares:       1000,
		BlockIOShares:   1000,
		MemoryMegabytes: 1000,
		Namespace:       cluster.NewNamespace("testnamespace"),
		Local:           local,
		ClusterBackend:  testtools.TestBackend{},
	}
	return agent, local
}

func TestEmptyAgentJoinSync(t *testing.T) {
	agent, _ := getTestingAgent()
	err := agent.joinCluster()
	if err != nil {
		t.Fatal(err)
	}

	err = agent.syncState()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAgentTracksAssignedJobs(t *testing.T) {
	agent, local := getTestingAgent()
	err := agent.joinCluster()
	if err != nil {
		t.Fatal(err)
	}

	job, err := cluster.LoadJob("testjob", unitFile)
	if err != nil {
		t.Fatal("unable to load job", err)
	}

	err = agent.Namespace.CreateJob(agent.ClusterBackend, job)
	if err != nil {
		t.Fatal("unable to create job", err)
	}

	if agent.Node == nil {
		t.Fatal("agent.Node not initialized")
	}

	err = agent.Node.AssignJob(agent.ClusterBackend, job.ID)
	if err != nil {
		t.Fatal("unable to assign job", err)
	}

	err = agent.syncState()
	if err != nil {
		t.Fatal(err)
	}

	_, ok := local["testjob"]
	if !ok {
		t.Fatal("local job not created")
	}

	err = agent.Node.UnassignJob(agent.ClusterBackend, job.ID)
	if err != nil {
		t.Fatal("unable to unassign job", err)
	}

	err = agent.syncState()
	if err != nil {
		t.Fatal(err)
	}

	_, ok = local["testjob"]
	if ok {
		t.Fatal("local job not destroyed")
	}

	err = agent.watchClusterOnce()
	if err != nil {
		t.Fatal(err)
	}
}

// simple mock local backend for testing
type TestLocal map[string]cluster.Job

func (t TestLocal) Connect() error {
	return nil
}

func (t TestLocal) Disconnect() {}

func (t TestLocal) GetManagedJobs() ([]cluster.Job, error) {
	ret := make([]cluster.Job, 0)
	for _, j := range t {
		ret = append(ret, j)
	}
	return ret, nil
}

func (t TestLocal) CreateJob(job *cluster.Job) error {
	t[job.ID] = *job
	return nil
}

func (t TestLocal) StartJob(job *cluster.Job) error {
	j := t[job.ID]
	j.IsRunning = true
	return nil
}

func (t TestLocal) StopJob(job *cluster.Job) error {
	j := t[job.ID]
	j.IsRunning = false
	return nil
}

func (t TestLocal) DestroyJob(job *cluster.Job) error {
	delete(t, job.ID)
	return nil
}

func (t TestLocal) GetLogs(job *cluster.Job, count int) (string, error) {
	return "", nil
}
