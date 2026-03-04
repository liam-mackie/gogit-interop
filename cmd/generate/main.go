package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	pkg := analyze()

	outputDir := filepath.Join("..", "shared")
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	fmt.Printf("Generating Go CGO wrappers to %s\n", outputDir)
	if err := generateGo(pkg, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating Go: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating C# interop to %s/dotnet/GoGit.Interop\n", outputDir)
	if err := generateCSharp(pkg, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating C#: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generation complete.")
}
