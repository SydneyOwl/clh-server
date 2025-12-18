package pkg

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gookit/slog"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type (
	errMsg error
)

type Model struct {
	viewport    viewport.Model
	messages    *[]string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	lock        *sync.RWMutex

	endFunc  func()
	writeCmd func(command string) (string, error)
}

func InitialModel(preExec func(), writeCmd func(command string) (string, error), endExec func()) Model {
	ta := textarea.New()
	ta.Placeholder = "Send a command..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Type command and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	slog.Std().Output = io.Discard

	go preExec()

	return Model{
		textarea:    ta,
		messages:    &[]string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		lock:        &sync.RWMutex{},
		endFunc:     endExec,
		writeCmd:    writeCmd,
	}
}

func (m Model) AppendInfo(info string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	*m.messages = append(*m.messages, info)
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(*m.messages) > 0 {
			m.lock.RLock()
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(*m.messages, "\n")))
			m.lock.RUnlock()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.endFunc()
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			*m.messages = append(*m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			rep, err := m.writeCmd(m.textarea.Value())

			m.lock.Lock()
			if err != nil {
				*m.messages = append(*m.messages, m.senderStyle.Render("Error: ")+err.Error())
			} else {
				*m.messages = append(*m.messages, rep)
			}
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(*m.messages, "\n")))
			m.lock.Unlock()
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
