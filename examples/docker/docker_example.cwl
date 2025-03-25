#!/usr/bin/env cwl-runner

cwlVersion: v1.0
class: CommandLineTool

requirements:
  DockerRequirement:
    dockerPull: "ubuntu:20.04"

baseCommand: ["echo"]

inputs:
  message:
    type: string
    inputBinding:
      position: 1

outputs:
  output:
    type: stdout

stdout: output.txt