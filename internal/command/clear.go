package command

import (
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/st3iny/nextcloud-status-command/internal/ocs"
)

type errorMsg struct {
	err error
}

func RunClear() error {
	auth, err := ocs.LoadAuth()
	if err != nil {
		return fmt.Errorf("Failed to load auth: %s", err)
	}

	errChan := make(chan error, 1)
	err = spinner.New().
		Title("Clearing your status message ...").
		Action(func() {
			errChan <- ocs.ClearStatusMessage(auth)
		}).
		Run()
	if err != nil {
		return fmt.Errorf("Failed to render spinner: %s", err)
	}

	return <-errChan
}
