package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func buildDiffTable(diff PackageDiff, strict bool) *table.Table {
	white := lipgloss.Color("15")
	green := lipgloss.Color("10")
	red := lipgloss.Color("9")

	headerStyle := lipgloss.NewStyle().
		Foreground(white).
		Padding(0, 1).
		Bold(true)

	installStyle := lipgloss.NewStyle().
		Foreground(green).
		Padding(0, 1).
		Width(24)

	removeStyle := lipgloss.NewStyle().
		Foreground(red).
		Padding(0, 1).
		Width(24)

	t := table.New().
		Border(lipgloss.MarkdownBorder()).
		BorderTop(false).
		BorderBottom(false).
		Headers("To Install", "To Remove").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col == 0 {
				return installStyle
			}
			return removeStyle
		})

	maxLen := len(diff.ToAdd)
	if strict && len(diff.ToRemove) > maxLen {
		maxLen = len(diff.ToRemove)
	}

	for i := 0; i < maxLen; i++ {
		var addPkg, removePkg string
		if i < len(diff.ToAdd) {
			addPkg = diff.ToAdd[i]
		}
		if strict && i < len(diff.ToRemove) {
			removePkg = diff.ToRemove[i]
		}
		t.Row(addPkg, removePkg)
	}

	return t
}

func displayWithOptionalPager(out *bytes.Buffer) {
	if cfg.Pager != nil && len(*cfg.Pager) > 0 {
		pagerCommand := (*cfg.Pager)[0]
		pagerArgs := (*cfg.Pager)[1:]

		cmd := exec.Command(pagerCommand, pagerArgs...)
		cmd.Stdin = out
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Print(out.String())
		}
	} else {
		fmt.Print(out.String())
	}
}
