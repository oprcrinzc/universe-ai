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

	// 2b. Draw 3D Spacetime Curvature Potential Grid and/or Gravitational Wave Metric Ripples
	if cfg.ShowPotentialGrid || cfg.ShowGravitationalWaves {
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
				float32(math.Sin(tilt)*math.Cos(rotRad)),
				float32(math.Cos(tilt)),
				float32(math.Sin(tilt)*math.Sin(rotRad)),
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

// StaticBaseSpacetimeSpan defines the fixed, invariant world span of the spacetime curvature grid (5x expanded: 1,500 units)
const StaticBaseSpacetimeSpan float32 = 1500.0

// getStaticSpacetimeSpan returns the invariant global physical span of the spacetime grid, scaled only by user multiplier
func getStaticSpacetimeSpan(cfg *Config) float32 {
	scale := cfg.SpacetimeScale
	if scale <= 0 {
		scale = 1.0
	}
	return StaticBaseSpacetimeSpan * scale
}

// SpacetimeResolutions defines user-selectable subdivisions per axis for the spacetime grid
var SpacetimeResolutions = []int{32, 48, 64, 80, 96, 120, 160, 200, 256, 384, 512}

// DecreaseResolution steps the spacetime curvature grid resolution down by one tier
func DecreaseResolution(cfg *Config) {
	current := cfg.SpacetimeResolution
	if current <= 0 {
		current = 120
	}
	target := SpacetimeResolutions[0]
	for i := len(SpacetimeResolutions) - 1; i >= 0; i-- {
		if SpacetimeResolutions[i] < current {
			target = SpacetimeResolutions[i]
			break
		}
	}
	cfg.SpacetimeResolution = target
	numVerts := (target + 1) * (target + 1)
	cfg.NotificationText = fmt.Sprintf("Spacetime Grid Resolution: %dx%d (%d Vertices)", target, target, numVerts)
	cfg.NotificationTimer = 2.2
}

// IncreaseResolution steps the spacetime curvature grid resolution up by one tier
func IncreaseResolution(cfg *Config) {
	current := cfg.SpacetimeResolution
	if current <= 0 {
		current = 120
	}
	target := SpacetimeResolutions[len(SpacetimeResolutions)-1]
	for i := 0; i < len(SpacetimeResolutions); i++ {
		if SpacetimeResolutions[i] > current {
			target = SpacetimeResolutions[i]
			break
		}
	}
	cfg.SpacetimeResolution = target
	numVerts := (target + 1) * (target + 1)
	cfg.NotificationText = fmt.Sprintf("Spacetime Grid Resolution: %dx%d (%d Vertices)", target, target, numVerts)
	cfg.NotificationTimer = 2.2
}

// Heatmap2DResolutions defines the progressive resolution tiers for the 2D gravitational heatmap (columns)
var Heatmap2DResolutions = []int{16, 24, 32, 48, 64, 80, 96, 128, 160, 200}

func getHeatmapTierName(res int) string {
	switch {
	case res <= 16:
		return "Low / 16 (Fast)"
	case res <= 24:
		return "Medium-Low / 24"
	case res <= 32:
		return "Balanced / 32"
	case res <= 48:
		return "Standard / 48"
	case res <= 64:
		return "High Detail / 64"
	case res <= 80:
		return "Very High / 80"
	case res <= 96:
		return "Ultra / 96"
	case res <= 128:
		return "Extreme / 128"
	case res <= 160:
		return "Super Sampled / 160"
	default:
		return "Max Quality / 200"
	}
}

// DecreaseHeatmap2DResolution steps the 2D gravitational heatmap resolution down by one tier
func DecreaseHeatmap2DResolution(cfg *Config) {
	current := cfg.Heatmap2DResolution
	if current <= 0 {
		current = 64
	}
	target := Heatmap2DResolutions[0]
	for i := len(Heatmap2DResolutions) - 1; i >= 0; i-- {
		if Heatmap2DResolutions[i] < current {
			target = Heatmap2DResolutions[i]
			break
		}
	}
	cfg.Heatmap2DResolution = target
	cfg.NotificationText = fmt.Sprintf("2D Heatmap Resolution: %d Columns (%s)", target, getHeatmapTierName(target))
	cfg.NotificationTimer = 2.2
}

// IncreaseHeatmap2DResolution steps the 2D gravitational heatmap resolution up by one tier
func IncreaseHeatmap2DResolution(cfg *Config) {
	current := cfg.Heatmap2DResolution
	if current <= 0 {
		current = 64
	}
	target := Heatmap2DResolutions[len(Heatmap2DResolutions)-1]
	for i := 0; i < len(Heatmap2DResolutions); i++ {
		if Heatmap2DResolutions[i] > current {
			target = Heatmap2DResolutions[i]
			break
		}
	}
	cfg.Heatmap2DResolution = target
	cfg.NotificationText = fmt.Sprintf("2D Heatmap Resolution: %d Columns (%s)", target, getHeatmapTierName(target))
	cfg.NotificationTimer = 2.2
}

// drawSpacetimePotentialGrid renders an Einsteinian gravitational potential well grid on the horizontal plane
func drawSpacetimePotentialGrid(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera // Spacetime grid is anchored in global world coordinates; camera does not shift or stretch it!

	res := cfg.SpacetimeResolution
	if res <= 0 {
		res = 120 // Default: 120x120 subdivisions (14,641 vertices)
	} else if res < 16 {
		res = 16
	} else if res > 512 {
		res = 512
	}

	totalSpan := getStaticSpacetimeSpan(cfg)
	spacing := totalSpan / float32(res)
	halfDim := totalSpan * 0.5

	// GLOBAL POSITION: anchored firmly at world coordinates (0, 0, 0)
	baseX := float32(0.0)
	baseZ := float32(0.0)

	// Determine system mass scale
	var maxMass float64 = 1.0
	for _, b := range state.Bodies {
		if b.Mass > maxMass {
			maxMass = b.Mass
		}
	}
	minSignifMass := math.Max(0.0001, maxMass*1e-7)

	// Gather celestial bodies that curve spacetime
	heavyBodies := make([]*Body, 0, 64)
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
	if len(state.Bodies) <= 80 {
		for _, b := range state.Bodies {
			if selectedIncluded && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= minSignifMass || b.IsStar || b.TextureType == TextureBlackHole {
				heavyBodies = append(heavyBodies, b)
				if len(heavyBodies) >= 64 {
					break
				}
			}
		}
	} else {
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

	// Precompute body parameters for multi-scale general relativistic curvature (only if Spacetime is enabled)
	type precomputedBody struct {
		x, y, z    float32
		bodyType   int // 0: minor/planet, 1: star, 2: white dwarf, 3: neutron star/pulsar, 4: black hole
		wellDepth  float32
		wellRadius float32
		cutoffSq   float32
	}
	var sources []precomputedBody
	if cfg.ShowPotentialGrid {
		sources = make([]precomputedBody, len(heavyBodies))
		for k, b := range heavyBodies {
			bType := 0
			var wDepth float32
			var wRad float32
			var cSq float32

			if b.TextureType == TextureBlackHole {
				// Black Hole: Deep asymptotic Schwarzschild funnel with sharp event horizon
				bType = 4
				wDepth = float32(math.Min(58.0, 44.0+math.Log10(b.Mass+1.0)*3.0))
				wRad = float32(math.Max(float64(b.Radius*1.5), 2.2))
				cSq = wRad * wRad * 120.0
			} else if b.TextureType == TextureNeutronStar || b.IsPulsar {
				// Neutron Star / Pulsar: Extremely compact superdense funnel
				bType = 3
				wDepth = 38.0
				wRad = float32(math.Max(float64(b.Radius*1.8), 2.5))
				cSq = wRad * wRad * 100.0
			} else if b.TextureType == TextureWhiteDwarf {
				// White Dwarf: Dense degenerate stellar remnant
				bType = 2
				wDepth = 32.0
				wRad = float32(math.Max(float64(b.Radius*2.0), 2.8))
				cSq = wRad * wRad * 90.0
			} else if b.IsStar || b.Mass >= 200.0 {
				// Stellar bowl (Sun, Red Giant, Protostar)
				bType = 1
				wDepth = float32(math.Min(35.0, 22.0+math.Pow(b.Mass/1000.0, 0.3)*3.0))
				wRad = float32(math.Max(float64(b.Radius*3.5), float64(totalSpan*0.035)))
				cSq = 1e12 // Stars have global reach across wide domain
			} else {
				// Planets and minor celestial bodies (Earth, Gas Giant, Moon, Comet, Asteroid)
				bType = 0
				wDepth = float32(math.Min(18.0, math.Pow(b.Mass, 0.45)*4.2))
				if b.ID == state.SelectedBodyID {
					wDepth *= 1.2
				}
				wRad = float32(math.Max(float64(b.Radius*2.4), float64(spacing*1.2)))
				cSq = wRad * wRad * 80.0
			}

			sources[k] = precomputedBody{
				x:          b.Position.X,
				y:          b.Position.Y,
				z:          b.Position.Z,
				bodyType:   bType,
				wellDepth:  wDepth,
				wellRadius: wRad,
				cutoffSq:   cSq,
			}
		}
	}

	totalPoints := (res + 1) * (res + 1)
	if len(gridDepthBuf) < totalPoints {
		gridDepthBuf = make([]float32, totalPoints)
		gridColorBuf = make([]rl.Color, totalPoints)
	}

	curTime := rl.GetTime()
	gwActive := cfg.ShowGravitationalWaves

	clampU8 := func(v int) uint8 {
		if v > 255 {
			return 255
		}
		if v < 0 {
			return 0
		}
		return uint8(v)
	}

	// Calculate vertex depths, relativistic time dilation colors, and GW ripples
	idx := 0
	for i := 0; i <= res; i++ {
		gx := baseX - halfDim + float32(i)*spacing
		for j := 0; j <= res; j++ {
			gz := baseZ - halfDim + float32(j)*spacing
			var totalDepth float32 = 0.0

			// 1. Spacetime Curvature Potential Wells (Only computed if ShowPotentialGrid is ON)
			if cfg.ShowPotentialGrid {
				for k := 0; k < len(sources); k++ {
					dx := gx - sources[k].x
					dz := gz - sources[k].z
					dy := sources[k].y // accounts for body vertical height Y in 3D
					rSq := dx*dx + dz*dz + dy*dy*0.75

					// Spatial influence cutoff: skip negligible contributions for 4x performance boost
					if rSq > sources[k].cutoffSq {
						continue
					}

					switch sources[k].bodyType {
					case 4:
						// Black Hole: steep asymptotic Schwarzschild throat
						rs2 := sources[k].wellRadius * sources[k].wellRadius
						throat := float32(math.Pow(float64(rs2/(rSq+rs2)), 1.35))
						totalDepth -= sources[k].wellDepth * throat
					case 3, 2:
						// Compact relativistic remnant (Neutron star / White dwarf)
						rw2 := sources[k].wellRadius * sources[k].wellRadius
						funnel := float32(math.Pow(float64(rw2/(rSq+rw2)), 1.15))
						totalDepth -= sources[k].wellDepth * funnel
					case 1:
						// Star: broad gravitational bowl
						r := float32(math.Sqrt(float64(rSq) + 1.8))
						rScale := sources[k].wellRadius
						starFract := rScale / (r + rScale)
						totalDepth -= sources[k].wellDepth * starFract
					default:
						// Planet / Minor: localized Lorentzian depression
						radSq := sources[k].wellRadius * sources[k].wellRadius
						wellFract := radSq / (rSq + radSq)
						totalDepth -= sources[k].wellDepth * wellFract
					}
				}

				if totalDepth < -52.0 {
					totalDepth = -52.0
				}
			}

			// 2. Dynamic Gravitational Wave Metric Ripples (Only computed if ShowGravitationalWaves is ON)
			var gwDisp, gwIntensity float32
			if gwActive {
				gwDisp, gwIntensity = ComputeGravitationalWaveDisplacement(state, cfg, gx, gz, curTime)
				totalDepth += gwDisp
			}

			// 3. Adaptive Vertex Coloring
			var col rl.Color
			if cfg.ShowPotentialGrid {
				depthRatio := float32(math.Min(1.0, math.Abs(float64(totalDepth))/36.0))
				// Relativistic time dilation coloring:
				// Vacuum (flat): deep cosmos blue
				// Mid-depth (planetary wells): vivid teal/emerald
				// Extreme depth (star/black hole core): luminous gold/magenta
				if depthRatio < 0.4 {
					t := depthRatio / 0.4
					col = rl.NewColor(
						uint8(25+t*25),
						uint8(70+t*130),
						uint8(160+t*50),
						uint8(80+t*80),
					)
				} else if depthRatio < 0.75 {
					t := (depthRatio - 0.4) / 0.35
					col = rl.NewColor(
						uint8(50+t*190),
						uint8(200+t*20),
						uint8(210-t*150),
						uint8(160+t*60),
					)
				} else {
					t := (depthRatio - 0.75) / 0.25
					col = rl.NewColor(
						uint8(240+t*15),
						uint8(220-t*90),
						uint8(60+t*120),
						uint8(220+t*35),
					)
				}

				// Gravitational Wave crest & trough luminous interference fringe
				if gwActive && gwIntensity > 0.005 {
					cAdd := int(math.Min(130.0, float64(gwIntensity*70.0)))
					if gwDisp > 0 {
						// Crest: vivid cyan / electric turquoise pulse
						col.G = clampU8(int(col.G) + cAdd)
						col.B = clampU8(int(col.B) + cAdd*2)
					} else {
						// Trough: deep violet / ultraviolet pulse
						col.R = clampU8(int(col.R) + cAdd*2)
						col.B = clampU8(int(col.B) + cAdd)
					}
				}
			} else {
				// Pure Gravitational Waves Mode (Spacetime Potential Wells OFF, GW Ripples ON)
				amp := float32(math.Min(1.0, math.Abs(float64(gwDisp))/1.8))
				cAdd := int(amp * 180.0)
				if gwDisp > 0 {
					// Wave Crest: glowing electric cyan / neon turquoise
					col = rl.NewColor(
						clampU8(24+cAdd/3),
						clampU8(80+cAdd),
						clampU8(170+cAdd),
						clampU8(110+int(amp*135)),
					)
				} else {
					// Wave Trough: vibrant cosmic magenta / violet
					col = rl.NewColor(
						clampU8(70+cAdd),
						clampU8(30+cAdd/4),
						clampU8(160+cAdd),
						clampU8(110+int(amp*135)),
					)
				}
			}

			gridDepthBuf[idx] = totalDepth
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

	// Local High-Definition Spacetime Trampoline for selected body (planet, star, pulsar, or black hole)
	if state.SelectedBodyID != -1 {
		var sel *Body
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				sel = b
				break
			}
		}
		if sel != nil {
			drawLocalBodySpacetimeWell(sel, cfg)
		}
	}
}

// drawLocalBodySpacetimeWell renders a localized high-definition Einsteinian spacetime curvature trampoline
// directly beneath the selected body, enabling close-up inspection of its gravitational potential well.
func drawLocalBodySpacetimeWell(body *Body, cfg *Config) {
	const localRes = 24
	localSpan := float32(math.Max(float64(body.Radius*8.0), 10.0))
	localSpacing := localSpan / float32(localRes)
	halfLocal := localSpan * 0.5

	cPos := body.Position
	baseY := cPos.Y - body.Radius*0.2

	wellRadius := body.Radius * 2.2
	radSq := wellRadius * wellRadius
	maxDepth := float32(math.Min(24.0, math.Max(4.0, math.Pow(body.Mass+1.0, 0.35)*5.0)))

	// For black holes: draw event horizon, photon sphere, and ISCO rings
	if body.TextureType == TextureBlackHole {
		rs := body.Radius * 1.5
		rl.DrawCircle3D(rl.NewVector3(cPos.X, baseY-maxDepth*0.9, cPos.Z), rs, rl.NewVector3(0, 1, 0), 0, rl.NewColor(20, 20, 30, 240))
		rl.DrawCircle3D(rl.NewVector3(cPos.X, baseY-maxDepth*0.6, cPos.Z), rs*1.5, rl.NewVector3(0, 1, 0), 0, rl.NewColor(255, 200, 60, 220))  // Photon Sphere
		rl.DrawCircle3D(rl.NewVector3(cPos.X, baseY-maxDepth*0.3, cPos.Z), rs*3.0, rl.NewVector3(0, 1, 0), 0, rl.NewColor(120, 220, 255, 180)) // ISCO
	} else {
		// Draw concentric geodesic rings
		pulse := float32(math.Sin(float64(rl.GetTime()*4.0)))*0.5 + 0.5
		for rIdx := 1; rIdx <= 4; rIdx++ {
			ringR := body.Radius * float32(rIdx) * 1.6
			ringY := baseY - maxDepth*(radSq/(ringR*ringR+radSq))
			ringCol := rl.NewColor(80, 220, 255, uint8(100+rIdx*25))
			if rIdx == 1 {
				ringCol = rl.NewColor(255, 215, 60, uint8(160+pulse*60))
			}
			rl.DrawCircle3D(rl.NewVector3(cPos.X, ringY, cPos.Z), ringR, rl.NewVector3(0, 1, 0), 0, ringCol)
		}
	}

	// Draw localized mesh lines
	for i := 0; i <= localRes; i++ {
		lx := -halfLocal + float32(i)*localSpacing
		for j := 0; j < localRes; j++ {
			lz1 := -halfLocal + float32(j)*localSpacing
			lz2 := lz1 + localSpacing

			r1Sq := lx*lx + lz1*lz1
			r2Sq := lx*lx + lz2*lz2
			d1 := -maxDepth * (radSq / (r1Sq + radSq))
			d2 := -maxDepth * (radSq / (r2Sq + radSq))

			p1 := rl.NewVector3(cPos.X+lx, baseY+d1, cPos.Z+lz1)
			p2 := rl.NewVector3(cPos.X+lx, baseY+d2, cPos.Z+lz2)
			rl.DrawLine3D(p1, p2, rl.NewColor(60, 190, 240, 160))
		}
	}

	for j := 0; j <= localRes; j++ {
		lz := -halfLocal + float32(j)*localSpacing
		for i := 0; i < localRes; i++ {
			lx1 := -halfLocal + float32(i)*localSpacing
			lx2 := lx1 + localSpacing

			r1Sq := lx1*lx1 + lz*lz
			r2Sq := lx2*lx2 + lz*lz
			d1 := -maxDepth * (radSq / (r1Sq + radSq))
			d2 := -maxDepth * (radSq / (r2Sq + radSq))

			p1 := rl.NewVector3(cPos.X+lx1, baseY+d1, cPos.Z+lz)
			p2 := rl.NewVector3(cPos.X+lx2, baseY+d2, cPos.Z+lz)
			rl.DrawLine3D(p1, p2, rl.NewColor(60, 190, 240, 160))
		}
	}
}

// drawGravitationalVectorField renders a true 3D volumetric multi-elevation lattice of gravitational force vectors
func drawGravitationalVectorField(state *SimState, cfg *Config, camera *OrbitCamera) {
	_ = camera
	if len(state.Bodies) == 0 {
		return
	}

	totalSpan := getStaticSpacetimeSpan(cfg)
	halfSpan := totalSpan * 0.5

	var maxMass float64 = 1.0
	for _, b := range state.Bodies {
		if b.Mass > maxMass {
			maxMass = b.Mass
		}
	}
	minSignifMass := math.Max(0.0001, maxMass*1e-7)

	// Gather heavy bodies
	heavy := make([]*Body, 0, 48)
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				heavy = append(heavy, b)
				break
			}
		}
	}

	if len(state.Bodies) <= 80 {
		for _, b := range state.Bodies {
			if state.SelectedBodyID != -1 && b.ID == state.SelectedBodyID {
				continue
			}
			if b.Mass >= minSignifMass || b.IsStar || b.TextureType == TextureBlackHole {
				heavy = append(heavy, b)
				if len(heavy) >= 48 {
					break
				}
			}
		}
	} else {
		for _, b := range state.Bodies {
			if state.SelectedBodyID != -1 && b.ID == state.SelectedBodyID {
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

	// 3D Volumetric Elevations: Multi-elevation lattice with dense 5-plane configuration
	yElevation := totalSpan * 0.075
	var yPlanes []struct {
		y     float32
		steps int
		alpha uint8
	}

	densityScale := float32(1.0)
	if cfg.DenseVectorField {
		// Ultra-Dense 5-elevation volumetric lattice: 3,448 vectors
		densityScale = 0.72
		yPlanes = []struct {
			y     float32
			steps int
			alpha uint8
		}{
			{-yElevation * 1.8, 20, 115},
			{-yElevation * 0.9, 26, 165},
			{0.0, 36, 245}, // High-resolution primary orbital plane: 36x36 = 1,296 vectors
			{yElevation * 0.9, 26, 165},
			{yElevation * 1.8, 20, 115},
		}
	} else {
		// Standard 3-elevation volumetric lattice: 1,432 vectors
		densityScale = 0.95
		yPlanes = []struct {
			y     float32
			steps int
			alpha uint8
		}{
			{-yElevation, 18, 145},
			{0.0, 28, 235},
			{yElevation, 18, 145},
		}
	}

	draw3DArrow := func(pStart rl.Vector3, dirX, dirY, dirZ, arrowLen float32, col rl.Color) {
		pEnd := rl.NewVector3(pStart.X+dirX*arrowLen, pStart.Y+dirY*arrowLen, pStart.Z+dirZ*arrowLen)
		rl.DrawLine3D(pStart, pEnd, col)

		// 3D orthogonal coordinate frame for 4-fin arrowhead
		dir := rl.NewVector3(dirX, dirY, dirZ)
		var up rl.Vector3
		if float32(math.Abs(float64(dirY))) < 0.9 {
			up = rl.NewVector3(0, 1, 0)
		} else {
			up = rl.NewVector3(1, 0, 0)
		}
		u := rl.Vector3Normalize(rl.Vector3CrossProduct(dir, up))
		v := rl.Vector3CrossProduct(dir, u)

		wLen := arrowLen * 0.28
		wSpread := arrowLen * 0.12
		baseCenter := rl.Vector3Subtract(pEnd, rl.Vector3Scale(dir, wLen))

		w1 := rl.Vector3Add(baseCenter, rl.Vector3Scale(u, wSpread))
		w2 := rl.Vector3Subtract(baseCenter, rl.Vector3Scale(u, wSpread))
		w3 := rl.Vector3Add(baseCenter, rl.Vector3Scale(v, wSpread))
		w4 := rl.Vector3Subtract(baseCenter, rl.Vector3Scale(v, wSpread))

		rl.DrawLine3D(pEnd, w1, col)
		rl.DrawLine3D(pEnd, w2, col)
		rl.DrawLine3D(pEnd, w3, col)
		rl.DrawLine3D(pEnd, w4, col)
	}

	for _, plane := range yPlanes {
		steps := plane.steps
		stepSize := totalSpan / float32(steps-1)
		py := plane.y

		for ix := 0; ix < steps; ix++ {
			px := -halfSpan + float32(ix)*stepSize
			for iz := 0; iz < steps; iz++ {
				pz := -halfSpan + float32(iz)*stepSize

				var gx, gy, gz float32
				for k := 0; k < len(sources); k++ {
					dx := sources[k].x - px
					dy := sources[k].y - py
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

					arrowLen := (1.0 + 3.0*(gLen/(gLen+0.35))) * densityScale
					pStart := rl.NewVector3(px, py, pz)

					ratio := float32(math.Min(1.0, math.Max(0.0, (math.Log10(float64(gLen))+2.5)/3.0)))
					var col rl.Color
					if ratio < 0.33 {
						t := ratio / 0.33
						col = rl.NewColor(
							uint8(40+t*20),
							uint8(160+t*60),
							uint8(255-t*90),
							plane.alpha,
						)
					} else if ratio < 0.66 {
						t := (ratio - 0.33) / 0.33
						col = rl.NewColor(
							uint8(60+t*195),
							uint8(220-t*20),
							uint8(165-t*115),
							plane.alpha,
						)
					} else {
						t := (ratio - 0.66) / 0.34
						col = rl.NewColor(
							uint8(255),
							uint8(200-t*140),
							uint8(50+t*50),
							plane.alpha,
						)
					}

					draw3DArrow(pStart, dirX, dirY, dirZ, arrowLen, col)
				}
			}
		}
	}

	// 3D Inward Gravitational Inflow Vectors around Selected Body / Stars
	for _, b := range heavy {
		if b.IsStar || b.ID == state.SelectedBodyID {
			rAttract := b.Radius * 3.5
			if rAttract < 4.0 {
				rAttract = 4.0
			}
			fluxR := rAttract * 1.8
			fluxLen := rAttract * 0.65

			// 8 spherical radial points pulling inward in 3D
			angles := []struct{ th, ph float64 }{
				{0, 0}, {math.Pi * 0.5, 0}, {math.Pi, 0}, {math.Pi * 1.5, 0},
				{0.78, 0.78}, {2.35, 0.78}, {3.92, -0.78}, {5.49, -0.78},
			}
			col := rl.NewColor(255, 200, 60, 210)
			for _, ang := range angles {
				cosP := float32(math.Cos(ang.ph))
				sinP := float32(math.Sin(ang.ph))
				cosT := float32(math.Cos(ang.th))
				sinT := float32(math.Sin(ang.th))

				spX := b.Position.X + fluxR*cosT*cosP
				spY := b.Position.Y + fluxR*sinP
				spZ := b.Position.Z + fluxR*sinT*cosP

				inX := -cosT * cosP
				inY := -sinP
				inZ := -sinT * cosP

				pStart := rl.NewVector3(spX, spY, spZ)
				draw3DArrow(pStart, inX, inY, inZ, fluxLen, col)
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
	primary, secondary := GetLagrangePair(state)
	if primary == nil || secondary == nil {
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

		// If this point is selected, draw animated pulsating beacon rings!
		if state.SelectedLagrangeIndex == i+1 {
			pulse := float32(math.Sin(float64(rl.GetTime()*6.0)))*0.5 + 0.5
			selR := markerSize * (3.0 + pulse*1.2)
			rl.DrawCircle3D(p, selR, rl.NewVector3(0, 1, 0), 0, rl.NewColor(col.R, col.G, col.B, 255))
			rl.DrawCircle3D(p, selR*0.7, rl.NewVector3(0, 1, 0), 0, rl.NewColor(col.R, col.G, col.B, 180))
			rl.DrawSphere(p, markerSize*0.5, rl.ColorAlpha(col, 0.95))
		}

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
	primary, secondary := GetLagrangePair(state)
	if primary == nil || secondary == nil {
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
		isSelected := state.SelectedLagrangeIndex == i+1
		if isSelected {
			tag = "★ " + tag
		}
		tagW := int32(MeasureTextBoldUI(tag, FontSizeTelemetry)) + 12
		tagH := int32(18)
		tagX := int32(screenPos.X) - tagW/2
		tagY := int32(screenPos.Y) - 18

		if isSelected {
			rl.DrawRectangle(tagX-2, tagY-2, tagW+4, tagH+4, rl.NewColor(20, 45, 80, 240))
			rl.DrawRectangleLinesEx(rl.NewRectangle(float32(tagX-2), float32(tagY-2), float32(tagW+4), float32(tagH+4)), 1.5, rl.Gold)
			DrawTextBoldUI(tag, tagX+6, tagY+2, FontSizeTelemetry, rl.Gold)
		} else {
			rl.DrawRectangle(tagX, tagY, tagW, tagH, rl.NewColor(12, 16, 26, 210))
			rl.DrawRectangleLines(tagX, tagY, tagW, tagH, rl.Fade(col, 0.8))
			DrawTextBoldUI(tag, tagX+6, tagY+2, FontSizeTelemetry, col)
		}
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

	var maxMass float64 = 1.0
	for _, b := range state.Bodies {
		if b.Mass > maxMass {
			maxMass = b.Mass
		}
	}
	minSignifMass := math.Max(0.0001, maxMass*1e-7)

	// Gather heavy bodies for potential evaluation
	heavy := make([]*Body, 0, 48)
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				heavy = append(heavy, b)
				break
			}
		}
	}
	for _, b := range state.Bodies {
		if (state.SelectedBodyID != -1 && b.ID == state.SelectedBodyID) || b.Mass < minSignifMass {
			continue
		}
		heavy = append(heavy, b)
		if len(heavy) >= 48 {
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
