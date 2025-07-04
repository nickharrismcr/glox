package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var screenWidth int
var screenHeight int

const textureSpeed = 1
const imagePath = "img.jpg"
const segmentCount = 256

type Triangle struct {
	P1 rl.Vector2 // Top vertex
	P2 rl.Vector2 // Bottom left vertex
	P3 rl.Vector2 // Bottom right vertex
}

// TextureSampler holds texture info and sampling rectangle
type TextureSampler struct {
	TextureWidth  float32
	TextureHeight float32
	SampleX       float32
	SampleY       float32
	SampleWidth   float32
	SampleHeight  float32
	VelX          float32
	VelY          float32
}

func makeTextureSampler(texture rl.Texture2D, sampleWidth, sampleHeight float32) *TextureSampler {
	// Initialize texture sampler
	return &TextureSampler{
		TextureWidth:  float32(texture.Width),
		TextureHeight: float32(texture.Height),
		SampleX:       0,
		SampleY:       0,
		SampleWidth:   sampleWidth,
		SampleHeight:  sampleHeight,
		VelX:          textureSpeed,
		VelY:          textureSpeed,
	}
}

func (t *TextureSampler) Update() {

	// Update position
	t.SampleX += t.VelX
	t.SampleY += t.VelY

	// Bounce off left/right edges
	if t.SampleX <= 0 {
		t.SampleX = 0
		t.VelX = -t.VelX
	} else if t.SampleX >= t.TextureWidth-t.SampleWidth {
		t.SampleX = t.TextureWidth - t.SampleWidth
		t.VelX = -t.VelX
	}

	// Bounce off top/bottom edges
	if t.SampleY <= 0 {
		t.SampleY = 0
		t.VelY = -t.VelY
	} else if t.SampleY >= t.TextureHeight-t.SampleHeight {
		t.SampleY = t.TextureHeight - t.SampleHeight
		t.VelY = -t.VelY
	}
}

func (t *TextureSampler) Get() rl.Rectangle {
	// Always return the same rectangle - we'll handle flipping with transforms
	return rl.Rectangle{
		X:      t.SampleX,
		Y:      t.SampleY,
		Width:  t.SampleWidth,
		Height: t.SampleHeight,
	}
}

func drawSegment(renderTexture rl.RenderTexture2D,
	texture rl.Texture2D,
	sampler *TextureSampler,
	position rl.Vector2,
	rotation float32,
	flip bool) {

	// Step 1: Create the triangle mask in the render texture
	rl.BeginTextureMode(renderTexture)
	rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0}) // Clear with transparent

	v1 := rl.Vector2{X: sampler.SampleWidth / 2, Y: -12}
	v2 := rl.Vector2{X: -5, Y: sampler.SampleHeight + 5}
	v3 := rl.Vector2{X: sampler.SampleWidth + 5, Y: sampler.SampleHeight + 5}

	// Draw the main triangle
	rl.DrawTriangle(v1, v2, v3, rl.White)

	// Then multiply the sampled texture with this mask
	rl.BeginBlendMode(rl.BlendMultiplied)
	sourceRect := sampler.Get()
	if flip {
		sourceRect.Width = -sourceRect.Width
	}
	destRect := rl.Rectangle{X: 0, Y: 0, Width: sampler.SampleWidth, Height: sampler.SampleHeight}
	rl.DrawTexturePro(texture, sourceRect, destRect, rl.Vector2{X: 0, Y: 0}, 0, rl.White)
	rl.EndBlendMode()
	rl.EndTextureMode()

	renderDestRect := rl.Rectangle{X: position.X, Y: position.Y, Width: sampler.SampleWidth, Height: sampler.SampleHeight}
	origin := rl.Vector2{X: sampler.SampleWidth / 2.0, Y: 0} // Origin at the tip position
	renderSourceRect := rl.Rectangle{X: 0, Y: 0, Width: sampler.SampleWidth, Height: -sampler.SampleHeight}
	rl.DrawTexturePro(renderTexture.Texture, renderSourceRect, renderDestRect, origin, rotation, rl.White)
}

func getWidth(segmentCount int, screenHeight int) (float32, float32) {
	// Calculate triangle dimensions based on segment count
	segmentAngle := 2.0 * math.Pi / float32(segmentCount)
	radius := float32(screenHeight)
	triangleWidth := 2.0 * radius * float32(math.Tan(float64(segmentAngle)/2.0))
	triangleWidth *= 1.1 // generous overlap
	return radius, triangleWidth
}

func kaleido(segmentCount int, centerX, centerY float32, texture rl.Texture2D, sampler *TextureSampler, renderTexture rl.RenderTexture2D) {

	for i := 0; i < segmentCount; i++ {
		angle := float32(i) * 2.0 * math.Pi / float32(segmentCount)
		segmentRotation := angle * 180.0 / math.Pi // Convert to degrees
		segmentRotation += 20
		flip := (i % 2) == 0
		drawSegment(renderTexture, texture, sampler, rl.Vector2{X: centerX, Y: centerY}, segmentRotation, flip)
	}
}

func main() {
	// Initialize window
	rl.InitWindow(0, 0, "Triangle Segment Demo")
	rl.ToggleFullscreen()
	screenHeight = rl.GetScreenHeight() // Update height after fullscreen toggle
	screenWidth = rl.GetScreenWidth()   // Update width after fullscreen toggle
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Load image and create texture
	image := rl.LoadImage(imagePath)
	defer rl.UnloadImage(image)
	texture := rl.LoadTextureFromImage(image)
	defer rl.UnloadTexture(texture)

	radius, triangleWidth := getWidth(segmentCount, screenHeight)
	sampler := makeTextureSampler(texture, triangleWidth, radius)
	renderTexture := rl.LoadRenderTexture(int32(sampler.SampleWidth), int32(sampler.SampleHeight))
	defer rl.UnloadRenderTexture(renderTexture)

	centerX := float32(screenWidth / 2)
	centerY := float32(screenHeight / 2)

	for !rl.WindowShouldClose() {

		sampler.Update()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		kaleido(segmentCount, centerX, centerY, texture, sampler, renderTexture)

		rl.EndDrawing()
	}
}
