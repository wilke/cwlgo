// Package cwlgo provides functionality for parsing and executing
// Common Workflow Language (CWL) CommandLineTool descriptions.
package cwlgo

import (
	"fmt"
	"os"
	"path/filepath"
)

// CommandLineTool represents a CWL CommandLineTool document
type CommandLineTool struct {
	// Required fields
	CWLVersion  string      `yaml:"cwlVersion" json:"cwlVersion"`
	Class       string      `yaml:"class" json:"class"`                                 // Must be "CommandLineTool"
	BaseCommand interface{} `yaml:"baseCommand,omitempty" json:"baseCommand,omitempty"` // String or []string

	// Optional fields
	Inputs             map[string]CommandInputParameter  `yaml:"inputs" json:"inputs"`
	Outputs            map[string]CommandOutputParameter `yaml:"outputs" json:"outputs"`
	ID                 string                            `yaml:"id,omitempty" json:"id,omitempty"`
	Requirements       []map[string]interface{}          `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	Hints              []Hint                            `yaml:"hints,omitempty" json:"hints,omitempty"`
	Label              string                            `yaml:"label,omitempty" json:"label,omitempty"`
	Doc                string                            `yaml:"doc,omitempty" json:"doc,omitempty"`
	Arguments          []CommandLineBinding              `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Stdin              string                            `yaml:"stdin,omitempty" json:"stdin,omitempty"`
	Stdout             string                            `yaml:"stdout,omitempty" json:"stdout,omitempty"`
	Stderr             string                            `yaml:"stderr,omitempty" json:"stderr,omitempty"`
	SuccessCodes       []int                             `yaml:"successCodes,omitempty" json:"successCodes,omitempty"`
	TemporaryFailCodes []int                             `yaml:"temporaryFailCodes,omitempty" json:"temporaryFailCodes,omitempty"`
	PermanentFailCodes []int                             `yaml:"permanentFailCodes,omitempty" json:"permanentFailCodes,omitempty"`
}

// CommandInputParameter represents an input parameter for a CommandLineTool
type CommandInputParameter struct {
	ID             string              `yaml:"id,omitempty" json:"id,omitempty"`
	Label          string              `yaml:"label,omitempty" json:"label,omitempty"`
	Doc            string              `yaml:"doc,omitempty" json:"doc,omitempty"`
	Type           interface{}         `yaml:"type" json:"type"` // Can be string, []string, or InputRecordSchema, etc.
	Default        interface{}         `yaml:"default,omitempty" json:"default,omitempty"`
	Format         interface{}         `yaml:"format,omitempty" json:"format,omitempty"` // Can be string or Expression
	Binding        *CommandLineBinding `yaml:"inputBinding,omitempty" json:"inputBinding,omitempty"`
	SecondaryFiles interface{}         `yaml:"secondaryFiles,omitempty" json:"secondaryFiles,omitempty"`
}

// CommandOutputParameter represents an output parameter for a CommandLineTool
type CommandOutputParameter struct {
	ID             string                `yaml:"id,omitempty" json:"id,omitempty"`
	Label          string                `yaml:"label,omitempty" json:"label,omitempty"`
	Doc            string                `yaml:"doc,omitempty" json:"doc,omitempty"`
	Type           interface{}           `yaml:"type" json:"type"`                         // Can be string, []string, or OutputRecordSchema, etc.
	Format         interface{}           `yaml:"format,omitempty" json:"format,omitempty"` // Can be string or Expression
	Binding        *CommandOutputBinding `yaml:"outputBinding,omitempty" json:"outputBinding,omitempty"`
	SecondaryFiles interface{}           `yaml:"secondaryFiles,omitempty" json:"secondaryFiles,omitempty"`
}

// CommandLineBinding represents how to construct a command line argument
type CommandLineBinding struct {
	Position      int         `yaml:"position,omitempty" json:"position,omitempty"`
	Prefix        string      `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Separate      *bool       `yaml:"separate,omitempty" json:"separate,omitempty"`
	ItemSeparator string      `yaml:"itemSeparator,omitempty" json:"itemSeparator,omitempty"`
	ValueFrom     interface{} `yaml:"valueFrom,omitempty" json:"valueFrom,omitempty"` // String or Expression
	ShellQuote    *bool       `yaml:"shellQuote,omitempty" json:"shellQuote,omitempty"`
}

// CommandOutputBinding represents how to capture output from a command
type CommandOutputBinding struct {
	Glob         interface{} `yaml:"glob,omitempty" json:"glob,omitempty"` // String, Expression, or []string
	LoadContents *bool       `yaml:"loadContents,omitempty" json:"loadContents,omitempty"`
	OutputEval   interface{} `yaml:"outputEval,omitempty" json:"outputEval,omitempty"` // Expression
}

// Requirement represents a requirement that must be fulfilled to execute the tool
type Requirement interface {
	IsRequirement() bool
}

// Hint represents an optional hint for executing the tool
type Hint interface {
	IsHint() bool
}

// DockerRequirement specifies a Docker container to use
type DockerRequirement struct {
	Class           string `yaml:"class" json:"class"` // Must be "DockerRequirement"
	DockerPull      string `yaml:"dockerPull,omitempty" json:"dockerPull,omitempty"`
	DockerLoad      string `yaml:"dockerLoad,omitempty" json:"dockerLoad,omitempty"`
	DockerFile      string `yaml:"dockerFile,omitempty" json:"dockerFile,omitempty"`
	DockerImport    string `yaml:"dockerImport,omitempty" json:"dockerImport,omitempty"`
	DockerImageID   string `yaml:"dockerImageId,omitempty" json:"dockerImageId,omitempty"`
	DockerOutputDir string `yaml:"dockerOutputDirectory,omitempty" json:"dockerOutputDirectory,omitempty"`
}

// IsRequirement implements the Requirement interface
func (d DockerRequirement) IsRequirement() bool {
	return true
}

// EnvVarRequirement specifies environment variables to set
type EnvVarRequirement struct {
	Class  string           `yaml:"class" json:"class"` // Must be "EnvVarRequirement"
	EnvDef []EnvironmentDef `yaml:"envDef" json:"envDef"`
}

// EnvironmentDef represents an environment variable definition
type EnvironmentDef struct {
	Name  string      `yaml:"name" json:"name"`
	Value interface{} `yaml:"value" json:"value"` // String or Expression
}

// IsRequirement implements the Requirement interface
func (e EnvVarRequirement) IsRequirement() bool {
	return true
}

// ResourceRequirement specifies computational resource requirements
type ResourceRequirement struct {
	Class     string      `yaml:"class" json:"class"`                             // Must be "ResourceRequirement"
	CoresMin  interface{} `yaml:"coresMin,omitempty" json:"coresMin,omitempty"`   // Long or Expression
	CoresMax  interface{} `yaml:"coresMax,omitempty" json:"coresMax,omitempty"`   // Long or Expression
	RAMMin    interface{} `yaml:"ramMin,omitempty" json:"ramMin,omitempty"`       // Long or Expression
	RAMMax    interface{} `yaml:"ramMax,omitempty" json:"ramMax,omitempty"`       // Long or Expression
	TMPDirMin interface{} `yaml:"tmpdirMin,omitempty" json:"tmpdirMin,omitempty"` // Long or Expression
	TMPDirMax interface{} `yaml:"tmpdirMax,omitempty" json:"tmpdirMax,omitempty"` // Long or Expression
	OutDirMin interface{} `yaml:"outdirMin,omitempty" json:"outdirMin,omitempty"` // Long or Expression
	OutDirMax interface{} `yaml:"outdirMax,omitempty" json:"outdirMax,omitempty"` // Long or Expression
}

// IsRequirement implements the Requirement interface
func (r ResourceRequirement) IsRequirement() bool {
	return true
}

// SingularityRequirement specifies a Singularity/Apptainer container to use
type SingularityRequirement struct {
	Class                string `yaml:"class" json:"class"` // Must be "SingularityRequirement"
	SingularityPull      string `yaml:"singularityPull,omitempty" json:"singularityPull,omitempty"`
	SingularityLoad      string `yaml:"singularityLoad,omitempty" json:"singularityLoad,omitempty"`
	SingularityFile      string `yaml:"singularityFile,omitempty" json:"singularityFile,omitempty"`
	SingularityImport    string `yaml:"singularityImport,omitempty" json:"singularityImport,omitempty"`
	SingularityImageID   string `yaml:"singularityImageId,omitempty" json:"singularityImageId,omitempty"`
	SingularityOutputDir string `yaml:"singularityOutputDirectory,omitempty" json:"singularityOutputDirectory,omitempty"`
}

// IsRequirement implements the Requirement interface
func (s SingularityRequirement) IsRequirement() bool {
	return true
}

// Error types
var (
	ErrInvalidCWL = fmt.Errorf("invalid CWL document")
	ErrExecution  = fmt.Errorf("command execution error")
)

// CWLError represents an error that occurred during CWL processing
type CWLError struct {
	Err     error
	Message string
}

// Error implements the error interface
func (e *CWLError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error
func (e *CWLError) Unwrap() error {
	return e.Err
}

// ContainerConfig holds configuration for container execution
type ContainerConfig struct {
	Type      string   // "docker" or "singularity"
	Image     string   // Image name or path
	Pull      bool     // Whether to pull the image
	Load      string   // Path to image file to load
	File      string   // Dockerfile or Singularity definition file
	Import    string   // Path to archive to import
	ImageID   string   // Explicit image ID
	OutputDir string   // Output directory inside container
	Volumes   []string // Additional volumes to mount
	WorkDir   string   // Working directory inside container
	EnvVars   []string // Environment variables to set in container
}

// ExecutionContext holds the context for executing a CommandLineTool
type ExecutionContext struct {
	WorkingDir      string
	TempDir         string
	Inputs          map[string]interface{}
	OutputDir       string
	EnvironmentVars map[string]string
	Container       *ContainerConfig // Container configuration if using containers
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(workingDir string) (*ExecutionContext, error) {
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, &CWLError{Err: err, Message: "failed to get current working directory"}
		}
	}

	tempDir, err := os.MkdirTemp("", "cwlgo-")
	if err != nil {
		return nil, &CWLError{Err: err, Message: "failed to create temporary directory"}
	}

	outputDir := filepath.Join(workingDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, &CWLError{Err: err, Message: "failed to create output directory"}
	}

	return &ExecutionContext{
		WorkingDir:      workingDir,
		TempDir:         tempDir,
		Inputs:          make(map[string]interface{}),
		OutputDir:       outputDir,
		EnvironmentVars: make(map[string]string),
		Container:       nil, // Will be set if container execution is required
	}, nil
}

// Cleanup cleans up temporary resources
func (ctx *ExecutionContext) Cleanup() error {
	if ctx.TempDir != "" {
		if err := os.RemoveAll(ctx.TempDir); err != nil {
			return &CWLError{Err: err, Message: "failed to clean up temporary directory"}
		}
	}
	return nil
}
