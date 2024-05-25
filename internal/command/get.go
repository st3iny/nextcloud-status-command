package command

import (
	"fmt"

	"github.com/st3iny/nsc/internal/ocs"
)

func RunGet() error {
	auth, err := ocs.LoadAuth()
	if err != nil {
		return missingAuthError()
	}

	status, err := ocs.GetStatus(auth)
	if err != nil {
		return err
	}

	fmt.Printf("%s (%s) %s %s\n", status.User, status.Status, status.Icon, status.Message)
	return nil
}
