# OS Module Documentation

The `os` module provides directory and file system functionality similar to Python's `os` module.

## Usage

```lox
import os
```

## Directory Operations

### `os.listdir(path)` → list
Lists the contents of a directory.
- **path**: String path to the directory
- **Returns**: List of filenames and directory names

```lox
files = os.listdir(".")
for i in range(len(files)) {
    print files[i]
}
```

### `os.mkdir(path)` → bool
Creates a directory (including parent directories if needed).
- **path**: String path to create
- **Returns**: `true` on success

```lox
success = os.mkdir("new_directory")
```

### `os.rmdir(path)` → bool
Removes an empty directory.
- **path**: String path to remove
- **Returns**: `true` on success

```lox
success = os.rmdir("empty_directory")
```

### `os.getcwd()` → string
Gets the current working directory.
- **Returns**: String path of current directory

```lox
current_dir = os.getcwd()
print "Working in: " + current_dir
```

### `os.chdir(path)` → bool
Changes the current working directory.
- **path**: String path to change to
- **Returns**: `true` on success

```lox
success = os.chdir("../parent_directory")
```

## File Operations

### `os.open(path, mode)` → file
Opens a file for reading or writing.
- **path**: String path to the file
- **mode**: String mode ("r" for read, "w" for write, "a" for append)
- **Returns**: File object

```lox
file = os.open("data.txt", "r")
```

### `os.close(file)` → bool
Closes an open file.
- **file**: File object to close
- **Returns**: `true` on success

```lox
os.close(file)
```

### `os.readln(file)` → string
Reads a line from an open file.
- **file**: File object to read from
- **Returns**: String line (without newline characters)

```lox
line = os.readln(file)
```

### `os.write(file, text)` → bool
Writes text to an open file.
- **file**: File object to write to
- **text**: String text to write
- **Returns**: `true` on success

```lox
os.write(file, "Hello, world!")
```

### `os.remove(path)` → bool
Removes a file.
- **path**: String path to file
- **Returns**: `true` on success

```lox
success = os.remove("unwanted_file.txt")
```

## Path Testing

### `os.exists(path)` → bool
Checks if a path exists (file or directory).
- **path**: String path to check
- **Returns**: `true` if exists, `false` otherwise

```lox
if (os.exists("config.txt")) {
    print "Config file found"
}
```

### `os.isdir(path)` → bool
Checks if a path is a directory.
- **path**: String path to check
- **Returns**: `true` if directory, `false` otherwise

```lox
if (os.isdir("assets")) {
    print "Assets directory exists"
}
```

### `os.isfile(path)` → bool
Checks if a path is a regular file.
- **path**: String path to check
- **Returns**: `true` if file, `false` otherwise

```lox
if (os.isfile("main.lox")) {
    print "Main script found"
}
```

## Path Manipulation

### `os.join(path1, path2, ...)` → string
Joins path components with the appropriate separator.
- **path1, path2, ...**: String path components
- **Returns**: Joined path string

```lox
full_path = os.join("assets", "images", "sprite.png")
// Result: "assets/images/sprite.png"
```

### `os.dirname(path)` → string
Gets the directory portion of a path.
- **path**: String path
- **Returns**: Directory portion

```lox
dir = os.dirname("assets/images/sprite.png")
// Result: "assets/images"
```

### `os.basename(path)` → string
Gets the filename portion of a path.
- **path**: String path
- **Returns**: Filename portion

```lox
filename = os.basename("assets/images/sprite.png")
// Result: "sprite.png"
```

### `os.splitext(path)` → [name, extension]
Splits a filename into name and extension.
- **path**: String path
- **Returns**: List with [name, extension]

```lox
parts = os.splitext("sprite.png")
name = parts[0]      // "sprite"
extension = parts[1] // ".png"
```

## Examples

### Directory Traversal
```lox
import os

fun list_directory_recursive(path, level) {
    indent = ""
    i = 0
    while (i < level) {
        indent = indent + "  "
        i = i + 1
    }
    
    files = os.listdir(path)
    i = 0
    while (i < len(files)) {
        file = files[i]
        full_path = os.join(path, file)
        
        if (os.isdir(full_path)) {
            print indent + "[DIR]  " + file
            list_directory_recursive(full_path, level + 1)
        } else {
            print indent + "[FILE] " + file
        }
        i = i + 1
    }
}

list_directory_recursive(".", 0)
```

### File Utility Functions
```lox
import os

fun find_files_with_extension(directory, extension) {
    results = []
    files = os.listdir(directory)
    
    i = 0
    while (i < len(files)) {
        file = files[i]
        full_path = os.join(directory, file)
        
        if (os.isfile(full_path)) {
            parts = os.splitext(file)
            if (parts[1] == extension) {
                append(results, full_path)
            }
        }
        i = i + 1
    }
    
    return results
}

// Find all .lox files in current directory
lox_files = find_files_with_extension(".", ".lox")
print "Found Lox files:"
i = 0
while (i < len(lox_files)) {
    print "  " + lox_files[i]
    i = i + 1
}
```

### Working with Paths
```lox
import os

// Build platform-independent paths
config_path = os.join("config", "settings.txt")
assets_path = os.join("assets", "images", "player.png")

// Check if required files exist
if (!os.exists(config_path)) {
    print "Config file missing: " + config_path
}

if (!os.exists(assets_path)) {
    print "Asset file missing: " + assets_path
}

// Create directories if needed
if (!os.exists("output")) {
    os.mkdir("output")
    print "Created output directory"
}
```

### File I/O Operations
```lox
import os

// Writing to a file
fun write_config(filename, data) {
    file = os.open(filename, "w")
    if (file != nil) {
        os.write(file, "# Configuration file\n")
        os.write(file, "version = 1.0\n")
        os.write(file, "debug = true\n")
        os.close(file)
        print "Config written to " + filename
    }
}

// Reading from a file
fun read_config(filename) {
    if (!os.exists(filename)) {
        print "Config file not found: " + filename
        return
    }
    
    file = os.open(filename, "r")
    if (file != nil) {
        print "Reading config from " + filename + ":"
        
        // Read all lines (simplified - would need EOF handling)
        i = 0
        while (i < 10) {  // Limit to prevent infinite loop
            try {
                line = os.readln(file)
                if (line != nil and line != "") {
                    print "  " + line
                }
                i = i + 1
            } catch (EOFError) {
                break
            }
        }
        
        os.close(file)
    }
}

// Example usage
config_file = "app_config.txt"
write_config(config_file, "")
read_config(config_file)

// Clean up
if (os.exists(config_file)) {
    os.remove(config_file)
    print "Cleaned up " + config_file
}
```

### Log File Management
```lox
import os

// Create a simple logging system
fun log_message(message) {
    log_dir = "logs"
    if (!os.exists(log_dir)) {
        os.mkdir(log_dir)
    }
    
    log_file = os.join(log_dir, "app.log")
    file = os.open(log_file, "a")  // Append mode
    if (file != nil) {
        // In a real implementation, you'd add timestamps
        os.write(file, "[LOG] " + message + "\n")
        os.close(file)
    }
}

// Usage
log_message("Application started")
log_message("Processing data...")
log_message("Operation completed")

// Read back the log
if (os.exists(os.join("logs", "app.log"))) {
    print "Log contents:"
    file = os.open(os.join("logs", "app.log"), "r")
    if (file != nil) {
        i = 0
        while (i < 20) {  // Safety limit
            try {
                line = os.readln(file)
                if (line != nil and line != "") {
                    print line
                }
                i = i + 1
            } catch (EOFError) {
                break
            }
        }
        os.close(file)
    }
}
```

## Notes

- All path operations use forward slashes (`/`) as separators for cross-platform compatibility
- Directory creation with `mkdir` will create parent directories if they don't exist
- Error handling: Functions return `false` or raise runtime errors on failure
- Paths can be absolute or relative to the current working directory
- File operations:
  - Always `close()` files after opening them to prevent resource leaks
  - File modes: "r" (read), "w" (write/create), "a" (append)
  - `readln()` reads one line at a time and may raise `EOFError` at end of file
  - `write()` expects string input and handles `\n` escape sequences
- The `os` module consolidates all file system operations in one place
