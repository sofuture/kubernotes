package cluster

import (
	"fmt"
	"log"
)

// Represents the assigned state of a job to a node.
type JobStatus struct {
	ID          string
	Node        string
	IsScheduled bool
}

// Stop a job from running on the node that it's scheduled on. This method is
// really naive, and we assume that any given job is only running in a single
// place, if at all.
func (n *Namespace) Unschedule(backend Backend, job *Job) error {
	// loop through all nodes in the namespace
	log.Println("getting nodes in namespace that might be running this job")
	nodes, err := n.GetNodes(backend)
	if err != nil {
		return err
	}

	found := false

	for _, node := range nodes {
		log.Println("checking if", node.Name, "is running this job")

		// see if any of them are assigned the provided job
		for i, id := range node.JobIDs {

			// if they are assigned this job, remove it, and save the node
			if id == job.ID {
				node.JobIDs = append(node.JobIDs[:i], node.JobIDs[i+1:]...)
				err = node.UnassignJob(backend, job.ID)
				if err != nil {
					return fmt.Errorf("unable to unschedule job %v", err)
				}

				// we're being a little naive about jobs only running in a single place
				// but that's how it is, so bounce out of here
				found = true
				break
			}
		}

		// again, we're only assuming the job exists in a single place, so bounce out
		if found {
			break
		}
	}

	// successfully unscheduled
	if found {
		return nil
	}

	return fmt.Errorf("unable to unschedule job")
}

// Schedule a job on the cluster. Find a node with available resources, and assign
// it the job, saving the node in the process.
func (n *Namespace) Schedule(backend Backend, job *Job) (*JobStatus, error) {
	status := &JobStatus{
		ID:          job.ID,
		IsScheduled: false,
	}

	// loop through all the nodes in the namespace
	log.Println("getting nodes in namespace available for scheduling")
	nodes, err := n.GetNodes(backend)
	if err != nil {
		return nil, err
	}

	// this is a really not great way to do this, but it's the simplest way
	// to ensure we only run each job once. a better approach would be to store
	// the assignment state of the job.
	alreadyRunning := false
	for _, node := range nodes {
		for _, jobID := range node.JobIDs {
			if jobID == job.ID {
				alreadyRunning = true
				break
			}
		}
	}

	// if the job is already scheduled, bail with an error
	if alreadyRunning {
		status.IsScheduled = true
		return status, fmt.Errorf("%s already scheduled", job.ID)
	}

	for _, node := range nodes {
		// grab the free resources (available minus used by jobs)
		log.Println("determining free resources for", node.Name)
		resources, err := node.GetFreeResources(backend)
		if err != nil {
			return nil, err
		}

		// A more interesting/advanced scheduler could do something like look
		// for resource availability across all nodes, and see if moving jobs
		// elsewhere would enable an otherwise overloaded node to handle a job.
		//
		// For now, we'll only consider intra-node resouce availability.
		if resources.CPUShares >= job.CPUShares &&
			resources.BlockIOShares >= job.BlockIOWeight &&
			resources.MemoryMegabytes >= job.MemoryLimitMegabytes {

			log.Println("node", node.Name, "able to run job", job.ID)
			err = node.AssignJob(backend, job.ID)
			if err != nil {
				// if we have trouble scheduling on a node we can look for others
				continue
			}
			status.Node = node.Name
			status.IsScheduled = true
			break
		} else {
			log.Println("node", node.Name, "NOT able to run job", job.ID)
		}
	}

	return status, nil
}
