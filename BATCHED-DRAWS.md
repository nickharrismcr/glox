# Batched Drawing System Design

## Overview

This document describes the design and implementation of a batched drawing system for the GLox interpreter. The system follows the **established builtin object pattern** used throughout GLox (like `FloatArrayObject`, `WindowObject`, etc.) and will dramatically improve rendering performance by reducing individual draw calls and leveraging OpenGL/Raylib's batching capabilities.

 

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

Following the existing builtin object pattern (like `FloatArrayObject`, `WindowObject`, etc.):

```go
// Core batch entry structure
type BatchEntry struct {
    Position core.Vec3Object  // World position (reuse existing Vec3)
    Size     core.Vec3Object  // Dimensions (width, height, depth)
    Color    core.Vec4Object  // RGBA color (0-255 range)
    Rotation core.Vec3Object  // Rotation angles (optional, for future)
}

// Internal batch data container
type DrawBatch struct {
    BatchType string        // "cube", "sphere", "plane", etc.
    Entries   []BatchEntry  // Array of draw data
    Capacity  int          // Pre-allocated capacity for performance
}

// Main batch object following the standard pattern
type BatchObject struct {
    core.BuiltInObject
    Value   *DrawBatch
    Methods map[int]*core.BuiltInObject
}
```

### Batch Management Functions

```go
// Create batch object (follows standard constructor pattern)
func BatchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
    if argCount != 1 {
        vm.RunTimeError("batch() expects 1 argument")
        return core.NIL_VALUE
    }
    
    batchTypeVal := vm.Stack(arg_stackptr)
    if !batchTypeVal.IsStringObject() {
        vm.RunTimeError("batch() argument must be a string")
        return core.NIL_VALUE
    }
    
    batchType := batchTypeVal.Obj.(*core.StringObject).Value
    batchObj := MakeBatchObject(batchType)
    RegisterAllBatchMethods(batchObj)
    return core.MakeObjectValue(batchObj, true)
}

// Constructor following standard pattern
func MakeBatchObject(batchType string) *BatchObject {
    return &BatchObject{
        BuiltInObject: core.BuiltInObject{},
        Value: &DrawBatch{
            BatchType: batchType,
            Entries:   make([]BatchEntry, 0, 1000), // Pre-allocate capacity
            Capacity:  1000,
        },
    }
}

// Standard object interface implementations
func (o *BatchObject) String() string {
    return fmt.Sprintf("<Batch %s [%d entries]>", o.Value.BatchType, len(o.Value.Entries))
}

func (o *BatchObject) GetType() core.ObjectType {
    return core.OBJECT_NATIVE
}

func (o *BatchObject) GetNativeType() core.NativeType {
    return core.NATIVE_BATCH  // New type to be added
}

func (o *BatchObject) GetMethod(stringId int) *core.BuiltInObject {
    return o.Methods[stringId]
}

func (o *BatchObject) RegisterMethod(name string, method *core.BuiltInObject) {
    if o.Methods == nil {
        o.Methods = make(map[int]*core.BuiltInObject)
    }
    o.Methods[core.InternName(name)] = method
}

func (o *BatchObject) IsBuiltIn() bool {
    return true
}

// Core batch operations (internal methods)
func (batch *DrawBatch) Add(pos *core.Vec3Object, size *core.Vec3Object, color *core.Vec4Object) int {
    entry := BatchEntry{
        Position: *pos,
        Size:     *size,
        Color:    *color,
        Rotation: *core.MakeVec3Object(0, 0, 0), // Default no rotation
    }
    batch.Entries = append(batch.Entries, entry)
    return len(batch.Entries) - 1
}

func (batch *DrawBatch) SetPosition(index int, pos *core.Vec3Object) error {
    if index < 0 || index >= len(batch.Entries) {
        return fmt.Errorf("index out of range: %d", index)
    }
    batch.Entries[index].Position = *pos
    return nil
}

func (batch *DrawBatch) SetColor(index int, color *core.Vec4Object) error {
    if index < 0 || index >= len(batch.Entries) {
        return fmt.Errorf("index out of range: %d", index)
    }
    batch.Entries[index].Color = *color
    return nil
}

func (batch *DrawBatch) SetSize(index int, size *core.Vec3Object) error {
    if index < 0 || index >= len(batch.Entries) {
        return fmt.Errorf("index out of range: %d", index)
    }
    batch.Entries[index].Size = *size
    return nil
}

// Clear all entries for next frame
func (batch *DrawBatch) Clear() {
    batch.Entries = batch.Entries[:0] // Keep capacity, reset length
}
```

## Lox API Design

### Basic Usage

The batch objects are created directly and follow the standard builtin object pattern:

```lox
// Create a batch for cubes
var cube_batch = batch("cube")

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

### Batch Object Methods (in batch_methods.go)

Following the existing pattern like `farray_methods.go`:

```lox
// Core batch operations
cube_batch.add(position, size, color)          // Returns index
cube_batch.set_position(index, vec3)           // Update position by index
cube_batch.set_color(index, vec4)              // Update color by index
cube_batch.set_size(index, vec3)               // Update size by index
cube_batch.get_position(index)                 // Get position by index
cube_batch.get_color(index)                    // Get color by index
cube_batch.get_size(index)                     // Get size by index

// Batch information
cube_batch.count()                             // Number of entries
cube_batch.capacity()                          // Current capacity

// Batch management
cube_batch.clear()                             // Remove all entries
cube_batch.reserve(size)                       // Pre-allocate space
cube_batch.is_valid_index(index)               // Check if index exists

// Draw operation
cube_batch.draw()                              // Render all entries in batch
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
    var cube_batch = batch("cube")
    
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
    
    draw(renderer) {        // Initialize batch once
        if (!this.batch_initialized) {
            this.cube_batch = batch("cube")
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
- Add `NATIVE_BATCH` type to `src/core/object.go`
- Implement `src/builtin/obj_builtin_batch.go` with core BatchObject
- Implement `src/builtin/batch_methods.go` with Lox method bindings
- Register `BatchBuiltIn` constructor in `src/builtin/builtin.go`
- Add basic methods: `add()`, `set_position()`, `set_color()`, `draw()`, `clear()`
- Convert cube city example to use batching
- Performance testing and validation

### Phase 2: Extended API
- Add indexed access methods: `get_position()`, `get_color()`, `set_size()`
- Implement batch management: `reserve()`, `count()`, `capacity()`, `is_valid_index()`
- Add utility methods for common operations
- Error handling and bounds checking

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

#### New Files to Create
- `src/builtin/obj_builtin_batch.go` - Core `BatchObject` implementation
- `src/builtin/batch_methods.go` - Lox method bindings for batch operations

#### Files to Modify
- `src/core/object.go` - Add `NATIVE_BATCH` type constant
- `src/builtin/builtin.go` - Register `BatchBuiltIn` constructor function
- `src/builtin/builtin_draw.go` - Add batch rendering implementation

#### Implementation Structure

**obj_builtin_batch.go** (following `obj_builtin_farray.go` pattern):
```go
package builtin

import (
    "fmt"
    "glox/src/core"
    rl "github.com/gen2brain/raylib-go/raylib"
)

// Constructor function (follows standard pattern)
func BatchBuiltIn(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
    if argCount != 1 {
        vm.RunTimeError("batch() expects 1 argument")
        return core.NIL_VALUE
    }
    
    batchTypeVal := vm.Stack(arg_stackptr)
    if !batchTypeVal.IsStringObject() {
        vm.RunTimeError("batch() argument must be a string")
        return core.NIL_VALUE
    }
    
    batchType := batchTypeVal.Obj.(*core.StringObject).Value
    batchObj := MakeBatchObject(batchType)
    RegisterAllBatchMethods(batchObj)
    return core.MakeObjectValue(batchObj, true)
}

// Internal data structures
type BatchEntry struct {
    Position core.Vec3Object
    Size     core.Vec3Object
    Color    core.Vec4Object
    Rotation core.Vec3Object
}

type DrawBatch struct {
    BatchType string
    Entries   []BatchEntry
    Capacity  int
}

// Main object (follows standard pattern)
type BatchObject struct {
    core.BuiltInObject
    Value   *DrawBatch
    Methods map[int]*core.BuiltInObject
}

// Standard interface implementations
func MakeBatchObject(batchType string) *BatchObject { /* ... */ }
func (o *BatchObject) String() string { /* ... */ }
func (o *BatchObject) GetType() core.ObjectType { return core.OBJECT_NATIVE }
func (o *BatchObject) GetNativeType() core.NativeType { return core.NATIVE_BATCH }
func (o *BatchObject) GetMethod(stringId int) *core.BuiltInObject { /* ... */ }
func (o *BatchObject) RegisterMethod(name string, method *core.BuiltInObject) { /* ... */ }
func (o *BatchObject) IsBuiltIn() bool { return true }

// Utility functions
func IsBatchObject(v core.Value) bool { /* ... */ }
func AsBatch(v core.Value) *BatchObject { /* ... */ }
```

**batch_methods.go** (following `farray_methods.go` pattern):
```go
package builtin

import "glox/src/core"

func RegisterAllBatchMethods(o *BatchObject) {
    o.RegisterMethod("add", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Add entry to batch, return index
        },
    })
    
    o.RegisterMethod("set_position", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Update position by index
        },
    })
    
    o.RegisterMethod("set_color", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Update color by index
        },
    })
    
    o.RegisterMethod("draw", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Render all entries in batch
        },
    })
    
    o.RegisterMethod("clear", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Clear all entries
        },
    })
    
    o.RegisterMethod("count", &core.BuiltInObject{
        Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
            // Return number of entries
        },
    })
    
    // ... additional methods
}
```

### Core Type System Update

**src/core/object.go** - Add new native type:
```go
const (
    NATIVE_FLOAT_ARRAY NativeType = iota
    NATIVE_VEC2
    NATIVE_VEC3
    NATIVE_VEC4
    NATIVE_WINDOW
    NATIVE_IMAGE
    NATIVE_TEXTURE
    NATIVE_RENDER_TEXTURE
    NATIVE_CAMERA
    NATIVE_SHADER
    NATIVE_BATCH  // <- Add this new type
)
```

### Integration Pattern

The batch system follows the **standard GLox builtin pattern**:

1. **Constructor Function** (`BatchBuiltIn`) - Creates new batch objects
2. **Object Structure** (`BatchObject`) - Wraps internal data with methods map
3. **Method Registration** (`RegisterAllBatchMethods`) - Binds Lox methods
4. **Standard Interfaces** - Implements `Object`, `NativeObject` interfaces
5. **Type Safety** - Uses existing `Vec3Object`, `Vec4Object` types internally

This approach ensures:
- **Consistency** with existing builtin objects
- **Type Safety** using established Vec3/Vec4 objects  
- **Memory Management** following existing patterns
- **Easy Extension** can add new batch types and methods
- **Clean Integration** works seamlessly with existing renderer

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

### Rendering Implementation

The `draw()` method in `batch_methods.go` performs the actual batched rendering:

```go
o.RegisterMethod("draw", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if len(o.Value.Entries) == 0 {
            return core.NIL_VALUE
        }
        
        // Batch render based on type
        switch o.Value.BatchType {
        case "cube":
            for _, entry := range o.Value.Entries {
                pos := rl.Vector3{
                    X: float32(entry.Position.X), 
                    Y: float32(entry.Position.Y), 
                    Z: float32(entry.Position.Z),
                }
                size := rl.Vector3{
                    X: float32(entry.Size.X), 
                    Y: float32(entry.Size.Y), 
                    Z: float32(entry.Size.Z),
                }
                color := rl.Color{
                    R: uint8(entry.Color.X), 
                    G: uint8(entry.Color.Y), 
                    B: uint8(entry.Color.Z), 
                    A: uint8(entry.Color.W),
                }
                
                rl.DrawCube(pos, size.X, size.Y, size.Z, color)
            }
        case "sphere":
            // Similar implementation for spheres
            for _, entry := range o.Value.Entries {
                // Draw sphere with entry data
            }
        case "plane":
            // Similar implementation for planes
            for _, entry := range o.Value.Entries {
                // Draw plane with entry data
            }
        default:
            vm.RunTimeError(fmt.Sprintf("Unknown batch type: %s", o.Value.BatchType))
        }
        
        return core.NIL_VALUE
    },
})
```

**Key Benefits of This Approach:**

1. **Single Loop**: All primitives of the same type rendered in one tight loop
2. **Minimal State Changes**: Raylib can optimize consecutive similar draw calls
3. **Memory Locality**: Entry data stored contiguously for better cache performance
4. **Type-Specific Optimization**: Each primitive type can have specialized rendering
5. **GPU Batching**: Raylib's internal batching systems can optimize further

**Integration with 3D Mode:**

The batch draw calls work within existing 3D rendering context:

```lox
win.begin_3d(camera)
// Multiple batch.draw() calls here - all batched efficiently
cube_batch.draw()
sphere_batch.draw() 
plane_batch.draw()
win.end_3d()
```

This maintains compatibility with existing rendering while dramatically improving performance.
