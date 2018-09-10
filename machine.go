package dsm

import (
	"encoding/json"
	"log"
	"time"
)

const(
	MACHINE_PREFIX = "machine-"
	MACHINE_TIMEOUT = 600 * time.Second
)

// registry of state machine types mapped to their corresponding tick function to drive the machine's logic:
var machineHandler map[string]func(*StateMachine) bool

func init() {
	machineHandler = make(map[string]func(*StateMachine) bool)
}

func RegisterMachineHandler(typeName string, tick func(*StateMachine) bool) {
	machineHandler[typeName] = tick
}

// generic state machine
type StateMachine struct {
	State   string
	Payload json.RawMessage
	Key     string
	MType   string // machine type
}

// Submit the state machine for execution.  Return the uuid of the submitted machine.
func (machine *StateMachine) Submit(client Client) string {

	machineId := GenerateUUID()
	machineKey := getMachineKey(machineId)

	machine.Key = machineKey
	machine.Put(client)

	return machineId
}

func (machine *StateMachine) Put(client Client) {

	machineValue := machine.serialize()

	err := client.Put(machine.Key, machineValue, MACHINE_TIMEOUT)
	if err != nil {
		log.Fatal("Machine put fail", err)
	}
}

// deserialize the json representation
func Deserialize(buf []byte) *StateMachine {

	var machine StateMachine
	err := json.Unmarshal(buf, &machine)
	if err != nil {
		log.Fatal("Deserialization failure: ", err)
	}
	log.Print("Deserialized machine: ", machine)
	return &machine
}

func getMachineKey(machineId string) string {
	return MACHINE_PREFIX + machineId
}

// serialize to json for the purpose of persisting state in etcd
func (machine *StateMachine) serialize() string {
	serializedMachine, err := json.Marshal(machine)

	if err != nil {
		log.Fatal("Error serializing machine: ", err)
	}

	return string(serializedMachine)
}
