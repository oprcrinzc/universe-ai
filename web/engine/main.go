//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"
)

var (
	simState  *SimState
	simConfig *Config
)

func main() {
	simConfig = &Config{
		G:                   1.0,
		TimeScale:           1.0,
		TimeScaleStep:       0.5,
		SubSteps:            2,
		Softening:           0.8,
		Collision:           CollisionMerge,
		Paused:              false,
		EnableBarnesHut:     true,
		BarnesHutTheta:      0.7,
		EnableRelativity:    false,
		SpeedOfLight:        180.0,
		EnableRocheLimit:    true,
		ParticleGlowMode:    false,
		ShowPotentialGrid:   false,
		SpacetimeScale:      1.0,
		SpacetimeResolution: 80,
		ShowVectorField:     false,
		Show2DCircles:       true,
		Heatmap2DResolution: 64,
	}

	simState = &SimState{
		SelectedBodyID: -1,
	}

	LoadPreset(simState, simConfig, PresetSolarSystem)

	// Expose JS API
	jsGlobal := js.Global()
	gravitySim := jsGlobal.Get("Object").New()

	gravitySim.Set("init", js.FuncOf(jsInit))
	gravitySim.Set("step", js.FuncOf(jsStep))
	gravitySim.Set("getBodiesJSON", js.FuncOf(jsGetBodiesJSON))
	gravitySim.Set("getBodiesBuffer", js.FuncOf(jsGetBodiesBuffer))
	gravitySim.Set("selectBody", js.FuncOf(jsSelectBody))
	gravitySim.Set("applyThrust", js.FuncOf(jsApplyThrust))
	gravitySim.Set("spawnBody", js.FuncOf(jsSpawnBody))
	gravitySim.Set("spawnCluster", js.FuncOf(jsSpawnCluster))
	gravitySim.Set("deleteBody", js.FuncOf(jsDeleteBody))
	gravitySim.Set("setConfig", js.FuncOf(jsSetConfig))
	gravitySim.Set("getConfig", js.FuncOf(jsGetConfig))
	gravitySim.Set("getStats", js.FuncOf(jsGetStats))
	gravitySim.Set("getSpacetimeGrid", js.FuncOf(jsGetSpacetimeGrid))
	gravitySim.Set("getVectorField", js.FuncOf(jsGetVectorField))
	gravitySim.Set("getLagrangePoints", js.FuncOf(jsGetLagrangePoints))
	gravitySim.Set("spawnLagrangeProbe", js.FuncOf(jsSpawnLagrangeProbe))
	gravitySim.Set("getGWStats", js.FuncOf(jsGetGWStats))

	jsGlobal.Set("GravitySim", gravitySim)

	// Keep wasm module alive
	select {}
}

func jsInit(this js.Value, args []js.Value) any {
	preset := PresetSolarSystem
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		preset = PresetType(args[0].Int())
	}
	LoadPreset(simState, simConfig, preset)
	return true
}

func jsStep(this js.Value, args []js.Value) any {
	dt := float32(0.0166)
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		dt = float32(args[0].Float())
	}
	if dt > 0.05 {
		dt = 0.05
	}

	start := time.Now()
	UpdatePhysics(simState, simConfig, dt)
	simState.PhysicsTimeMs = float64(time.Since(start).Microseconds()) / 1000.0
	return simState.PhysicsTimeMs
}

func jsGetBodiesJSON(this js.Value, args []js.Value) any {
	data, err := json.Marshal(simState.Bodies)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// jsGetBodiesBuffer returns a packed Float32Array for ultra-low overhead WebGL updates:
// Stride = 16 floats per body:
// 0: ID
// 1: Pos.X, 2: Pos.Y, 3: Pos.Z
// 4: Vel.X, 5: Vel.Y, 6: Vel.Z
// 7: Radius
// 8: Mass
// 9: Color.R, 10: Color.G, 11: Color.B, 12: Color.A
// 13: Flags (1=Star, 2=BlackHole, 4=Pulsar, 8=Comet, 16=Stationary, 32=Selected)
// 14: TextureType
// 15: Speed
func jsGetBodiesBuffer(this js.Value, args []js.Value) any {
	n := len(simState.Bodies)
	const stride = 16
	totalFloats := n * stride
	buffer := make([]float32, totalFloats)

	for i, b := range simState.Bodies {
		offset := i * stride
		buffer[offset+0] = float32(b.ID)
		buffer[offset+1] = b.Position.X
		buffer[offset+2] = b.Position.Y
		buffer[offset+3] = b.Position.Z
		buffer[offset+4] = b.Velocity.X
		buffer[offset+5] = b.Velocity.Y
		buffer[offset+6] = b.Velocity.Z
		buffer[offset+7] = b.Radius
		buffer[offset+8] = float32(b.Mass)
		buffer[offset+9] = float32(b.Color.R) / 255.0
		buffer[offset+10] = float32(b.Color.G) / 255.0
		buffer[offset+11] = float32(b.Color.B) / 255.0
		buffer[offset+12] = float32(b.Color.A) / 255.0

		var flags float32 = 0
		if b.IsStar {
			flags += 1
		}
		if b.IsBlackHole {
			flags += 2
		}
		if b.IsPulsar {
			flags += 4
		}
		if b.IsComet {
			flags += 8
		}
		if b.IsStationary {
			flags += 16
		}
		if b.ID == simState.SelectedBodyID {
			flags += 32
		}
		buffer[offset+13] = flags

		buffer[offset+14] = float32(b.CelestialIcon)
		spd := float32(math.Sqrt(float64(b.Velocity.X*b.Velocity.X + b.Velocity.Y*b.Velocity.Y + b.Velocity.Z*b.Velocity.Z)))
		buffer[offset+15] = spd
	}

	return float32SliceToJSFloat32Array(buffer)
}

func float32SliceToJSFloat32Array(s []float32) js.Value {
	if len(s) == 0 {
		return js.Global().Get("Float32Array").New(0)
	}
	bytes := float32SliceToBytes(s)
	uint8Arr := js.Global().Get("Uint8Array").New(len(bytes))
	js.CopyBytesToJS(uint8Arr, bytes)
	return js.Global().Get("Float32Array").New(uint8Arr.Get("buffer"), uint8Arr.Get("byteOffset"), len(s))
}

func float32SliceToBytes(s []float32) []byte {
	bytes := make([]byte, len(s)*4)
	for i, f := range s {
		bits := math.Float32bits(f)
		bytes[i*4+0] = byte(bits)
		bytes[i*4+1] = byte(bits >> 8)
		bytes[i*4+2] = byte(bits >> 16)
		bytes[i*4+3] = byte(bits >> 24)
	}
	return bytes
}

func jsSelectBody(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].IsNull() || args[0].IsUndefined() || args[0].Type() != js.TypeNumber {
		simState.SelectedBodyID = -1
		return nil
	}
	id := args[0].Int()
	simState.SelectedBodyID = id

	var body *Body
	for _, b := range simState.Bodies {
		if b.ID == id {
			body = b
			break
		}
	}
	if body == nil {
		return nil
	}

	elem := CalculateOrbitalElements(body, simState.Bodies, simConfig.G)
	data, _ := json.Marshal(map[string]any{
		"id":            body.ID,
		"name":          body.Name,
		"mass":          body.Mass,
		"radius":        body.Radius,
		"is_stationary": body.IsStationary,
		"pos":           body.Position,
		"vel":           body.Velocity,
		"elements":      elem,
	})
	return string(data)
}

func jsApplyThrust(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return false
	}
	id := args[0].Int()
	action := args[1].String()
	amount := float32(0.10)
	if len(args) > 2 {
		amount = float32(args[2].Float())
	}

	var body *Body
	for _, b := range simState.Bodies {
		if b.ID == id {
			body = b
			break
		}
	}
	if body == nil {
		return false
	}

	switch action {
	case "prograde":
		spd := Vector3Length(body.Velocity)
		if spd > 1e-4 {
			dir := Vector3Normalize(body.Velocity)
			body.Velocity = Vector3Add(body.Velocity, Vector3Scale(dir, spd*amount))
		}
	case "retrograde":
		spd := Vector3Length(body.Velocity)
		if spd > 1e-4 {
			dir := Vector3Normalize(body.Velocity)
			body.Velocity = Vector3Subtract(body.Velocity, Vector3Scale(dir, spd*amount))
		}
	case "circularize":
		elem := CalculateOrbitalElements(body, simState.Bodies, simConfig.G)
		if elem.PrimaryID > 0 && elem.CircularSpeed > 0 {
			var primary *Body
			for _, b := range simState.Bodies {
				if b.ID == elem.PrimaryID {
					primary = b
					break
				}
			}
			if primary != nil {
				rVec := Vector3Subtract(body.Position, primary.Position)
				rLen := Vector3Length(rVec)
				if rLen > 1e-3 {
					// Perpendicular in plane: cross with Y or normal
					tangent := Vector3Normalize(Vector3{-rVec.Z, 0, rVec.X})
					body.Velocity = Vector3Add(primary.Velocity, Vector3Scale(tangent, elem.CircularSpeed))
				}
			}
		}
	case "stop":
		body.Velocity = Vector3{0, 0, 0}
	case "toggle_stationary":
		body.IsStationary = !body.IsStationary
		if body.IsStationary {
			body.Velocity = Vector3{0, 0, 0}
		}
	case "inclination_pos":
		body.Velocity.Y += amount * 5.0
	case "inclination_neg":
		body.Velocity.Y -= amount * 5.0
	case "scale_mass":
		body.Mass *= float64(amount)
		if body.Mass < 0.001 {
			body.Mass = 0.001
		}
	}
	return true
}

func jsSpawnBody(this js.Value, args []js.Value) any {
	if len(args) < 11 {
		return -1
	}
	name := args[0].String()
	pos := Vector3{float32(args[1].Float()), float32(args[2].Float()), float32(args[3].Float())}
	vel := Vector3{float32(args[4].Float()), float32(args[5].Float()), float32(args[6].Float())}
	mass := args[7].Float()
	radius := float32(args[8].Float())
	col := Color{uint8(args[9].Int()), uint8(args[10].Int()), uint8(args[11].Int()), 255}

	isStar := false
	if len(args) > 12 {
		isStar = args[12].Bool()
	}
	isBlackHole := false
	if len(args) > 13 {
		isBlackHole = args[13].Bool()
	}

	texType := 1 // Earth
	if isStar {
		texType = 6
	}
	if isBlackHole {
		texType = 7
	}

	b := addBody(simState, name, pos, vel, mass, radius, col, false, isStar, texType)
	b.IsBlackHole = isBlackHole
	return b.ID
}

func jsSpawnCluster(this js.Value, args []js.Value) any {
	if len(args) < 7 {
		return 0
	}
	clusterType := args[0].String()
	center := Vector3{float32(args[1].Float()), float32(args[2].Float()), float32(args[3].Float())}
	vel := Vector3{float32(args[4].Float()), float32(args[5].Float()), float32(args[6].Float())}

	count := 50
	if clusterType == "galaxy150" {
		count = 150
		addGalaxyDisk(simState, center, vel, 2500.0, count, 24.0, simConfig.G, Color{160, 220, 255, 255})
		return count
	} else if clusterType == "collapse100" {
		count = 100
		cloudRadius := 8.5
		totalMass := 600.0
		partMass := totalMass / float64(count)
		rSource := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < count; i++ {
			u := rSource.Float64()
			r := cloudRadius * math.Pow(u, 0.45)
			theta := rSource.Float64() * 2.0 * math.Pi
			phi := math.Acos(2.0*rSource.Float64() - 1.0)
			ox := float32(r * math.Sin(phi) * math.Cos(theta) * 1.3)
			oy := float32(r * math.Sin(phi) * math.Sin(theta) * 0.7)
			oz := float32(r * math.Cos(phi) * 1.1)

			vDisp := float32(0.12)
			vx := vel.X + float32(rSource.Float64()-0.5)*vDisp
			vy := vel.Y + float32(rSource.Float64()-0.5)*vDisp*0.5
			vz := vel.Z + float32(rSource.Float64()-0.5)*vDisp

			var col Color
			roll := rSource.Float64()
			if roll < 0.3 {
				col = Color{100, 210, 255, 255}
			} else if roll < 0.7 {
				col = Color{255, 210, 100, 255}
			} else {
				col = Color{255, 140, 60, 255}
			}

			addBody(simState, fmt.Sprintf("Collapse Particle #%d", i+1),
				Vector3{center.X + ox, center.Y + oy, center.Z + oz}, Vector3{vx, vy, vz},
				partMass*(0.8+rSource.Float64()*0.4), 0.28, col, false, true, 1)
		}
		return count
	}

	rSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < count; i++ {
		r := 2.0 + math.Pow(rSource.Float64(), 1.5)*18.0
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := (rSource.Float64() - 0.5) * 0.5
		x := center.X + float32(r*math.Cos(theta)*math.Cos(phi))
		y := center.Y + float32(r*math.Sin(phi))
		z := center.Z + float32(r*math.Sin(theta)*math.Cos(phi))

		vCirc := float32(math.Sqrt((simConfig.G * 400.0) / (r + 1.0)))
		vx := vel.X - vCirc*float32(math.Sin(theta)) + float32(rSource.Float64()-0.5)*0.5
		vy := vel.Y + float32(rSource.Float64()-0.5)*0.3
		vz := vel.Z + vCirc*float32(math.Cos(theta)) + float32(rSource.Float64()-0.5)*0.5

		addBody(simState, fmt.Sprintf("Cluster Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			0.5, 0.3, Color{255, uint8(180 + rSource.Intn(75)), 120, 255}, false, true, 6)
	}
	return count
}

func jsDeleteBody(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].IsNull() || args[0].IsUndefined() || args[0].Type() != js.TypeNumber {
		return false
	}
	id := args[0].Int()
	newBodies := make([]*Body, 0, len(simState.Bodies))
	for _, b := range simState.Bodies {
		if b.ID != id {
			newBodies = append(newBodies, b)
		}
	}
	simState.Bodies = newBodies
	if simState.SelectedBodyID == id {
		simState.SelectedBodyID = -1
	}
	return true
}

func jsSetConfig(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return false
	}
	key := args[0].String()
	val := args[1]

	switch key {
	case "paused":
		simConfig.Paused = val.Bool()
	case "time_scale":
		simConfig.TimeScale = val.Float()
	case "time_scale_step":
		simConfig.TimeScaleStep = val.Float()
	case "g":
		simConfig.G = val.Float()
	case "enable_barnes_hut":
		simConfig.EnableBarnesHut = val.Bool()
	case "barnes_hut_theta":
		simConfig.BarnesHutTheta = val.Float()
	case "enable_relativity":
		simConfig.EnableRelativity = val.Bool()
	case "enable_roche_limit":
		simConfig.EnableRocheLimit = val.Bool()
	case "collision":
		simConfig.Collision = CollisionMode(val.Int())
	case "particle_glow_mode":
		simConfig.ParticleGlowMode = val.Bool()
	case "show_potential_grid":
		simConfig.ShowPotentialGrid = val.Bool()
	case "show_gravitational_waves":
		simConfig.ShowGravitationalWaves = val.Bool()
	case "spacetime_scale":
		simConfig.SpacetimeScale = float32(val.Float())
	case "spacetime_resolution":
		simConfig.SpacetimeResolution = val.Int()
	case "show_vector_field":
		simConfig.ShowVectorField = val.Bool()
	case "show_vectors":
		simConfig.ShowVectors = val.Bool()
	case "show_2d_circles":
		simConfig.Show2DCircles = val.Bool()
	case "heatmap_2d_resolution":
		simConfig.Heatmap2DResolution = val.Int()
	}
	return true
}

func jsGetConfig(this js.Value, args []js.Value) any {
	data, _ := json.Marshal(simConfig)
	return string(data)
}

func jsGetStats(this js.Value, args []js.Value) any {
	data, _ := json.Marshal(map[string]any{
		"body_count":       len(simState.Bodies),
		"time":             simState.Time,
		"collisions":       simState.CollisionCount,
		"merges":           simState.MergeCount,
		"tidal_breakups":   simState.TidalBreakupCount,
		"physics_time_ms":  simState.PhysicsTimeMs,
		"interactions_sec": simState.InteractionsPerSec,
	})
	return string(data)
}

func jsGetGWStats(this js.Value, args []js.Value) any {
	data, _ := json.Marshal(map[string]any{
		"strain": simState.LastGWStrain,
		"freq":   simState.LastGWFreq,
		"power":  simState.LastGWPower,
		"bursts": len(simState.GWBursts),
	})
	return string(data)
}

func jsGetSpacetimeGrid(this js.Value, args []js.Value) any {
	centerX := float32(0)
	centerZ := float32(0)
	span := float32(300.0)
	steps := 120

	if len(args) >= 4 {
		centerX = float32(args[0].Float())
		centerZ = float32(args[1].Float())
		span = float32(args[2].Float())
		steps = args[3].Int()
	}

	if steps < 36 {
		steps = 120
	}

	halfSpan := span * 0.5
	stepSize := span / float32(steps)
	totalPoints := (steps + 1) * (steps + 1)
	depths := make([]float32, totalPoints)

	// Precompute source bodies if Spacetime potential grid is enabled
	type precomputedBody struct {
		x, z        float32
		scaledMassG float64
	}
	var sources []precomputedBody

	if simConfig.ShowPotentialGrid {
		heavyBodies := make([]*Body, 0, 32)
		selectedIncluded := false

		if simState.SelectedBodyID != -1 {
			for _, b := range simState.Bodies {
				if b.ID == simState.SelectedBodyID {
					heavyBodies = append(heavyBodies, b)
					selectedIncluded = true
					break
				}
			}
		}

		if len(simState.Bodies) <= 50 {
			for _, b := range simState.Bodies {
				if selectedIncluded && b.ID == simState.SelectedBodyID {
					continue
				}
				if b.Mass >= 0.05 || b.IsStar || b.IsBlackHole {
					heavyBodies = append(heavyBodies, b)
				}
			}
		} else {
			for _, b := range simState.Bodies {
				if selectedIncluded && b.ID == simState.SelectedBodyID {
					continue
				}
				if b.Mass >= 15.0 || b.IsBlackHole || b.IsStar {
					heavyBodies = append(heavyBodies, b)
					if len(heavyBodies) >= 32 {
						break
					}
				}
			}
		}
		if len(heavyBodies) == 0 && len(simState.Bodies) > 0 {
			heavyBodies = append(heavyBodies, simState.Bodies[0])
		}

		sources = make([]precomputedBody, len(heavyBodies))
		for k, b := range heavyBodies {
			sm := math.Pow(b.Mass, 0.48) * 8.0
			if b.ID == simState.SelectedBodyID {
				sm *= 1.35
			}
			sources[k] = precomputedBody{
				x:           b.Position.X,
				z:           b.Position.Z,
				scaledMassG: simConfig.G * sm,
			}
		}
	}

	curTime := simState.Time
	gwActive := simConfig.ShowGravitationalWaves

	// Align vertex ordering with Three.js PlaneGeometry(span, span, steps, steps) rotated -PI/2
	// Outer loop j corresponds to Z axis, inner loop i corresponds to X axis
	idx := 0
	for j := 0; j <= steps; j++ {
		gz := centerZ - halfSpan + float32(j)*stepSize
		for i := 0; i <= steps; i++ {
			gx := centerX - halfSpan + float32(i)*stepSize
			var totalDepth float32 = 0.0

			// 1. Spacetime Curvature Potential Wells (Only if ShowPotentialGrid is ON)
			if simConfig.ShowPotentialGrid {
				var potential float64
				for k := 0; k < len(sources); k++ {
					dx := float64(gx - sources[k].x)
					dz := float64(gz - sources[k].z)
					r := math.Sqrt(dx*dx + dz*dz + 1.8)
					potential += sources[k].scaledMassG / r
				}
				totalDepth = -float32(math.Min(38.0, potential*0.14))
			}

			// 2. Dynamic Gravitational Wave Metric Ripples (Only if ShowGravitationalWaves is ON)
			if gwActive {
				gwDisp, _ := ComputeGravitationalWaveDisplacement(simState, simConfig, gx, gz, curTime)
				totalDepth += gwDisp
			}

			depths[idx] = totalDepth
			idx++
		}
	}

	return float32SliceToJSFloat32Array(depths)
}

func jsGetVectorField(this js.Value, args []js.Value) any {
	span := float32(300.0)
	steps := 32

	if len(args) >= 3 {
		span = float32(args[2].Float())
	}
	if len(args) >= 4 {
		steps = args[3].Int()
	}

	if steps < 8 {
		steps = 32
	}
	if span < 10.0 {
		span = 300.0
	}

	totalPoints := steps * steps
	const floatsPerVector = 36 // 6 vertices (shaft 2, wing1 2, wing2 2) * 6 floats (x, y, z, r, g, b)
	buffer := make([]float32, totalPoints*floatsPerVector)

	if len(simState.Bodies) == 0 {
		return float32SliceToJSFloat32Array(buffer)
	}

	stepSize := span / float32(steps-1)
	halfSpan := span * 0.5
	baseX := float32(0.0)
	baseZ := float32(0.0)

	heavy := make([]*Body, 0, 32)
	selectedIncluded := false
	if simState.SelectedBodyID != -1 {
		for _, b := range simState.Bodies {
			if b.ID == simState.SelectedBodyID {
				heavy = append(heavy, b)
				selectedIncluded = true
				break
			}
		}
	}

	if len(simState.Bodies) <= 50 {
		for _, b := range simState.Bodies {
			if selectedIncluded && b.ID == simState.SelectedBodyID {
				continue
			}
			if b.Mass >= 0.05 || b.IsStar || b.IsBlackHole {
				heavy = append(heavy, b)
			}
		}
	} else {
		for _, b := range simState.Bodies {
			if selectedIncluded && b.ID == simState.SelectedBodyID {
				continue
			}
			if b.Mass >= 15.0 || b.IsBlackHole || b.IsStar {
				heavy = append(heavy, b)
				if len(heavy) >= 32 {
					break
				}
			}
		}
	}
	if len(heavy) == 0 && len(simState.Bodies) > 0 {
		heavy = append(heavy, simState.Bodies[0])
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
			scaledG: float32(simConfig.G * b.Mass),
		}
	}

	eps2 := float32(simConfig.Softening*simConfig.Softening + 4.0)

	outIdx := 0
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

				arrowLen := 1.2 + 3.3*(gLen/(gLen+0.35))
				startX := px
				startY := float32(0.0)
				startZ := pz

				endX := px + dirX*arrowLen
				endY := dirY * arrowLen
				endZ := pz + dirZ*arrowLen

				ratio := float32(math.Min(1.0, math.Max(0.0, (math.Log10(float64(gLen))+2.5)/3.0)))
				var cr, cg, cb float32
				if ratio < 0.33 {
					t := ratio / 0.33
					cr = (40.0 + t*20.0) / 255.0
					cg = (160.0 + t*60.0) / 255.0
					cb = (255.0 - t*90.0) / 255.0
				} else if ratio < 0.66 {
					t := (ratio - 0.33) / 0.33
					cr = (60.0 + t*195.0) / 255.0
					cg = (220.0 - t*20.0) / 255.0
					cb = (165.0 - t*115.0) / 255.0
				} else {
					t := (ratio - 0.66) / 0.34
					cr = 1.0
					cg = (200.0 - t*140.0) / 255.0
					cb = (50.0 + t*50.0) / 255.0
				}

				sideX := -dirZ * 0.28 * arrowLen
				sideZ := dirX * 0.28 * arrowLen
				backX := dirX * 0.32 * arrowLen
				backY := dirY * 0.32 * arrowLen
				backZ := dirZ * 0.32 * arrowLen

				w1x := endX - backX + sideX
				w1y := endY - backY
				w1z := endZ - backZ + sideZ

				w2x := endX - backX - sideX
				w2y := endY - backY
				w2z := endZ - backZ - sideZ

				// Segment 1: shaft (start -> end)
				buffer[outIdx+0] = startX
				buffer[outIdx+1] = startY
				buffer[outIdx+2] = startZ
				buffer[outIdx+3] = cr
				buffer[outIdx+4] = cg
				buffer[outIdx+5] = cb
				buffer[outIdx+6] = endX
				buffer[outIdx+7] = endY
				buffer[outIdx+8] = endZ
				buffer[outIdx+9] = cr
				buffer[outIdx+10] = cg
				buffer[outIdx+11] = cb

				// Segment 2: wing 1 (end -> w1)
				buffer[outIdx+12] = endX
				buffer[outIdx+13] = endY
				buffer[outIdx+14] = endZ
				buffer[outIdx+15] = cr
				buffer[outIdx+16] = cg
				buffer[outIdx+17] = cb
				buffer[outIdx+18] = w1x
				buffer[outIdx+19] = w1y
				buffer[outIdx+20] = w1z
				buffer[outIdx+21] = cr
				buffer[outIdx+22] = cg
				buffer[outIdx+23] = cb

				// Segment 3: wing 2 (end -> w2)
				buffer[outIdx+24] = endX
				buffer[outIdx+25] = endY
				buffer[outIdx+26] = endZ
				buffer[outIdx+27] = cr
				buffer[outIdx+28] = cg
				buffer[outIdx+29] = cb
				buffer[outIdx+30] = w2x
				buffer[outIdx+31] = w2y
				buffer[outIdx+32] = w2z
				buffer[outIdx+33] = cr
				buffer[outIdx+34] = cg
				buffer[outIdx+35] = cb
			} else {
				// Zero-length segment collapsed to needle root
				for v := 0; v < 6; v++ {
					buffer[outIdx+v*6+0] = px
					buffer[outIdx+v*6+1] = 0
					buffer[outIdx+v*6+2] = pz
					buffer[outIdx+v*6+3] = 0
					buffer[outIdx+v*6+4] = 0
					buffer[outIdx+v*6+5] = 0
				}
			}
			outIdx += floatsPerVector
		}
	}

	return float32SliceToJSFloat32Array(buffer)
}
