package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// OrbitCamera manages 3D orbital viewpoint and smooth body tracking
type OrbitCamera struct {
	Camera    rl.Camera3D
	Target    rl.Vector3
	Distance  float32
	Azimuth   float32 // Yaw angle in radians
	Elevation float32 // Pitch angle in radians
	MinDist   float32
	MaxDist   float32
}

// NewOrbitCamera creates a default orbital camera
func NewOrbitCamera() *OrbitCamera {
	cam := &OrbitCamera{
		Distance:  120.0,
		Azimuth:   0.8,
		Elevation: 0.45,
		MinDist:   2.0,
		MaxDist:   1500.0,
	}

	cam.Camera = rl.Camera3D{
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}
	cam.UpdatePosition()
	return cam
}

// UpdatePosition recalculates Camera.Position from Target, Azimuth, Elevation, Distance
func (c *OrbitCamera) UpdatePosition() {
	cosElev := float32(math.Cos(float64(c.Elevation)))
	sinElev := float32(math.Sin(float64(c.Elevation)))
	sinAzim := float32(math.Sin(float64(c.Azimuth)))
	cosAzim := float32(math.Cos(float64(c.Azimuth)))

	c.Camera.Target = c.Target
	c.Camera.Position = rl.NewVector3(
		c.Target.X+c.Distance*cosElev*sinAzim,
		c.Target.Y+c.Distance*sinElev,
		c.Target.Z+c.Distance*cosElev*cosAzim,
	)
}

// Update handles mouse and keyboard inputs for camera rotation, zoom, pan, and tracking
func (c *OrbitCamera) Update(state *SimState, uiHandledMouse bool) {
	// If following a selected body, smoothly interpolate target
	if state.FollowSelected && state.SelectedBodyID != -1 {
		var selected *Body
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				selected = b
				break
			}
		}
		if selected != nil {
			c.Target.X += (selected.Position.X - c.Target.X) * 0.15
			c.Target.Y += (selected.Position.Y - c.Target.Y) * 0.15
			c.Target.Z += (selected.Position.Z - c.Target.Z) * 0.15
		} else {
			state.FollowSelected = false
		}
	}

	// Zoom with mouse wheel (if mouse not consumed by UI)
	if !uiHandledMouse {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			c.Distance *= float32(math.Pow(0.88, float64(wheel)))
			if c.Distance < c.MinDist {
				c.Distance = c.MinDist
			}
			if c.Distance > c.MaxDist {
				c.Distance = c.MaxDist
			}
		}

		// Rotate view with Right Mouse Button drag
		if rl.IsMouseButtonDown(rl.MouseRightButton) {
			mouseDelta := rl.GetMouseDelta()
			c.Azimuth -= mouseDelta.X * 0.006
			c.Elevation += mouseDelta.Y * 0.006

			// Clamp pitch to avoid flipping
			const maxElev = 1.52 // ~87 degrees
			if c.Elevation > maxElev {
				c.Elevation = maxElev
			} else if c.Elevation < -maxElev {
				c.Elevation = -maxElev
			}
		}

		// Pan target with Middle Mouse Button drag or Shift + Right Mouse Button
		if rl.IsMouseButtonDown(rl.MouseMiddleButton) || (rl.IsKeyDown(rl.KeyLeftShift) && rl.IsMouseButtonDown(rl.MouseRightButton)) {
			state.FollowSelected = false // Unlink follow on manual pan
			mouseDelta := rl.GetMouseDelta()
			panSpeed := c.Distance * 0.0015

			// Right vector: cross product of forward and world up
			cosAzim := float32(math.Cos(float64(c.Azimuth)))
			sinAzim := float32(math.Sin(float64(c.Azimuth)))

			rightX := cosAzim
			rightZ := -sinAzim

			c.Target.X -= (rightX * mouseDelta.X) * panSpeed
			c.Target.Z -= (rightZ * mouseDelta.X) * panSpeed
			c.Target.Y += mouseDelta.Y * panSpeed
		}
	}

	// Keyboard controls for pan
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X -= sinAzim * speed
		c.Target.Z -= cosAzim * speed
		state.FollowSelected = false
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X += sinAzim * speed
		c.Target.Z += cosAzim * speed
		state.FollowSelected = false
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X -= cosAzim * speed
		c.Target.Z += sinAzim * speed
		state.FollowSelected = false
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X += cosAzim * speed
		c.Target.Z -= sinAzim * speed
		state.FollowSelected = false
	}
	if rl.IsKeyDown(rl.KeyE) || rl.IsKeyDown(rl.KeyPageUp) {
		c.Target.Y += c.Distance * 0.01
		state.FollowSelected = false
	}
	if rl.IsKeyDown(rl.KeyQ) || rl.IsKeyDown(rl.KeyPageDown) {
		c.Target.Y -= c.Distance * 0.01
		state.FollowSelected = false
	}

	// Focus shortcut (F)
	if rl.IsKeyPressed(rl.KeyF) && state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				c.Target = b.Position
				break
			}
		}
	}

	// Reset camera view (R)
	if rl.IsKeyPressed(rl.KeyR) {
		c.Target = rl.NewVector3(0, 0, 0)
		c.Distance = 120.0
		c.Azimuth = 0.8
		c.Elevation = 0.45
		state.FollowSelected = false
	}

	c.UpdatePosition()
}
