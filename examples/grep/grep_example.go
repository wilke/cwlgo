package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/user/cwlgo"
)

func main() {
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Path to the CWL file
	cwlFile := filepath.Join(currentDir, "grep.cwl")

	// Path to the sample file
	sampleFile := filepath.Join(currentDir, "sample.txt")

	// Check if the files exist
	if _, err := os.Stat(cwlFile); os.IsNotExist(err) {
		log.Fatalf("CWL file not found: %s", cwlFile)
	}
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		log.Fatalf("Sample file not found: %s", sampleFile)
	}

	// Create a new parser
	parser := cwlgo.NewParser()

	// Parse the CWL file
	tool, err := parser.ParseFile(cwlFile)
	if err != nil {
		log.Fatalf("Failed to parse CWL file: %v", err)
	}

	fmt.Printf("Parsed CWL tool: %s\n", tool.ID)
	fmt.Printf("Base command: %v\n", tool.BaseCommand)

	// Print inputs
	fmt.Println("Inputs:")
	for id, input := range tool.Inputs {
		fmt.Printf("  %s: %v\n", id, input.Type)
	}

	// Print outputs
	fmt.Println("Outputs:")
	for id, output := range tool.Outputs {
		fmt.Printf("  %s: %v\n", id, output.Type)
	}

	// Create inputs
	inputs := map[string]interface{}{
		"pattern": "sample",
		"file": map[string]interface{}{
			"class": "File",
			"path":  sampleFile,
		},
		"invert": false,
	}

	// Create a new executor
	executor := cwlgo.NewExecutor()

	// Execute the tool
	fmt.Println("\nExecuting grep tool...")

	result, err := executor.Execute(context.Background(), tool, inputs)
	if err != nil {
		log.Fatalf("Failed to execute tool: %v", err)
	}

	// Print results
	fmt.Printf("\nExit code: %d\n", result.ExitCode)
	fmt.Println("\nMatches:")
	fmt.Println(result.Stdout)

	// Print output files
	fmt.Println("Output files:")
	for id, path := range result.OutputFiles {
		fmt.Printf("  %s: %s\n", id, path)
	}

	// If we have a count output, print it
	if count, ok := result.OutputFiles["count"]; ok {
		fmt.Printf("\nNumber of matching lines: %s\n", count)
	}

	// Try with inverted search
	fmt.Println("\n\nExecuting grep tool with inverted search...")
	inputs["invert"] = true
	result, err = executor.Execute(context.Background(), tool, inputs)
	if err != nil {
		log.Fatalf("Failed to execute tool: %v", err)
	}

	// Print results
	fmt.Printf("\nExit code: %d\n", result.ExitCode)
	fmt.Println("\nNon-matching lines:")
	fmt.Println(result.Stdout)
}
