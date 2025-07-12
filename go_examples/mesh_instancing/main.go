package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const MAX_INSTANCES = 5000

//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------

// holds camera, mesh, shader, and material needed for instance rendering
type Context struct {
	camera rl.Camera
	shader rl.Shader
}

func MakeContext() *Context {
	context := &Context{
		camera: rl.Camera{
			Position:   rl.NewVector3(-100.0, 30.0, -50.0),
			Target:     rl.NewVector3(0.0, 0.0, 0.0),
			Up:         rl.NewVector3(0.0, 1.0, 0.0),
			Fovy:       45.0,
			Projection: rl.CameraPerspective,
		},

		shader: rl.LoadShader("glsl330/base_lighting_instanced.vs", "glsl330/lighting.fs"),
	}

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

	return context
}

// ---------------------------------------------------------------------------------------------
// Model represents a 3D model with its mesh and material
type Model struct {
	mesh     rl.Mesh
	material rl.Material
}

func MakeModel(context *Context, mesh rl.Mesh, texture rl.Texture2D) *Model {
	rv := &Model{
		mesh:     mesh,
		material: rl.LoadMaterialDefault(),
	}
	rv.material.Shader = context.shader
	rv.material.Maps.Texture = texture
	rv.material.Maps.Color = rl.White
	rv.material.Maps.Value = 1.0
	mmap := rv.material.GetMap(rl.MapDiffuse)
	mmap.Color = rl.White
	return rv
}

//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------

// Instance represents a single drawn instance of a mesh with its transformation matrices
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

//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------

// Instances holds a list of Instance objects
type Instances struct {
	list []*Instance
}

func MakeInstances(max int) *Instances {
	return &Instances{list: make([]*Instance, 0, max)}
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

// ---------------------------------------------------------------------------------------------
func AddRandomInstancePositions(batch *Batch, count int) {

	var spread int32 = 500

	for i := 0; i < count; i++ {
		x := float32(rl.GetRandomValue(-spread, spread))
		y := float32(rl.GetRandomValue(-spread, spread))
		z := float32(rl.GetRandomValue(-spread, spread))

		// xa := float32(rl.GetRandomValue(0, 360))
		// ya := float32(rl.GetRandomValue(0, 360))
		// za := float32(rl.GetRandomValue(0, 360))
		// axis := rl.Vector3Normalize(rl.NewVector3(xa, ya, za))
		// angle := float32(rl.GetRandomValue(0, 10)) * rl.Deg2rad

		batch.AddInstance(x, y, z, 0, 0, 0, 0)
	}
}

// ---------------------------------------------------------------------------------------------
// Batch holds a model, a collection of draw instances and their transforms for rendering
type Batch struct {
	instances  *Instances
	transforms []rl.Matrix // Transform matrices for instancing
	model      *Model      // Model to be instanced
}

func MakeBatch(model *Model, maxInstances int) *Batch {
	return &Batch{
		instances:  MakeInstances(maxInstances),
		transforms: make([]rl.Matrix, maxInstances), // Initialize transform matrices
		model:      model,
	}
}

func (b *Batch) AddInstance(x, y, z, axisX, axisY, axisZ, angle float32) {
	b.instances.AddInstance(x, y, z, axisX, axisY, axisZ, angle)
}

func (b *Batch) MakeTransforms() {
	for i := 0; i < len(b.instances.list); i++ {
		instance := b.instances.GetInstance(i)
		if instance != nil {
			b.transforms[i] = rl.MatrixMultiply(instance.rotation, instance.translation)
		} else {
			b.transforms[i] = rl.MatrixIdentity() // Default identity matrix if instance is nil
		}
	}
}

//---------------------------------------------------------------------------------------------

// Create a procedural texture by drawing a circle on a render texture
func createCircleTexture(size int32, circleColor rl.Color, backgroundColor rl.Color) rl.Texture2D {

	renderTexture := rl.LoadRenderTexture(size, size)
	rl.BeginTextureMode(renderTexture)
	{
		rl.ClearBackground(backgroundColor)
		radius := float32(size) / 5.0
		center := rl.NewVector2(float32(size)/2.0, float32(size)/2.0)
		rl.DrawCircleV(center, radius, circleColor)
	}
	rl.EndTextureMode()
	image := rl.LoadImageFromTexture(renderTexture.Texture)
	texture := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	rl.UnloadRenderTexture(renderTexture)
	return texture
}

//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------
//---------------------------------------------------------------------------------------------

func main() {

	rl.SetConfigFlags(rl.FlagMsaa4xHint) // Enable Multi Sampling Anti Aliasing 4x (if available)
	rl.InitWindow(0, 0, "raylib [shaders] example - mesh instancing")
	rl.ToggleFullscreen()

	context := MakeContext()
	cube := rl.GenMeshCube(3.0, 3.0, 3.0)
	var batches []*Batch

	for range 10 {
		colour := rl.NewColor(uint8(rl.GetRandomValue(0, 255)), uint8(rl.GetRandomValue(0, 255)), uint8(rl.GetRandomValue(0, 255)), 255)
		texture := createCircleTexture(100, colour, rl.Black)
		model := MakeModel(context, cube, texture)
		batch := MakeBatch(model, MAX_INSTANCES)
		AddRandomInstancePositions(batch, MAX_INSTANCES)
		batch.MakeTransforms()
		batches = append(batches, batch)
	}

	for !rl.WindowShouldClose() {

		// Update the light shader with the camera view position
		//rl.SetShaderValue(context.shader, context.shader.GetLocation(rl.ShaderLocVectorView),
		//	[]float32{context.camera.Position.X, context.camera.Position.Y, context.camera.Position.Z}, rl.ShaderUniformVec3)

		// could update transform list to pass to DrawMeshInstanced
		// for i := 0; i < MAX_INSTANCES; i++ {
		// 	instance := instances.GetInstance(i)
		// 	transforms[i] = rl.MatrixMultiply(instance.rotation, instance.translation)
		// 	//transforms[i] = rl.MatrixMultiply(transforms[i], rl.MatrixTranslate(0.0, y, 0.0))
		// }

		rl.UpdateCamera(&context.camera, rl.CameraOrbital) // orbital camera movement

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode3D(context.camera)
			for _, batch := range batches {
				rl.DrawMeshInstanced(batch.model.mesh, batch.model.material, batch.transforms, MAX_INSTANCES)
			}
			rl.EndMode3D()

			rl.DrawFPS(10, 10)
		}
		rl.EndDrawing()

	}

	rl.CloseWindow() // Close window and OpenGL context

}
