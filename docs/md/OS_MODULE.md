# OS Module Documentation

The `os` module provides directory and file system functionality similar to Python's `os` module.

## Usage

```lox
import os
```

## Directory Operations

### `os.listdir(path)` → list
Lists the contents of a directory. Raises `RunTimeError` if the directory doesn't exist or can't be read.
- **path**: String path to the directory
- **Returns**: List of filenames and directory names

```lox
foreach (name in os.listdir(".")) {
    print name
}
```

### `os.mkdir(path)` → bool
Creates a directory, including any missing parent directories. Raises `RunTimeError` on failure.
- **path**: String path to create
- **Returns**: `true` on success

```lox
os.mkdir("new_directory")
```

### `os.rmdir(path)` → bool
Removes an empty directory. Raises `RunTimeError` on failure (including a non-empty directory).
- **path**: String path to remove
- **Returns**: `true` on success

```lox
os.rmdir("empty_directory")
```

### `os.getcwd()` → string
Gets the current working directory.
- **Returns**: String path of current directory

```lox
print "Working in: " & os.getcwd()
```

### `os.chdir(path)` → bool
Changes the current working directory. Raises `RunTimeError` on failure.
- **path**: String path to change to
- **Returns**: `true` on success

```lox
os.chdir("../parent_directory")
```

## File Operations

### `os.open(path, mode)` → file
Opens a file for reading or writing. Raises `RunTimeError` if the file can't be opened (e.g. mode `"r"` on a path that doesn't exist) — it never returns `nil`, so there's no need to check the result.
- **path**: String path to the file
- **mode**: String mode (`"r"` read, `"w"` write/truncate, `"a"` append)
- **Returns**: File object

```lox
var file = os.open("data.txt", "r")
```

### `os.close(file)` → bool
Closes an open file, flushing any buffered writes first.
- **file**: File object to close
- **Returns**: `true` on success

```lox
os.close(file)
```

### `os.readln(file)` → string
Reads one line from an open file (without the trailing newline). Raises [`EOFError`](../language-reference.html#exceptions) once the end of the file has been reached.
- **file**: File object to read from
- **Returns**: String line

```lox
var line = os.readln(file)
```

If the file ends with a trailing newline, the read that consumes that final
newline returns `""` rather than raising `EOFError` immediately — `EOFError`
only fires on the *next* call after that. A loop that checks for `""` as well
as catching `EOFError` (see the streaming example below) handles this
correctly either way.

### `os.write(file, text)` → bool
Writes text to an open file. `\n` in the string is written as a real newline.
- **file**: File object to write to
- **text**: String text to write
- **Returns**: `true` on success

```lox
os.write(file, "Hello, world!\n")
```

### `os.read_all(path)` → string
Reads an entire file in one call, given its path directly — no `open`/`close` needed. The simpler choice whenever the whole file (not line-by-line streaming) is what's wanted.
- **path**: String path to the file
- **Returns**: The file's full contents as a string

```lox
var contents = os.read_all("data.txt")
```

### `os.remove(path)` → bool
Removes a file. Raises `RunTimeError` on failure.
- **path**: String path to file
- **Returns**: `true` on success

```lox
os.remove("unwanted_file.txt")
```

## Path Testing

### `os.exists(path)` → bool
Checks if a path exists (file or directory). Never raises for a missing path.
- **path**: String path to check
- **Returns**: `true` if it exists, `false` otherwise

```lox
if (os.exists("config.txt")) {
    print "Config file found"
}
```

### `os.isdir(path)` → bool
Checks if a path is a directory. Never raises for a missing path.
- **path**: String path to check
- **Returns**: `true` if directory, `false` otherwise

```lox
if (os.isdir("assets")) {
    print "Assets directory exists"
}
```

### `os.isfile(path)` → bool
Checks if a path is a regular file. Never raises for a missing path.
- **path**: String path to check
- **Returns**: `true` if file, `false` otherwise

```lox
if (os.isfile("main.lox")) {
    print "Main script found"
}
```

## Path Manipulation

These are pure string operations — they don't touch the filesystem or check whether anything actually exists at the resulting path.

### `os.join(path1, path2, ...)` → string
Joins path components with `/`.
- **path1, path2, ...**: String path components
- **Returns**: Joined path string

```lox
var full_path = os.join("assets", "images", "sprite.png")
// "assets/images/sprite.png"
```

### `os.dirname(path)` → string
Gets the directory portion of a path.
- **path**: String path
- **Returns**: Directory portion

```lox
var dir = os.dirname("assets/images/sprite.png")
// "assets/images"
```

### `os.basename(path)` → string
Gets the filename portion of a path.
- **path**: String path
- **Returns**: Filename portion

```lox
var filename = os.basename("assets/images/sprite.png")
// "sprite.png"
```

### `os.splitext(path)` → [name, extension]
Splits a filename into name and extension.
- **path**: String path
- **Returns**: A 2-element list `[name, extension]`; `extension` is `""` if there is none

```lox
var parts = os.splitext("sprite.png")
var name = parts[0]      // "sprite"
var extension = parts[1] // ".png"
```

## Examples

### Directory traversal
```lox
import os

func list_directory_recursive(path, level) {
    var indent = "  " * level

    foreach (name in os.listdir(path)) {
        var full_path = os.join(path, name)
        if (os.isdir(full_path)) {
            print indent & "[DIR]  " & name
            list_directory_recursive(full_path, level + 1)
        } else {
            print indent & "[FILE] " & name
        }
    }
}

list_directory_recursive(".", 0)
```

### Finding files by extension
```lox
import os

func find_files_with_extension(directory, extension) {
    var results = []
    foreach (name in os.listdir(directory)) {
        var full_path = os.join(directory, name)
        if (os.isfile(full_path) and os.splitext(name)[1] == extension) {
            results.append(full_path)
        }
    }
    return results
}

foreach (path in find_files_with_extension(".", ".lox")) {
    print "  " & path
}
```

### Working with paths
```lox
import os

// Build platform-independent paths
var config_path = os.join("config", "settings.txt")
var assets_path = os.join("assets", "images", "player.png")

if (!os.exists(config_path)) {
    print "Config file missing: " & config_path
}
if (!os.exists(assets_path)) {
    print "Asset file missing: " & assets_path
}

if (!os.exists("output")) {
    os.mkdir("output")
    print "Created output directory"
}
```

### Writing and reading a config file
```lox
import os

func write_config(filename) {
    var file = os.open(filename, "w")
    os.write(file, "# Configuration file\n")
    os.write(file, "version = 1.0\n")
    os.write(file, "debug = true\n")
    os.close(file)
}

// Line-by-line streaming, for large files or when you need to stop early --
// readln() raises EOFError once the file is exhausted (see the note above
// about the one "" read that can precede it).
func read_config_streaming(filename) {
    var file = os.open(filename, "r")
    try {
        while (true) {
            var line = os.readln(file)
            if (line != "") {
                print "  " & line
            }
        }
    } except EOFError as e {
        // reached the end -- expected, not an error
    } finally {
        os.close(file)
    }
}

// Simpler alternative when you just want the whole file at once.
func read_config_whole(filename) {
    print os.read_all(filename)
}

var config_file = "app_config.txt"
write_config(config_file)
print "Reading config from " & config_file & ":"
read_config_streaming(config_file)
os.remove(config_file)
```

### Log file management
```lox
import os

func log_message(message) {
    var log_dir = "logs"
    if (!os.exists(log_dir)) {
        os.mkdir(log_dir)
    }
    var log_file = os.join(log_dir, "app.log")
    var file = os.open(log_file, "a")
    os.write(file, "[LOG] " & message & "\n")
    os.close(file)
}

log_message("Application started")
log_message("Processing data...")
log_message("Operation completed")

print "Log contents:"
print os.read_all(os.join("logs", "app.log"))
```

## Notes

- All path operations use forward slashes (`/`) as separators, regardless of platform.
- `mkdir` creates any missing parent directories (like `mkdir -p`).
- Most operations **raise `RunTimeError` on failure** rather than returning `false` — see each function above for which ones. `exists`/`isdir`/`isfile` are the exception: they return `false` for a missing path instead of raising.
- Paths can be absolute or relative to the current working directory.
- Always `os.close()` a file once done with it (or run cleanup in a `finally` block) to avoid leaking file descriptors and to make sure buffered writes are flushed.
- `readln()` raises `EOFError` at end of file; `read_all()` is simpler when the whole file, not line-by-line access, is what's needed.
