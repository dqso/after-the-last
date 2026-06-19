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

// MemoryLayer maintains a world-space color snapshot of what the player has seen.
// Each pixel stores:
//
//	RGB = the world color captured the last time it was inside the FOV cone
//	A   = memory strength in [0, 1] (1 = vivid, 0 = forgotten)
//
// Inside the cone the stored color tracks the live world; outside it stays
// frozen, so a moving object (e.g. a blinking button) keeps its last-seen color
// in memory even while it really changes. Uses a ping-pong double buffer because
// a shader cannot read and write the same image.
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

// Update runs the memory shader: inside the FOV cone it refreshes stored colors
// to the live world snapshot and strengthens memory; outside it freezes the
// colors and decays memory. worldColor is a 1:1 world-pixel snapshot of the
// current world (same size as the memory buffers). eyeWX/eyeWY are world-pixel
// coordinates; dt is seconds; screenW/screenH bound the update to the visible area.
func (m *MemoryLayer) Update(worldColor *ebiten.Image, eyeWX, eyeWY, dirAngle, dt float64, screenW, screenH int) {
	src := m.bufs[m.cur]
	dst := m.bufs[1-m.cur]
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

	// Limit memory updates to the screen-diagonal radius in world pixels.
	maxViewRadius := math.Sqrt(float64(screenW*screenW+screenH*screenH)) / (2.0 * Scale)

	opts := &ebiten.DrawRectShaderOptions{}
	// Copy raw shader output (straight RGBA) so color and strength survive the
	// ping-pong intact instead of being alpha-blended.
	opts.Blend = ebiten.BlendCopy
	opts.Images[0] = src        // previous memory (frozen color + strength)
	opts.Images[1] = worldColor // live world color snapshot
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

// Image returns the current memory texture (world-space; RGB = frozen color, A = strength).
func (m *MemoryLayer) Image() *ebiten.Image { return m.bufs[m.cur] }

// DrawToScreen projects the world-space memory texture onto dst (screen-sized)
// using the camera transform so it aligns with the rendered world. BlendCopy
// preserves the raw RGBA (straight color + strength) for the FOV shader.
func (m *MemoryLayer) DrawToScreen(dst *ebiten.Image, cam *Camera, screenW, screenH int) {
	dst.Clear()
	hw := float64(screenW) / 2
	hh := float64(screenH) / 2
	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendCopy
	op.GeoM.Scale(Scale, Scale)
	op.GeoM.Translate(-cam.X*Scale+hw, -cam.Y*Scale+hh)
	dst.DrawImage(m.Image(), op)
}
