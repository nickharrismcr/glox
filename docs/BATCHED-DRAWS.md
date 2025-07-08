# Batched Drawing System Design

## Overview

This document describes the design and implementation of a batched drawing system for the GLox interpreter. The system will dramatically improve rendering performance by reducing individual draw calls and leveraging OpenGL/Raylib's batching capabilities.

## Current Performance Issues

The current cube city example generates approximately **5,000+ individual draw calls per frame**:
- 49 stacks × ~12 cubes average × 5 faces each = ~2,940 face draws
- 49 stacks × ~12 cubes × 1 base cube = ~588 base cube draws  
- Plus boundaries, ground plane, etc.

Each draw call has significant CPU-GPU communication overhead, limiting scalability.

## Proposed Solution

Implement a batching system where the Go backend maintains lists of draw data that can be rendered in single, optimized OpenGL calls using Raylib's batching capabilities.

## Architecture

### Go Backend Implementation

```go
// Core batch entry structure
type BatchEntry struct {
    Position vec3  // World position
    Size     vec3  // Dimensions (width, height, depth)
    Color    vec4  // RGBA color (0-255 range)
    Rotation vec3  // Rotation angles (optional, for future)
}

// Batch container for specific primitive types
type DrawBatch struct {
    BatchType string        // "cube", "sphere", "plane", etc.
    Entries   []BatchEntry  // Array of draw data
    Capacity  int          // Pre-allocated capacity for performance
}

// Global batch registry
var globalBatches = make(map[string]*DrawBatch)
```

### Batch Management Functions

```go
// Create or retrieve existing batch
func createBatch(batchType string) *DrawBatch {
    if batch, exists := globalBatches[batchType]; exists {
        return batch
    }
    
    batch := &DrawBatch{
        BatchType: batchType,
        Entries:   make([]BatchEntry, 0, 1000), // Pre-allocate capacity
        Capacity:  1000,
    }
    globalBatches[batchType] = batch
    return batch
}

// Add entry to batch, returns index for later access
func (batch *DrawBatch) Add(pos vec3, size vec3, color vec4) int {
    entry := BatchEntry{
        Position: pos,
        Size:     size,
        Color:    color,
    }
    batch.Entries = append(batch.Entries, entry)
    return len(batch.Entries) - 1
}

// Modify existing entry by index
func (batch *DrawBatch) SetPosition(index int, pos vec3) {
    if index >= 0 && index < len(batch.Entries) {
        batch.Entries[index].Position = pos
    }
}

func (batch *DrawBatch) SetColor(index int, color vec4) {
    if index >= 0 && index < len(batch.Entries) {
        batch.Entries[index].Color = color
    }
}

// Generic set method using field names (recommended approach)
func (batch *DrawBatch) Set(index int, field string, value interface{}) error {
    if index < 0 || index >= len(batch.Entries) {
        return fmt.Errorf("index out of range: %d", index)
    }
    
    entry := &batch.Entries[index]
    
    switch field {
    case "position":
        if pos, ok := value.(vec3); ok {
            entry.Position = pos
        } else {
            return fmt.Errorf("invalid type for position")
        }
    case "size":
        if size, ok := value.(vec3); ok {
            entry.Size = size
        } else {
            return fmt.Errorf("invalid type for size")
        }
    case "color":
        if color, ok := value.(vec4); ok {
            entry.Color = color
        } else {
            return fmt.Errorf("invalid type for color")
        }
    default:
        return fmt.Errorf("unknown field: %s", field)
    }
    
    return nil
}

// Get method for retrieving current values
func (batch *DrawBatch) Get(index int, field string) (interface{}, error) {
    if index < 0 || index >= len(batch.Entries) {
        return nil, fmt.Errorf("index out of range: %d", index)
    }
    
    entry := &batch.Entries[index]
    
    switch field {
    case "position":
        return entry.Position, nil
    case "size":
        return entry.Size, nil
    case "color":
        return entry.Color, nil
    default:
        return nil, fmt.Errorf("unknown field: %s", field)
    }
}

// Clear all entries for next frame
func (batch *DrawBatch) Clear() {
    batch.Entries = batch.Entries[:0] // Keep capacity, reset length
}

// Perform batched draw using Raylib
func (batch *DrawBatch) Draw() {
    if len(batch.Entries) == 0 {
        return
    }
    
    rl.Begin3dMode(getCurrentCamera()) // If not already in 3D mode
    
    switch batch.BatchType {
    case "cube":
        for _, entry := range batch.Entries {
            rl.DrawCube(
                rl.Vector3{X: entry.Position.X, Y: entry.Position.Y, Z: entry.Position.Z},
                entry.Size.X, entry.Size.Y, entry.Size.Z,
                rl.Color{R: uint8(entry.Color.X), G: uint8(entry.Color.Y), 
                        B: uint8(entry.Color.Z), A: uint8(entry.Color.W)},
            )
        }
    case "sphere":
        // Similar implementation for spheres
    case "plane":
        // Similar implementation for planes
    }
    
    // Note: rl.End3dMode() called by main render loop
}
```

## Lox API Design

### Basic Usage

```lox
// Create or get a batch for cubes
var cube_batch = win.create_batch("cube")

// Add primitives (returns index for later modification)
var base_cube_idx = cube_batch.add(vec3(0, 0, 0), vec3(2, 2, 2), vec4(0, 0, 0, 255))
var face_cube_idx = cube_batch.add(vec3(0, 0, 1), vec3(1.8, 1.8, 0.1), vec4(255, 0, 0, 255))

// Modify existing entries by index
cube_batch.set_position(base_cube_idx, vec3(1, 1, 1))
cube_batch.set_color(face_cube_idx, vec4(0, 255, 0, 255))

// Draw entire batch in one optimized call
cube_batch.draw()

// Clear for next frame (optional - can reuse data)
cube_batch.clear()
```

### Position Update Strategy

The optimal approach for handling position updates uses **index + field name** for maximum flexibility:

```lox
// Add entry, get index for later updates
var cube_idx = cube_batch.add(vec3(0, 0, 0), vec3(2, 2, 2), vec4(255, 0, 0, 255))

// Update specific fields by index + field name
cube_batch.set(cube_idx, "position", vec3(1, 1, 1))
cube_batch.set(cube_idx, "color", vec4(0, 255, 0, 255))
cube_batch.set(cube_idx, "size", vec3(3, 3, 3))

// Get current values
var pos = cube_batch.get(cube_idx, "position")
var color = cube_batch.get(cube_idx, "color")
```

**Alternative: Typed Methods** (more verbose but type-safe)
```lox
cube_batch.set_position(cube_idx, vec3(1, 1, 1))
cube_batch.set_color(cube_idx, vec4(0, 255, 0, 255))
cube_batch.set_size(cube_idx, vec3(3, 3, 3))
```

**Recommendation**: Use index + field name approach for flexibility and ease of extension.

### Advanced API Methods

```lox
// Batch information
var count = cube_batch.count()        // Number of entries
var capacity = cube_batch.capacity()  // Current capacity

// Batch management
cube_batch.reserve(5000)              // Pre-allocate space
cube_batch.clear()                    // Remove all entries
cube_batch.remove(index)              // Remove specific entry

// Validation
var valid = cube_batch.is_valid_index(cube_idx)  // Check if index exists

// Bulk operations
cube_batch.set_all_colors(vec4(255, 0, 0, 255))  // Set same color for all
cube_batch.translate_all(vec3(0, 1, 0))           // Move all entries

// Efficient relative updates for moving objects
cube_batch.translate(cube_idx, vec3(0.1, 0, 0))  // Relative movement
cube_batch.scale(cube_idx, 1.2)                  // Scale single entry
```

## Integration with Current Code

### Before (Current Implementation)

```lox
// In Cube.draw() - 5 individual draw calls per cube
class Cube {
    draw(renderer) {
        // Black base cube
        renderer.cube(this.position, vec3(this.size, this.size, this.size), vec4(0, 0, 0, 255))
        
        // 4 colored face overlays
        var halfSize = this.size / 2
        var faceOffset = 0.15
        
        var frontPos = vec3(this.position.x, this.position.y, this.position.z + halfSize + faceOffset)
        renderer.cube(frontPos, vec3(this.size * 0.8, this.size * 0.8, 0.02), this.frontColor)
        
        // ... 3 more face draws
    }
}

// In CubeStackScene.draw() - calls draw() for each cube individually
foreach (var cube in this.cubes) {
    cube.draw(renderer)  // 5 draw calls × 1000+ cubes = 5000+ calls
}
```

### After (Batched Implementation)

```lox
// Modified Cube class - stores batch indices for updates
class Cube {
    init(x, y, z, size) {
        this.position = vec3(x, y, z)
        this.size = size
        this.rotation = vec3(0, 0, 0)
        
        // Store batch indices for later updates
        this.batch_indices = []  // [base_idx, front_idx, back_idx, left_idx, right_idx]
        
        // ...existing color generation...
    }
    
    addToBatch(cube_batch) {
        // Add black base cube to batch
        var base_idx = cube_batch.add(this.position, vec3(this.size, this.size, this.size), vec4(0, 0, 0, 255))
        this.batch_indices.append(base_idx)
        
        // Add colored face overlays to batch
        var halfSize = this.size / 2
        var faceOffset = 0.15
        
        // Front face
        var frontPos = vec3(this.position.x, this.position.y, this.position.z + halfSize + faceOffset)
        var front_idx = cube_batch.add(frontPos, vec3(this.size * 0.8, this.size * 0.8, 0.02), this.frontColor)
        this.batch_indices.append(front_idx)
        
        // Back face
        var backPos = vec3(this.position.x, this.position.y, this.position.z - halfSize - faceOffset)
        var back_idx = cube_batch.add(backPos, vec3(this.size * 0.8, this.size * 0.8, 0.02), this.backColor)
        this.batch_indices.append(back_idx)
        
        // ...add left and right faces...
    }
    
    updateBatch(cube_batch, deltaTime) {
        // Update rotation (for animation)
        this.rotation.y = this.rotation.y + deltaTime * 0.3
        this.rotation.x = this.rotation.x + deltaTime * 0.1
        
        // Update positions in batch if needed (for rotation effects)
        var halfSize = this.size / 2
        var faceOffset = 0.15
        
        // Update base cube position (if cube moves)
        cube_batch.set(this.batch_indices[0], "position", this.position)
        
        // Update face positions (accounting for rotation if implemented)
        var frontPos = vec3(this.position.x, this.position.y, this.position.z + halfSize + faceOffset)
        cube_batch.set(this.batch_indices[1], "position", frontPos)
        
        // ...update other face positions...
    }
}

// In CubeStackScene.draw() - single batched draw call
draw(renderer) {
    // Create batch for this frame
    var cube_batch = renderer.create_batch("cube")
    
    // Add all cubes to batch (no drawing yet)
    foreach (var cube in this.cubes) {
        cube.addToBatch(cube_batch)
    }
    
    // Single optimized draw call for ALL cubes (~5000 primitives in one call!)
    cube_batch.draw()
    
    // Optional: clear for next frame if not reusing data
    cube_batch.clear()
}

// Alternative: Persistent batch with updates (more efficient)
class CubeStackScene {
    init() {
        // ...existing code...
        this.cube_batch = null  // Will be created once and reused
        this.batch_initialized = false
    }
    
    draw(renderer) {
        // Initialize batch once
        if (!this.batch_initialized) {
            this.cube_batch = renderer.create_batch("cube")
            foreach (var cube in this.cubes) {
                cube.addToBatch(this.cube_batch)
            }
            this.batch_initialized = true
        }
        
        // Update positions/rotations in existing batch
        foreach (var cube in this.cubes) {
            cube.updateBatch(this.cube_batch, 0.016)  // deltaTime
        }
        
        // Single draw call
        this.cube_batch.draw()
    }
}
```

## Performance Benefits

### Expected Improvements

- **Draw Call Reduction**: From 5,000+ calls to 1-5 batched calls per frame
- **CPU Usage**: Significant reduction in CPU-GPU communication overhead
- **Scalability**: Can easily support 10x-100x more objects
- **Memory Efficiency**: Better GPU memory access patterns
- **Frame Rate**: Substantial FPS improvements, especially with many objects

### Benchmarking Targets

Current performance with ~1,000 cubes:
- Target: Maintain 60+ FPS with 10,000+ cubes
- Memory: Efficient reuse of batch arrays
- Latency: Sub-millisecond batch preparation time

## Implementation Phases

### Phase 1: Basic Cube Batching
- Implement core BatchEntry and DrawBatch structures
- Add basic Lox API: `create_batch()`, `add()`, `draw()`, `clear()`
- Convert cube city example to use batching
- Performance testing and validation

### Phase 2: Extended API
- Add indexed access: `set_position()`, `set_color()`, etc.
- Implement batch management: `reserve()`, `remove()`, `count()`
- Add bulk operations for common use cases

### Phase 3: Multiple Primitive Types
- Extend to spheres, planes, lines, etc.
- Optimize per-primitive-type batching
- Add primitive-specific batch methods

### Phase 4: Advanced Features
- Rotation support in batch entries
- Texture coordinate batching
- Material/shader batching
- Instanced rendering optimization

## Technical Considerations

### Memory Management
- Pre-allocate batch arrays to avoid frequent allocations
- Implement capacity growth strategies (2x growth)
- Consider memory pooling for very large scenes

### Threading
- Batch preparation could be threaded for large datasets
- Ensure thread-safety for batch modifications
- Consider producer-consumer patterns for dynamic scenes

### Backwards Compatibility
- Maintain existing immediate-mode drawing API
- Batched API as opt-in enhancement
- Gradual migration path for existing code

### Error Handling
- Bounds checking for indexed access
- Graceful degradation if batching fails
- Clear error messages for invalid operations

## Future Enhancements

- **Automatic Batching**: Transparent batching of immediate-mode calls
- **Spatial Culling**: Only batch visible objects
- **Level-of-Detail**: Different batch strategies based on distance
- **GPU Instancing**: Hardware-accelerated instanced rendering
- **Compute Shaders**: GPU-based batch preparation

## Files to Modify

### Go Backend Files
- `src/builtin/builtin_draw.go` - Core batching implementation
- `src/builtin/obj_builtin_window.go` - Window batch management methods
- `src/core/object.go` - Batch object type definitions

### Documentation
- `BUILTINS.md` - Document new batch API methods
- Update existing drawing examples

### Test Files
- Create batch performance test scripts
- Convert existing examples to demonstrate batching

## Success Metrics

- **Performance**: 10x+ improvement in draw call efficiency
- **Scalability**: Support for 10,000+ objects at 60+ FPS
- **API Usability**: Clean, intuitive Lox API
- **Backwards Compatibility**: No breaking changes to existing code
- **Memory Efficiency**: Minimal memory overhead for batching
