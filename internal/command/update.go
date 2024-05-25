package command

import (
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/st3iny/nsc/internal/emoji"
	"github.com/st3iny/nsc/internal/ocs"
)

func RunUpdate() error {
	auth, err := ocs.LoadAuth()
	if err != nil {
		return missingAuthError()
	}

	p := tea.NewProgram(newUpdateModel())
	m, err := p.Run()
	model := m.(updateModel)
	if err != nil {
		return fmt.Errorf("Failed to render form: %s", err)
	}

	if model.form.State != huh.StateCompleted {
		return nil
	}

	return spinner.New().
		Title("Updating your status ...").
		Action(func() {
			updateStatusAndMessage(auth, model)
		}).
		Run()
}

type updateModel struct {
	form *huh.Form
}

func newUpdateModel() updateModel {
	var emojiOptions = []huh.Option[string]{huh.NewOption("none", "")}
	for _, e := range emoji.Emojis {
		option := huh.NewOption(fmt.Sprintf("%s %s", e.Emoji, e.Description), e.Emoji)
		emojiOptions = append(emojiOptions, option)
	}

	return updateModel{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("status").
					Options(huh.NewOptions("online", "away", "dnd", "invisible")...).
					Title("Choose a status"),
				huh.NewSelect[string]().
					Key("emoji").
					Options(emojiOptions...).
					Height(10).
					Title("Choose an emoji (type / to search)"),
				huh.NewText().
					Key("message").
					Lines(1).
					Placeholder("Status message ...").
					Title("Type a status message"),
			),
		),
	}
}

func (m updateModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m updateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m updateModel) View() string {
	return m.form.View()
}

func updateStatusAndMessage(auth ocs.Auth, model updateModel) {
	status := model.form.GetString("status")
	message := model.form.GetString("message")
	emoji := model.form.GetString("emoji")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err := ocs.UpdateStatus(auth, ocs.Status{
			StatusType: status,
		})
		if err != nil {
			fmt.Println(err)
		}

		wg.Done()
	}()

	go func() {
		err := ocs.UpdateStatusMessage(auth, ocs.StatusMessage{
			ClearAt:    time.Now().Unix() + 3600,
			Message:    message,
			StatusIcon: emoji,
		})
		if err != nil {
			fmt.Println(err)
		}

		wg.Done()
	}()

	wg.Wait()
}

func missingAuthError() error {
	return fmt.Errorf(
		"Not authenticated to a Nextcloud server\n"+
			"Please run \"%s auth\" first",
		os.Args[0],
	)
}