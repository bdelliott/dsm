package dsm

import (
	"log"
	"testing"
)

func TestGenerateUUID(t *testing.T) {

	uuid := GenerateUUID()
	log.Print(uuid)

	if uuid == "" {
		t.Error("Empty uuid")
	}
	if len(uuid) < 30 {
		t.Error("UUID may be malformatted: ", uuid)
	}
}
