package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gmherb/envtab/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Generate Markdown documentation for all commands under ./docs
	outputDir := "./docs"

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("failed to create docs directory %s: %v", outputDir, err)
	}

	root := cmd.GetRootCmd()

	if err := doc.GenMarkdownTree(root, outputDir); err != nil {
		log.Fatalf("failed to generate markdown docs: %v", err)
	}

	// Also generate a single top-level README-style command doc if desired
	readmePath := filepath.Join(outputDir, "envtab.md")
	f, err := os.Create(readmePath)
	if err != nil {
		log.Fatalf("failed to create top-level envtab doc: %v", err)
	}
	defer f.Close()

	if err := doc.GenMarkdown(root, f); err != nil {
		log.Fatalf("failed to generate top-level envtab markdown: %v", err)
	}
}


