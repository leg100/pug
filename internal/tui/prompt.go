package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Height of prompt including borders
const PromptHeight = 3

// PromptMsg enables the prompt widget.
type PromptMsg struct {
	// Prompt to display to the user.
	Prompt string
	// Set initial value for the user to edit.
	InitialValue string
	// Action to carry out when key is pressed.
	Action PromptAction
	// Key that when pressed triggers the action and closes the prompt.
	Key key.Binding
	// Cancel is a key that when pressed skips the action and closes the prompt.
	Cancel key.Binding
	// CancelAnyOther, if true, checks if any key other than that specified in
	// Key is pressed. If so then the action is skipped and the prompt is
	// closed. Overrides Cancel key binding.
	CancelAnyOther bool
	// Set placeholder text in prompt
	Placeholder string
}

type PromptAction func(text string) tea.Cmd

// YesNoPrompt sends a message to enable the prompt widget, specifically
// asking the user for a yes/no answer. If yes is given then the action is
// invoked.
func YesNoPrompt(prompt string, action tea.Cmd) tea.Cmd {
	return CmdHandler(PromptMsg{
		Prompt: fmt.Sprintf("%s (y/N): ", prompt),
		Action: func(_ string) tea.Cmd {
			return action
		},
		Key: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		CancelAnyOther: true,
	})
}

func NewPrompt(msg PromptMsg) (*Prompt, tea.Cmd) {
	model := textinput.New()
	model.Prompt = msg.Prompt
	model.SetValue(msg.InitialValue)
	model.Placeholder = msg.Placeholder
	model.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	blink := model.Focus()

	prompt := Prompt{
		model:          model,
		action:         msg.Action,
		trigger:        msg.Key,
		cancel:         msg.Cancel,
		cancelAnyOther: msg.CancelAnyOther,
	}
	return &prompt, blink
}

// Prompt is a widget that prompts the user for input and triggers an action.
type Prompt struct {
	model          textinput.Model
	action         PromptAction
	trigger        key.Binding
	cancel         key.Binding
	cancelAnyOther bool
}

// HandleKey handles the user key press, and returns a command to be run, and
// whether the prompt should be closed.
func (p *Prompt) HandleKey(msg tea.KeyMsg) (closePrompt bool, cmd tea.Cmd) {
	switch {
	case key.Matches(msg, p.trigger):
		cmd = p.action(p.model.Value())
		closePrompt = true
	case key.Matches(msg, p.cancel), p.cancelAnyOther:
		cmd = ReportInfo("canceled operation")
		closePrompt = true
	default:
		p.model, cmd = p.model.Update(msg)
	}
	return
}

// HandleBlink handles the bubbletea blink message.
func (p *Prompt) HandleBlink(msg tea.Msg) (cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ignore key presses, they're handled by HandleKey above.
	default:
		// The blink message type is unexported so we just send unknown types to
		// the model.
		p.model, cmd = p.model.Update(msg)
	}
	return
}

func (p *Prompt) View(width int) string {
	paddedBorder := ThickBorder.BorderForeground(Red).Padding(0, 1)
	paddedBorderWidth := paddedBorder.GetHorizontalBorderSize() + paddedBorder.GetHorizontalPadding()
	// Set available width for user entered value before it horizontally
	// scrolls.
	p.model.Width = max(0, width-lipgloss.Width(p.model.Prompt)-paddedBorderWidth)
	// Render a prompt, surrounded by a padded red border, spanning the width of the
	// terminal, accounting for width of border. Inline and MaxWidth ensures the
	// prompt remains on a single line.
	content := Regular.Inline(true).MaxWidth(width - paddedBorderWidth).Render(p.model.View())
	return paddedBorder.Width(width - paddedBorder.GetHorizontalBorderSize()).Render(content)
}

func (p *Prompt) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		p.trigger,
	}
	if p.cancelAnyOther {
		bindings = append(bindings, key.NewBinding(key.WithHelp("n", "cancel")))
	} else {
		bindings = append(bindings, p.cancel)
	}
	return bindings
}
