package game

import (
	"math"

	"github.com/dqso/after-the-last/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

// alphaQuantum is one step of the 8-bit alpha channel the memory strength is
// stored in. A per-frame decay smaller than this would round back to the same
// stored value and freeze the memory forever, so decay is accumulated and only
// applied in whole multiples of this step (see Update).
const alphaQuantum = 1.0 / 255.0

const (
	memoryFadeInRate = 6.0 // strength gained per second when inside FOV (higher = remembered after a shorter glance)

	// memoryInitialDim is the fraction of brightness lost the instant a pixel
	// leaves the FOV cone (0.4 = darken by 40% immediately). The remaining
	// strength then decays toward 0 (full black) at a rate set by the player's
	// memory stat (see memoryDecayRate).
	memoryInitialDim = 0.3

	// memoryMidForgetSeconds is the time to fully forget a remembered area when
	// the memory stat sits at its midpoint (50). The forgetting time scales
	// hyperbolically with the stat, so this value anchors the overall pace.
	memoryMidForgetSeconds = 4.0
)

// memoryDecayRate maps the player's memory stat [0,100] to a strength-decay rate
// (memory strength lost per second for out-of-FOV pixels). It is a hyperbola in
// the "time to forget" domain:
//
//	stat = 0   -> instant clear  (decay rate +Inf; memory does not work at all)
//	stat = 50  -> forgets over memoryMidForgetSeconds
//	stat = 100 -> never forgets  (decay rate 0)
//
// Forgetting time T(stat) = memoryMidForgetSeconds * stat / (100 - stat); the
// decay rate is full strength (1.0) divided by that time. The curve stays gentle
// near the midpoint and shoots up steeply only as the stat approaches 0 or 100.
func memoryDecayRate(stat float64) float64 {
	stat = math.Max(0, math.Min(100, stat))
	switch {
	case stat >= 100:
		return 0 // never forgets
	case stat <= 0:
		return math.Inf(1) // instant clear
	}
	return 1.0 / memoryForgetSeconds(stat)
}

// memoryForgetSeconds returns how long (seconds) a remembered area takes to fade
// fully to black at the given memory stat: 0 at stat=0 (instant) up to +Inf at
// stat=100 (never). Mainly used for debug readout.
func memoryForgetSeconds(stat float64) float64 {
	stat = math.Max(0, math.Min(100, stat))
	switch {
	case stat >= 100:
		return math.Inf(1)
	case stat <= 0:
		return 0
	}
	return memoryMidForgetSeconds * stat / (100 - stat)
}

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

	// decayAccum carries fractional decay between frames so that sub-quantum
	// per-frame decay still adds up to a real, applied alpha step over time.
	decayAccum float64
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
// current world (same size as the memory buffers). losTex is the world-pixel
// line-of-sight mask (image2) from the visibility pass: the shader multiplies the
// cone by it so areas hidden behind walls (or out of range) are not learned.
// eyeWX/eyeWY are world-pixel coordinates; dt is seconds; memoryStat is the
// player's [0,100] memory stat controlling forgetting speed.
func (m *MemoryLayer) Update(worldColor, losTex *ebiten.Image, eyeWX, eyeWY, dirAngle, dt, memoryStat float64) {
	src := m.bufs[m.cur]
	dst := m.bufs[1-m.cur]
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()

	// Strength lost this frame. Accumulate fractional decay and only emit it in
	// whole alpha-quantum steps; otherwise a sub-quantum delta rounds away when
	// written back to the 8-bit buffer and the memory freezes (never forgets).
	rate := memoryDecayRate(memoryStat)
	var decayDelta float64
	if math.IsInf(rate, 1) {
		decayDelta = 1.0 // stat 0: clear memory in a single frame
		m.decayAccum = 0
	} else {
		m.decayAccum += rate * dt
		if steps := math.Floor(m.decayAccum / alphaQuantum); steps >= 1 {
			decayDelta = steps * alphaQuantum
			m.decayAccum -= decayDelta
		}
	}

	opts := &ebiten.DrawRectShaderOptions{}
	// Copy raw shader output (straight RGBA) so color and strength survive the
	// ping-pong intact instead of being alpha-blended.
	opts.Blend = ebiten.BlendCopy
	opts.Images[0] = src        // previous memory (frozen color + strength)
	opts.Images[1] = worldColor // live world color snapshot
	opts.Images[2] = losTex     // line-of-sight mask
	opts.Uniforms = map[string]any{
		"EyeWorldPos":       []float32{float32(eyeWX), float32(eyeWY)},
		"DirAngle":          float32(dirAngle),
		"FOVHalfAngle":      float32(fovHalfAngle),
		"DecayDelta":        float32(decayDelta),
		"FadeInDelta":       float32(memoryFadeInRate * dt),
		"MemoryMaxStrength": float32(1.0 - memoryInitialDim),
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
	projectWorldToScreen(dst, m.Image(), cam, screenW, screenH)
}
