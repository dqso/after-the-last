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

// Draw overlays FOV darkness onto dst.
// memoryScreenTex must be screen-sized; its R channel holds memory strength [0, 1].
// Pixels outside the FOV are darkened proportionally to how much memory has faded.
func (f *FOVRenderer) Draw(dst *ebiten.Image, memoryScreenTex *ebiten.Image, playerSX, playerSY, dirAngle float64) {
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = memoryScreenTex
	opts.Uniforms = map[string]any{
		"PlayerScreenPos": []float32{float32(playerSX), float32(playerSY)},
		"DirAngle":        float32(dirAngle),
		"FOVHalfAngle":    float32(fovHalfAngle),
	}
	dst.DrawRectShader(w, h, f.shader, opts)
}
