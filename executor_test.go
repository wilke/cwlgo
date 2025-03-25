package cwlgo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	// Create a simple CommandLineTool for testing
	tool := &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "CommandLineTool",
		BaseCommand: "echo",
		Inputs: map[string]CommandInputParameter{
			"message": {
				Type: "string",
				Binding: &CommandLineBinding{
					Position: 1,
				},
			},
		},
		Outputs: map[string]CommandOutputParameter{
			"output": {
				Type: "stdout",
			},
		},
		Stdout: "output.txt",
	}

	// Create inputs
	inputs := map[string]interface{}{
		"message": "Hello, CWL!",
	}

	// Create an executor
	executor := NewExecutor()

	// Execute the tool
	result, err := executor.Execute(context.Background(), tool, inputs)
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}

	// Verify the results
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "Hello, CWL!") {
		t.Errorf("Expected stdout to contain 'Hello, CWL!', got %s", result.Stdout)
	}
}

func TestBuildCommandLine(t *testing.T) {
	// Create a CommandLineTool with various input bindings
	tool := &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "CommandLineTool",
		BaseCommand: "grep",
		Inputs: map[string]CommandInputParameter{
			"pattern": {
				Type: "string",
				Binding: &CommandLineBinding{
					Position: 1,
				},
			},
			"file": {
				Type: "File",
				Binding: &CommandLineBinding{
					Position: 2,
				},
			},
			"invert": {
				Type: "boolean",
				Binding: &CommandLineBinding{
					Position: 0,
					Prefix:   "-v",
				},
			},
		},
		Arguments: []CommandLineBinding{
			{
				Position:  0,
				ValueFrom: "-n",
			},
		},
	}

	// Create inputs
	inputs := map[string]interface{}{
		"pattern": "test",
		"file": map[string]interface{}{
			"class": "File",
			"path":  "file.txt",
		},
		"invert": true,
	}

	// Create an executor
	executor := NewExecutor()

	// Create execution context
	execCtx, err := NewExecutionContext("")
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer execCtx.Cleanup()

	// Set inputs
	execCtx.Inputs = inputs

	// Build command line
	cmdArgs, err := executor.BuildCommandLine(tool, execCtx)
	if err != nil {
		t.Fatalf("Failed to build command line: %v", err)
	}

	// Verify the command line arguments
	expectedArgs := []string{"grep", "-n", "-v", "test", "file.txt"}
	if len(cmdArgs) != len(expectedArgs) {
		t.Fatalf("Expected %d command line arguments, got %d: %v", len(expectedArgs), len(cmdArgs), cmdArgs)
	}

	for i, arg := range expectedArgs {
		if cmdArgs[i] != arg {
			t.Errorf("Expected argument %d to be %s, got %s", i, arg, cmdArgs[i])
		}
	}
}

func TestProcessRequirements(t *testing.T) {
	// Create a CommandLineTool with requirements
	tool := &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "CommandLineTool",
		BaseCommand: "echo",
		Requirements: []map[string]interface{}{
			{
				"class": "EnvVarRequirement",
				"envDef": []interface{}{
					map[string]interface{}{
						"name":  "TEST_ENV",
						"value": "test_value",
					},
				},
			},
			{
				"class":    "ResourceRequirement",
				"coresMin": float64(2),
				"ramMin":   float64(1024),
			},
		},
	}

	// Create an executor
	executor := NewExecutor()

	// Create execution context
	execCtx, err := NewExecutionContext("")
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer execCtx.Cleanup()

	// Process requirements
	err = executor.processRequirements(tool, execCtx)
	if err != nil {
		t.Fatalf("Failed to process requirements: %v", err)
	}

	// Verify environment variables
	if execCtx.EnvironmentVars["TEST_ENV"] != "test_value" {
		t.Errorf("Expected environment variable TEST_ENV=test_value, got %s", execCtx.EnvironmentVars["TEST_ENV"])
	}

	// Test with excessive resource requirements
	executor.MaxCores = 1 // Set lower than required
	err = executor.processRequirements(tool, execCtx)
	if err == nil {
		t.Error("Expected error for excessive resource requirements, got nil")
	}
}

func TestProcessOutputs(t *testing.T) {
	// Create a temporary directory for outputs
	tempDir, err := os.MkdirTemp("", "cwlgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test output file
	outputFile := filepath.Join(tempDir, "output.txt")
	if err := os.WriteFile(outputFile, []byte("test output"), 0644); err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	// Create a CommandLineTool with outputs
	tool := &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "CommandLineTool",
		BaseCommand: "echo",
		Outputs: map[string]CommandOutputParameter{
			"output": {
				Type: "File",
				Binding: &CommandOutputBinding{
					Glob: "output.txt",
				},
			},
		},
	}

	// Create an executor
	executor := NewExecutor()

	// Create execution context
	execCtx, err := NewExecutionContext("")
	if err != nil {
		t.Fatalf("Failed to create execution context: %v", err)
	}
	defer execCtx.Cleanup()

	// Override output directory for testing
	execCtx.OutputDir = tempDir

	// Create a dummy execution result
	result := &ExecuteResult{
		ExitCode: 0,
		Stdout:   "test output",
	}

	// Process outputs
	outputFiles, err := executor.processOutputs(tool, execCtx, result)
	if err != nil {
		t.Fatalf("Failed to process outputs: %v", err)
	}

	// Verify output files
	if len(outputFiles) != 1 {
		t.Fatalf("Expected 1 output file, got %d", len(outputFiles))
	}

	if outputPath, ok := outputFiles["output"]; !ok {
		t.Error("Expected output 'output' not found")
	} else if outputPath != outputFile {
		t.Errorf("Expected output path %s, got %s", outputFile, outputPath)
	}
}
