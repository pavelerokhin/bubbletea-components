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

	selected  = make(map[int]int)
	ancestors = make(map[int]int)
)

type item struct {
	ID        int
	Ancestors []item
	Title     string
}

func (i item) FilterValue() string { return i.Title }

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 1 }

func (d itemDelegate) Spacing() int { return 0 }

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	selectionSign := " "

	if val, exist := selected[i.ID]; exist {
		if val == 1 {
			selectionSign = "-"
		} else {
			selectionSign = "+"
		}
	}

	listItemStr := ""
	if val := ancestors[i.ID]; val > 0 {
		listItemStr = fmt.Sprintf("%d. [%s] %s (%d items)", index+1, selectionSign, i.Title, val)
	} else {
		listItemStr = fmt.Sprintf("%d. [%s] %s", index+1, selectionSign, i.Title)
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(listItemStr))
}

type model struct {
	level int
	lists []list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, l := range m.lists {
			l.SetWidth(msg.Width)
		}
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case " ":
			// select in taxonomy
			i, ok := m.lists[m.level].SelectedItem().(item)
			if ok {
				toggleItemAndAncestors(&i)
				m.controlTaxonomy()
			}
			return m, nil

		case "enter":
			i, ok := m.lists[m.level].SelectedItem().(item)
			if ok {
				var anc []item
				for _, a := range i.Ancestors {
					anc = append(anc, a)
				}
				if len(anc) > 0 {
					m.level++
					m.lists = append(m.lists, constructList(anc, m.level))
				}
			}
			return m, nil

		case "backspace":
			if m.level > 0 {
				m.lists = m.lists[:m.level]
				m.level--
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.lists[m.level], cmd = m.lists[m.level].Update(msg)
	return m, cmd
}

func (m model) View() string {
	out := ""
	for i, l := range m.lists {
		if i == 0 {
			out = lipgloss.JoinHorizontal(lipgloss.Left, out, l.View())
		} else {
			out = lipgloss.JoinHorizontal(lipgloss.Left, out, l.View(), "    ")
		}
	}
	return out
}

const defaultWidth = 30

func constructList(taxonomy []item, level int) list.Model {
	var castedTaxonomy []list.Item
	for _, t := range taxonomy {
		castedTaxonomy = append(castedTaxonomy, t)
	}

	l := list.New(castedTaxonomy, itemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	if level != 0 {
		l.SetShowHelp(false)
	}

	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func main() {
	taxonomy := []item{
		item{
			Ancestors: []item{
				{
					Ancestors: []item{
						{
							ID:        5,
							Ancestors: nil,
							Title:     "x",
						},
						{
							ID:        6,
							Ancestors: nil,
							Title:     "y",
						},
					},
					ID:    1,
					Title: "A",
				},
				{
					Ancestors: nil,
					ID:        2,
					Title:     "B",
				},
				{
					Ancestors: nil,
					ID:        3,
					Title:     "C",
				},
			},
			ID:    0,
			Title: "category 1",
		},
		item{
			Ancestors: nil,
			ID:        4,
			Title:     "category 2",
		},
	}

	setAncestors(taxonomy)
	l := constructList(taxonomy, 0)
	m := model{lists: []list.Model{l}}

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// check if not all the elements of sub-levels of the taxonomy are selected
func (m *model) controlTaxonomy() {
	for j := len(m.lists) - 1; j >= 0; j-- {
		l := m.lists[j]
		for _, i := range l.Items() {
			it := i.(item)
			if len(it.Ancestors) > 0 {
				sum := 0
				for _, a := range it.Ancestors {
					sum += selected[a.ID]
				}
				if sum == 0 {
					delete(selected, it.ID)
				} else if sum < len(it.Ancestors)*2 {
					selected[it.ID] = 1
				} else {
					selected[it.ID] = 2
				}
			}
		}
	}
}

func toggleItemAndAncestors(i *item) {
	if _, exist := selected[i.ID]; exist {
		deleteItemAndAncestors(i)
	} else {
		selectItemAndAncestors(i)
	}
}

func deleteItemAndAncestors(i *item) {
	delete(selected, i.ID)
	// delete ancestors
	for _, a := range i.Ancestors {
		deleteItemAndAncestors(&a)
	}
}

func selectItemAndAncestors(i *item) {
	selected[i.ID] = 2
	// select ancestors
	for _, a := range i.Ancestors {
		selectItemAndAncestors(&a)
	}
}

func setAncestors(taxonomy []item) {
	for _, i := range taxonomy {
		ancestors[i.ID] = len(i.Ancestors)
		setAncestors(i.Ancestors)
	}
}
