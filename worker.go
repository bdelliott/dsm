package dsm

import (
	"log"
	"time"
)

const WORKER_SLEEP_DURATION = 1 * time.Second

var workerDone bool

func init() {
	workerDone = false
}

// Run a persistent worker process that polls for state machine instances that are ready to be worked
func RunWorker(client Client) {

	for {
		serializedMachines, err := client.GetByPrefix(MACHINE_PREFIX)
		if err != nil {
			log.Print("Failed to get prefix:", MACHINE_PREFIX)

		} else {

			machines := make([]*StateMachine, len(serializedMachines))

			for i, serializedMachine := range serializedMachines {
				machines[i] = Deserialize(serializedMachine)
			}

			log.Print("There are ", len(machines), " state machine instances")
			attemptWork(client, machines)
		}

		if workerDone {
			log.Print("Worker received quit signal")
			break
		} else {
			time.Sleep(WORKER_SLEEP_DURATION)
		}
	}
}

// Find an unlocked machine and attempt to advance its state
func attemptWork(client Client, machines []*StateMachine) {
	for _, machine := range machines {
		attemptWorkOnMachine(client, machine)
	}
}

// Try to work on the given machine
func attemptWorkOnMachine(client Client, machine *StateMachine) {

	lockSuccess, unlockFunction := client.Lock(machine.Key)
	defer unlockFunction()

	if lockSuccess {
		// invoke registered handler for this type of machine:
		machineType := machine.MType
		done := machineHandler[machineType](machine)

		if done {
			log.Print("State machine ", machine.Key, " is complete.")
			defer client.Delete(machine.Key)

		} else {
			// machine state and payload have been mutated, save the updates:
			machine.Put(client)
			log.Print("Saved update to machine ", machine.Key)
		}
	}

}

// Cause main worker loop to exit
func KillWorker() {
	workerDone = true
}
