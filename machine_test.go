package dsm

import (
	"strings"
	"testing"
)

const (
	TEST_MACHINE_TYPE = "test-machine"
	TEST_STATE        = "test-state"
)

func TestRegisterMachineHandler(t *testing.T) {

	called := false

	// register a machine with a dummy "tick handler"
	RegisterMachineHandler(TEST_MACHINE_TYPE, func(machine *StateMachine) bool {
		called = true
		return true
	})

	machine := &StateMachine{}
	if !machineHandler[TEST_MACHINE_TYPE](machine) {
		t.Error("Expected a true response")
	}

	if !called {
		t.Error("Handler didn't register properly.")
	}

}

func TestSubmit(t *testing.T) {

	machine := &StateMachine{State: TEST_STATE}
	client := NewTestClient()
	machineId := machine.Submit(client)

	machineKey := getMachineKey(machineId)

	if !strings.HasPrefix(machineKey, MACHINE_PREFIX) {
		t.Error("Malformed machine key")
	}

	serializedMachine, err := client.Get(machineKey)

	if err != nil {
		t.Error("Get failure: ", err)
	}

	machine2 := Deserialize(serializedMachine)
	if machine.State != machine2.State {
		t.Error("Machines don't match")
	}

}
