# CWLGo

CWLGo is a Go library for parsing and executing Common Workflow Language (CWL) CommandLineTool descriptions.

## Features

- Parse CWL CommandLineTool descriptions from YAML or JSON files
- Execute command-line tools with the specified inputs
- Handle input and output bindings
- Support for Docker containers
- Support for Singularity/Apptainer containers
- Support for environment variables and resource requirements

## Installation

```bash
go get github.com/wilke/cwlgo
```

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/user/cwlgo"
)

func main() {
	// Create a new parser
	parser := cwlgo.NewParser()
	
	// Parse a CWL file
	tool, err := parser.ParseFile("path/to/tool.cwl")
	if err != nil {
		log.Fatalf("Failed to parse CWL file: %v", err)
	}
	
	// Create inputs
	inputs := map[string]interface{}{
		"input1": "value1",
		"input2": "value2",
	}
	
	// Create a new executor
	executor := cwlgo.NewExecutor()
	
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
	}
}
```

### Running the Examples

#### Echo Example
```bash
cd examples/echo
go run simple.go echo.cwl message="Hello, CWL!"
```

#### Grep Example
```bash
cd examples/grep
go run grep_example.go
```

You can also use the Makefile:
```bash
# Run the echo example
make echo-example

# Run the grep example
make grep-example
```

## Supported CWL Features

- CommandLineTool class
- Basic input and output bindings
- Environment variables
- Resource requirements
- Docker containers
- Singularity/Apptainer containers

## Container Support

### Docker

CWLGo supports executing tools inside Docker containers when specified in the CWL file using the DockerRequirement:

```yaml
requirements:
  DockerRequirement:
    dockerPull: "ubuntu:20.04"
```

The executor will automatically:
- Mount the working directory and output directory
- Map the command line arguments
- Handle environment variables
- Clean up containers after execution

### Singularity/Apptainer

CWLGo also supports Singularity/Apptainer containers:

```yaml
requirements:
  SingularityRequirement:
    singularityPull: "docker://ubuntu:20.04"
```

The executor will automatically detect whether Singularity or Apptainer is installed on your system.

## Limitations

- Limited support for CWL expressions
- No support for Workflows yet
- Limited support for complex data types

## License

MIT
