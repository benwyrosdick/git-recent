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
	branches []string
	cursor   int
	offset   int
	remote   bool
	err      error
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
		branches: branches,
		cursor:   0,
		offset:   0,
		remote:   remote,
		err:      err,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

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
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if len(m.branches) == 0 {
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

	s += "\n(j/k to move, enter to select, q to quit)\n"
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

	if len(finalModel.branches) > 0 {
		selectedBranch := finalModel.branches[finalModel.cursor]
		fmt.Printf("Checking out: %s\n", selectedBranch)
		if err := checkoutBranch(selectedBranch, finalModel.remote); err != nil {
			fmt.Printf("Failed to checkout branch: %v\n", err)
			os.Exit(1)
		}
	}
}
