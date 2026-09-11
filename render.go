package main

import (
	"fmt"
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Render3DScene renders all 3D objects within the camera perspective
func Render3DScene(state *SimState, cfg *Config, camera *OrbitCamera) {
	startRender := time.Now()
	rl.BeginMode3D(camera.Camera)

	// 1. Draw Starfield
	if cfg.ShowStarfield {
		for i, starPos := range state.Stars {
			// Offset by camera target so stars stay at infinity
			pos := rl.Vector3Add(starPos, camera.Target)
			rl.DrawCube(pos, state.StarSizes[i], state.StarSizes[i], state.StarSizes[i], state.StarColors[i])
		}
	}

	// 2. Draw Reference Grid on XZ plane
	if cfg.ShowGrid {
		rl.DrawGrid(60, 4.0)
	}

	// 2b. Draw 3D Spacetime Curvature Potential Grid
	if cfg.ShowPotentialGrid {
		drawSpacetimePotentialGrid(state, cfg, camera)
	}

	// 2c. Draw Gravitational Vector Field
	if cfg.ShowVectorField {
		drawGravitationalVectorField(state, cfg, camera)
	}

	// 2d. Draw 3D Gravitational Heatmap Plane Grid
	if cfg.Show3DHeatmapPlane {
		draw3DGravitationalHeatmapPlane(state, cfg, camera)
	}

	// 2e. Draw 3D Lagrange Equilibrium Points & Orbit Geometry
	if cfg.ShowLagrangePoints {
		draw3DLagrangePoints(state, cfg, camera)
	}

	// 3. Draw Orbit Trails (optimized: skip background particles in massive swarms)
	if cfg.ShowTrails {
		bodyCount := len(state.Bodies)
		for _, b := range state.Bodies {
			// In massive swarms, skip trails on tiny background particles to maintain 144 FPS
			if bodyCount > 100 && b.ID != state.SelectedBodyID && b.Mass < 5.0 && !b.IsStar {
				continue
			}

			trailLen := len(b.Trail)
			if trailLen < 2 {
				continue
			}

			col := b.Color
			for i := 0; i < trailLen-1; i++ {
				// Fade alpha towards the tail
				alpha := float32(i) / float32(trailLen)
				trailCol := rl.Fade(col, alpha*0.85)
				rl.DrawLine3D(b.Trail[i], b.Trail[i+1], trailCol)
			}
			// Connect last trail point to current position
			if trailLen > 0 {
				rl.DrawLine3D(b.Trail[trailLen-1], b.Position, col)
			}
		}
	}

	// 4. Draw Celestial Bodies
	bodyCount := len(state.Bodies)
	for _, b := range state.Bodies {
		isSelected := (b.ID == state.SelectedBodyID)

		drawColor := b.Color
		if cfg.ParticleGlowMode && !b.IsStar && b.TextureType != TextureBlackHole {
			drawColor = getBodyGlowColor(b)
		}

		// Celestial body sphere (textured or untextured, with fast LOD batching for 10k particle swarms)
		if bodyCount > 250 && !isSelected && b.Radius < 0.65 && !b.IsStar && b.TextureType != TextureBlackHole {
			cubeSize := b.Radius * 1.6
			rl.DrawCube(b.Position, cubeSize, cubeSize, cubeSize, drawColor)
		} else if cfg.ShowTextures && b.TextureType != TextureNone {
			DrawTexturedCelestialSphere(b, isSelected)
		} else {
			rl.DrawSphere(b.Position, b.Radius, drawColor)
		}

		// Glowing star aura (only for prominent stars or small simulations)
		if b.IsStar && (len(state.Bodies) < 150 || b.Radius >= 0.8) {
			rl.DrawSphere(b.Position, b.Radius*1.35, rl.Fade(b.Color, 0.28))
			rl.DrawSphere(b.Position, b.Radius*1.85, rl.Fade(b.Color, 0.12))
		}

		// Pulsar relativistic polar jet beams
		if b.IsPulsar {
			rotRad := float64(b.RotationAngle) * (math.Pi / 180.0)
			tilt := 0.35 // magnetic axis inclination
			jetDir := rl.NewVector3(
				float32(math.Sin(tilt) * math.Cos(rotRad)),
				float32(math.Cos(tilt)),
				float32(math.Sin(tilt) * math.Sin(rotRad)),
			)
			jetLen := b.Radius * 14.0

			tipNorth := rl.Vector3Add(b.Position, rl.Vector3Scale(jetDir, jetLen))
			tipSouth := rl.Vector3Subtract(b.Position, rl.Vector3Scale(jetDir, jetLen))

			rl.DrawLine3D(b.Position, tipNorth, rl.NewColor(160, 230, 255, 240))
			rl.DrawLine3D(b.Position, tipSouth, rl.NewColor(160, 230, 255, 240))

			// Pulsing nodes along the jet
			for step := 1; step <= 4; step++ {
				distFrac := float32(step) * 0.25 * jetLen
				nodeNorth := rl.Vector3Add(b.Position, rl.Vector3Scale(jetDir, distFrac))
				nodeSouth := rl.Vector3Subtract(b.Position, rl.Vector3Scale(jetDir, distFrac))
				nodeRadius := b.Radius * (0.2 + float32(step)*0.15)

				rl.DrawSphere(nodeNorth, nodeRadius, rl.Fade(rl.SkyBlue, 0.25))
				rl.DrawSphere(nodeSouth, nodeRadius, rl.Fade(rl.Purple, 0.25))
			}
		}

		// Comet dynamic ion & dust tail (points away from nearest star)
		if b.IsComet {
			var nearestStar *Body
			var minStarDist float32 = 1e9
			for _, other := range state.Bodies {
				if other.IsStar && other.ID != b.ID {
					d := rl.Vector3Distance(b.Position, other.Position)
					if d < minStarDist {
						minStarDist = d
						nearestStar = other
					}
				}
			}

			if nearestStar != nil {
				awayDir := rl.Vector3Normalize(rl.Vector3Subtract(b.Position, nearestStar.Position))
				tailLen := float32(18.0)
				tailTip := rl.Vector3Add(b.Position, rl.Vector3Scale(awayDir, tailLen))

				// Wide fan dust tail
				rl.DrawLine3D(b.Position, tailTip, rl.NewColor(220, 245, 255, 180))
				for t := 1; t <= 5; t++ {
					tFrac := float32(t) / 5.0
					pos := rl.Vector3Add(b.Position, rl.Vector3Scale(awayDir, tailLen*tFrac))
					spread := float32(t) * 0.4
					rl.DrawSphere(pos, spread, rl.Fade(rl.SkyBlue, 0.12*(1.0-tFrac*0.5)))
				}
			}
		}

		// Accretion disk for Black Holes
		if b.TextureType == TextureBlackHole {
			diskR1 := b.Radius * 1.8
			diskR2 := b.Radius * 4.2
			const segs = 32
			for s := 0; s < segs; s++ {
				a1 := float64(s) * (2.0 * math.Pi / segs)
				a2 := float64(s+1) * (2.0 * math.Pi / segs)
				p1 := rl.NewVector3(b.Position.X+diskR1*float32(math.Cos(a1)), b.Position.Y, b.Position.Z+diskR1*float32(math.Sin(a1)))
				p2 := rl.NewVector3(b.Position.X+diskR2*float32(math.Cos(a1)), b.Position.Y, b.Position.Z+diskR2*float32(math.Sin(a1)))
				p3 := rl.NewVector3(b.Position.X+diskR2*float32(math.Cos(a2)), b.Position.Y, b.Position.Z+diskR2*float32(math.Sin(a2)))

				rl.DrawLine3D(p1, p2, rl.NewColor(180, 100, 240, 90))
				rl.DrawLine3D(p2, p3, rl.NewColor(220, 140, 255, 120))
			}
		}

		// Stationary indicator (lock ring)
		if b.IsStationary {
			rl.DrawSphereWires(b.Position, b.Radius*1.15, 6, 8, rl.NewColor(255, 60, 60, 200))
		}

		// Highlight selected body & Keplerian orbit preview
		if isSelected {
			rl.DrawSphereWires(b.Position, b.Radius*1.35, 14, 14, rl.Yellow)

			// Drop altitude projection line to XZ plane (Y=0)
			dropPoint := rl.NewVector3(b.Position.X, 0, b.Position.Z)
			rl.DrawLine3D(b.Position, dropPoint, rl.Fade(rl.Yellow, 0.5))
			rl.DrawCircle3D(dropPoint, b.Radius*0.8, rl.NewVector3(1, 0, 0), 90, rl.Fade(rl.Yellow, 0.3))

			// Projected Keplerian circular reference track around primary
			primary := FindPrimaryBody(state.Bodies, b)
			if primary != nil && !b.IsStationary {
				orbitDist := rl.Vector3Distance(b.Position, primary.Position)
				rl.DrawCircle3D(primary.Position, orbitDist, rl.NewVector3(0, 1, 0), 0, rl.Fade(rl.SkyBlue, 0.25))
			}
		}

		// Velocity vector (Green)
		if cfg.ShowVectors && !b.IsStationary {
			vLen := rl.Vector3Length(b.Velocity)
			if vLen > 0.01 {
				vEnd := rl.Vector3Add(b.Position, rl.Vector3Scale(b.Velocity, 1.2))
				rl.DrawLine3D(b.Position, vEnd, rl.Lime)
			}
		}

		// Gravitational Force vector (Orange/Red)
		if cfg.ShowForces && !b.IsStationary {
			fLen := rl.Vector3Length(b.NetForce)
			if fLen > 0.001 {
				scale := float32(2.5) / float32(math.Sqrt(float64(fLen))+1.0)
				fEnd := rl.Vector3Add(b.Position, rl.Vector3Scale(b.NetForce, scale))
				rl.DrawLine3D(b.Position, fEnd, rl.Orange)
			}
		}
	}

	// 5. Draw interactive slingshot spawn vector
	if state.IsDraggingSpawn {
		rl.DrawSphere(state.SpawnWorldPos, 1.2, rl.Fade(rl.SkyBlue, 0.8))
		endPos := rl.Vector3Add(state.SpawnWorldPos, state.DragVelocity)
		rl.DrawLine3D(state.SpawnWorldPos, endPos, rl.Yellow)
		rl.DrawSphere(endPos, 0.4, rl.Gold)
	}

	// 6. Draw interactive force vector on selected body
	if state.IsDraggingForce && state.SelectedBodyID != -1 {
		var sel *Body
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				sel = b
				break
			}
		}
		if sel != nil {
			endPos := rl.Vector3Add(sel.Position, state.ForceVector)
			rl.DrawLine3D(sel.Position, endPos, rl.Red)
			rl.DrawSphere(endPos, 0.5, rl.Orange)
		}
	}

	rl.EndMode3D()

	// Record render pass execution duration
	cfg.RenderTimeMs = float32(time.Since(startRender).Seconds() * 1000.0)

	// 7. Draw floating 2D labels in screen space
	if cfg.ShowLabels {
		drawScreenLabels(state, camera)
	}

	// 8. Draw 3D Lagrange Point Billboard Badges in screen space
	if cfg.ShowLagrangePoints && cfg.Show2DLabels {
		draw3DLagrangeScreenBadges(state, cfg, camera)
	}
}

// getBodyGlowColor computes dynamic kinetic energy / velocity color glow
func getBodyGlowColor(b *Body) rl.Color {
	speed := rl.Vector3Length(b.Velocity)
	t := speed / 22.0
	if t > 1.0 {
		t = 1.0
	}
	r := uint8(float32(b.Color.R)*(1.0-t) + 120.0*t)
	g := uint8(float32(b.Color.G)*(1.0-t) + 230.0*t)
	bVal := uint8(float32(b.Color.B)*(1.0-t) + 255.0*t)
	return rl.NewColor(r, g, bVal, 255)
}

var (
	gridDepthBuf []float32
	gridColorBuf []rl.Color
)

// StaticBaseSpacetimeSpan defines the fixed, invariant world span of the spacetime curvature grid
const StaticBaseSpacetimeSpan float32 = 300.0

// getStaticSpacetimeSpan returns the invariant global physical span of the spacetime grid, scaled only by user multiplier
func getStaticSpacetimeSpan(cfg *Config) float32 {
	scale := cfg.SpacetimeScale
	if scale <= 0 {
		scale = 1.0
	}
	return StaticBaseSpacetimeSpan * scale
}

// drawSpacetimePotentialGrid renders an Einsteinian gravitational potential well grid on the horizontal plane
func drawSpacetimePotentialGrid(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera // Spacetime grid is anchored in global world coordinates; camera does not shift or stretch it!

	res := cfg.SpacetimeResolution
	if res < 36 {
		res = 120 // Ultra-high-resolution default: 120x120 subdivisions (14,641 vertices)
	}

	totalSpan := getStaticSpacetimeSpan(cfg)
	spacing := totalSpan / float32(res)
	halfDim := totalSpan * 0.5

	// GLOBAL POSITION: anchored firmly at world coordinates (0, 0, 0)
	baseX := float32(0.0)
	baseZ := float32(0.0)

	// Gather celestial bodies that curve spacetime
	heavyBodies := make([]*Body, 0, 32)
	selectedIncluded := false

	// 1. Always prioritize the currently selected body so its gravity well is never missing
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				heavyBodies = append(heavyBodies, b)
				selectedIncluded = true
				break
			}
		}
	}

	// 2. Include all major planets, stars, black holes, and significant masses
	if len(state.Bodies) <= 50 {
		// In solar systems and standard scenes, include all planets (Earth, Mars, Venus, etc.) and moons
		for _, b := range state.Bodies {
			if selectedIncluded && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= 0.05 || b.IsStar || b.TextureType == TextureBlackHole {
				heavyBodies = append(heavyBodies, b)
			}
		}
	} else {
		// In massive swarms (10K particles), include stars, black holes, and anchors up to 32 bodies
		for _, b := range state.Bodies {
			if selectedIncluded && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= 15.0 || b.TextureType == TextureBlackHole || b.IsStar {
				heavyBodies = append(heavyBodies, b)
				if len(heavyBodies) >= 32 {
					break
				}
			}
		}
	}
	if len(heavyBodies) == 0 && len(state.Bodies) > 0 {
		heavyBodies = append(heavyBodies, state.Bodies[0])
	}

	// Precompute body scaled masses once per frame to eliminate millions of math.Pow calls
	type precomputedBody struct {
		x, z        float32
		scaledMassG float64
	}
	sources := make([]precomputedBody, len(heavyBodies))
	for k, b := range heavyBodies {
		sm := math.Pow(b.Mass, 0.48) * 8.0
		if b.ID == state.SelectedBodyID {
			sm *= 1.35 // Extra distinct well for inspected body
		}
		sources[k] = precomputedBody{
			x:           b.Position.X,
			z:           b.Position.Z,
			scaledMassG: cfg.G * sm,
		}
	}

	totalPoints := (res + 1) * (res + 1)
	if len(gridDepthBuf) < totalPoints {
		gridDepthBuf = make([]float32, totalPoints)
		gridColorBuf = make([]rl.Color, totalPoints)
	}

	// Calculate vertex depths and colors
	idx := 0
	for i := 0; i <= res; i++ {
		gx := baseX - halfDim + float32(i)*spacing
		for j := 0; j <= res; j++ {
			gz := baseZ - halfDim + float32(j)*spacing
			var potential float64
			for k := 0; k < len(sources); k++ {
				dx := float64(gx - sources[k].x)
				dz := float64(gz - sources[k].z)
				r := math.Sqrt(dx*dx + dz*dz + 1.8)
				potential += sources[k].scaledMassG / r
			}
			depth := -float32(math.Min(38.0, potential*0.14))
			depthRatio := float32(math.Min(1.0, math.Abs(float64(depth))/28.0))
			col := rl.NewColor(
				uint8(35+depthRatio*220),
				uint8(85+depthRatio*140),
				uint8(170-depthRatio*95),
				uint8(85+depthRatio*135),
			)
			gridDepthBuf[idx] = depth
			gridColorBuf[idx] = col
			idx++
		}
	}

	// Draw lines along Z axis
	for i := 0; i <= res; i++ {
		gx := baseX - halfDim + float32(i)*spacing
		for j := 0; j < res; j++ {
			gz1 := baseZ - halfDim + float32(j)*spacing
			gz2 := gz1 + spacing
			idx1 := i*(res+1) + j
			idx2 := idx1 + 1
			p1 := rl.NewVector3(gx, gridDepthBuf[idx1], gz1)
			p2 := rl.NewVector3(gx, gridDepthBuf[idx2], gz2)
			rl.DrawLine3D(p1, p2, gridColorBuf[idx1])
		}
	}

	// Draw lines along X axis
	for j := 0; j <= res; j++ {
		gz := baseZ - halfDim + float32(j)*spacing
		for i := 0; i < res; i++ {
			gx1 := baseX - halfDim + float32(i)*spacing
			gx2 := gx1 + spacing
			idx1 := i*(res+1) + j
			idx2 := (i+1)*(res+1) + j
			p1 := rl.NewVector3(gx1, gridDepthBuf[idx1], gz)
			p2 := rl.NewVector3(gx2, gridDepthBuf[idx2], gz)
			rl.DrawLine3D(p1, p2, gridColorBuf[idx1])
		}
	}
}

// drawGravitationalVectorField displays a high-resolution global 3D grid of local gravitational acceleration vectors
func drawGravitationalVectorField(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera // Anchored globally at origin (0,0,0) in sync with the spacetime curvature grid
	if len(state.Bodies) == 0 {
		return
	}

	totalSpan := getStaticSpacetimeSpan(cfg)
	const steps = 32 // High resolution: 32x32 = 1,024 3D vectors
	stepSize := totalSpan / float32(steps-1)
	halfSpan := totalSpan * 0.5

	baseX := float32(0.0)
	baseZ := float32(0.0)

	// Gather celestial bodies that generate gravitational acceleration
	heavy := make([]*Body, 0, 32)
	selectedIncluded := false
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				heavy = append(heavy, b)
				selectedIncluded = true
				break
			}
		}
	}

	if len(state.Bodies) <= 50 {
		for _, b := range state.Bodies {
			if selectedIncluded && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= 0.05 || b.IsStar || b.TextureType == TextureBlackHole {
				heavy = append(heavy, b)
			}
		}
	} else {
		for _, b := range state.Bodies {
			if selectedIncluded && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= 15.0 || b.TextureType == TextureBlackHole || b.IsStar {
				heavy = append(heavy, b)
				if len(heavy) >= 32 {
					break
				}
			}
		}
	}
	if len(heavy) == 0 && len(state.Bodies) > 0 {
		heavy = append(heavy, state.Bodies[0])
	}

	type precomputedMass struct {
		x, y, z float32
		scaledG float32
	}
	sources := make([]precomputedMass, len(heavy))
	for k, b := range heavy {
		sources[k] = precomputedMass{
			x:       b.Position.X,
			y:       b.Position.Y,
			z:       b.Position.Z,
			scaledG: float32(cfg.G * b.Mass),
		}
	}

	eps2 := float32(cfg.Softening*cfg.Softening + 4.0)

	for ix := 0; ix < steps; ix++ {
		px := baseX - halfSpan + float32(ix)*stepSize
		for iz := 0; iz < steps; iz++ {
			pz := baseZ - halfSpan + float32(iz)*stepSize
			var gx, gy, gz float32

			for k := 0; k < len(sources); k++ {
				dx := sources[k].x - px
				dy := sources[k].y - 0
				dz := sources[k].z - pz
				distSq := dx*dx + dy*dy + dz*dz + eps2
				dist := float32(math.Sqrt(float64(distSq)))
				invDist3 := 1.0 / (distSq * dist)
				f := sources[k].scaledG * invDist3
				gx += dx * f
				gy += dy * f
				gz += dz * f
			}

			gLen := float32(math.Sqrt(float64(gx*gx + gy*gy + gz*gz)))
			if gLen > 0.0001 {
				dirX := gx / gLen
				dirY := gy / gLen
				dirZ := gz / gLen

				// Logarithmic / sigmoid length scaling: min 1.2 units up to 4.5 units
				arrowLen := 1.2 + 3.3*(gLen/(gLen+0.35))
				pStart := rl.NewVector3(px, 0, pz)
				pEnd := rl.NewVector3(px+dirX*arrowLen, dirY*arrowLen, pz+dirZ*arrowLen)

				// Color gradient based on gravitational field strength
				ratio := float32(math.Min(1.0, math.Max(0.0, (math.Log10(float64(gLen))+2.5)/3.0)))
				var col rl.Color
				if ratio < 0.33 {
					t := ratio / 0.33
					col = rl.NewColor(
						uint8(40+t*20),
						uint8(160+t*60),
						uint8(255-t*90),
						uint8(140+t*40),
					)
				} else if ratio < 0.66 {
					t := (ratio - 0.33) / 0.33
					col = rl.NewColor(
						uint8(60+t*195),
						uint8(220-t*20),
						uint8(165-t*115),
						uint8(180+t*40),
					)
				} else {
					t := (ratio - 0.66) / 0.34
					col = rl.NewColor(
						uint8(255),
						uint8(200-t*140),
						uint8(50+t*50),
						uint8(220+t*35),
					)
				}

				rl.DrawLine3D(pStart, pEnd, col)

				// Draw 3D arrowhead wings
				sideX := -dirZ * 0.28 * arrowLen
				sideZ := dirX * 0.28 * arrowLen
				wing1 := rl.NewVector3(pEnd.X-dirX*0.25*arrowLen+sideX*0.4, pEnd.Y-dirY*0.25*arrowLen, pEnd.Z-dirZ*0.25*arrowLen+sideZ*0.4)
				wing2 := rl.NewVector3(pEnd.X-dirX*0.25*arrowLen-sideX*0.4, pEnd.Y-dirY*0.25*arrowLen, pEnd.Z-dirZ*0.25*arrowLen-sideZ*0.4)
				rl.DrawLine3D(pEnd, wing1, col)
				rl.DrawLine3D(pEnd, wing2, col)
			}
		}
	}
}

// drawScreenLabels draws nameplates and telemetry above celestial bodies with occlusion culling to prevent overlapping
func drawScreenLabels(state *SimState, camera *OrbitCamera) {
	bodyCount := len(state.Bodies)
	drawnRects := make([]rl.Rectangle, 0, 32)
	var selectedBody *Body

	for _, b := range state.Bodies {
		if b.ID == state.SelectedBodyID {
			selectedBody = b
			continue
		}

		// Don't draw if behind camera
		camToBody := rl.Vector3Subtract(b.Position, camera.Camera.Position)
		camForward := rl.Vector3Subtract(camera.Camera.Target, camera.Camera.Position)
		if rl.Vector3DotProduct(camToBody, camForward) <= 0 {
			continue
		}

		// Clutter culling: in swarms, only label major bodies
		if bodyCount > 40 && b.Mass < 20.0 {
			continue
		}
		if len(drawnRects) >= 16 {
			break
		}

		screenPos := rl.GetWorldToScreen(b.Position, camera.Camera)
		textX := int32(screenPos.X) + 12
		textY := int32(screenPos.Y) - 10

		label := b.Name
		if b.IsStationary {
			label += " [FIXED]"
		}

		labelW := int32(MeasureTextBoldUI(label, FontSizeButton)) + 14
		boxRec := rl.NewRectangle(float32(textX-4), float32(textY-3), float32(labelW), 22)

		// Overlap culling
		collides := false
		for _, prev := range drawnRects {
			if rl.CheckCollisionRecs(boxRec, prev) {
				collides = true
				break
			}
		}
		if collides {
			continue
		}

		drawnRects = append(drawnRects, boxRec)
		rl.DrawRectangle(textX-4, textY-3, labelW, 22, rl.NewColor(12, 16, 26, 175))
		rl.DrawRectangleLines(textX-4, textY-3, labelW, 22, rl.NewColor(50, 70, 105, 200))
		DrawTextBoldUI(label, textX+3, textY+3, FontSizeButton, rl.White)
	}

	// Always draw selected body label on top
	if selectedBody != nil {
		camToBody := rl.Vector3Subtract(selectedBody.Position, camera.Camera.Position)
		camForward := rl.Vector3Subtract(camera.Camera.Target, camera.Camera.Position)
		if rl.Vector3DotProduct(camToBody, camForward) > 0 {
			screenPos := rl.GetWorldToScreen(selectedBody.Position, camera.Camera)
			textX := int32(screenPos.X) + 12
			textY := int32(screenPos.Y) - 12

			label := selectedBody.Name
			if selectedBody.IsStationary {
				label += " [FIXED]"
			}

			speed := rl.Vector3Length(selectedBody.Velocity)
			info := fmt.Sprintf("M: %.2f | V: %.2f", selectedBody.Mass, speed)

			labelW := int32(MeasureTextBoldUI(label, FontSizeButton))
			infoW := int32(MeasureTextUI(info, FontSizeTelemetry))
			maxW := labelW
			if infoW > maxW {
				maxW = infoW
			}
			boxW := maxW + 16

			rl.DrawRectangle(textX-4, textY-4, boxW, 40, rl.NewColor(12, 18, 30, 225))
			rl.DrawRectangleLines(textX-4, textY-4, boxW, 40, rl.Gold)

			DrawTextBoldUI(label, textX+4, textY+2, FontSizeButton, rl.Yellow)
			DrawTextUI(info, textX+4, textY+21, FontSizeTelemetry, rl.SkyBlue)
		}
	}
}

// draw3DLagrangePoints renders 3D diamond wireframe markers, halo libration rings, equilateral orbital triangle lines,
// and collinear axis connecting Primary, Secondary, and L1-L5 in 3D space.
func draw3DLagrangePoints(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera
	if len(state.Bodies) < 2 {
		return
	}

	var primary, secondary *Body
	for _, b := range state.Bodies {
		if primary == nil || b.Mass > primary.Mass {
			secondary = primary
			primary = b
		} else if secondary == nil || b.Mass > secondary.Mass {
			secondary = b
		}
	}

	if primary == nil || secondary == nil || secondary.Mass < 0.001 {
		return
	}

	pts := ComputeLagrangePoints(primary, secondary, cfg.G)
	colors := [5]rl.Color{
		rl.NewColor(255, 220, 50, 240),  // L1 Gold
		rl.NewColor(80, 220, 255, 240),  // L2 Sky Cyan
		rl.NewColor(255, 120, 80, 240),  // L3 Coral
		rl.NewColor(120, 255, 180, 240), // L4 Emerald
		rl.NewColor(255, 180, 120, 240), // L5 Amber
	}

	// 1. Collinear axis line through L3 -> Primary -> L1 -> Secondary -> L2
	axisCol := rl.NewColor(120, 160, 220, 90)
	rl.DrawLine3D(pts[2].Position, primary.Position, axisCol)
	rl.DrawLine3D(primary.Position, pts[0].Position, axisCol)
	rl.DrawLine3D(pts[0].Position, secondary.Position, axisCol)
	rl.DrawLine3D(secondary.Position, pts[1].Position, axisCol)

	// 2. Equilateral triangle geometry lines
	// L4 (Trojan leading): Primary -> L4 -> Secondary
	l4Col := rl.NewColor(80, 240, 160, 130)
	rl.DrawLine3D(primary.Position, pts[3].Position, l4Col)
	rl.DrawLine3D(secondary.Position, pts[3].Position, l4Col)

	// L5 (Greek trailing): Primary -> L5 -> Secondary
	l5Col := rl.NewColor(255, 180, 80, 130)
	rl.DrawLine3D(primary.Position, pts[4].Position, l5Col)
	rl.DrawLine3D(secondary.Position, pts[4].Position, l5Col)

	// Base marker size proportional to primary radius
	markerSize := float32(math.Max(0.7, math.Min(3.2, float64(primary.Radius*0.35))))

	// 3. Render each Lagrange equilibrium point
	for i := 0; i < 5; i++ {
		p := pts[i].Position
		col := colors[i]

		// Libration halo ring on orbital plane
		haloR := markerSize * 2.2
		rl.DrawCircle3D(p, haloR, rl.NewVector3(0, 1, 0), 0, rl.Fade(col, 0.3))

		// Glowing core sphere
		rl.DrawSphere(p, markerSize*0.3, col)

		// 3D Wireframe Octahedron (Diamond)
		s := markerSize
		top := rl.NewVector3(p.X, p.Y+s, p.Z)
		bot := rl.NewVector3(p.X, p.Y-s, p.Z)
		px := rl.NewVector3(p.X+s, p.Y, p.Z)
		nx := rl.NewVector3(p.X-s, p.Y, p.Z)
		pz := rl.NewVector3(p.X, p.Y, p.Z+s)
		nz := rl.NewVector3(p.X, p.Y, p.Z-s)

		// Upper pyramid
		rl.DrawLine3D(top, px, col)
		rl.DrawLine3D(top, nx, col)
		rl.DrawLine3D(top, pz, col)
		rl.DrawLine3D(top, nz, col)

		// Lower pyramid
		rl.DrawLine3D(bot, px, col)
		rl.DrawLine3D(bot, nx, col)
		rl.DrawLine3D(bot, pz, col)
		rl.DrawLine3D(bot, nz, col)

		// Equator
		rl.DrawLine3D(px, pz, col)
		rl.DrawLine3D(pz, nx, col)
		rl.DrawLine3D(nx, nz, col)
		rl.DrawLine3D(nz, px, col)
	}
}

// draw3DLagrangeScreenBadges projects HUD tags [L1]-[L5] in 2D screen space above their 3D coordinates
func draw3DLagrangeScreenBadges(state *SimState, cfg *Config, camera *OrbitCamera) {
	if len(state.Bodies) < 2 {
		return
	}

	var primary, secondary *Body
	for _, b := range state.Bodies {
		if primary == nil || b.Mass > primary.Mass {
			secondary = primary
			primary = b
		} else if secondary == nil || b.Mass > secondary.Mass {
			secondary = b
		}
	}

	if primary == nil || secondary == nil || secondary.Mass < 0.001 {
		return
	}

	pts := ComputeLagrangePoints(primary, secondary, cfg.G)
	labels := [5]string{"L1", "L2", "L3", "L4 Trojan", "L5 Greek"}
	colors := [5]rl.Color{
		rl.NewColor(255, 220, 50, 255),
		rl.NewColor(80, 220, 255, 255),
		rl.NewColor(255, 120, 80, 255),
		rl.NewColor(120, 255, 180, 255),
		rl.NewColor(255, 180, 120, 255),
	}

	camForward := rl.Vector3Subtract(camera.Camera.Target, camera.Camera.Position)

	for i := 0; i < 5; i++ {
		p := pts[i].Position
		camToPt := rl.Vector3Subtract(p, camera.Camera.Position)
		if rl.Vector3DotProduct(camToPt, camForward) <= 0 {
			continue // Behind camera
		}

		screenPos := rl.GetWorldToScreen(p, camera.Camera)
		// Only render within visible screen viewport
		if screenPos.X < 20 || screenPos.X > float32(rl.GetScreenWidth())-20 ||
			screenPos.Y < 50 || screenPos.Y > float32(rl.GetScreenHeight())-50 {
			continue
		}

		tag := labels[i]
		col := colors[i]
		tagW := int32(MeasureTextBoldUI(tag, FontSizeTelemetry)) + 12
		tagH := int32(18)
		tagX := int32(screenPos.X) - tagW/2
		tagY := int32(screenPos.Y) - 18

		rl.DrawRectangle(tagX, tagY, tagW, tagH, rl.NewColor(12, 16, 26, 210))
		rl.DrawRectangleLines(tagX, tagY, tagW, tagH, rl.Fade(col, 0.8))
		DrawTextBoldUI(tag, tagX+6, tagY+2, FontSizeTelemetry, col)
	}
}

var (
	heatmapPlaneColBuf []rl.Color
)

// draw3DGravitationalHeatmapPlane renders a continuous thermodynamic gravitational potential heatmap overlay on the Y=0 plane like a grid
func draw3DGravitationalHeatmapPlane(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera
	if len(state.Bodies) == 0 {
		return
	}

	const res = 64
	totalSpan := getStaticSpacetimeSpan(cfg)
	spacing := totalSpan / float32(res)
	halfDim := totalSpan * 0.5

	totalVerts := (res + 1) * (res + 1)
	if len(heatmapPlaneColBuf) < totalVerts {
		heatmapPlaneColBuf = make([]rl.Color, totalVerts)
	}

	// Gather heavy bodies for potential evaluation
	heavy := make([]*Body, 0, 32)
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				heavy = append(heavy, b)
				break
			}
		}
	}
	for _, b := range state.Bodies {
		if (state.SelectedBodyID != -1 && b.ID == state.SelectedBodyID) || b.Mass < 0.05 {
			continue
		}
		heavy = append(heavy, b)
		if len(heavy) >= 28 {
			break
		}
	}
	if len(heavy) == 0 && len(state.Bodies) > 0 {
		heavy = append(heavy, state.Bodies[0])
	}

	type heatSource struct {
		x, y, z float32
		scaledG float32
	}
	sources := make([]heatSource, len(heavy))
	for k, b := range heavy {
		sources[k] = heatSource{
			x:       b.Position.X,
			y:       b.Position.Y,
			z:       b.Position.Z,
			scaledG: float32(b.Mass * cfg.G),
		}
	}

	softSq := float32(cfg.Softening*cfg.Softening + 0.25)

	// Compute thermodynamic potential color for each grid vertex
	idx := 0
	for i := 0; i <= res; i++ {
		gx := -halfDim + float32(i)*spacing
		for j := 0; j <= res; j++ {
			gz := -halfDim + float32(j)*spacing

			var pot float32 = 0.0
			for k := range sources {
				dx := gx - sources[k].x
				dy := -sources[k].y
				dz := gz - sources[k].z
				dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz + softSq)))
				pot += sources[k].scaledG / dist
			}

			// Thermodynamic potential logarithmic ratio [0.0, 1.0]
			ratio := float32(math.Min(1.0, math.Max(0.0, (math.Log10(float64(pot)+1.0)-0.05)/2.4)))

			// Color gradient: Deep Void Navy -> Azure -> Cyan -> Emerald -> Gold -> Crimson -> Core White
			var col rl.Color
			if ratio < 0.2 {
				t := ratio / 0.2
				col = rl.NewColor(
					uint8(15+t*15),
					uint8(30+t*60),
					uint8(75+t*115),
					uint8(100+t*40),
				)
			} else if ratio < 0.45 {
				t := (ratio - 0.2) / 0.25
				col = rl.NewColor(
					uint8(30+t*10),
					uint8(90+t*120),
					uint8(190+t*40),
					uint8(140+t*40),
				)
			} else if ratio < 0.70 {
				t := (ratio - 0.45) / 0.25
				col = rl.NewColor(
					uint8(40+t*195),
					uint8(210+t*25),
					uint8(230-t*180),
					uint8(180+t*40),
				)
			} else if ratio < 0.90 {
				t := (ratio - 0.70) / 0.20
				col = rl.NewColor(
					uint8(235+t*20),
					uint8(235-t*160),
					uint8(50-t*20),
					uint8(220+t*25),
				)
			} else {
				t := (ratio - 0.90) / 0.10
				col = rl.NewColor(
					255,
					uint8(75+t*180),
					uint8(30+t*225),
					255,
				)
			}

			heatmapPlaneColBuf[idx] = col
			idx++
		}
	}

	planeY := float32(-0.02) // subtle offset to prevent z-fighting with standard grid

	// Draw lines along Z axis
	for i := 0; i <= res; i++ {
		gx := -halfDim + float32(i)*spacing
		for j := 0; j < res; j++ {
			gz1 := -halfDim + float32(j)*spacing
			gz2 := gz1 + spacing
			idx1 := i*(res+1) + j
			p1 := rl.NewVector3(gx, planeY, gz1)
			p2 := rl.NewVector3(gx, planeY, gz2)
			rl.DrawLine3D(p1, p2, heatmapPlaneColBuf[idx1])
		}
	}

	// Draw lines along X axis
	for j := 0; j <= res; j++ {
		gz := -halfDim + float32(j)*spacing
		for i := 0; i < res; i++ {
			gx1 := -halfDim + float32(i)*spacing
			gx2 := gx1 + spacing
			idx1 := i*(res+1) + j
			p1 := rl.NewVector3(gx1, planeY, gz)
			p2 := rl.NewVector3(gx2, planeY, gz)
			rl.DrawLine3D(p1, p2, heatmapPlaneColBuf[idx1])
		}
	}
}

