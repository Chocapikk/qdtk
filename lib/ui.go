package lib

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	// Colors
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
	white   = color.New(color.FgWhite, color.Bold).SprintFunc()
	dim     = color.New(color.Faint).SprintFunc()

	// Bold variants
	boldYellow = color.New(color.FgYellow, color.Bold).SprintFunc()
	boldGreen  = color.New(color.FgGreen, color.Bold).SprintFunc()
	boldCyan   = color.New(color.FgCyan, color.Bold).SprintFunc()
	boldRed    = color.New(color.FgRed, color.Bold).SprintFunc()
)

const banner = `
                ╭──────────────────────────────────────╮
                │              %s              │
                │    %s    │
                ╰──────────────────────────────────────╯
`

func PrintBanner() {
	title := boldCyan("Q D T K")
	subtitle := dim("Qdrant ToolKit v1.0.0")
	fmt.Printf(banner, title, subtitle)
}

func PrintSection(title string) {
	fmt.Printf("\n%s %s\n", boldYellow("▸"), white(title))
	fmt.Println(dim(strings.Repeat("─", 50)))
}

func PrintSuccess(msg string) {
	fmt.Printf("%s %s\n", boldGreen("✓"), msg)
}

func PrintError(msg string) {
	fmt.Printf("%s %s\n", boldRed("✗"), msg)
}

func PrintInfo(msg string) {
	fmt.Printf("%s %s\n", boldCyan("ℹ"), msg)
}

func PrintWarning(msg string) {
	fmt.Printf("%s %s\n", boldYellow("⚠"), msg)
}

func PrintItem(label string, value interface{}) {
	fmt.Printf("  %s %v\n", dim(label+":"), cyan(value))
}

func PrintCollectionHeader(name string, points int64) {
	status := green("●")
	if points == 0 {
		status = dim("○")
	}
	fmt.Printf("  %s %s %s\n", status, boldYellow(name), dim(fmt.Sprintf("(%d points)", points)))
}

func PrintTableHeader(columns ...string) {
	var header string
	for _, col := range columns {
		header += fmt.Sprintf("%-20s", white(col))
	}
	fmt.Println(header)
	fmt.Println(dim(strings.Repeat("─", len(columns)*20)))
}

func PrintTableRow(values ...interface{}) {
	var row string
	for i, val := range values {
		if i == 0 {
			row += fmt.Sprintf("%-20s", yellow(val))
		} else {
			row += fmt.Sprintf("%-20v", val)
		}
	}
	fmt.Println(row)
}

func PrintDivider() {
	fmt.Println(dim(strings.Repeat("─", 50)))
}

func PrintTotal(label string, value interface{}) {
	fmt.Printf("  %s %v\n", boldYellow(label+":"), boldCyan(value))
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
