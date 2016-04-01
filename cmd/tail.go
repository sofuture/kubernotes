package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sofuture/kubernotes/cluster"
)

func Tail(etcdServers []string, namespace string, jobID string, count int) error {

	// connect to etcd
	etcd, err := cluster.NewEtcd(etcdServers)
	if err != nil {
		return err
	}

	// find which node is running the job, if any
	c := cluster.NewNamespace(namespace)
	node, err := c.GetNodeRunningJob(etcd, jobID)
	if err != nil {
		return err
	}

	if node != nil {
		url := fmt.Sprintf("http://%s/logs?job=%s&count=%d", node.Endpoint, jobID, count)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("problem retrieving logs %v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("problem retrieving logs %v", err)
		}
		fmt.Print(string(body))
	} else {
		return fmt.Errorf("could not find job running")
	}

	return nil
}
