package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const MAX_INSTANCES = 100000

type Instance struct {
	translation rl.Matrix
	rotation    rl.Matrix
}

func makeInstance(x, y, z, axisX, axisY, axisZ, angle float32) Instance {
	translation := rl.MatrixTranslate(x, y, z)
	axis := rl.Vector3Normalize(rl.NewVector3(axisX, axisY, axisZ))
	rotation := rl.MatrixRotate(axis, angle*rl.Deg2rad)
	return Instance{translation: translation, rotation: rotation}
}

func main() {
	var (
		screenWidth   = int32(800) // Framebuffer width
		screenHeight  = int32(450) // Framebuffer height
		fps           = 60         // Frames per second
		framesCounter = 0
		instances     = make([]*Instance, 0, MAX_INSTANCES) // Slice to hold instances
		transforms    = make([]rl.Matrix, MAX_INSTANCES)    // Transform matrices for instancing
	)

	rl.SetConfigFlags(rl.FlagMsaa4xHint) // Enable Multi Sampling Anti Aliasing 4x (if available)
	rl.InitWindow(screenWidth, screenHeight, "raylib [shaders] example - mesh instancing")

	// Define the camera to look into our 3d world
	camera := rl.Camera{
		Position:   rl.NewVector3(-100.0, 30.0, -50.0),
		Target:     rl.NewVector3(0.0, 0.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	cube := rl.GenMeshCube(1.0, 1.0, 1.0)

	// Scatter random cubes around
	for i := 0; i < MAX_INSTANCES; i++ {
		x := float32(rl.GetRandomValue(-50, 50))
		y := float32(rl.GetRandomValue(-50, 50))
		z := float32(rl.GetRandomValue(-50, 50))

		xa := float32(rl.GetRandomValue(0, 360))
		ya := float32(rl.GetRandomValue(0, 360))
		za := float32(rl.GetRandomValue(0, 360))
		axis := rl.Vector3Normalize(rl.NewVector3(xa, ya, za))
		angle := float32(rl.GetRandomValue(0, 10)) * rl.Deg2rad

		instance := makeInstance(x, y, z, axis.X, axis.Y, axis.Z, angle)
		instances = append(instances, &instance)
	}

	shader := rl.LoadShader("glsl330/base_lighting_instanced.vs", "glsl330/lighting.fs")
	shader.UpdateLocation(rl.ShaderLocMatrixMvp, rl.GetShaderLocation(shader, "mvp"))
	shader.UpdateLocation(rl.ShaderLocVectorView, rl.GetShaderLocation(shader, "viewPos"))
	shader.UpdateLocation(rl.ShaderLocMatrixModel, rl.GetShaderLocationAttrib(shader, "instanceTransform"))

	// ambient light level
	ambientLoc := rl.GetShaderLocation(shader, "ambient")
	rl.SetShaderValue(shader, ambientLoc, []float32{0.2, 0.2, 0.2, 1.0}, rl.ShaderUniformVec4)
	NewLight(LightTypeDirectional, rl.NewVector3(50.0, 50.0, 0.0), rl.Vector3Zero(), rl.White, shader)

	material := rl.LoadMaterialDefault()
	material.Shader = shader
	mmap := material.GetMap(rl.MapDiffuse)
	mmap.Color = rl.Red

	rl.SetTargetFPS(int32(fps))
	for !rl.WindowShouldClose() {
		// Update
		//----------------------------------------------------------------------------------

		framesCounter++

		// Update the light shader with the camera view position
		rl.SetShaderValue(shader, shader.GetLocation(rl.ShaderLocVectorView),
			[]float32{camera.Position.X, camera.Position.Y, camera.Position.Z}, rl.ShaderUniformVec3)

		// Apply per-instance transformations
		for i := 0; i < MAX_INSTANCES; i++ {
			instance := instances[i]
			transforms[i] = rl.MatrixMultiply(instance.rotation, instance.translation)
			//transforms[i] = rl.MatrixMultiply(transforms[i], rl.MatrixTranslate(0.0, y, 0.0))
		}

		rl.UpdateCamera(&camera, rl.CameraOrbital) // Update camera with orbital camera mode
		//----------------------------------------------------------------------------------

		// Draw
		//----------------------------------------------------------------------------------
		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.RayWhite)

			rl.BeginMode3D(camera)
			//rl.DrawMesh(cube, material, rl.MatrixIdentity())
			rl.DrawMeshInstanced(cube, material, transforms, MAX_INSTANCES)
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
