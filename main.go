package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	dirclone "amalitech.org/subsys/cmd/dir_clone"
	dirconfig "amalitech.org/subsys/cmd/dir_config"
	dirinit "amalitech.org/subsys/cmd/dir_init"
	dirsnap "amalitech.org/subsys/cmd/dir_snap"
	dirsubmission "amalitech.org/subsys/cmd/dir_submission"
)

type Command int

const (
	Init Command = iota
	Config
	Snap
	Submit
	Clone
)

func (c Command) String() string {
	switch c {
	case Init:
		return "init"
	case Config:
		return "config"
	case Snap:
		return "snap"
	case Submit:
		return "submit"
	case Clone:
		return "clone"
	default:
		return "unknown"
	}
}

func Greet() string {
	return "Welcome to subsys v0.0.1, an assignment submission platform\nCommands:\nsubsys init - This command is for initialising a new subsys directory\n\nsubsys config - This command is for configuring your directory\nFlags: --code 'Your assignmnent code' --student_id 'Your student ID'\n\nsubsys snap - This command is for making a snapshot of your work, it's what is going to be submitted\nFlags: --name 'Name of the snapshot to create'\n\nsubsys submit - This command allows you to specify a snapshot to submit or submit all snapshots if you don't specify a snapshot\nFlags: --name 'Name of the snapshot to submit'\n\nsubsys clone - This command allows a lecture to download student's snapshots and run them locally"
}

func allowedCommands() string {
	var commands []string
	for _, command := range []Command{Init, Config, Snap, Submit, Clone} {
		commands = append(commands, command.String())
	}

	return strings.Join(commands, ", ")
}

func CommandFromString(commandStr string) (Command, error) {
	for _, command := range []Command{Init, Config, Snap, Submit, Clone} {
		if command.String() == commandStr {
			return command, nil
		}
	}

	return Command(0), fmt.Errorf("unknown command: %s", commandStr)
}

func main() {
	var command Command
	if len(os.Args) > 1 {
		cmd, err := CommandFromString(os.Args[1])
		if err == nil {
			command = cmd
		} else {
			log.Fatalf("Subsys doesn't have the command %s. Allowed commands are %v", os.Args[1], allowedCommands())
		}
	} else {
		fmt.Println(Greet())
		return
	}

	switch command {
	case Init:
		initializer, err := dirinit.NewDirectoryInitializer()
		if err != nil {
			log.Fatalf("Error creating SubmissionInitializer: %v", err)
		}

		err = initializer.Initialize()
		if err != nil {
			log.Fatalf("Error initializing submission: %v", err)
		}

	case Config:
		configurator := dirconfig.NewConfigurator()

		err := configurator.ConfigureDirectory()
		if err != nil {
			log.Fatalf("Error configuring directory: %v\n", err)
		}

	case Snap:
		snapshotManager := dirsnap.NewSnapshotManager()

		err := snapshotManager.CreateSnapshot()
		if err != nil {
			log.Fatalf("Error making a new snaphot: %v\n", err)
		}

	case Submit:
		submissionManager := dirsubmission.NewSubmissionInitializer()

		err := submissionManager.SubmitSnapshots()
		if err != nil {
			log.Fatalf("Error submitting snapshot: %v\n", err)
		}
	case Clone:
		cloneManager := dirclone.NewCloneManager()

		err := cloneManager.CloneSnapshot()
		if err != nil {
			log.Fatalf("Error cloning submission: %v\n", err)
		}
	}

}
