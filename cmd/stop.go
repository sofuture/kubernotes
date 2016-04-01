package cmd

import (
	"fmt"
	"log"

	"github.com/sofuture/kubernotes/cluster"
)

func Stop(etcdServers []string, namespace string, jobID string) error {
	etcd, err := cluster.NewEtcd(etcdServers)
	if err != nil {
		return err
	}

	c := cluster.NewNamespace(namespace)
	job, err := c.GetJob(etcd, jobID)
	if err != nil {
		return err
	}

	log.Println("unscheduling job:", jobID)
	err = c.Unschedule(etcd, job)
	if err != nil {
		return fmt.Errorf("unable to unschedule job %v", err)
	}

	return nil
}
