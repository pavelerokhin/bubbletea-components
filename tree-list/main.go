package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	Ancestors []item
	Title     string
	Selected  bool
}

func (i item) FilterValue() string { return i.Title }
func (i item) Convert() list.Item  { return i }

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 1 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str))
}

type model struct {
	taxonomy list.Model
	level    int
	lists    []list.Model
	choice   item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		detailLevel := len(m.lists)
		for _, l := range m.lists {
			l.SetWidth(msg.Width / detailLevel)
		}

		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.lists[m.level].SelectedItem().(item)
			if ok {
				var ancestors []list.Item
				for _, a := range i.Ancestors {
					ancestors = append(ancestors, a.Convert())
				}
				m.lists = append(m.lists, constructList(ancestors))
				m.level++
			}
			return m, nil

			//case "backspace":
			//	i, ok := m.list.SelectedItem().(item)
			//	if ok {
			//		lists = append(lists, constructList(i.Ancestors.(list.Item)))
			//	}
			//	return m, nil
		}
	}

	var cmd tea.Cmd
	m.lists[m.level], cmd = m.lists[m.level].Update(msg)
	return m, cmd
}

func (m model) View() string {
	out := ""
	for _, l := range m.lists {
		out = lipgloss.JoinHorizontal(lipgloss.Left, out, "\n"+l.View())
	}
	return out
	//return "\n" + m.list.View()
}

const defaultWidth = 20

func constructList(taxonomy []list.Item) list.Model {
	l := list.New(taxonomy, itemDelegate{}, defaultWidth, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func main() {
	taxonomy := []list.Item{
		item{
			Ancestors: []item{
				{
					Ancestors: nil,
					Title:     "A",
				},
				{
					Ancestors: nil,
					Title:     "B",
				},
				{
					Ancestors: nil,
					Title:     "C",
				},
			},
			Title: "category 1",
		},
		item{
			Ancestors: nil,
			Title:     "category 2",
		},
	}

	l := constructList(taxonomy)

	m := model{taxonomy: l, lists: []list.Model{l}}

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
