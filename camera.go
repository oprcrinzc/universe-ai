package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// OrbitCamera manages 3D orbital viewpoint and smooth body tracking
type OrbitCamera struct {
	Camera      rl.Camera3D
	Target      rl.Vector3
	Distance    float32
	Azimuth     float32 // Yaw angle in radians
	Elevation   float32 // Pitch angle in radians
	MinDist     float32
	MaxDist     float32
	CurrentView string // "top", "side", "front", "iso", or ""
}

// NewOrbitCamera creates a default orbital camera
func NewOrbitCamera() *OrbitCamera {
	cam := &OrbitCamera{
		Distance:    120.0,
		Azimuth:     0.8,
		Elevation:   0.45,
		MinDist:     2.0,
		MaxDist:     1500.0,
		CurrentView: "iso",
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
func (c *OrbitCamera) Update(state *SimState, cfg *Config, uiHandledMouse bool) {
	// Cinematic Camera Auto-Orbit Mode
	if cfg.CinematicCamera {
		state.CinematicTime += 0.016
		c.Azimuth += 0.003
		c.Elevation = 0.42 + 0.10*float32(math.Sin(float64(state.CinematicTime)*0.25))

		if !state.FollowSelected && len(state.Bodies) > 0 {
			var com rl.Vector3
			var totM float64
			for _, b := range state.Bodies {
				if b.Mass >= 20.0 || b.IsStar || b.TextureType == TextureBlackHole {
					com.X += b.Position.X * float32(b.Mass)
					com.Y += b.Position.Y * float32(b.Mass)
					com.Z += b.Position.Z * float32(b.Mass)
					totM += b.Mass
				}
			}
			if totM > 0 {
				invM := float32(1.0 / totM)
				tX := com.X * invM
				tY := com.Y * invM
				tZ := com.Z * invM
				c.Target.X += (tX - c.Target.X) * 0.05
				c.Target.Y += (tY - c.Target.Y) * 0.05
				c.Target.Z += (tZ - c.Target.Z) * 0.05
			}
		}
	}

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
	} else if state.FollowSelected && state.SelectedLagrangeIndex > 0 {
		primary, secondary := GetLagrangePair(state)
		if primary != nil && secondary != nil {
			pts := ComputeLagrangePoints(primary, secondary, cfg.G)
			idx := state.SelectedLagrangeIndex - 1
			if idx >= 0 && idx < 5 {
				ptPos := pts[idx].Position
				c.Target.X += (ptPos.X - c.Target.X) * 0.15
				c.Target.Y += (ptPos.Y - c.Target.Y) * 0.15
				c.Target.Z += (ptPos.Z - c.Target.Z) * 0.15
			}
		}
	} else if state.FollowBarycenter && len(state.Bodies) > 0 {
		var com rl.Vector3
		var totM float64
		for _, b := range state.Bodies {
			com.X += b.Position.X * float32(b.Mass)
			com.Y += b.Position.Y * float32(b.Mass)
			com.Z += b.Position.Z * float32(b.Mass)
			totM += b.Mass
		}
		if totM > 0 {
			invM := float32(1.0 / totM)
			tX := com.X * invM
			tY := com.Y * invM
			tZ := com.Z * invM
			c.Target.X += (tX - c.Target.X) * 0.15
			c.Target.Y += (tY - c.Target.Y) * 0.15
			c.Target.Z += (tZ - c.Target.Z) * 0.15
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
			if mouseDelta.X != 0 || mouseDelta.Y != 0 {
				c.CurrentView = ""
			}
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
			state.FollowBarycenter = false
			c.CurrentView = ""
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
		state.FollowBarycenter = false
		c.CurrentView = ""
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X += sinAzim * speed
		c.Target.Z += cosAzim * speed
		state.FollowSelected = false
		state.FollowBarycenter = false
		c.CurrentView = ""
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X -= cosAzim * speed
		c.Target.Z += sinAzim * speed
		state.FollowSelected = false
		state.FollowBarycenter = false
		c.CurrentView = ""
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		cosAzim := float32(math.Cos(float64(c.Azimuth)))
		sinAzim := float32(math.Sin(float64(c.Azimuth)))
		speed := c.Distance * 0.01
		c.Target.X += cosAzim * speed
		c.Target.Z -= sinAzim * speed
		state.FollowSelected = false
		state.FollowBarycenter = false
		c.CurrentView = ""
	}
	if rl.IsKeyDown(rl.KeyE) || rl.IsKeyDown(rl.KeyPageUp) {
		c.Target.Y += c.Distance * 0.01
		state.FollowSelected = false
		state.FollowBarycenter = false
		c.CurrentView = ""
	}
	if rl.IsKeyDown(rl.KeyQ) || rl.IsKeyDown(rl.KeyPageDown) {
		c.Target.Y -= c.Distance * 0.01
		state.FollowSelected = false
		state.FollowBarycenter = false
		c.CurrentView = ""
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
		c.Reset(state, cfg)
	}

	// Lock to nearest body shortcut (N)
	if rl.IsKeyPressed(rl.KeyN) {
		c.LockToNearest(state, cfg)
	}

	c.UpdatePosition()
}

// Reset restores camera position, target, distance, angles, and FOV to origin defaults
func (c *OrbitCamera) Reset(state *SimState, cfg *Config) {
	c.Target = rl.NewVector3(0, 0, 0)
	c.Distance = 120.0
	c.Azimuth = 0.8
	c.Elevation = 0.45
	c.Camera.Fovy = 45.0
	c.CurrentView = "iso"
	if state != nil {
		state.SelectedBodyID = -1
		state.FollowSelected = false
		state.FollowBarycenter = false
	}
	if cfg != nil {
		cfg.CinematicCamera = false
		cfg.NotificationText = "Camera view reset to origin"
		cfg.NotificationTimer = 2.5
	}
	c.UpdatePosition()
}

// LockToNearest finds the celestial body closest to the camera and smoothly tracks it.
// If already locked to a body, it finds the next nearest neighbor to easily cycle through bodies.
func (c *OrbitCamera) LockToNearest(state *SimState, cfg *Config) *Body {
	if state == nil || len(state.Bodies) == 0 {
		if cfg != nil {
			cfg.NotificationText = "No celestial bodies in scene to lock on"
			cfg.NotificationTimer = 2.0
		}
		return nil
	}

	camPos := c.Camera.Position
	var nearest *Body
	var minDist float32 = float32(math.MaxFloat32)

	// If already following a body and multiple bodies exist, skip the currently followed one
	skipCurrent := state.FollowSelected && state.SelectedBodyID != -1 && len(state.Bodies) > 1

	for _, b := range state.Bodies {
		if skipCurrent && b.ID == state.SelectedBodyID {
			continue
		}
		dx := b.Position.X - camPos.X
		dy := b.Position.Y - camPos.Y
		dz := b.Position.Z - camPos.Z
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist < minDist {
			minDist = dist
			nearest = b
		}
	}

	// Fallback to currently selected body if none other was picked
	if nearest == nil {
		for _, b := range state.Bodies {
			dx := b.Position.X - camPos.X
			dy := b.Position.Y - camPos.Y
			dz := b.Position.Z - camPos.Z
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
			if dist < minDist {
				minDist = dist
				nearest = b
			}
		}
	}

	if nearest != nil {
		state.SelectedBodyID = nearest.ID
		state.FollowSelected = true
		state.FollowBarycenter = false
		c.Target = nearest.Position

		// Comfortably adjust distance based on body radius
		idealDist := nearest.Radius * 5.0
		if idealDist < 10.0 {
			idealDist = 10.0
		}
		if c.Distance > idealDist*6.0 || c.Distance < nearest.Radius*1.5 {
			c.Distance = idealDist * 2.0
			if c.Distance > c.MaxDist {
				c.Distance = c.MaxDist
			}
			if c.Distance < c.MinDist {
				c.Distance = c.MinDist
			}
		}

		if cfg != nil {
			cfg.NotificationText = fmt.Sprintf("Locked Camera -> %s (Dist: %.1f)", nearest.Name, minDist)
			cfg.NotificationTimer = 2.5
		}
		c.UpdatePosition()
	}

	return nearest
}

// LockToHeaviest finds the body with maximum mass (star, black hole) and locks camera to it
func (c *OrbitCamera) LockToHeaviest(state *SimState, cfg *Config) *Body {
	if state == nil || len(state.Bodies) == 0 {
		if cfg != nil {
			cfg.NotificationText = "No celestial bodies in scene"
			cfg.NotificationTimer = 2.0
		}
		return nil
	}

	var heaviest *Body
	var maxMass float64 = -1.0

	for _, b := range state.Bodies {
		if b.Mass > maxMass {
			maxMass = b.Mass
			heaviest = b
		}
	}

	if heaviest != nil {
		state.SelectedBodyID = heaviest.ID
		state.FollowSelected = true
		state.FollowBarycenter = false
		c.Target = heaviest.Position

		idealDist := heaviest.Radius * 5.0
		if idealDist < 15.0 {
			idealDist = 15.0
		}
		if c.Distance > idealDist*6.0 || c.Distance < heaviest.Radius*1.5 {
			c.Distance = idealDist * 2.5
		}

		if cfg != nil {
			cfg.NotificationText = fmt.Sprintf("Locked Camera -> %s (Mass: %.1f)", heaviest.Name, heaviest.Mass)
			cfg.NotificationTimer = 2.5
		}
		c.UpdatePosition()
	}

	return heaviest
}

// LockToBarycenter tracks the center of mass of the celestial system
func (c *OrbitCamera) LockToBarycenter(state *SimState, cfg *Config) {
	if state == nil || len(state.Bodies) == 0 {
		if cfg != nil {
			cfg.NotificationText = "No celestial bodies in scene"
			cfg.NotificationTimer = 2.0
		}
		return
	}

	state.FollowBarycenter = true
	state.FollowSelected = false

	var com rl.Vector3
	var totM float64
	for _, b := range state.Bodies {
		com.X += b.Position.X * float32(b.Mass)
		com.Y += b.Position.Y * float32(b.Mass)
		com.Z += b.Position.Z * float32(b.Mass)
		totM += b.Mass
	}
	if totM > 0 {
		invM := float32(1.0 / totM)
		c.Target = rl.NewVector3(com.X*invM, com.Y*invM, com.Z*invM)
	}

	if cfg != nil {
		cfg.NotificationText = "Tracking System Barycenter (Center of Mass)"
		cfg.NotificationTimer = 2.5
	}
	c.UpdatePosition()
}

// FrameAll fits all celestial bodies currently in the scene into the camera frustum
func (c *OrbitCamera) FrameAll(state *SimState, cfg *Config) {
	if state == nil || len(state.Bodies) == 0 {
		c.Reset(state, cfg)
		return
	}

	var minX, maxX float32 = 1e9, -1e9
	var minY, maxY float32 = 1e9, -1e9
	var minZ, maxZ float32 = 1e9, -1e9

	for _, b := range state.Bodies {
		if b.Position.X < minX {
			minX = b.Position.X
		}
		if b.Position.X > maxX {
			maxX = b.Position.X
		}
		if b.Position.Y < minY {
			minY = b.Position.Y
		}
		if b.Position.Y > maxY {
			maxY = b.Position.Y
		}
		if b.Position.Z < minZ {
			minZ = b.Position.Z
		}
		if b.Position.Z > maxZ {
			maxZ = b.Position.Z
		}
	}

	center := rl.NewVector3(
		(minX+maxX)*0.5,
		(minY+maxY)*0.5,
		(minZ+maxZ)*0.5,
	)

	var maxR float32 = 10.0
	for _, b := range state.Bodies {
		dx := b.Position.X - center.X
		dy := b.Position.Y - center.Y
		dz := b.Position.Z - center.Z
		d := float32(math.Sqrt(float64(dx*dx+dy*dy+dz*dz))) + b.Radius
		if d > maxR {
			maxR = d
		}
	}

	c.Target = center
	halfFovRad := float64(c.Camera.Fovy*0.5) * math.Pi / 180.0
	neededDist := float32(float64(maxR) / math.Sin(halfFovRad) * 1.2)
	if neededDist < c.MinDist {
		neededDist = c.MinDist
	}
	if neededDist > c.MaxDist {
		neededDist = c.MaxDist
	}
	c.Distance = neededDist
	state.FollowSelected = false
	state.FollowBarycenter = false
	c.UpdatePosition()

	if cfg != nil {
		cfg.NotificationText = fmt.Sprintf("Framed all %d bodies in scene", len(state.Bodies))
		cfg.NotificationTimer = 2.5
	}
}

// SetViewPreset sets vantage angle and marks the preset active
func (c *OrbitCamera) SetViewPreset(preset string) {
	c.CurrentView = preset
	switch preset {
	case "top":
		c.Azimuth = 0.0
		c.Elevation = 1.50
	case "side":
		c.Azimuth = 0.0
		c.Elevation = 0.0
	case "front":
		c.Azimuth = float32(math.Pi / 2.0)
		c.Elevation = 0.0
	case "iso":
		c.Azimuth = 0.8
		c.Elevation = 0.45
	}
	c.UpdatePosition()
}

// SetViewAngle sets orientation angles and updates camera position
func (c *OrbitCamera) SetViewAngle(azimuth, elevation float32) {
	c.CurrentView = ""
	c.Azimuth = azimuth
	const maxElev = 1.52
	if elevation > maxElev {
		elevation = maxElev
	} else if elevation < -maxElev {
		elevation = -maxElev
	}
	c.Elevation = elevation
	c.UpdatePosition()
}

// SetDistance sets camera distance clamped to min/max
func (c *OrbitCamera) SetDistance(dist float32) {
	if dist < c.MinDist {
		dist = c.MinDist
	}
	if dist > c.MaxDist {
		dist = c.MaxDist
	}
	c.Distance = dist
	c.UpdatePosition()
}

// SetFov sets camera vertical field of view in degrees
func (c *OrbitCamera) SetFov(fov float32) {
	if fov < 20.0 {
		fov = 20.0
	}
	if fov > 100.0 {
		fov = 100.0
	}
	c.Camera.Fovy = fov
}
