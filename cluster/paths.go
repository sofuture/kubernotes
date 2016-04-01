package cluster

import (
	"fmt"
)

func getNamespacePath(clusterName string) string {
	return fmt.Sprintf("/kubernotes/clusters/%s", clusterName)
}

func getJobPath(clusterName string, jobID string) string {
	return fmt.Sprintf("%s/jobs/%s", getNamespacePath(clusterName), jobID)
}

func getNodesPath(clusterName string) string {
	return fmt.Sprintf("%s/nodes", getNamespacePath(clusterName))
}

func getNodePath(clusterName string, nodeName string) string {
	return fmt.Sprintf("%s/%s", getNodesPath(clusterName), nodeName)
}

func getNodeChangesPath(clusterName string, nodeName string) string {
	return getNodePath(clusterName, nodeName)
}
