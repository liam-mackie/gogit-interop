package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	pkg := analyzePackages()

	goOutputDir := filepath.Join("..", "shared")
	csOutputDir := filepath.Join("..", "dotnet", "GoGitDotNet")
	if len(os.Args) > 1 {
		goOutputDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		csOutputDir = os.Args[2]
	}

	fmt.Printf("Generating Go CGO wrappers to %s\n", goOutputDir)
	if err := generateGo(pkg, goOutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating Go: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating C# interop to %s\n", csOutputDir)
	if err := generateCSharp(pkg, csOutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating C#: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generation complete.")
}
