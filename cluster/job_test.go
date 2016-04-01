package cluster

import (
	"testing"
)

var unitFile = `[Unit]
Description=foobar service

[Service]
ExecStart=/bin/bash -c "while true; do echo 'foo'; sleep 1; done"
MemoryLimit=1M
CPUShares=10
BlockIOWeight=10

[Install]
WantedBy=multi-user.target
`

var badUnitFile = `[Unit]
Description=foobar service

[Service]
ExecStart=/bin/bash -c "while true; do echo 'foo'; sleep 1; done"
MemoryLimit=foo
CPUShares=lol
BlockIOWeight=wat

[Install]
WantedBy=multi-user.target
`

func TestLoadJob(t *testing.T) {
	// parse a working unit file and make sure we load the right stuff
	job, err := LoadJob("foobar", unitFile)
	if err != nil {
		t.Fatal("error parsing unit file", err)
	}

	if job.ID != "foobar" {
		t.Fatal("got unexpected job name")
	}

	if job.MemoryLimitMegabytes != 1 {
		t.Fatal("memory limit not parsed correctly")
	}

}

func TestLoadBadJobWithDefaultLimits(t *testing.T) {
	// parse a broken unit file, make sure it errors
	job, err := LoadJob("foobar", badUnitFile)
	if err != nil {
		t.Fatal("error parsing unit file", err)
	}

	if job.ID != "foobar" {
		t.Fatal("got unexpected job name")
	}

	// make sure we get default values out of the broken parsing
	if job.CPUShares != DefaultCPUShares {
		t.Fatal("cpu shares did not get default values")
	}

	if job.BlockIOWeight != DefaultBlockIOWeight {
		t.Fatal("block io weight did not get default values")
	}

	if job.MemoryLimitMegabytes != DefaultMemoryLimitMegabytes {
		t.Fatal("memory limit did not get default values")
	}
}

func TestSerializeDeserializeJob(t *testing.T) {
	// serialize and deserialize a job
	j := &Job{
		ID:                   "foo",
		UnitFile:             "bar",
		CPUShares:            4,
		BlockIOWeight:        5,
		MemoryLimitMegabytes: 6,
	}

	json, err := j.Serialize()
	if err != nil {
		t.Fatal("problem serializing job", err)
	}

	j1 := &Job{}
	err = j1.Deserialize(json)
	if err != nil {
		t.Fatal("problem deserializing job", err)
	}

	if j.ID != j1.ID || j.UnitFile != j1.UnitFile || j.CPUShares != j1.CPUShares ||
		j.BlockIOWeight != j1.BlockIOWeight || j.MemoryLimitMegabytes != j1.MemoryLimitMegabytes {
		t.Fatal("deserialized job differs from original")
	}

}
