package command

import (
	"fmt"
	"time"

	"github.com/st3iny/nextcloud-status-command/internal/ocs"
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

	clearAt := "never"
	if status.ClearAt > 0 {
		clearAt = time.Unix(status.ClearAt, 0).String()
	}

	if status.Icon == "" {
		fmt.Printf("%s (%s) %s\n", status.User, status.Status, status.Message)
	} else {
		fmt.Printf("%s (%s) %s %s\n", status.User, status.Status, status.Icon, status.Message)
	}
	fmt.Printf("clear at %s\n", clearAt)
	return nil
}
