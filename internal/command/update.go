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

	defaultStatus := statusOnline
	defaultEmoji := ""
	defaultMessage := ""
	defaultTimeoutKey := timeoutNever

	statusValue := flag.String("status", defaultStatus, fmt.Sprintf(
		"your status [options: %s]",
		strings.Join(statusOptions, ", "),
	))
	emojiValue := flag.String("emoji", defaultEmoji, "your status emoji")
	messageValue := flag.String("message", defaultMessage, "your status message")
	timeoutKey := flag.String("timeout", defaultTimeoutKey, fmt.Sprintf(
		"timeout after which to delete your status [options: %s]",
		strings.Join(timeoutOptions, ", "),
	))
	submit := flag.Bool("submit", false, "skip the form and submit your status directly")
	empty := flag.Bool("empty", false, "do not prefill all fields with values from your current status")
	flag.Parse()

	auth, err := ocs.LoadAuth()
	if err != nil {
		return missingAuthError()
	}

	var timeoutValue int64
	if *empty || *statusValue != defaultStatus || *emojiValue != defaultEmoji || *messageValue != defaultMessage || *timeoutKey != defaultTimeoutKey {
		timeoutValue = timeoutKeyToValue(*timeoutKey)
	} else {
		statusChannel := make(chan *ocs.UserStatus, 1)
		errorChannel := make(chan error, 1)

		spinner.New().
			Title("Fetching your current status ...").
			Action(func() {
				status, err := ocs.GetStatus(auth)
				if err != nil {
					errorChannel <- err
					return
				}

				statusChannel <- status
			}).
			Run()

		select {
		case err := <-errorChannel:
			return fmt.Errorf("Failed to fetch current status: %s", err)
		case status := <-statusChannel:
			*statusValue = status.Status
			*emojiValue = status.Icon
			*messageValue = status.Message
			timeoutValue = status.ClearAt
		}
	}

	if !*submit {
		model := newUpdateModel(statusValue, emojiValue, messageValue, &timeoutValue)
		p := tea.NewProgram(model)
		m, err := p.Run()
		if err != nil {
			return fmt.Errorf("Failed to render form: %s", err)
		}

		model = m.(updateModel)
		if model.form.State != huh.StateCompleted {
			return nil
		}

		*statusValue = model.form.GetString("status")
		*messageValue = model.form.GetString("message")
		*emojiValue = model.form.GetString("emoji")
		timeoutValue = model.form.Get("timeout").(int64)
	}

	return spinner.New().
		Title("Updating your status ...").
		Action(func() {
			updateStatus(auth, *statusValue, *messageValue, *emojiValue, timeoutValue)
		}).
		Run()
}

type updateModel struct {
	form *huh.Form
}

func newUpdateModel(statusValue, emojiValue, messageValue *string, timeoutValue *int64) updateModel {
	emojiOptions := []huh.Option[string]{huh.NewOption("none", "")}
	for _, e := range emoji.Emojis {
		if len(e.Emoji) > 4 {
			continue
		}

		option := huh.NewOption(fmt.Sprintf("%s %s", e.Emoji, e.Description), e.Emoji)
		emojiOptions = append(emojiOptions, option)
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
					Options(timeoutOptions(timeoutValue)...).
					Title("Delete status after").
					Value(timeoutValue),
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

func updateStatus(auth ocs.Auth, status, message, emoji string, timeout int64) {
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

func timeoutOptions(timeoutValue *int64) []huh.Option[int64] {
	now := time.Now()
	nowUnix := now.Unix()
	startOfTodayUnix := time.
		Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).
		Unix()
	daysUntilSunday := int64(daysFromStartOfDayUntilEndOfSunday(now))
	options := []huh.Option[int64]{
		huh.NewOption(timeoutNever, int64(0)),
		huh.NewOption(timeout30Minutes, nowUnix+1800),
		huh.NewOption(timeout1Hour, nowUnix+3600),
		huh.NewOption(timeout4Hours, nowUnix+4*3600),
		huh.NewOption(timeoutToday, startOfTodayUnix+24*3600),
		huh.NewOption(timeoutThisWeek, startOfTodayUnix+daysUntilSunday*24*3600),
	}

	if timeoutValue == nil {
		return options
	}

	needsCustomOption := true
	for _, option := range options {
		if option.Value == *timeoutValue {
			needsCustomOption = false
			break
		}
	}

	if needsCustomOption {
		options = append(options, huh.NewOption(
			fmt.Sprintf("custom (%s)", time.Unix(*timeoutValue, 0)),
			*timeoutValue,
		))
	}

	return options
}

func timeoutKeyToValue(timeoutKey string) int64 {
	for _, option := range timeoutOptions(nil) {
		if option.Key == timeoutKey {
			return option.Value
		}
	}

	return 0
}
