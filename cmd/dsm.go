package main

import (
	"encoding/json"
	"flag"
	"github.com/bdelliott/dsm"
	"log"
)


const (
	INITIAL = "INIT"
	SEASONED = "SEASONED"
	GRILLHOT = "GRILLHOT"
	SIDEONE = "SIDEONE"
	COOKED = "COOKED"
	DONE = "DONE"
)

// This particular state machine tracks trivial state pertinent to the backyard grillmaster:
type SteakGrillState struct {
	Seasoned bool
	GrillHot bool
	CookedSide1 bool
	CookedSide2 bool
	Rested bool
}

// logic to advance the application's state machine.  return true if the state machine is at its terminal (completed)
// state
func tick(machine *dsm.StateMachine) bool {

	log.Print("tick: ", machine.Key)

	steakState := SteakGrillState{}

	if len(machine.Payload) > 0 {
		err := json.Unmarshal(machine.Payload, &steakState)
		if err != nil {
			log.Fatal("Failed to demarshal")
		}
	}

	switch machine.State {
	case INITIAL:
		log.Print("Season steak!")
		steakState.Seasoned = true
		machine.State = SEASONED

	case SEASONED:
		log.Print("Now heat that grill")
		steakState.GrillHot = true // (magic instant heating grill)
		machine.State = GRILLHOT

	case GRILLHOT:
		log.Print("Cook first side") // (magic instant cook!)
		steakState.CookedSide1 = true
		machine.State = SIDEONE

	case SIDEONE:
		log.Print("Flip and cook second side") // (magic instant cook!)
		steakState.CookedSide2 = true
		machine.State = COOKED

	case COOKED:
		log.Print("Rest steak") // (magic instant rest!)
		steakState.Rested = true
		machine.State = DONE
	}

	log.Printf("Updated payload: %v", steakState)

	payload, err := json.Marshal(steakState)
	if err != nil {
		log.Fatal("Failed to json encode steak state: ", err)
	}
	machine.Payload = payload

	return machine.State == DONE
}

// Example application utilizing distributed state machinery
func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	worker := flag.Bool("worker", false, "Run a state machine worker process")
	etcdUrl := flag.String("etcdurl", "http://etcd:2379", "URL to etcd cluster")

	flag.Parse()

	client := dsm.NewClient(etcdUrl)
	defer client.Close()

	// register this particular type of state machine (to allow for multiple types):
	dsm.RegisterMachineHandler("steak", tick)

	if *worker {
		log.Print("Starting state machine worker process")
		dsm.RunWorker(client)

	} else {
		// Client, will submit a state machine instance:
		submitStateMachine(client)
	}
}

// submit an instance of a state machine to the distributed processing machinery:
func submitStateMachine(client dsm.Client) {

	// initialize the state machine targeting a nice grilled steak dinner:
	grillMaster := SteakGrillState{}

	payload, err := json.Marshal(grillMaster)
	if err != nil {
		log.Fatal("json encode failure", err)
	}

	machine := &dsm.StateMachine{State: INITIAL, Payload: payload, MType: "steak"}
	machineId := machine.Submit(client)
	log.Print("Submitted state machine instance has id:", machineId)
}
