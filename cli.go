package main

import (
	"os"

	"github.com/voxelbrain/goptions"

	"github.com/sofuture/kubernotes/cmd"
)

// run agent
type AgentOptions struct {
	Bind            string `goptions:"-b, --bind, description='bind for agent to listen on'"`
	NodeName        string `goptions:"-n, --name, obligatory, description='node name'"`
	CPUShares       int    `goptions:"-c, --cpu, description='cpu shares available to scheduler'"`
	BlockIOShares   int    `goptions:"-i, --io, description='block io shares available to scheduler'"`
	MemoryMegabytes int    `goptions:"-m, --memory, description='memory megabytes available to scheduler'"`
}

// create jobs
type CreateOptions struct {
	Name     string   `goptions:"-n, --name, obligatory, description='unique name of job'"`
	UnitFile *os.File `goptions:"-f, --file, obligatory, description='service unit file'"`
}

// destroy jobs
type DestroyOptions struct {
	Name string `goptions:"-n, --name, obligatory, description='job to destroy'"`
}

// list jobs
type ListOptions struct {
	Name string `goptions:"-n, --name, description='job to view status for'"`
}

// start jobs
type StartOptions struct {
	Name string `goptions:"-n, --name, obligatory, description='job to start'"`
}

// display cluster status
type StatusOptions struct{}

// stop jobs
type StopOptions struct {
	Name string `goptions:"-n, --name, obligatory, description='job to stop'"`
}

// get output from jobs
type TailOptions struct {
	Name  string `goptions:"-n, --name, obligatory, description='job to watch'"`
	Count int    `goptions:"-c, --count, description='number of lines to display'"`
}

// full cli options struct
type Cli struct {
	EtcdServers []string      `goptions:"-e, --etcd, description='etcd servers to connect to'"`
	Namespace   string        `goptions:"-c, --namespace, description='cluster namespace'"`
	Help        goptions.Help `goptions:"-h, --help, description='Show this help'"`

	Verb    goptions.Verbs
	Agent   AgentOptions   `goptions:"agent"`
	Create  CreateOptions  `goptions:"create"`
	Destroy DestroyOptions `goptions:"destroy"`
	List    ListOptions    `goptions:"list"`
	Start   StartOptions   `goptions:"start"`
	Status  StatusOptions  `goptions:"status"`
	Stop    StopOptions    `goptions:"stop"`
	Tail    TailOptions    `goptions:"tail"`
}

func runCli() (err error) {
	options := &Cli{
		EtcdServers: []string{"http://localhost:2379"},
		Namespace:   "kubernotes",
		Agent: AgentOptions{
			Bind:            "127.0.0.1:10004",
			CPUShares:       4000,
			BlockIOShares:   4000,
			MemoryMegabytes: 4000,
		},
		Tail: TailOptions{
			Count: 20,
		},
	}

	goptions.ParseAndFail(options)

	switch options.Verb {
	case "agent":
		err = cmd.Agent(options.EtcdServers, options.Namespace, options.Agent.Bind,
			options.Agent.NodeName, options.Agent.CPUShares, options.Agent.BlockIOShares,
			options.Agent.MemoryMegabytes)
	case "status":
		err = cmd.Status(options.EtcdServers, options.Namespace)
	case "list":
		err = cmd.List(options.EtcdServers, options.Namespace, options.List.Name)
	case "create":
		err = cmd.Create(options.EtcdServers, options.Namespace, options.Create.Name, options.Create.UnitFile)
	case "destroy":
		err = cmd.Destroy(options.EtcdServers, options.Namespace, options.Destroy.Name)
	case "start":
		err = cmd.Start(options.EtcdServers, options.Namespace, options.Start.Name)
	case "stop":
		err = cmd.Stop(options.EtcdServers, options.Namespace, options.Stop.Name)
	case "tail":
		err = cmd.Tail(options.EtcdServers, options.Namespace, options.Tail.Name, options.Tail.Count)
	default:
		goptions.PrintHelp()
	}

	return err
}
