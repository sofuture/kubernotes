package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sofuture/kubernotes/cluster"
)

func Create(etcdServers []string, namespace string, jobName string, unitFile *os.File) error {
	log.Println("storing job", jobName, "in cluster")
	etcd, err := cluster.NewEtcd(etcdServers)
	if err != nil {
		return err
	}

	unitBytes, err := ioutil.ReadAll(unitFile)
	if err != nil {
		return err
	}

	job, err := cluster.LoadJob(jobName, string(unitBytes))
	if err != nil {
		return fmt.Errorf("error parsing provided unit file %v", err)
	}

	c := cluster.NewNamespace(namespace)
	err = c.CreateJob(etcd, job)
	if err != nil {
		return err
	}

	log.Println("stored job", jobName)
	return nil
}
