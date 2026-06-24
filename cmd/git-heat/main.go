// git-heat is a CLI tool that visualizes code ownership and contribution
// patterns directly in the terminal.
//
// Usage:
//
//	git-heat files              # heatmap of all files
//	git-heat dir [path]         # directory-level heatmap
//	git-heat author [name]      # files owned by author
//	git-heat summary            # quick overview
//
// See git-heat --help for full usage.
package main

import (
	"github.com/Shashank0701-byte/git-heat/internal/cmd"
)

func main() {
	cmd.Execute()
}
