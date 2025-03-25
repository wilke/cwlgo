package cwlgo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// Executor handles execution of CommandLineTools
type Executor struct {
	// Configuration options for the executor
	DockerEnabled bool
	MaxCores      int
	MaxRAM        int64 // in MiB
	// Add more configuration options as needed
}

// NewExecutor creates a new executor with default settings
func NewExecutor() *Executor {
	return &Executor{
		DockerEnabled: true,
		MaxCores:      4,
		MaxRAM:        8192, // 8 GiB
	}
}

// ExecuteResult contains the results of executing a CommandLineTool
type ExecuteResult struct {
	ExitCode    int
	Stdout      string
	Stderr      string
	OutputFiles map[string]string // Output ID -> File path
}

// Execute executes a CommandLineTool with the given inputs
func (e *Executor) Execute(ctx context.Context, tool *CommandLineTool, inputs map[string]interface{}) (*ExecuteResult, error) {
	// Create execution context
	execCtx, err := NewExecutionContext("")
	if err != nil {
		return nil, err
	}
	defer execCtx.Cleanup()

	// Set inputs
	execCtx.Inputs = inputs

	// Process requirements
	if err := e.processRequirements(tool, execCtx); err != nil {
		return nil, err
	}

	// Build command line
	cmdArgs, err := e.BuildCommandLine(tool, execCtx)
	if err != nil {
		return nil, err
	}

	// Execute command
	result, err := e.runCommand(ctx, tool, cmdArgs, execCtx)
	if err != nil {
		return nil, err
	}

	// Process outputs
	outputFiles, err := e.processOutputs(tool, execCtx, result)
	if err != nil {
		return nil, err
	}

	result.OutputFiles = outputFiles
	return result, nil
}

// processRequirements processes the requirements of a CommandLineTool
func (e *Executor) processRequirements(tool *CommandLineTool, ctx *ExecutionContext) error {
	for _, reqMap := range tool.Requirements {
		// Get the class of the requirement
		class, ok := reqMap["class"].(string)
		if !ok {
			return &CWLError{
				Err:     ErrExecution,
				Message: "requirement must have a 'class' field",
			}
		}

		switch class {
		case "DockerRequirement":
			if !e.DockerEnabled {
				return &CWLError{
					Err:     ErrExecution,
					Message: "Docker is required but not enabled",
				}
			}
			// Process Docker requirement
			// This would involve setting up Docker container execution
			// For now, we'll just log it
			dockerPull, _ := reqMap["dockerPull"].(string)
			fmt.Printf("Docker requirement: %s\n", dockerPull)

		case "EnvVarRequirement":
			// Process environment variables
			envDefList, ok := reqMap["envDef"].([]interface{})
			if !ok {
				return &CWLError{
					Err:     ErrExecution,
					Message: "EnvVarRequirement must have an 'envDef' field",
				}
			}

			for _, envDefInterface := range envDefList {
				envDefMap, ok := envDefInterface.(map[string]interface{})
				if !ok {
					return &CWLError{
						Err:     ErrExecution,
						Message: "envDef items must be objects",
					}
				}

				name, ok := envDefMap["name"].(string)
				if !ok {
					return &CWLError{
						Err:     ErrExecution,
						Message: "envDef items must have a 'name' field",
					}
				}

				value, ok := envDefMap["value"]
				if !ok {
					return &CWLError{
						Err:     ErrExecution,
						Message: "envDef items must have a 'value' field",
					}
				}

				// For simplicity, we'll only handle string values for now
				if strVal, ok := value.(string); ok {
					ctx.EnvironmentVars[name] = strVal
				} else {
					// In a real implementation, we would evaluate expressions here
					return &CWLError{
						Err:     ErrExecution,
						Message: fmt.Sprintf("unsupported environment variable value type for %s", name),
					}
				}
			}

		case "ResourceRequirement":
			// Process resource requirements
			// For now, we'll just check if they're within our limits
			if coresMin, ok := reqMap["coresMin"].(float64); ok {
				if int(coresMin) > e.MaxCores {
					return &CWLError{
						Err:     ErrExecution,
						Message: fmt.Sprintf("required cores (%f) exceeds maximum (%d)", coresMin, e.MaxCores),
					}
				}
			}

			if ramMin, ok := reqMap["ramMin"].(float64); ok {
				if int64(ramMin) > e.MaxRAM {
					return &CWLError{
						Err:     ErrExecution,
						Message: fmt.Sprintf("required RAM (%f MiB) exceeds maximum (%d MiB)", ramMin, e.MaxRAM),
					}
				}
			}

		default:
			// Unknown requirement type
			return &CWLError{
				Err:     ErrExecution,
				Message: fmt.Sprintf("unsupported requirement class: %s", class),
			}
		}
	}

	return nil
}

// CommandArg represents a command line argument with its position
type CommandArg struct {
	Position int
	Args     []string
}

// BuildCommandLine builds the command line arguments for a CommandLineTool
func (e *Executor) BuildCommandLine(tool *CommandLineTool, ctx *ExecutionContext) ([]string, error) {
	var cmdArgs []string
	var posArgs []CommandArg

	// Add base command
	switch cmd := tool.BaseCommand.(type) {
	case string:
		cmdArgs = append(cmdArgs, cmd)
	case []interface{}:
		for _, c := range cmd {
			if strCmd, ok := c.(string); ok {
				cmdArgs = append(cmdArgs, strCmd)
			} else {
				return nil, &CWLError{
					Err:     ErrExecution,
					Message: fmt.Sprintf("invalid base command element type: %T", c),
				}
			}
		}
	default:
		return nil, &CWLError{
			Err:     ErrExecution,
			Message: fmt.Sprintf("invalid base command type: %T", cmd),
		}
	}

	// Collect arguments with positions
	for _, arg := range tool.Arguments {
		var argStrings []string

		// Handle arguments with valueFrom
		if valueFrom, ok := arg.ValueFrom.(string); ok {
			if arg.Prefix != "" {
				// Handle separate flag
				separate := true
				if arg.Separate != nil {
					separate = *arg.Separate
				}

				if separate {
					argStrings = append(argStrings, arg.Prefix, valueFrom)
				} else {
					argStrings = append(argStrings, arg.Prefix+valueFrom)
				}
			} else {
				argStrings = append(argStrings, valueFrom)
			}
		} else if arg.Prefix != "" {
			// Handle arguments with just a prefix (flags)
			argStrings = append(argStrings, arg.Prefix)
		} else {
			// In a real implementation, we would evaluate expressions here
			return nil, &CWLError{
				Err:     ErrExecution,
				Message: "unsupported argument value type",
			}
		}

		posArgs = append(posArgs, CommandArg{
			Position: arg.Position,
			Args:     argStrings,
		})
	}

	// Collect input bindings with positions
	for inputID, inputParam := range tool.Inputs {
		if inputParam.Binding == nil {
			continue
		}

		inputValue, ok := ctx.Inputs[inputID]
		if !ok {
			// Check if there's a default value
			if inputParam.Default != nil {
				inputValue = inputParam.Default
			} else {
				return nil, &CWLError{
					Err:     ErrExecution,
					Message: fmt.Sprintf("missing required input: %s", inputID),
				}
			}
		}

		// Get the value to use for the command line
		var cmdValue string
		var skipArg bool

		// Process the input binding based on type
		switch v := inputValue.(type) {
		case string:
			// Handle string values
			cmdValue = v
		case bool:
			// Handle boolean values
			if !v {
				// Skip the argument if false
				skipArg = true
			} else {
				// For true boolean values, we don't need a value
				cmdValue = ""
			}
		case float64:
			// Handle numeric values
			cmdValue = fmt.Sprintf("%g", v)
		case int:
			// Handle integer values
			cmdValue = fmt.Sprintf("%d", v)
		case map[string]interface{}:
			// Handle complex types like File
			if class, ok := v["class"].(string); ok && class == "File" {
				if path, ok := v["path"].(string); ok {
					cmdValue = path
				} else {
					return nil, &CWLError{
						Err:     ErrExecution,
						Message: fmt.Sprintf("File input %s missing path", inputID),
					}
				}
			} else {
				return nil, &CWLError{
					Err:     ErrExecution,
					Message: fmt.Sprintf("unsupported input value type for %s: %T", inputID, inputValue),
				}
			}
		default:
			// In a real implementation, we would handle more types and evaluate expressions
			return nil, &CWLError{
				Err:     ErrExecution,
				Message: fmt.Sprintf("unsupported input value type for %s: %T", inputID, inputValue),
			}
		}

		// Skip this argument if needed (e.g., for false boolean values)
		if skipArg {
			continue
		}

		// Add the argument to the command line
		if inputParam.Binding.Prefix != "" {
			// Handle separate flag
			separate := true
			if inputParam.Binding.Separate != nil {
				separate = *inputParam.Binding.Separate
			}

			// Special handling for boolean values
			boolVal, isBool := inputValue.(bool)
			if cmdValue == "" && isBool && boolVal {
				// For true boolean values, just add the flag
				posArgs = append(posArgs, CommandArg{
					Position: inputParam.Binding.Position,
					Args:     []string{inputParam.Binding.Prefix},
				})
			} else if separate {
				posArgs = append(posArgs, CommandArg{
					Position: inputParam.Binding.Position,
					Args:     []string{inputParam.Binding.Prefix, cmdValue},
				})
			} else {
				posArgs = append(posArgs, CommandArg{
					Position: inputParam.Binding.Position,
					Args:     []string{inputParam.Binding.Prefix + cmdValue},
				})
			}
		} else {
			posArgs = append(posArgs, CommandArg{
				Position: inputParam.Binding.Position,
				Args:     []string{cmdValue},
			})
		}

	}

	// Sort arguments by position
	sort.Slice(posArgs, func(i, j int) bool {
		return posArgs[i].Position < posArgs[j].Position
	})

	// Append sorted arguments to command line
	for _, arg := range posArgs {
		cmdArgs = append(cmdArgs, arg.Args...)
	}

	return cmdArgs, nil
}

// runCommand executes the command with the given arguments
func (e *Executor) runCommand(ctx context.Context, tool *CommandLineTool, cmdArgs []string, execCtx *ExecutionContext) (*ExecuteResult, error) {
	if len(cmdArgs) == 0 {
		return nil, &CWLError{
			Err:     ErrExecution,
			Message: "empty command",
		}
	}

	// Create command
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = execCtx.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for name, value := range execCtx.EnvironmentVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
	}

	// Set up stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Handle stdin if specified
	if tool.Stdin != "" {
		// In a real implementation, we would evaluate expressions here
		stdinFile, err := os.Open(tool.Stdin)
		if err != nil {
			return nil, &CWLError{
				Err:     err,
				Message: fmt.Sprintf("failed to open stdin file: %s", tool.Stdin),
			}
		}
		defer stdinFile.Close()
		cmd.Stdin = stdinFile
	}

	// Handle stdout if specified
	if tool.Stdout != "" {
		stdoutPath := filepath.Join(execCtx.OutputDir, tool.Stdout)
		stdoutFile, err := os.Create(stdoutPath)
		if err != nil {
			return nil, &CWLError{
				Err:     err,
				Message: fmt.Sprintf("failed to create stdout file: %s", stdoutPath),
			}
		}
		defer stdoutFile.Close()

		// Use MultiWriter to capture stdout both in memory and in file
		cmd.Stdout = io.MultiWriter(&stdout, stdoutFile)
	}

	// Handle stderr if specified
	if tool.Stderr != "" {
		stderrPath := filepath.Join(execCtx.OutputDir, tool.Stderr)
		stderrFile, err := os.Create(stderrPath)
		if err != nil {
			return nil, &CWLError{
				Err:     err,
				Message: fmt.Sprintf("failed to create stderr file: %s", stderrPath),
			}
		}
		defer stderrFile.Close()

		// Use MultiWriter to capture stderr both in memory and in file
		cmd.Stderr = io.MultiWriter(&stderr, stderrFile)
	}

	// Run the command
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		// Check if it's an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()

			// Check if the exit code is in the success codes
			isSuccess := false
			for _, code := range tool.SuccessCodes {
				if exitCode == code {
					isSuccess = true
					break
				}
			}

			if isSuccess {
				// This is a successful exit code
				err = nil
			}
		} else {
			return nil, &CWLError{
				Err:     err,
				Message: "command execution failed",
			}
		}
	}

	return &ExecuteResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, err
}

// processOutputs processes the outputs of a CommandLineTool
func (e *Executor) processOutputs(tool *CommandLineTool, ctx *ExecutionContext, result *ExecuteResult) (map[string]string, error) {
	outputFiles := make(map[string]string)

	for outputID, outputParam := range tool.Outputs {
		if outputParam.Binding == nil {
			continue
		}

		// Process the output binding
		// For simplicity, we'll only handle glob patterns for now
		if glob, ok := outputParam.Binding.Glob.(string); ok {
			// Expand the glob pattern
			pattern := filepath.Join(ctx.OutputDir, glob)
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, &CWLError{
					Err:     err,
					Message: fmt.Sprintf("failed to expand glob pattern: %s", pattern),
				}
			}

			if len(matches) > 0 {
				// For simplicity, we'll just use the first match
				outputFiles[outputID] = matches[0]
			}
		} else if globList, ok := outputParam.Binding.Glob.([]interface{}); ok {
			// Handle list of glob patterns
			for _, g := range globList {
				if globStr, ok := g.(string); ok {
					pattern := filepath.Join(ctx.OutputDir, globStr)
					matches, err := filepath.Glob(pattern)
					if err != nil {
						return nil, &CWLError{
							Err:     err,
							Message: fmt.Sprintf("failed to expand glob pattern: %s", pattern),
						}
					}

					if len(matches) > 0 {
						// For simplicity, we'll just use the first match
						outputFiles[outputID] = matches[0]
						break
					}
				}
			}
		} else {
			// In a real implementation, we would evaluate expressions here
			return nil, &CWLError{
				Err:     ErrExecution,
				Message: fmt.Sprintf("unsupported glob pattern type for output %s", outputID),
			}
		}
	}

	return outputFiles, nil
}
