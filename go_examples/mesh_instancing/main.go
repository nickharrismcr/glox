package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const MAX_INSTANCES = 100

// holds camera, mesh, shader, and material needed for instance rendering
type Context struct {
	camera   rl.Camera
	cube     rl.Mesh
	shader   rl.Shader
	material rl.Material
}

func MakeContext() *Context {
	return &Context{
		camera: rl.Camera{
			Position:   rl.NewVector3(-100.0, 30.0, -50.0),
			Target:     rl.NewVector3(0.0, 0.0, 0.0),
			Up:         rl.NewVector3(0.0, 1.0, 0.0),
			Fovy:       45.0,
			Projection: rl.CameraPerspective,
		},
		cube:     rl.GenMeshCube(10.0, 10.0, 10.0),
		shader:   rl.LoadShader("glsl330/base_lighting_instanced.vs", "glsl330/lighting.fs"),
		material: rl.LoadMaterialDefault(),
	}
}

// Instance represents a single instance of a mesh with its transformation matrices
type Instance struct {
	translation rl.Matrix
	rotation    rl.Matrix
}

func MakeInstance(x, y, z, axisX, axisY, axisZ, angle float32) Instance {
	translation := rl.MatrixTranslate(x, y, z)
	axis := rl.Vector3Normalize(rl.NewVector3(axisX, axisY, axisZ))
	rotation := rl.MatrixRotate(axis, angle*rl.Deg2rad)
	return Instance{translation: translation, rotation: rotation}
}

// Instances holds a list of Instance objects
type Instances struct {
	list []*Instance
}

func MakeInstances() *Instances {
	return &Instances{list: make([]*Instance, 0, MAX_INSTANCES)}
}

func (i *Instances) AddInstance(x, y, z, axisX, axisY, axisZ, angle float32) {
	instance := MakeInstance(x, y, z, axisX, axisY, axisZ, angle)
	i.list = append(i.list, &instance)
}

func (i *Instances) GetInstance(index int) *Instance {
	if index < 0 || index >= len(i.list) {
		return nil // Handle out of bounds
	}
	return i.list[index]
}

func AddRandomCubes(instances *Instances, count int) {
	for i := 0; i < count; i++ {
		x := float32(rl.GetRandomValue(-50, 50))
		y := float32(rl.GetRandomValue(-50, 50))
		z := float32(rl.GetRandomValue(-50, 50))

		xa := float32(rl.GetRandomValue(0, 360))
		ya := float32(rl.GetRandomValue(0, 360))
		za := float32(rl.GetRandomValue(0, 360))
		axis := rl.Vector3Normalize(rl.NewVector3(xa, ya, za))
		angle := float32(rl.GetRandomValue(0, 10)) * rl.Deg2rad

		instances.AddInstance(x, y, z, axis.X, axis.Y, axis.Z, angle)
	}
}

func main() {
	var (
		screenWidth   = int32(800) // Framebuffer width
		screenHeight  = int32(450) // Framebuffer height
		fps           = 60         // Frames per second
		framesCounter = 0
		instances     = MakeInstances()
		transforms    = make([]rl.Matrix, MAX_INSTANCES) // Transform matrices for instancing
	)

	rl.SetConfigFlags(rl.FlagMsaa4xHint) // Enable Multi Sampling Anti Aliasing 4x (if available)
	rl.InitWindow(screenWidth, screenHeight, "raylib [shaders] example - mesh instancing")

	context := MakeContext()

	// Scatter random cubes around
	AddRandomCubes(instances, MAX_INSTANCES)

	mvp := rl.GetShaderLocation(context.shader, "mvp")
	viewPos := rl.GetShaderLocation(context.shader, "viewPos")
	transform := rl.GetShaderLocationAttrib(context.shader, "instanceTransform")
	context.shader.UpdateLocation(rl.ShaderLocMatrixMvp, mvp)
	context.shader.UpdateLocation(rl.ShaderLocVectorView, viewPos)
	context.shader.UpdateLocation(rl.ShaderLocMatrixModel, transform)

	// ambient light level
	ambientLoc := rl.GetShaderLocation(context.shader, "ambient")
	rl.SetShaderValue(context.shader, ambientLoc, []float32{10.0, 10.0, 10.0, 10.0}, rl.ShaderUniformVec4)
	//NewLight(LightTypeDirectional, rl.NewVector3(50.0, 50.0, 0.0), rl.Vector3Zero(), rl.White, context.shader)

	context.material.Shader = context.shader
	mmap := context.material.GetMap(rl.MapDiffuse)
	mmap.Color = rl.Red

	rl.SetTargetFPS(int32(fps))
	for !rl.WindowShouldClose() {
		// Update
		//----------------------------------------------------------------------------------

		framesCounter++

		// Update the light shader with the camera view position
		rl.SetShaderValue(context.shader, context.shader.GetLocation(rl.ShaderLocVectorView),
			[]float32{context.camera.Position.X, context.camera.Position.Y, context.camera.Position.Z}, rl.ShaderUniformVec3)

		// Apply per-instance transformations
		for i := 0; i < MAX_INSTANCES; i++ {
			instance := instances.GetInstance(i)
			transforms[i] = rl.MatrixMultiply(instance.rotation, instance.translation)
			//transforms[i] = rl.MatrixMultiply(transforms[i], rl.MatrixTranslate(0.0, y, 0.0))
		}

		rl.UpdateCamera(&context.camera, rl.CameraOrbital) // Update camera with orbital camera mode
		//----------------------------------------------------------------------------------

		// Draw
		//----------------------------------------------------------------------------------
		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.RayWhite)

			rl.BeginMode3D(context.camera)
			//rl.DrawMesh(cube, material, rl.MatrixIdentity())
			rl.DrawMeshInstanced(context.cube, context.material, transforms, MAX_INSTANCES)
			rl.EndMode3D()

			rl.DrawFPS(10, 10)
		}
		rl.EndDrawing()
		//----------------------------------------------------------------------------------
	}

	// De-Initialization
	//--------------------------------------------------------------------------------------
	rl.CloseWindow() // Close window and OpenGL context
	//--------------------------------------------------------------------------------------
}
