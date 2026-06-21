package game

import (
	"math"

	"github.com/dqso/after-the-last/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// VisibilityRenderer computes the line-of-sight mask: a world-pixel texture whose
// red channel is 1 where a straight ray from the player's eye reaches the pixel
// within range (no wall in between). It is the single pass that performs the
// expensive ray march; the memory and FOV passes just read the mask and apply the
// cheap directional cone themselves.
type VisibilityRenderer struct {
	shader *ebiten.Shader
}

func NewVisibilityRenderer() (*VisibilityRenderer, error) {
	s, err := ebiten.NewShader(assets.VisibilityShader)
	if err != nil {
		return nil, err
	}
	return &VisibilityRenderer{shader: s}, nil
}

// Draw writes the line-of-sight mask into dst (must be world-pixel sized, matching
// collisionTex). collisionTex (image0) is the world-pixel sight-blocker texture.
// eyeWX/eyeWY are the eye in world-pixel coordinates; screenW/screenH bound the
// view radius to the visible area.
func (v *VisibilityRenderer) Draw(dst, collisionTex *ebiten.Image, eyeWX, eyeWY float64, screenW, screenH int) {
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

	// Limit visibility to the screen-diagonal radius in world pixels.
	maxViewRadius := math.Sqrt(float64(screenW*screenW+screenH*screenH)) / (2.0 * Scale)

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Blend = ebiten.BlendCopy // keep the raw mask value, no alpha blending
	opts.Images[0] = collisionTex
	opts.Uniforms = map[string]any{
		"EyeWorldPos":   []float32{float32(eyeWX), float32(eyeWY)},
		"MaxViewRadius": float32(maxViewRadius),
	}
	dst.DrawRectShader(w, h, v.shader, opts)
}
