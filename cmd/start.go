package cmd

import (
	"fmt"
	"log"

	"github.com/sofuture/kubernotes/cluster"
)

func Start(etcdServers []string, namespace string, jobID string) error {
	etcd, err := cluster.NewEtcd(etcdServers)
	if err != nil {
		return err
	}

	c := cluster.NewNamespace(namespace)
	job, err := c.GetJob(etcd, jobID)
	if err != nil {
		return err
	}

	log.Println("scheduling job:", jobID)
	status, err := c.Schedule(etcd, job)
	if err != nil {
		return fmt.Errorf("unable to schedule job %v", err)
	}

	if status.IsScheduled {
		log.Println("scheduled on node:", status.Node)
	} else {
		log.Println("unable to find resources to run job", status.ID)
	}

	return nil
}
