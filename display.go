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

	actionInstall := lipgloss.NewStyle().
		Foreground(green).
		Padding(0, 1).
		Width(10)

	actionRemove := lipgloss.NewStyle().
		Foreground(red).
		Padding(0, 1).
		Width(10)

	packageStyle := lipgloss.NewStyle().
		Foreground(white).
		Padding(0, 1).
		Width(28)

	reasonStyle := lipgloss.NewStyle().
		Foreground(white).
		Padding(0, 1).
		Width(40)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("Action", "Package", "Reason").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			switch col {
			case 0: // Action column
				return lipgloss.NewStyle().Padding(0, 1).Width(10)
			case 1:
				return packageStyle
			case 2:
				return reasonStyle
			default:
				return lipgloss.NewStyle()
			}
		})

	// To Install
	for _, pkg := range diff.ToAdd {
		t.Row(
			actionInstall.Render("INSTALL"),
			pkg,
			"Missing from system",
		)
	}

	// Strict removals
	if strict {
		for _, pkg := range diff.ToRemove {
			t.Row(
				actionRemove.Render("REMOVE"),
				pkg,
				"Not in definitions (strict mode)",
			)
		}
	}

	// Ditto-managed removals
	for _, pkg := range diff.ToRemoveFromDitto {
		t.Row(
			actionRemove.Render("REMOVE"),
			pkg,
			"No longer managed by Ditto",
		)
	}

	return t
}

// displayWithOptionalPager renders output via pager (if configured), or directly to stdout.
func displayWithOptionalPager(appCtx *AppContext, out *bytes.Buffer) {
	pager := appCtx.Config.Pager

	if pager != nil && len(*pager) > 0 {
		pagerCommand := (*pager)[0]
		pagerArgs := (*pager)[1:]

		cmd := exec.Command(pagerCommand, pagerArgs...)
		cmd.Stdin = out
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			// Fallback to direct output
			fmt.Print(out.String())
		}
	} else {
		fmt.Print(out.String())
	}
}
