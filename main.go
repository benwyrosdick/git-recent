package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	branches        []string
	allBranches     []string // original unfiltered list
	cursor          int
	offset          int
	remote          bool
	selected        bool
	err             error
	filterMode      bool
	filterText      string
	filteredApplied bool // tracks if we're showing a filtered list
}

func getRecentBranches(remote bool) ([]string, error) {
	var cmd *exec.Cmd
	if remote {
		cmd = exec.Command("git", "for-each-ref", "--sort=-committerdate", "refs/remotes/", "--format=%(refname:short)")
	} else {
		cmd = exec.Command("git", "for-each-ref", "--sort=-committerdate", "refs/heads/", "--format=%(refname:short)")
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var filtered []string
	for _, b := range branches {
		if b != "" && !strings.HasSuffix(b, "/HEAD") {
			filtered = append(filtered, b)
		}
	}
	return filtered, nil
}

func initialModel(remote bool) model {
	branches, err := getRecentBranches(remote)
	return model{
		branches:        branches,
		allBranches:     branches,
		cursor:          0,
		offset:          0,
		remote:          remote,
		selected:        false,
		err:             err,
		filterMode:      false,
		filterText:      "",
		filteredApplied: false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle filter mode
		if m.filterMode {
			switch msg.String() {
			case "esc":
				// Cancel filter mode and restore original list
				m.filterMode = false
				m.filterText = ""
				m.branches = m.allBranches
				m.cursor = 0
				m.offset = 0
				m.filteredApplied = false
			case "enter":
				// Keep the filtered list and exit filter mode
				m.filterMode = false
				m.filteredApplied = true
			case "backspace":
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
				}
			default:
				// Add character to filter
				if len(msg.String()) == 1 {
					m.filterText += msg.String()
					m.applyFilter()
				}
			}
			return m, nil
		}

		// Normal mode
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			// Clear filter if one is applied, otherwise quit
			if m.filteredApplied {
				m.branches = m.allBranches
				m.filterText = ""
				m.cursor = 0
				m.offset = 0
				m.filteredApplied = false
			} else {
				return m, tea.Quit
			}

		case "/":
			// Enter filter mode
			m.filterMode = true
			m.filterText = ""

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}

		case "down", "j":
			if m.cursor < len(m.branches)-1 {
				m.cursor++
				if m.cursor >= m.offset+10 {
					m.offset++
				}
			}

		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) applyFilter() {
	if m.filterText == "" {
		m.branches = m.allBranches
		m.cursor = 0
		m.offset = 0
		return
	}

	var filtered []string
	filterLower := strings.ToLower(m.filterText)
	for _, branch := range m.allBranches {
		if strings.Contains(strings.ToLower(branch), filterLower) {
			filtered = append(filtered, branch)
		}
	}
	m.branches = filtered
	m.cursor = 0
	m.offset = 0
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if len(m.branches) == 0 {
		if m.filterMode {
			return fmt.Sprintf("No branches match filter.\n\nFilter: /%s_\n\n(type to filter, enter to keep, esc to cancel)\n", m.filterText)
		}
		return "No branches found.\n"
	}

	s := "Select a branch to checkout:\n\n"

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	end := m.offset + 10
	if end > len(m.branches) {
		end = len(m.branches)
	}

	for i := m.offset; i < end; i++ {
		branch := m.branches[i]
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render("›")
			branch = selectedStyle.Render(branch)
		}
		s += fmt.Sprintf("%s %s\n", cursor, branch)
	}

	s += "\n"

	if m.filterMode {
		s += fmt.Sprintf("Filter: /%s_\n", m.filterText)
		s += "(type to filter, enter to keep, esc to cancel)\n"
	} else if m.filteredApplied {
		s += fmt.Sprintf("[Filtered: %s] ", m.filterText)
		s += "(/ to filter, esc to clear, j/k to move, enter to select, q to quit)\n"
	} else {
		s += "(/ to filter, j/k to move, enter to select, q to quit)\n"
	}

	return s
}

func checkoutBranch(branch string, remote bool) error {
	var cmd *exec.Cmd
	if remote {
		parts := strings.SplitN(branch, "/", 2)
		if len(parts) == 2 {
			localBranch := parts[1]
			cmd = exec.Command("git", "checkout", localBranch)
		} else {
			cmd = exec.Command("git", "checkout", "--track", branch)
		}
	} else {
		cmd = exec.Command("git", "checkout", branch)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	remote := flag.Bool("r", false, "list remote branches")
	flag.BoolVar(remote, "remote", false, "list remote branches")
	flag.Parse()

	p := tea.NewProgram(initialModel(*remote))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	finalModel := m.(model)
	if finalModel.err != nil {
		fmt.Printf("Error: %v\n", finalModel.err)
		os.Exit(1)
	}

	if finalModel.selected && len(finalModel.branches) > 0 {
		selectedBranch := finalModel.branches[finalModel.cursor]
		fmt.Printf("Checking out: %s\n", selectedBranch)
		if err := checkoutBranch(selectedBranch, finalModel.remote); err != nil {
			fmt.Printf("Failed to checkout branch: %v\n", err)
			os.Exit(1)
		}
	}
}
