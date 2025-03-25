package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	cwlgo "../../" // Local import of the cwlgo package
)

func main() {
	// Check if a message was provided
	message := "Hello from Docker container!"
	if len(os.Args) > 1 {
		message = os.Args[1]
	}

	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a new parser
	parser := cwlgo.NewParser()

	// Parse the CWL file
	cwlFile := filepath.Join(dir, "docker_example.cwl")
	tool, err := parser.ParseFile(cwlFile)
	if err != nil {
		log.Fatalf("Failed to parse CWL file: %v", err)
	}

	// Create inputs
	inputs := map[string]interface{}{
		"message": message,
	}

	// Create a new executor
	executor := cwlgo.NewExecutor()

	// Make sure Docker is enabled
	executor.DockerEnabled = true

	fmt.Println("Executing tool in Docker container...")

	// Execute the tool
	result, err := executor.Execute(context.Background(), tool, inputs)
	if err != nil {
		log.Fatalf("Failed to execute tool: %v", err)
	}

	// Print results
	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Stdout: %s\n", result.Stdout)

	// Access output files
	for id, path := range result.OutputFiles {
		fmt.Printf("Output %s: %s\n", id, path)

		// Read and print the output file content
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read output file: %v", err)
		}
		fmt.Printf("Content: %s\n", content)
	}
}
