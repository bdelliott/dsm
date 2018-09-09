package dsm

import (
	"github.com/nu7hatch/gouuid"
	"log"
)

func GenerateUUID() string {
	u4, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Failed to generate a UUID")
	}

	return u4.String()
}