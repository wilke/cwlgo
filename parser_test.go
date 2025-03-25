package cwlgo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAML(t *testing.T) {
	// Create a temporary file with CWL content
	content := `
cwlVersion: v1.2
class: CommandLineTool
baseCommand: echo
id: echo-tool

inputs:
  message:
    type: string
    inputBinding:
      position: 1

outputs:
  output:
    type: stdout

stdout: output.txt
`
	tempDir, err := os.MkdirTemp("", "cwlgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.cwl")
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Create a parser
	parser := NewParser()

	// Parse the file
	tool, err := parser.ParseFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to parse CWL file: %v", err)
	}

	// Verify the parsed content
	if tool.CWLVersion != "v1.2" {
		t.Errorf("Expected CWLVersion v1.2, got %s", tool.CWLVersion)
	}

	if tool.Class != "CommandLineTool" {
		t.Errorf("Expected Class CommandLineTool, got %s", tool.Class)
	}

	if tool.ID != "echo-tool" {
		t.Errorf("Expected ID echo-tool, got %s", tool.ID)
	}

	baseCmd, ok := tool.BaseCommand.(string)
	if !ok {
		t.Fatalf("Expected BaseCommand to be a string")
	}
	if baseCmd != "echo" {
		t.Errorf("Expected BaseCommand echo, got %s", baseCmd)
	}

	if len(tool.Inputs) != 1 {
		t.Fatalf("Expected 1 input, got %d", len(tool.Inputs))
	}

	input, ok := tool.Inputs["message"]
	if !ok {
		t.Fatalf("Expected input 'message' not found")
	}

	inputType, ok := input.Type.(string)
	if !ok {
		t.Fatalf("Expected input type to be a string")
	}
	if inputType != "string" {
		t.Errorf("Expected input type string, got %s", inputType)
	}

	if len(tool.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(tool.Outputs))
	}

	output, ok := tool.Outputs["output"]
	if !ok {
		t.Fatalf("Expected output 'output' not found")
	}

	outputType, ok := output.Type.(string)
	if !ok {
		t.Fatalf("Expected output type to be a string")
	}
	if outputType != "stdout" {
		t.Errorf("Expected output type stdout, got %s", outputType)
	}

	if tool.Stdout != "output.txt" {
		t.Errorf("Expected stdout output.txt, got %s", tool.Stdout)
	}
}

func TestParseJSON(t *testing.T) {
	// Create a temporary file with CWL content in JSON
	content := `{
  "cwlVersion": "v1.2",
  "class": "CommandLineTool",
  "baseCommand": "echo",
  "id": "echo-tool",
  "inputs": {
    "message": {
      "type": "string",
      "inputBinding": {
        "position": 1
      }
    }
  },
  "outputs": {
    "output": {
      "type": "stdout"
    }
  },
  "stdout": "output.txt"
}`

	tempDir, err := os.MkdirTemp("", "cwlgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// Create a parser
	parser := NewParser()

	// Parse the file
	tool, err := parser.ParseFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to parse CWL file: %v", err)
	}

	// Verify the parsed content
	if tool.CWLVersion != "v1.2" {
		t.Errorf("Expected CWLVersion v1.2, got %s", tool.CWLVersion)
	}

	if tool.Class != "CommandLineTool" {
		t.Errorf("Expected Class CommandLineTool, got %s", tool.Class)
	}

	if tool.ID != "echo-tool" {
		t.Errorf("Expected ID echo-tool, got %s", tool.ID)
	}

	baseCmd, ok := tool.BaseCommand.(string)
	if !ok {
		t.Fatalf("Expected BaseCommand to be a string")
	}
	if baseCmd != "echo" {
		t.Errorf("Expected BaseCommand echo, got %s", baseCmd)
	}
}

func TestValidateCommandLineTool(t *testing.T) {
	// Test with invalid tool (missing required fields)
	parser := NewParser()

	// Missing cwlVersion
	tool := &CommandLineTool{
		Class:       "CommandLineTool",
		BaseCommand: "echo",
	}

	err := parser.validateCommandLineTool(tool)
	if err == nil {
		t.Error("Expected error for missing cwlVersion, got nil")
	}

	// Invalid class
	tool = &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "InvalidClass",
		BaseCommand: "echo",
	}

	err = parser.validateCommandLineTool(tool)
	if err == nil {
		t.Error("Expected error for invalid class, got nil")
	}

	// Missing baseCommand
	tool = &CommandLineTool{
		CWLVersion: "v1.2",
		Class:      "CommandLineTool",
	}

	err = parser.validateCommandLineTool(tool)
	if err == nil {
		t.Error("Expected error for missing baseCommand, got nil")
	}

	// Valid tool
	tool = &CommandLineTool{
		CWLVersion:  "v1.2",
		Class:       "CommandLineTool",
		BaseCommand: "echo",
	}

	err = parser.validateCommandLineTool(tool)
	if err != nil {
		t.Errorf("Expected no error for valid tool, got %v", err)
	}
}
