# Proposed Window Methods Fixes

This file contains suggested fixes for the inconsistent window method APIs.

## Current Issues Summary

The window methods have inconsistent parameter patterns:
- Some methods correctly take individual x,y parameters (like `rectangle`)
- Other methods incorrectly require Vec2 objects (like `line`, `circle`, texture methods)
- The `text` method has hardcoded size and color
- Color handling is inconsistent (Vec4 vs encoded RGB)

## Proposed Fixes

### 1. Line Methods - Fix to use x,y parameters

**Current problematic implementation:**
```go
// line(vec2_start, vec2_end, color_vec4)
o.RegisterMethod("line", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        l1 := vm.Stack(arg_stackptr)
        if l1.Type != core.VAL_VEC2 {
            vm.RunTimeError("Expected Vec2 for line start position")
            return core.NIL_VALUE
        }
        l2 := vm.Stack(arg_stackptr + 1)
        if l2.Type != core.VAL_VEC2 {
            vm.RunTimeError("Expected Vec2 for line end position")
            return core.NIL_VALUE
        }
        // ... rest of implementation
    },
})
```

**Proposed fix:**
```go
// line(x1, y1, x2, y2, color_vec4)
o.RegisterMethod("line", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if argCount != 5 {
            vm.RunTimeError("line expects 5 arguments: x1, y1, x2, y2, color")
            return core.NIL_VALUE
        }
        
        x1Val := vm.Stack(arg_stackptr)
        y1Val := vm.Stack(arg_stackptr + 1)
        x2Val := vm.Stack(arg_stackptr + 2)
        y2Val := vm.Stack(arg_stackptr + 3)
        colVal := vm.Stack(arg_stackptr + 4)
        
        if colVal.Type != core.VAL_VEC4 {
            vm.RunTimeError("Expected Vec4 for line color")
            return core.NIL_VALUE
        }
        
        v4obj := colVal.Obj.(*core.Vec4Object)
        r := uint8(v4obj.X)
        g := uint8(v4obj.Y)
        b := uint8(v4obj.Z)
        a := uint8(v4obj.W)

        x1 := int32(x1Val.AsFloat())
        y1 := int32(y1Val.AsFloat())
        x2 := int32(x2Val.AsFloat())
        y2 := int32(y2Val.AsFloat())

        rl.DrawLine(x1, y1, x2, y2, rl.NewColor(r, g, b, a))
        return core.NIL_VALUE
    },
})
```

### 2. Circle Methods - Fix to use x,y parameters

**Current problematic implementation:**
```go
// circle_fill(vec2_position, radius, color_vec4)
```

**Proposed fix:**
```go
// circle_fill(x, y, radius, color_vec4)
o.RegisterMethod("circle_fill", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if argCount != 4 {
            vm.RunTimeError("circle_fill expects 4 arguments: x, y, radius, color")
            return core.NIL_VALUE
        }
        
        xVal := vm.Stack(arg_stackptr)
        yVal := vm.Stack(arg_stackptr + 1)
        radVal := vm.Stack(arg_stackptr + 2)
        colVal := vm.Stack(arg_stackptr + 3)
        
        if colVal.Type != core.VAL_VEC4 {
            vm.RunTimeError("Expected Vec4 for circle color")
            return core.NIL_VALUE
        }
        
        v4obj := colVal.Obj.(*core.Vec4Object)
        r := uint8(v4obj.X)
        g := uint8(v4obj.Y)
        b := uint8(v4obj.Z)
        a := uint8(v4obj.W)

        x := int32(xVal.AsFloat())
        y := int32(yVal.AsFloat())
        radius := float32(radVal.AsFloat())

        rl.DrawCircle(x, y, radius, rl.NewColor(r, g, b, a))
        return core.NIL_VALUE
    },
})
```

### 3. Triangle Method - Fix to use individual coordinates

**Proposed fix:**
```go
// triangle(x1, y1, x2, y2, x3, y3, color_vec4)
o.RegisterMethod("triangle", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if argCount != 7 {
            vm.RunTimeError("triangle expects 7 arguments: x1, y1, x2, y2, x3, y3, color")
            return core.NIL_VALUE
        }
        
        x1Val := vm.Stack(arg_stackptr)
        y1Val := vm.Stack(arg_stackptr + 1)
        x2Val := vm.Stack(arg_stackptr + 2)
        y2Val := vm.Stack(arg_stackptr + 3)
        x3Val := vm.Stack(arg_stackptr + 4)
        y3Val := vm.Stack(arg_stackptr + 5)
        colVal := vm.Stack(arg_stackptr + 6)

        if colVal.Type != core.VAL_VEC4 {
            vm.RunTimeError("Expected Vec4 for triangle color")
            return core.NIL_VALUE
        }

        v4obj := colVal.Obj.(*core.Vec4Object)
        r := uint8(v4obj.X)
        g := uint8(v4obj.Y)
        b := uint8(v4obj.Z)
        a := uint8(v4obj.W)

        x1 := float32(x1Val.AsFloat())
        y1 := float32(y1Val.AsFloat())
        x2 := float32(x2Val.AsFloat())
        y2 := float32(y2Val.AsFloat())
        x3 := float32(x3Val.AsFloat())
        y3 := float32(y3Val.AsFloat())
        
        rlv1 := rl.Vector2{X: x1, Y: y1}
        rlv2 := rl.Vector2{X: x2, Y: y2}
        rlv3 := rl.Vector2{X: x3, Y: y3}

        rl.DrawTriangle(rlv1, rlv2, rlv3, rl.NewColor(r, g, b, a))
        return core.NIL_VALUE
    },
})
```

### 4. Text Method - Add size and color parameters

**Proposed fix:**
```go
// text(text, x, y, size, color_vec4)
o.RegisterMethod("text", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if argCount != 5 {
            vm.RunTimeError("text expects 5 arguments: text, x, y, size, color")
            return core.NIL_VALUE
        }
        
        textVal := vm.Stack(arg_stackptr)
        xVal := vm.Stack(arg_stackptr + 1)
        yVal := vm.Stack(arg_stackptr + 2)
        sizeVal := vm.Stack(arg_stackptr + 3)
        colVal := vm.Stack(arg_stackptr + 4)

        if colVal.Type != core.VAL_VEC4 {
            vm.RunTimeError("Expected Vec4 for text color")
            return core.NIL_VALUE
        }

        v4obj := colVal.Obj.(*core.Vec4Object)
        r := uint8(v4obj.X)
        g := uint8(v4obj.Y)
        b := uint8(v4obj.Z)
        a := uint8(v4obj.W)

        text := textVal.AsString().Get()
        x := int32(xVal.AsFloat())
        y := int32(yVal.AsFloat())
        size := int32(sizeVal.AsFloat())

        rl.DrawText(text, x, y, size, rl.NewColor(r, g, b, a))
        return core.NIL_VALUE
    },
})
```

### 5. Texture Methods - Fix to use x,y parameters

**Proposed fix for draw_texture:**
```go
// draw_texture(texture, x, y, color_vec4)
o.RegisterMethod("draw_texture", &core.BuiltInObject{
    Function: func(argCount int, arg_stackptr int, vm core.VMContext) core.Value {
        if argCount != 4 {
            vm.RunTimeError("draw_texture expects 4 arguments: texture, x, y, color")
            return core.NIL_VALUE
        }
        
        textureVal := vm.Stack(arg_stackptr)
        xVal := vm.Stack(arg_stackptr + 1)
        yVal := vm.Stack(arg_stackptr + 2)
        colVal := vm.Stack(arg_stackptr + 3)
        
        if colVal.Type != core.VAL_VEC4 {
            vm.RunTimeError("Expected Vec4 for texture color")
            return core.NIL_VALUE
        }
        
        v4obj := colVal.Obj.(*core.Vec4Object)
        tint := rl.NewColor(uint8(v4obj.X), uint8(v4obj.Y), uint8(v4obj.Z), uint8(v4obj.W))

        x := float32(xVal.AsFloat())
        y := float32(yVal.AsFloat())

        to := textureVal.Obj.(*TextureObject)
        rect := to.Data.GetFrameRect()
        rl.DrawTextureRec(to.Data.Texture, rect, rl.Vector2{X: x, Y: y}, tint)
        to.Data.Animate()
        return core.NIL_VALUE
    },
})
```

## Implementation Strategy

1. **Create backward-compatible versions first** - Keep old method names working while adding new consistent ones
2. **Add proper parameter validation** - Check argument counts and types
3. **Standardize color handling** - Use Vec4(r, g, b, a) consistently with 0-255 values
4. **Update examples and documentation** - Show the improved API usage
5. **Consider deprecation warnings** - For the old inconsistent methods

## Benefits of These Changes

1. **Consistency** - All methods follow the same pattern for coordinates
2. **Easier to use** - No need to create Vec2 objects for simple coordinates
3. **Better performance** - Fewer object allocations for simple drawing operations
4. **More intuitive** - Matches common graphics API patterns (like HTML5 Canvas, Processing, etc.)
5. **Future-proof** - Easier to extend and maintain

## Migration Path

```lox
// Old way (current problematic API)
win.line(vec2(10, 20), vec2(30, 40), vec4(255, 0, 0, 255));
win.circle_fill(vec2(100, 100), 25, vec4(0, 255, 0, 255));

// New way (proposed improved API)
win.line(10, 20, 30, 40, vec4(255, 0, 0, 255));
win.circle_fill(100, 100, 25, vec4(0, 255, 0, 255));
```
