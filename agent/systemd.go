package agent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	systemd "github.com/coreos/go-systemd/dbus"

	"github.com/sofuture/kubernotes/cluster"
)

type Systemd struct {
	Namespace string
	NodeName  string

	conn *systemd.Conn
}

// Create a new Systemd job management backend.
func NewSystemd(namespace string, nodeName string) *Systemd {
	return &Systemd{
		Namespace: namespace,
		NodeName:  nodeName,
	}
}

// Connect to Systemd over dbus.
func (s *Systemd) Connect() (err error) {
	s.conn, err = systemd.NewSystemdConnection()
	return err
}

// Disconnect from Systemd.
func (s *Systemd) Disconnect() {
	s.conn.Close()
}

// Create a unit file on disk for a service that represents the given job. Reload Systemd.
func (s *Systemd) CreateJob(job *cluster.Job) error {
	path := s.getServicePath(job)

	// write unit file to disk
	err := ioutil.WriteFile(path, []byte(job.UnitFile), 0644)
	if err != nil {
		return fmt.Errorf("could not write job unit file %v", err)
	}

	return s.conn.Reload()
}

// Start the Systemd service for the specified job.
func (s *Systemd) StartJob(job *cluster.Job) error {
	// start the service
	_, err := s.conn.StartUnit(s.getServiceName(job), "replace", nil)
	return err
}

// Stop the Systemd service for the specified job.
func (s *Systemd) StopJob(job *cluster.Job) error {
	// stop the service
	_, err := s.conn.StopUnit(s.getServiceName(job), "replace", nil)
	return err
}

// Destroy a Systemd services unit file, and reload Systemd.
func (s *Systemd) DestroyJob(job *cluster.Job) error {
	path := s.getServicePath(job)

	// write unit file to disk
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("could not delete job unit file %v", err)
	}

	// reload systemd
	return s.conn.Reload()
}

// Get a list of all local Jobs that we are responsible for (belong to our kubernotes node).
func (s *Systemd) GetManagedJobs() ([]cluster.Job, error) {
	startsWith := fmt.Sprintf("kubernotes-%s-%s-", s.Namespace, s.NodeName)
	localJobs := make([]cluster.Job, 0)

	// get all units that Systemd knows about
	localUnits, err := s.conn.ListUnits()
	if err != nil {
		return nil, err
	}

	// filter the ones that this node manages
	for _, unit := range localUnits {
		if strings.HasPrefix(unit.Name, startsWith) {
			name := strings.TrimPrefix(unit.Name, startsWith)
			name = strings.TrimSuffix(name, ".service")
			localJobs = append(localJobs, cluster.Job{
				ID:        name,
				IsRunning: unit.SubState == "running",
			})
		}
	}

	return localJobs, nil
}

// Get a chunk of logs since the specified position. If the position is not specified, the last 20 lines will be returned.
func (s *Systemd) GetLogs(job *cluster.Job, count int) (string, error) {
	serviceName := s.getServiceName(job)

	cmd := exec.Command("journalctl", "-u", serviceName, "-n", strconv.Itoa(count))
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to get logs for job %s: %v", job.ID, err)
	}

	return out.String(), nil
}

func (s *Systemd) getServiceName(job *cluster.Job) string {
	return fmt.Sprintf("kubernotes-%s-%s-%s.service",
		s.Namespace,
		s.NodeName,
		job.ID)
}

func (s *Systemd) getServicePath(job *cluster.Job) string {
	return fmt.Sprintf("/lib/systemd/system/%s", s.getServiceName(job))
}
