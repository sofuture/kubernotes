package cluster

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/unit"
)

const (
	DefaultCPUShares            = 1000
	DefaultBlockIOWeight        = 1000
	DefaultMemoryLimitMegabytes = 100
)

// Represents a job that can be scheduled by Kubernotes.
type Job struct {
	ID                   string
	UnitFile             string
	CPUShares            int
	BlockIOWeight        int
	MemoryLimitMegabytes int
	IsRunning            bool
}

// Deserialize a Job from JSON string.
func (j *Job) Deserialize(jsonBlob string) error {
	err := json.Unmarshal([]byte(jsonBlob), j)
	return err
}

// Serialize a Job to JSON string.
func (j *Job) Serialize() (string, error) {
	jsonBlob, err := json.Marshal(j)
	return string(jsonBlob), err
}

// Parse a Systemd unit file into a Job.
func LoadJob(name string, unitFile string) (*Job, error) {
	job := &Job{
		ID:       name,
		UnitFile: unitFile,
	}

	reader := strings.NewReader(unitFile)

	opts, err := unit.Deserialize(reader)
	if err != nil {
		return nil, err
	}

	// we only explicitly care about these 3 resource limits currently
	for _, opt := range opts {
		if opt.Section == "Service" {
			switch opt.Name {
			case "MemoryLimit":
				// for simplicity's sake we will only support memory limits provided
				// in megabytes, with the suffix 'M'
				if strings.HasSuffix(opt.Value, "M") {
					limit := strings.TrimSuffix(opt.Value, "M")
					job.MemoryLimitMegabytes, _ = strconv.Atoi(limit)
				}
			case "CPUShares":
				job.CPUShares, _ = strconv.Atoi(opt.Value)
			case "BlockIOWeight":
				job.BlockIOWeight, _ = strconv.Atoi(opt.Value)
			default:
				continue
			}
		}
	}

	// also for simplicity's sake, we'll give all jobs a default value for limits
	// that are left unspecified
	if job.MemoryLimitMegabytes == 0 {
		job.MemoryLimitMegabytes = DefaultMemoryLimitMegabytes
	}
	if job.CPUShares == 0 {
		job.CPUShares = DefaultCPUShares
	}
	if job.BlockIOWeight == 0 {
		job.BlockIOWeight = DefaultBlockIOWeight
	}

	return job, nil
}
