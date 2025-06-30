#!/bin/bash
# Build and run GLox with proper environment setup

# Build the project
echo "Building GLox..."
bin/build.sh
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

# Set environment
source setenv

# Run the provided script or start REPL
if [ $# -eq 0 ]; then
    echo "Starting GLox REPL..."
    ./bin/glox
else
    echo "Running: $1"
    ./bin/glox "$1"
fi
