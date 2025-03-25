package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/user/cwlgo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run simple.go <cwl-file> [input-key=input-value ...]")
		os.Exit(1)
	}

	cwlFile := os.Args[1]

	// Parse inputs from command line
	inputs := make(map[string]interface{})
	for i := 2; i < len(os.Args); i++ {
		var key, value string
		arg := os.Args[i]

		// Split on first equals sign
		for j, c := range arg {
			if c == '=' {
				key = arg[:j]
				value = arg[j+1:]
				break
			}
		}

		if key != "" && value != "" {
			inputs[key] = value
		}
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

	// Create a new executor
	executor := cwlgo.NewExecutor()

	// Execute the tool
	fmt.Println("Executing tool...")
	result, err := executor.Execute(context.Background(), tool, inputs)
	if err != nil {
		log.Fatalf("Failed to execute tool: %v", err)
	}

	// Print results
	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Stdout: %s\n", result.Stdout)
	fmt.Printf("Stderr: %s\n", result.Stderr)

	// Print output files
	fmt.Println("Output files:")
	for id, path := range result.OutputFiles {
		fmt.Printf("  %s: %s\n", id, path)
	}
}
