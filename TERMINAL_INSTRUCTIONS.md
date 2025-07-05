# Terminal Instructions for GLox Development

## IMPORTANT: Always read this file before running terminal commands

When working in the GLox workspace:

1. **NEVER use PowerShell commands in bash terminal**
2. **ALWAYS use `./run.sh <script>` to run GLox scripts**
3. **Use `./setenv` (not setenv.ps1) to set up environment if needed**
4. **The working directory is already set to c:\Users\nicholas.harris\glox**
5. **ALWAYS reuse the same terminal/console for all commands** - This prevents exhausting available consoles in VS Code

### Correct command patterns:
- To run a GLox script: `./run.sh lox_examples/cube_stack_fly2.lox`
- To set environment: `./setenv`
- To build: `./bin/build.sh`

### Additional Notes:
- **Terminal Reuse**: VS Code has limited console instances. Always reuse the same terminal/console for multiple commands to prevent resource exhaustion.
- **Build/Run Workflow**: Use the same terminal session for building and running to maintain environment consistency.

### AVOID these patterns:
- `powershell -ExecutionPolicy Bypass -File .\setenv.ps1`
- `.\bin\glox.exe .\lox_examples\script.lox`
- Any PowerShell-specific syntax in bash

This will prevent command errors and ensure consistent execution environment.
