package pkg

// Deprecated code

//
//import (
//	"fmt"
//	"io"
//	"strings"
//	"sync"
//
//	"github.com/gookit/slog"
//
//	"github.com/charmbracelet/bubbles/textarea"
//	"github.com/charmbracelet/bubbles/viewport"
//	tea "github.com/charmbracelet/bubbletea"
//	"github.com/charmbracelet/lipgloss"
//)
//
//const gap = "\n\n"
//
//type (
//	errMsg error
//)
//
//// Custom message types for async operations
//type ResponseMsg struct {
//	Response string
//	Err      error
//}
//
//type commandMsg struct {
//	command string
//}
//
//type Model struct {
//	viewport    viewport.Model
//	messages    []string
//	textarea    textarea.Model
//	senderStyle lipgloss.Style
//	err         error
//	lock        *sync.RWMutex
//
//	endFunc  func()
//	writeCmd func(command string) (string, error)
//
//	ready       bool
//	initialized bool
//	loading     bool // Track if we're waiting for a response
//}
//
//func InitialModel(preExec func(), writeCmd func(command string) (string, error), endExec func()) Model {
//	ta := textarea.New()
//	ta.Placeholder = "Send a command..."
//	ta.Focus()
//
//	ta.Prompt = "┃ "
//	ta.CharLimit = 280
//
//	ta.SetWidth(30)
//	ta.SetHeight(3)
//
//	// Remove cursor line styling
//	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
//
//	ta.ShowLineNumbers = false
//
//	vp := viewport.New(30, 5)
//	vp.SetContent(`Type command and press Enter to send.`)
//
//	ta.KeyMap.InsertNewline.SetEnabled(false)
//
//	slog.Std().Output = io.Discard
//
//	go preExec()
//
//	return Model{
//		textarea:    ta,
//		messages:    []string{},
//		viewport:    vp,
//		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
//		err:         nil,
//		lock:        &sync.RWMutex{},
//		endFunc:     endExec,
//		writeCmd:    writeCmd,
//		ready:       false,
//		initialized: false,
//		loading:     false,
//	}
//}
//
//// Helper method to render all messages
//func (m *Model) renderMessages() string {
//	m.lock.RLock()
//	defer m.lock.RUnlock()
//
//	if len(m.messages) == 0 {
//		return "No messages yet..."
//	}
//
//	// Join messages with newlines
//	return strings.Join(m.messages, "\n")
//}
//
//// Async function to send command and get response
//func (m *Model) sendCommandAsync(command string) tea.Cmd {
//	return func() tea.Msg {
//		response, err := m.writeCmd(command)
//		return ResponseMsg{Response: response, Err: err}
//	}
//}
//
//func (m Model) Init() tea.Cmd {
//	return textarea.Blink
//}
//
//func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	var (
//		tiCmd tea.Cmd
//		vpCmd tea.Cmd
//		cmds  []tea.Cmd
//	)
//
//	m.textarea, tiCmd = m.textarea.Update(msg)
//	m.viewport, vpCmd = m.viewport.Update(msg)
//
//	if tiCmd != nil {
//		cmds = append(cmds, tiCmd)
//	}
//	if vpCmd != nil {
//		cmds = append(cmds, vpCmd)
//	}
//
//	switch msg := msg.(type) {
//	case tea.WindowSizeMsg:
//		// Handle initial window size
//		if !m.ready {
//			m.viewport = viewport.New(msg.Width, msg.Height-5)
//			m.viewport.HighPerformanceRendering = false
//			m.ready = true
//			m.viewport.SetContent(m.renderMessages())
//			m.viewport.GotoBottom()
//		} else {
//			// Update existing viewport size
//			m.viewport.Width = msg.Width
//			m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
//			m.textarea.SetWidth(msg.Width)
//			m.viewport.SetContent(m.renderMessages())
//		}
//
//	case tea.KeyMsg:
//		switch msg.Type {
//		case tea.KeyCtrlC, tea.KeyEsc:
//			if m.endFunc != nil {
//				m.endFunc()
//			}
//			return m, tea.Quit
//
//		case tea.KeyEnter:
//			command := strings.TrimSpace(m.textarea.Value())
//			if command == "" {
//				return m, tea.Batch(cmds...)
//			}
//
//			// Add user message
//			m.lock.Lock()
//			m.messages = append(m.messages, m.senderStyle.Render("You: ")+command)
//			m.lock.Unlock()
//
//			// Update viewport immediately
//			content := m.renderMessages()
//			m.viewport.SetContent(content)
//			m.viewport.GotoBottom()
//
//			// Reset textarea
//			m.textarea.Reset()
//
//			// Set loading state
//			m.loading = true
//
//			// Add loading message
//			m.lock.Lock()
//			m.messages = append(m.messages, "⏳ Processing...")
//			m.lock.Unlock()
//			content = m.renderMessages()
//			m.viewport.SetContent(content)
//			m.viewport.GotoBottom()
//
//			// Send command asynchronously and return cmd to Bubble Tea
//			return m, tea.Batch(
//				m.sendCommandAsync(command),
//				tea.Batch(cmds...),
//			)
//		}
//
//	case ResponseMsg:
//		// Handle the async response
//		m.loading = false
//
//		// Remove the "Processing..." message
//		m.lock.Lock()
//		if len(m.messages) > 0 && m.messages[len(m.messages)-1] == "⏳ Processing..." {
//			m.messages = m.messages[:len(m.messages)-1]
//		}
//
//		// Add response or error
//		if msg.Err != nil {
//			m.messages = append(m.messages, m.senderStyle.Render("Error: ")+msg.Err.Error())
//		} else if msg.Response != "" {
//			m.messages = append(m.messages, msg.Response)
//		} else {
//			m.messages = append(m.messages, "(No response)")
//		}
//		m.lock.Unlock()
//
//		// Update viewport
//		content := m.renderMessages()
//		m.viewport.SetContent(content)
//		m.viewport.GotoBottom()
//
//		return m, tea.Batch(cmds...)
//
//	case errMsg:
//		m.err = msg
//		return m, nil
//	}
//
//	return m, tea.Batch(cmds...)
//}
//
//func (m Model) View() string {
//	if !m.ready {
//		return "\n  Initializing..."
//	}
//
//	// Ensure viewport has content
//	if m.viewport.View() == "" {
//		m.viewport.SetContent(m.renderMessages())
//	}
//
//	// Calculate proper heights
//	//textareaHeight := lipgloss.Height(m.textarea.View())
//	//gapHeight := lipgloss.Height(gap)
//
//	// Adjust viewport height if needed
//	if m.viewport.Height <= 0 {
//		m.viewport.Height = 5
//	}
//
//	return fmt.Sprintf(
//		"%s%s%s",
//		m.viewport.View(),
//		gap,
//		m.textarea.View(),
//	)
//}
