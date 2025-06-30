# Terminal Instructions for GLox Development

## IMPORTANT: Always read this file before running terminal commands

When working in the GLox workspace:

1. **NEVER use PowerShell commands in bash terminal**
2. **ALWAYS use `./run.sh <script>` to run GLox scripts**
3. **Use `./setenv` (not setenv.ps1) to set up environment if needed**
4. **The working directory is already set to c:\Users\nicholas.harris\glox**

### Correct command patterns:
- To run a GLox script: `./run.sh lox_examples/cube_stack_fly2.lox`
- To set environment: `./setenv`
- To build: `./bin/build.sh`

### AVOID these patterns:
- `powershell -ExecutionPolicy Bypass -File .\setenv.ps1`
- `.\bin\glox.exe .\lox_examples\script.lox`
- Any PowerShell-specific syntax in bash

This will prevent command errors and ensure consistent execution environment.
