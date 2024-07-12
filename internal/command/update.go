package command

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/st3iny/nextcloud-status-command/internal/emoji"
	"github.com/st3iny/nextcloud-status-command/internal/ocs"
)

const (
	statusOnline    = "online"
	statusAway      = "away"
	statusDnd       = "dnd"
	statusInvisible = "invisible"

	timeoutNever     = "never"
	timeout30Minutes = "30 minutes"
	timeout1Hour     = "1 hour"
	timeout4Hours    = "4 hours"
	timeoutToday     = "today"
	timeoutThisWeek  = "this week"
)

func RunUpdate() error {
	statusOptions := []string{
		statusOnline,
		statusAway,
		statusDnd,
		statusInvisible,
	}
	timeoutOptions := []string{
		timeoutNever,
		timeout30Minutes,
		timeout1Hour,
		timeout4Hours,
		timeoutToday,
		timeoutThisWeek,
	}

	statusValue := flag.String("status", statusOnline, fmt.Sprintf(
		"your status [options: %s]",
		strings.Join(statusOptions, ", "),
	))
	emojiValue := flag.String("emoji", "", "your status emoji")
	messageValue := flag.String("message", "", "your status message")
	timeoutKey := flag.String("timeout", timeoutNever, fmt.Sprintf(
		"timeout after which to delete your status [options: %s]",
		strings.Join(timeoutOptions, ", "),
	))
	submit := flag.Bool("submit", false, "skip the form and submit your status directly")
	flag.Parse()

	auth, err := ocs.LoadAuth()
	if err != nil {
		return missingAuthError()
	}

	p := tea.NewProgram(newUpdateModel(statusValue, emojiValue, messageValue, timeoutKey))
	// Hackity hack: Simulate form submission by pressing enter once for each field
	if *submit {
		go func() {
			p.Send(tea.KeyMsg{Type: tea.KeyEnter})
			p.Send(tea.KeyMsg{Type: tea.KeyEnter})
			p.Send(tea.KeyMsg{Type: tea.KeyEnter})
			p.Send(tea.KeyMsg{Type: tea.KeyEnter})
		}()
	}
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

func newUpdateModel(statusValue, emojiValue, messageValue, timeoutKey *string) updateModel {
	emojiOptions := []huh.Option[string]{huh.NewOption("none", "")}
	for _, e := range emoji.Emojis {
		option := huh.NewOption(fmt.Sprintf("%s %s", e.Emoji, e.Description), e.Emoji)
		emojiOptions = append(emojiOptions, option)
	}

	now := time.Now()
	nowUnix := now.Unix()
	startOfTodayUnix := time.
		Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		Unix()
	daysUntilSunday := int64(daysFromStartOfDayUntilEndOfSunday(now))
	timeoutOptions := []huh.Option[int64]{
		huh.NewOption(timeoutNever, int64(0)),
		huh.NewOption(timeout30Minutes, nowUnix+1800),
		huh.NewOption(timeout1Hour, nowUnix+3600),
		huh.NewOption(timeout4Hours, nowUnix+4*3600),
		huh.NewOption(timeoutToday, startOfTodayUnix+24*3600),
		huh.NewOption(timeoutThisWeek, startOfTodayUnix+daysUntilSunday*24*3600),
	}

	timeoutValue := new(int64)
	for _, option := range timeoutOptions {
		if option.Key == *timeoutKey {
			*timeoutValue = option.Value
			break
		}
	}

	return updateModel{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("status").
					Options(huh.NewOptions(
						statusOnline,
						statusAway,
						statusDnd,
						statusInvisible,
					)...).
					Title("Choose a status").
					Value(statusValue),
				huh.NewSelect[string]().
					Key("emoji").
					Options(emojiOptions...).
					Height(10).
					Title("Choose an emoji (type / to search)").
					Value(emojiValue),
				huh.NewText().
					Key("message").
					Lines(1).
					Placeholder("Status message ...").
					Title("Type a status message").
					Value(messageValue),
				huh.NewSelect[int64]().
					Key("timeout").
					Options(timeoutOptions...).
					Title("Delete status after").
					Value(timeoutValue),
			),
		),
	}
}

func daysFromStartOfDayUntilEndOfSunday(date time.Time) int {
	var daysUntilEndOfSunday int
	weekday := int(date.Weekday())
	if weekday == 0 /* Sunday */ {
		daysUntilEndOfSunday = 1
	} else /* Not Sunday */ {
		daysUntilEndOfSunday = 7 - weekday + 1
	}
	return daysUntilEndOfSunday
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
	timeout := model.form.Get("timeout").(int64)

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
			ClearAt:    timeout,
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
