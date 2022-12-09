package main

import (
	"log"
	"os"
	"path"

	"github.com/kameshsampath/harness-cli/pkg/commands"
	"github.com/spf13/cobra/doc"
)

func main() {
	cwd, _ := os.Getwd()
	rootCmd := commands.NewRootCommand()
	err := doc.GenMarkdownTree(rootCmd, path.Join(cwd, "docs"))
	if err != nil {
		log.Fatal(err)
	}
}
