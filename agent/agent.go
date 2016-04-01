package agent

import (
	"fmt"
	"log"

	"github.com/sofuture/kubernotes/cluster"
)

type Agent struct {
	Bind            string
	NodeName        string
	CPUShares       int
	BlockIOShares   int
	MemoryMegabytes int
	Namespace       *cluster.Namespace
	Node            *cluster.Node

	ClusterBackend cluster.Backend
	Local          Local
}

func (a *Agent) Run() (err error) {
	// connect to Systemd
	err = a.Local.Connect()
	if err != nil {
		return fmt.Errorf("cannot connect to systemd, maybe you need to be root? %v", err)
	}

	// join cluster
	err = a.joinCluster()
	if err != nil {
		return err
	}

	// sync state with cluster
	err = a.syncState()
	if err != nil {
		return err
	}

	errs := make(chan error, 1)

	// listen for changes
	go func() {
		err = a.watchCluster()
		if err != nil {
			errs <- err
		}
	}()

	// spawn api
	go func() {
		err = a.SpawnAPI()
		if err != nil {
			errs <- err
		}
	}()

	return <-errs
}

func (a *Agent) joinCluster() error {
	var err error

	a.Node = &cluster.Node{
		Name:            a.NodeName,
		Namespace:       a.Namespace.GetName(),
		BlockIOShares:   a.BlockIOShares,
		CPUShares:       a.CPUShares,
		MemoryMegabytes: a.MemoryMegabytes,

		// We make an assumption here that the bind we are listening on
		// is routable. This means ":1234" or "0.0.0.0:1234" won't work.
		// We could be smarter about this, but it's simplest to make
		// them identical for the time being.
		Endpoint: a.Bind,
	}

	err = a.Namespace.CreateNode(a.ClusterBackend, a.Node)
	if err != nil {
		return fmt.Errorf("Could not join cluster: %v", err)
	}

	return nil
}

func (a *Agent) syncState() error {
	log.Println("updating local state to match cluster")

	err := a.Node.Load(a.ClusterBackend)
	if err != nil {
		return err
	}

	// get list of jobs we should be running
	log.Println("getting jobs we should be running according to cluster")
	clusterJobs, err := a.Node.GetJobs(a.ClusterBackend)
	if err != nil {
		return err
	}

	// get the local jobs we have (running or not)
	log.Println("getting local jobs we know about")
	localJobs, err := a.Local.GetManagedJobs()
	if err != nil {
		return err
	}

	// helper func to create a systemd job
	createJob := func(job *cluster.Job) error {
		log.Println("job", job.ID, "needs to be created locally")
		err = a.Local.CreateJob(job)
		if err != nil {
			log.Println("unable to create local job", job.ID)
			return err
		}
		return nil
	}

	// helper func to start a systemd job
	startJob := func(job *cluster.Job) {
		log.Println("starting local job", job.ID)
		err = a.Local.StartJob(job)
		if err != nil {
			log.Println("unable to start local job", job.ID)
		}
	}

	// make sure each job exists locally and is in correct state
	seenJobs := make(map[string]bool)

	// loop through jobs that we've been scheduled
	for _, clusterJob := range clusterJobs {
		log.Println("checking for cluster job", clusterJob.ID)
		seenJobs[clusterJob.ID] = true
		found := false

		// loop through local jobs to see if we know about this job
		// we're supposed to be running
		for _, localJob := range localJobs {

			// if we know about the job, make sure it's running
			if localJob.ID == clusterJob.ID {
				found = true
				if !clusterJob.IsRunning {
					startJob(&clusterJob)
				}
			}
		}

		// if we didn't find the job locally, we need to create and start it
		if !found {
			err = createJob(&clusterJob)
			if err != nil {
				// we're going to ignore failure to start, because we don't do anything
				// (like unschedule) with it currently
				startJob(&clusterJob)
			}
		}
	}

	// try to destroy orphaned local jobs
	for _, localJob := range localJobs {
		if _, shouldHave := seenJobs[localJob.ID]; !shouldHave {
			log.Println("destroying local job", localJob.ID)
			err = a.Local.StopJob(&localJob)
			if err != nil {
				log.Println("unable to stop local job", localJob.ID)
				err = nil
			}

			err = a.Local.DestroyJob(&localJob)
			if err != nil {
				log.Println("unable to destroy local job", localJob.ID)
				err = nil
			}
		}
	}

	return err
}

func (a *Agent) watchCluster() error {
	// listen for changes, then run syncState
	log.Println("listening for schedule changes")

	for {
		err := a.watchClusterOnce()
		if err != nil {
			return err
		}
	}
}

func (a *Agent) watchClusterOnce() (err error) {
	// listen for changes since we've last synced
	lastSeen := a.Node.LastModifiedIndex

	lastSeen, err = a.Node.WatchForChanges(a.ClusterBackend, lastSeen)
	if err != nil {
		return err
	}
	err = a.syncState()
	if err != nil {
		return err
	}

	return nil
}
