package game

import (
	"math"

	"github.com/dqso/after-the-last/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// fovHalfAngle is half the cone angle in radians (50 deg → total 100 deg FOV).
const fovHalfAngle = 50.0 * math.Pi / 180

// FOVRenderer draws a dark overlay with a cone cutout, revealing remembered areas.
type FOVRenderer struct {
	shader *ebiten.Shader
}

func NewFOVRenderer() (*FOVRenderer, error) {
	s, err := ebiten.NewShader(assets.FOVShader)
	if err != nil {
		return nil, err
	}
	return &FOVRenderer{shader: s}, nil
}

// Draw overlays FOV darkness onto dst. Both source images must be screen-sized
// (DrawRectShader requires every source image to match the rectangle):
//   - memoryScreenTex (image0): the memory layer projected to screen space.
//   - losScreenTex (image1): the line-of-sight mask projected to screen space.
//
// The shader applies only the cheap directional cone here; the wall ray march is
// already baked into the los mask by the visibility pass.
func (f *FOVRenderer) Draw(dst, memoryScreenTex, losScreenTex *ebiten.Image, playerSX, playerSY, dirAngle float64, screenW, screenH int) {
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = memoryScreenTex
	opts.Images[1] = losScreenTex
	opts.Uniforms = map[string]any{
		"PlayerScreenPos": []float32{float32(playerSX), float32(playerSY)},
		"DirAngle":        float32(dirAngle),
		"FOVHalfAngle":    float32(fovHalfAngle),
	}
	dst.DrawRectShader(screenW, screenH, f.shader, opts)
}

// projectWorldToScreen copies a world-pixel image into a screen-sized buffer
// using the camera transform, so it lines up with the rendered world. Nearest
// filtering (the default) keeps values crisp and BlendCopy preserves the raw
// RGBA for shader sampling.
func projectWorldToScreen(dst, worldImg *ebiten.Image, cam *Camera, screenW, screenH int) {
	dst.Clear()
	hw := float64(screenW) / 2
	hh := float64(screenH) / 2
	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendCopy
	op.GeoM.Scale(Scale, Scale)
	op.GeoM.Translate(-cam.X*Scale+hw, -cam.Y*Scale+hh)
	dst.DrawImage(worldImg, op)
}
