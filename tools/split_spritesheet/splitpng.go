package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FrameRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type FrameDef struct {
	Filename      string    `json:"filename"`
	AnimFrames    int       `json:"anim_frames"`
	TicksPerFrame int       `json:"ticks_per_frame"`
	Frame         FrameRect `json:"frame"`
}

type SpriteSheet struct {
	Frames []FrameDef `json:"frames"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: splitpng <input.png> <input.json> [output_dir]")
		os.Exit(1)
	}
	pngPath := os.Args[1]
	jsonPath := os.Args[2]
	outDir := "."
	if len(os.Args) > 3 {
		outDir = os.Args[3]
	}

	// Read JSON
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}
	var sheet SpriteSheet
	if err := json.Unmarshal(data, &sheet); err != nil {
		panic(err)
	}

	// Open PNG
	f, err := os.Open(pngPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	// Split frames
	for _, frame := range sheet.Frames {
		frames := frame.AnimFrames
		if frames < 1 {
			frames = 1
		}
		frameWidth := frame.Frame.W / frames
		frameHeight := frame.Frame.H

		// Create a new RGBA image to hold all animation frames horizontally
		outImg := image.NewRGBA(image.Rect(0, 0, frameWidth*frames, frameHeight))

		for i := 0; i < frames; i++ {
			srcRect := image.Rect(
				frame.Frame.X+i*frameWidth,
				frame.Frame.Y,
				frame.Frame.X+(i+1)*frameWidth,
				frame.Frame.Y+frameHeight,
			)
			dstPt := image.Pt(i*frameWidth, 0)
			draw.Draw(outImg, image.Rect(dstPt.X, dstPt.Y, dstPt.X+frameWidth, dstPt.Y+frameHeight), img, srcRect.Min, draw.Src)
		}

		outPath := filepath.Join(outDir, frame.Filename)
		outF, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(outF, outImg); err != nil {
			panic(err)
		}
		outF.Close()
		fmt.Printf("Wrote %s\n", outPath)
	}
}
