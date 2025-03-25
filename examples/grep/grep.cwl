#!/usr/bin/env cwl-runner

cwlVersion: v1.2
class: CommandLineTool
id: grep-tool

requirements:
  - class: EnvVarRequirement
    envDef:
      - name: LANG
        value: C

baseCommand: grep

inputs:
  pattern:
    type: string
    inputBinding:
      position: 1
  
  invert:
    type: boolean?
    inputBinding:
      prefix: -v
      position: 0
  
  file:
    type: File
    inputBinding:
      position: 2

arguments:
  - prefix: -n
    position: 0

outputs:
  matches:
    type: stdout
  
  count:
    type: int
    outputBinding:
      glob: output.txt
      loadContents: true
      outputEval: $(parseInt(self[0].contents.split('\n').length - 1))

stdout: output.txt