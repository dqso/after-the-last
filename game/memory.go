package game

import (
	"math"

	"github.com/dqso/after-the-last/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	memoryDecayRate  = 0.8  // strength lost per second when outside FOV
	memoryFadeInRate = 10.0 // strength gained per second when inside FOV
)

// MemoryLayer maintains a world-space image tracking where the player has looked.
// Each pixel's R channel holds memory strength in [0, 1]:
//
//	1 = fully remembered, 0 = completely forgotten.
//
// Uses a ping-pong double buffer because a shader cannot read and write the same image.
type MemoryLayer struct {
	bufs   [2]*ebiten.Image
	cur    int
	shader *ebiten.Shader
}

func NewMemoryLayer(worldPixW, worldPixH int) (*MemoryLayer, error) {
	s, err := ebiten.NewShader(assets.MemoryUpdateShader)
	if err != nil {
		return nil, err
	}
	return &MemoryLayer{
		bufs: [2]*ebiten.Image{
			ebiten.NewImage(worldPixW, worldPixH),
			ebiten.NewImage(worldPixW, worldPixH),
		},
		shader: s,
	}, nil
}

// Update runs the memory shader: strengthens pixels inside the FOV cone,
// decays pixels outside it. eyeWX/eyeWY are world-pixel coordinates; dt is seconds.
// screenW/screenH bound the update to the currently visible area.
func (m *MemoryLayer) Update(eyeWX, eyeWY, dirAngle, dt float64, screenW, screenH int) {
	src := m.bufs[m.cur]
	dst := m.bufs[1-m.cur]
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

	// Limit memory updates to the screen-diagonal radius in world pixels.
	maxViewRadius := math.Sqrt(float64(screenW*screenW+screenH*screenH)) / (2.0 * Scale)

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]any{
		"EyeWorldPos":   []float32{float32(eyeWX), float32(eyeWY)},
		"DirAngle":      float32(dirAngle),
		"FOVHalfAngle":  float32(fovHalfAngle),
		"DecayDelta":    float32(memoryDecayRate * dt),
		"FadeInDelta":   float32(memoryFadeInRate * dt),
		"MaxViewRadius": float32(maxViewRadius),
	}
	dst.DrawRectShader(w, h, m.shader, opts)
	m.cur = 1 - m.cur
}

// Image returns the current memory texture (world-space, R channel = strength 0–1).
func (m *MemoryLayer) Image() *ebiten.Image { return m.bufs[m.cur] }

// DrawToScreen projects the world-space memory texture onto dst (screen-sized)
// using the camera transform so it aligns with the rendered world.
func (m *MemoryLayer) DrawToScreen(dst *ebiten.Image, cam *Camera, screenW, screenH int) {
	dst.Clear()
	hw := float64(screenW) / 2
	hh := float64(screenH) / 2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(Scale, Scale)
	op.GeoM.Translate(-cam.X*Scale+hw, -cam.Y*Scale+hh)
	dst.DrawImage(m.Image(), op)
}
