package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TodoItem struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	ID        int    `json:"id"`
}

func (t TodoItem) FilterValue() string { return t.Text }
func (t TodoItem) Title() string       { return t.Text }
func (t TodoItem) Description() string {
	return ""
}

type todoMode int

const (
	browsing todoMode = iota
	adding
	editing
)

type TodoModel struct {
	list       list.Model
	textarea   textarea.Model
	mode       todoMode
	keys       TodoKeyMap
	todos      []TodoItem
	nextID     int
	editingIdx int
}

type TodoKeyMap struct {
	Add     key.Binding
	Delete  key.Binding
	Toggle  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func DefaultTodoKeys() TodoKeyMap {
	return TodoKeyMap{
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add todo"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle done"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

func NewTodoModel() TodoModel {
	items := []list.Item{}

	l := list.New(items, list.NewDefaultDelegate(), 50, 10)
	l.Title = "📝 Todo List"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	ta := textarea.New()
	ta.Placeholder = "Enter your todo item..."
	ta.Focus()
	ta.CharLimit = 100
	ta.SetWidth(50)
	ta.SetHeight(3)

	tm := TodoModel{
		list:     l,
		textarea: ta,
		mode:     browsing,
		keys:     DefaultTodoKeys(),
		todos:    []TodoItem{},
		nextID:   1,
	}

	tm.loadTodos()
	return tm
}

func (m TodoModel) Init() tea.Cmd {
	return nil
}

func (m TodoModel) Update(msg tea.Msg) (TodoModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 10)
		return m, nil

	case tea.KeyMsg:
		if m.mode == adding {
			switch msg.String() {
			case "enter":
				text := m.textarea.Value()
				if text != "" {
					m.addTodo(text)
					m.textarea.Reset()
				}
				m.mode = browsing
				return m, nil
			case "esc":
				m.textarea.Reset()
				m.mode = browsing
				return m, nil
			default:
				// Let textarea handle other keys when in adding mode
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		} else if m.mode == editing {
			switch msg.String() {
			case "enter":
				text := m.textarea.Value()
				if text != "" {
					m.updateTodo(m.editingIdx, text)
					m.textarea.Reset()
				}
				m.mode = browsing
				return m, nil
			case "esc":
				m.textarea.Reset()
				m.mode = browsing
				return m, nil
			default:
				// Let textarea handle other keys when in editing mode
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "a":
				m.mode = adding
				return m, m.textarea.Focus()
			case "d":
				if len(m.todos) > 0 {
					selected := m.list.Index()
					if selected >= 0 && selected < len(m.todos) {
						m.deleteTodo(selected)
					}
				}
				return m, nil
			case "e":
				if len(m.todos) > 0 {
					selected := m.list.Index()
					if selected >= 0 && selected < len(m.todos) {
						m.editTodo(selected)
						m.mode = editing
					}
				}
				return m, m.textarea.Focus()
			case "enter":
				if len(m.todos) > 0 {
					selected := m.list.Index()
					if selected >= 0 && selected < len(m.todos) {
						m.toggleTodo(selected)
					}
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.mode == adding || m.mode == editing {
		m.textarea, cmd = m.textarea.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m *TodoModel) addTodo(text string) {
	todo := TodoItem{
		Text:      text,
		Completed: false,
		ID:        m.nextID,
	}
	m.nextID++
	m.todos = append(m.todos, todo)
	m.updateList()
	m.saveTodos()
}

func (m *TodoModel) deleteTodo(index int) {
	if index >= 0 && index < len(m.todos) {
		m.todos = append(m.todos[:index], m.todos[index+1:]...)
		m.updateList()
		m.saveTodos()
	}
}

func (m *TodoModel) toggleTodo(index int) {
	if index >= 0 && index < len(m.todos) {
		m.todos[index].Completed = !m.todos[index].Completed
		m.updateList()
		m.saveTodos()
	}
}

func (m *TodoModel) editTodo(index int) {
	if index >= 0 && index < len(m.todos) {
		m.editingIdx = index
		m.textarea.SetValue(m.todos[index].Text)
	}
}

func (m *TodoModel) updateTodo(index int, text string) {
	if index >= 0 && index < len(m.todos) {
		m.todos[index].Text = text
		m.updateList()
		m.saveTodos()
	}
}

func (m *TodoModel) updateList() {
	items := make([]list.Item, len(m.todos))
	for i, todo := range m.todos {
		items[i] = todo
	}
	m.list.SetItems(items)
}

func (m *TodoModel) saveTodos() {
	saveTodosToFile(m.todos)
}

func (m *TodoModel) loadTodos() {
	todos, err := loadTodosFromFile()
	if err != nil {
		return
	}

	m.todos = todos

	maxID := 0
	for _, todo := range m.todos {
		if todo.ID > maxID {
			maxID = todo.ID
		}
	}
	m.nextID = maxID + 1

	m.updateList()
}

func (m TodoModel) View() string {
	width := 60 // Fixed width for consistent centering

	if m.mode == adding {
		addStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width)

		return addStyle.Render(fmt.Sprintf(
			"Add new todo:\n\n%s\n\n%s",
			m.textarea.View(),
			"Press Enter to confirm, Esc to cancel",
		))
	}

	if m.mode == editing {
		editStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width)

		return editStyle.Render(fmt.Sprintf(
			"Edit todo:\n\n%s\n\n%s",
			m.textarea.View(),
			"Press Enter to confirm, Esc to cancel",
		))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Align(lipgloss.Center).
		Width(width)

	help := helpStyle.Render("a: add • e: edit • enter: toggle • d: delete")

	// Debug: show todos directly if list is empty
	debugStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		MarginBottom(1).
		Align(lipgloss.Center).
		Width(width)

	var debugInfo string
	if len(m.todos) == 0 {
		debugInfo = debugStyle.Render("No todos yet. Press 'a' to add one.")
	} else {
		debugContent := ""
		for i, todo := range m.todos {
			marker := "  "
			if i == m.list.Index() {
				marker = "▶ "
			}
			debugContent += fmt.Sprintf("\n%s• %s", marker, todo.Text)
		}
		debugInfo = debugStyle.Render(debugContent)
	}

	listStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		debugInfo,
		"",
		listStyle.Render(m.list.View()),
		help,
	)
}
