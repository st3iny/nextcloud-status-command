package command

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/st3iny/nextcloud-status-command/internal/emoji"
	"github.com/st3iny/nextcloud-status-command/internal/ocs"
)

func RunAuth() error {
	auth, err := ocs.LoadAuth()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Warning: Failed to load existing auth data")
	}

	p := tea.NewProgram(newAuthModel(auth))
	m, err := p.Run()
	model := m.(authModel)
	if err != nil {
		return err
	}

	if model.form.State != huh.StateCompleted {
		return nil
	}

	auth = ocs.Auth{
		ServerBaseUrl: model.form.GetString("url"),
		User:          model.form.GetString("user"),
		Password:      model.form.GetString("password"),
	}
	err = ocs.SaveAuth(auth)
	if err != nil {
		return err
	}

	fmt.Println("Credentials were saved")
	return nil
}

func NewAuthProgram() *tea.Program {
	return nil
}

type authModel struct {
	form *huh.Form
}

func newAuthModel(auth ocs.Auth) authModel {
	var emojiOptions = []huh.Option[string]{huh.NewOption("none", "")}
	for _, e := range emoji.Emojis {
		option := huh.NewOption(fmt.Sprintf("%s %s", e.Emoji, e.Description), e.Emoji)
		emojiOptions = append(emojiOptions, option)
	}

	return authModel{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Key("url").
					Lines(1).
					Placeholder("URL ...").
					Title("Type your server's base URL").
					Value(&auth.ServerBaseUrl),
				huh.NewText().
					Key("user").
					Lines(1).
					Placeholder("Username ...").
					Title("Type your username").
					Value(&auth.User),
				huh.NewText().
					Key("password").
					Lines(1).
					Placeholder("Password ...").
					Title("Type your password").
					Value(&auth.Password),
			),
		),
	}
}

func (m authModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m authModel) View() string {
	return m.form.View()
}
