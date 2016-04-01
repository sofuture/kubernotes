package cmd

import (
	"github.com/sofuture/kubernotes/agent"
	"github.com/sofuture/kubernotes/cluster"
)

func Agent(etcdServers []string, namespace string, bind string, name string,
	cpuShares int, blockIOShares int, memoryMegabytes int) error {

	etcd, err := cluster.NewEtcd(etcdServers)
	if err != nil {
		return err
	}

	agent := agent.Agent{
		Bind:            bind,
		Namespace:       cluster.NewNamespace(namespace),
		CPUShares:       cpuShares,
		BlockIOShares:   blockIOShares,
		MemoryMegabytes: memoryMegabytes,
		NodeName:        name,
		ClusterBackend:  etcd,
		Local:           agent.NewSystemd(namespace, name),
	}
	return agent.Run()
}
