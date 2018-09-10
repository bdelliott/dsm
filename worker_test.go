package dsm

import "testing"

const NEW_STATE = "state2"

// basic happy path test of the worker
func TestRunWorker(t *testing.T) {
	client := NewTestClient()

	// seed some data
	machine := &StateMachine{State: TEST_STATE, MType: TEST_MACHINE_TYPE}

	// register a "tick" handler
	RegisterMachineHandler(TEST_MACHINE_TYPE, func(machine *StateMachine) bool {
		// just advance to a new state
		machine.State = NEW_STATE
		return false
	})
	machineId := machine.Submit(client)

	KillWorker() // (worker loop will only execute once)
	RunWorker(client)

	machineKey := getMachineKey(machineId)
	serializedMachine, err := client.Get(machineKey)

	if err != nil {
		t.Error("Get error ", err)
	}

	machine2 := Deserialize(serializedMachine)
	if machine2.State != NEW_STATE {
		t.Error("Machine state mismatch, expected ", NEW_STATE, ", but got ", machine2.State)
	}
}
