package cwlgo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parser handles parsing of CWL files
type Parser struct {
	// Configuration options for the parser
	StrictValidation bool
	// Add more configuration options as needed
}

// NewParser creates a new CWL parser with default settings
func NewParser() *Parser {
	return &Parser{
		StrictValidation: true,
	}
}

// ParseFile parses a CWL file and returns a CommandLineTool
func (p *Parser) ParseFile(filePath string) (*CommandLineTool, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, &CWLError{
			Err:     err,
			Message: fmt.Sprintf("failed to open CWL file: %s", filePath),
		}
	}
	defer file.Close()

	// Determine the file format based on extension
	var tool *CommandLineTool
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		tool, err = p.parseJSON(file)
	case ".yaml", ".yml":
		tool, err = p.parseYAML(file)
	default:
		// Try to parse as YAML first, then JSON if that fails
		tool, err = p.parseYAML(file)
		if err != nil {
			// Reset file position to beginning
			if _, seekErr := file.Seek(0, 0); seekErr != nil {
				return nil, &CWLError{
					Err:     seekErr,
					Message: "failed to reset file position",
				}
			}
			tool, err = p.parseJSON(file)
		}
	}

	if err != nil {
		return nil, err
	}

	// Validate the parsed tool
	if err := p.validateCommandLineTool(tool); err != nil {
		return nil, err
	}

	return tool, nil
}

// parseYAML parses a CWL document from YAML format
func (p *Parser) parseYAML(r io.Reader) (*CommandLineTool, error) {
	var tool CommandLineTool
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&tool); err != nil {
		return nil, &CWLError{
			Err:     err,
			Message: "failed to parse YAML",
		}
	}
	return &tool, nil
}

// parseJSON parses a CWL document from JSON format
func (p *Parser) parseJSON(r io.Reader) (*CommandLineTool, error) {
	var tool CommandLineTool
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&tool); err != nil {
		return nil, &CWLError{
			Err:     err,
			Message: "failed to parse JSON",
		}
	}
	return &tool, nil
}

// validateCommandLineTool validates a parsed CommandLineTool
func (p *Parser) validateCommandLineTool(tool *CommandLineTool) error {
	if tool == nil {
		return &CWLError{
			Err:     ErrInvalidCWL,
			Message: "tool is nil",
		}
	}

	// Check required fields
	if tool.CWLVersion == "" {
		return &CWLError{
			Err:     ErrInvalidCWL,
			Message: "cwlVersion is required",
		}
	}

	if tool.Class != "CommandLineTool" {
		return &CWLError{
			Err:     ErrInvalidCWL,
			Message: "class must be 'CommandLineTool'",
		}
	}

	// Validate baseCommand
	if tool.BaseCommand == nil {
		return &CWLError{
			Err:     ErrInvalidCWL,
			Message: "baseCommand is required",
		}
	}

	// Check if baseCommand is a string or []string
	switch cmd := tool.BaseCommand.(type) {
	case string:
		// Valid: baseCommand is a string
	case []interface{}:
		// Valid: baseCommand is an array
		for i, c := range cmd {
			if _, ok := c.(string); !ok {
				return &CWLError{
					Err:     ErrInvalidCWL,
					Message: fmt.Sprintf("baseCommand[%d] must be a string", i),
				}
			}
		}
	default:
		return &CWLError{
			Err:     ErrInvalidCWL,
			Message: "baseCommand must be a string or array of strings",
		}
	}

	// Additional validation can be added here

	return nil
}

// ParseRequirement parses a requirement from a map
func ParseRequirement(reqMap map[string]interface{}) (Requirement, error) {
	class, ok := reqMap["class"].(string)
	if !ok {
		return nil, &CWLError{
			Err:     ErrInvalidCWL,
			Message: "requirement must have a 'class' field",
		}
	}

	switch class {
	case "DockerRequirement":
		var req DockerRequirement
		req.Class = class

		if pull, ok := reqMap["dockerPull"].(string); ok {
			req.DockerPull = pull
		}
		if load, ok := reqMap["dockerLoad"].(string); ok {
			req.DockerLoad = load
		}
		if file, ok := reqMap["dockerFile"].(string); ok {
			req.DockerFile = file
		}
		if imp, ok := reqMap["dockerImport"].(string); ok {
			req.DockerImport = imp
		}
		if id, ok := reqMap["dockerImageId"].(string); ok {
			req.DockerImageID = id
		}
		if outDir, ok := reqMap["dockerOutputDirectory"].(string); ok {
			req.DockerOutputDir = outDir
		}

		return req, nil

	case "EnvVarRequirement":
		var req EnvVarRequirement
		req.Class = class

		envDefs, ok := reqMap["envDef"].([]interface{})
		if !ok {
			return nil, &CWLError{
				Err:     ErrInvalidCWL,
				Message: "EnvVarRequirement must have an 'envDef' field",
			}
		}

		for _, envDefInterface := range envDefs {
			envDefMap, ok := envDefInterface.(map[string]interface{})
			if !ok {
				return nil, &CWLError{
					Err:     ErrInvalidCWL,
					Message: "envDef items must be objects",
				}
			}

			name, ok := envDefMap["name"].(string)
			if !ok {
				return nil, &CWLError{
					Err:     ErrInvalidCWL,
					Message: "envDef items must have a 'name' field",
				}
			}

			value, ok := envDefMap["value"]
			if !ok {
				return nil, &CWLError{
					Err:     ErrInvalidCWL,
					Message: "envDef items must have a 'value' field",
				}
			}

			req.EnvDef = append(req.EnvDef, EnvironmentDef{
				Name:  name,
				Value: value,
			})
		}

		return req, nil

	case "ResourceRequirement":
		var req ResourceRequirement
		req.Class = class

		if coresMin, ok := reqMap["coresMin"]; ok {
			req.CoresMin = coresMin
		}
		if coresMax, ok := reqMap["coresMax"]; ok {
			req.CoresMax = coresMax
		}
		if ramMin, ok := reqMap["ramMin"]; ok {
			req.RAMMin = ramMin
		}
		if ramMax, ok := reqMap["ramMax"]; ok {
			req.RAMMax = ramMax
		}
		if tmpdirMin, ok := reqMap["tmpdirMin"]; ok {
			req.TMPDirMin = tmpdirMin
		}
		if tmpdirMax, ok := reqMap["tmpdirMax"]; ok {
			req.TMPDirMax = tmpdirMax
		}
		if outdirMin, ok := reqMap["outdirMin"]; ok {
			req.OutDirMin = outdirMin
		}
		if outdirMax, ok := reqMap["outdirMax"]; ok {
			req.OutDirMax = outdirMax
		}

		return req, nil

	default:
		// For unknown requirements, we could return a generic requirement
		// or an error depending on the strictness of validation
		return nil, &CWLError{
			Err:     ErrInvalidCWL,
			Message: fmt.Sprintf("unsupported requirement class: %s", class),
		}
	}
}
