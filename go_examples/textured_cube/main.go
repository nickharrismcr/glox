package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Initialize window
	rl.InitWindow(800, 600, "Textured Cube Example")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Load texture from defender pngs
	texture := rl.LoadTexture("lander.png")
	if texture.ID == 0 {
		log.Fatal("Failed to load texture")
	}
	defer rl.UnloadTexture(texture)

	// Setup 3D camera
	camera := rl.Camera3D{
		Position:   rl.Vector3{X: 4.0, Y: 4.0, Z: 4.0},
		Target:     rl.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
		Up:         rl.Vector3{X: 0.0, Y: 1.0, Z: 0.0},
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	// Cube rotation
	rotation := float32(0.0)
	// Main game loop - modify the drawing section:
	for !rl.WindowShouldClose() {
		// Update
		rotation += 1.0

		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		rl.PushMatrix()
		//rl.Translatef(x, y, z)
		rl.Rotatef(rotation, 1.0, 1.0, 0.0) // Rotate around diagonal axis

		// First pass: Draw black cube (background for transparent areas)
		rl.DrawCube(rl.Vector3{X: 0, Y: 0, Z: 0}, 1.99, 1.99, 1.99, rl.Black)

		// Second pass: Draw textured cube with alpha blending
		rl.BeginBlendMode(rl.BlendAlpha)
		DrawTexturedCube(texture, rl.Vector3{X: 0, Y: 0, Z: 0}, 2.0, 2.0, 2.0, rl.White)
		rl.EndBlendMode()
		rl.PopMatrix()
		// Draw reference objects
		rl.DrawGrid(10, 1.0)

		rl.EndMode3D()

		// Draw UI
		rl.DrawText("Textured Cube Example", 10, 10, 20, rl.DarkGray)
		rl.DrawText("Cube rotates automatically", 10, 40, 16, rl.Gray)
		rl.DrawFPS(10, 580)

		rl.EndDrawing()
	}
}

// DrawTexturedCube draws a cube with texture on all faces
func DrawTexturedCube(texture rl.Texture2D, position rl.Vector3, width, height, length float32, tint rl.Color) {

	// Set texture and enable texturing
	rl.SetTexture(texture.ID)

	rl.Begin(rl.Quads)
	rl.Color4ub(tint.R, tint.G, tint.B, tint.A)

	// Front Face
	rl.Normal3f(0.0, 0.0, 1.0)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(-width/2, -height/2, length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(width/2, -height/2, length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(width/2, height/2, length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(-width/2, height/2, length/2)

	// Back Face
	rl.Normal3f(0.0, 0.0, -1.0)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(-width/2, -height/2, -length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(-width/2, height/2, -length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(width/2, height/2, -length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(width/2, -height/2, -length/2)

	// Top Face
	rl.Normal3f(0.0, 1.0, 0.0)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(-width/2, height/2, -length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(-width/2, height/2, length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(width/2, height/2, length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(width/2, height/2, -length/2)

	// Bottom Face
	rl.Normal3f(0.0, -1.0, 0.0)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(-width/2, -height/2, -length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(width/2, -height/2, -length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(width/2, -height/2, length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(-width/2, -height/2, length/2)

	// Right Face
	rl.Normal3f(1.0, 0.0, 0.0)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(width/2, -height/2, -length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(width/2, height/2, -length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(width/2, height/2, length/2)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(width/2, -height/2, length/2)

	// Left Face
	rl.Normal3f(-1.0, 0.0, 0.0)
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(-width/2, -height/2, -length/2)
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(-width/2, -height/2, length/2)
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(-width/2, height/2, length/2)
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(-width/2, height/2, -length/2)

	rl.End()

	// Disable texturing
	rl.SetTexture(0)
}
