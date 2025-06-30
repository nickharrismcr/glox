# GLox VS Code Setup

This project is configured to use bash terminals instead of PowerShell for better compatibility with the Unix-style build scripts.

## Terminal Configuration

- **Default Shell**: Git Bash
- **Environment Setup**: Use `source setenv` before running GLox
- **Terminal Reuse**: All tasks are configured to use the same shared terminal

## Quick Commands

### Build and Run
```bash
# Build, set environment, and run a script
./run.sh script.lox

# Just build and start REPL
./run.sh
```

### Manual Commands
```bash
# Build the project
go build

# Set environment (required before running GLox)
source setenv

# Run a script
./glox script.lox

# Start REPL
./glox
```

## VS Code Tasks

- **Build GLox** - Compile the Go project
- **Setup Environment** - Run `source setenv`
- **Build and Setup** - Combined build and environment setup
- **Run Batch Demo** - Run the batched cubes demo
- **Test Batch Constants** - Quick test of batch functionality

All tasks are configured to reuse the same terminal for efficient development.

## Environment Variables

The `setenv` script sets:
- `LOX_PATH`: Project root directory
- `PATH`: Adds `$LOX_PATH/bin` for module resolution

## File Associations

- Lox syntax highlighting is handled by the installed Lox VS Code extension
- No manual file associations needed
