package game

// Camera holds the world-space position of the screen center.
type Camera struct {
	X, Y float64 // world coordinates of screen center
}

func NewCamera() *Camera {
	return &Camera{}
}

// Follow centers the camera on the given world position.
func (c *Camera) Follow(worldX, worldY float64) {
	c.X = worldX
	c.Y = worldY
}

// WorldToScreen converts world coordinates to screen coordinates (scale applied).
func (c *Camera) WorldToScreen(worldX, worldY float64, screenW, screenH int) (float64, float64) {
	sx := (worldX-c.X)*Scale + float64(screenW)/2
	sy := (worldY-c.Y)*Scale + float64(screenH)/2
	return sx, sy
}
