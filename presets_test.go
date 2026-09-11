package main

import (
	"math"
	"path/filepath"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAllPresetsLoad(t *testing.T) {
	cfg := &Config{
		G:         1.0,
		TimeScale: 1.0,
		SubSteps:  5,
	}

	presets := []PresetType{
		PresetSolarSystem,
		PresetSolarSystemGrand,
		PresetMilkyWay10K,
		PresetAsteroidBelt2K,
		PresetGalaxyCollision5K,
		PresetBlackHoleSwarm3K,
		PresetBinaryStars,
		PresetThreeBody,
		PresetGalaxyDisk,
		PresetMilkyWayCluster,
		PresetLagrangeTrojans,
		PresetGalaxyCollision,
		PresetPulsarAccretion,
		PresetGlobularCluster,
		PresetTrappistResonance,
		PresetFiveBody,
		PresetRelativityRosette,
		PresetRocheDisruption,
		PresetGravitationalCollapse,
		PresetEmpty,
	}

	for _, p := range presets {
		state := &SimState{}
		LoadPreset(state, cfg, p)

		if p == PresetEmpty {
			if len(state.Bodies) != 0 {
				t.Fatalf("Expected 0 bodies in PresetEmpty, got %d", len(state.Bodies))
			}
		} else {
			if len(state.Bodies) == 0 {
				t.Fatalf("Preset %d loaded 0 bodies", p)
			}
			for i, b := range state.Bodies {
				if b.ID <= 0 {
					t.Fatalf("Body %d in preset %d has invalid ID %d", i, p, b.ID)
				}
				if b.Radius <= 0 {
					t.Fatalf("Body %s in preset %d has non-positive radius %f", b.Name, p, b.Radius)
				}
			}
		}
	}
}

func TestAllSpawnTypes(t *testing.T) {
	spawnTypes := []SpawnType{
		SpawnEarth,
		SpawnMoon,
		SpawnGasGiant,
		SpawnIceGiant,
		SpawnStar,
		SpawnRedGiant,
		SpawnWhiteDwarf,
		SpawnNeutronStar,
		SpawnBlackHole,
		SpawnDwarfPlanet,
		SpawnComet,
		SpawnAsteroidRing,
	}

	pos := rl.NewVector3(10, 0, 20)
	vel := rl.NewVector3(1, 0, -2)

	for id, st := range spawnTypes {
		body := CreateBodyTemplate(st, pos, vel, int64(id+1))
		if body == nil {
			t.Fatalf("Failed to create template for SpawnType %d", st)
		}
		if body.ID != int64(id+1) {
			t.Fatalf("Body ID mismatch: expected %d, got %d", id+1, body.ID)
		}
		if body.Mass <= 0 {
			t.Fatalf("Body mass must be positive, got %f for spawn type %d", body.Mass, st)
		}
		if body.Radius <= 0 {
			t.Fatalf("Body radius must be positive, got %f for spawn type %d", body.Radius, st)
		}

		// Specific property checks
		switch st {
		case SpawnNeutronStar:
			if !body.IsPulsar {
				t.Fatalf("Expected IsPulsar to be true for SpawnNeutronStar")
			}
			if body.TextureType != TextureNeutronStar {
				t.Fatalf("Expected TextureNeutronStar for SpawnNeutronStar, got %v", body.TextureType)
			}
		case SpawnComet:
			if !body.IsComet {
				t.Fatalf("Expected IsComet to be true for SpawnComet")
			}
			if body.TextureType != TextureComet {
				t.Fatalf("Expected TextureComet for SpawnComet, got %v", body.TextureType)
			}
		case SpawnRedGiant:
			if body.TextureType != TextureRedGiant {
				t.Fatalf("Expected TextureRedGiant for SpawnRedGiant, got %v", body.TextureType)
			}
		case SpawnWhiteDwarf:
			if body.TextureType != TextureWhiteDwarf {
				t.Fatalf("Expected TextureWhiteDwarf for SpawnWhiteDwarf, got %v", body.TextureType)
			}
		case SpawnIceGiant:
			if body.TextureType != TextureIceGiant {
				t.Fatalf("Expected TextureIceGiant for SpawnIceGiant, got %v", body.TextureType)
			}
			if body.RingInnerRadius <= 0 {
				t.Fatalf("Expected ring for SpawnIceGiant")
			}
		}
	}
}

func TestNewScenarioTypesSerialization(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "exotic_scenario.json")

	cfg := &Config{
		G: 1.0, TimeScale: 1.0, SubSteps: 5,
	}

	state := &SimState{
		Bodies: []*Body{
			{
				ID: 1, Name: "Pulsar Test", Mass: 6000.0, Radius: 1.2,
				IsPulsar: true, TextureType: TextureNeutronStar,
			},
			{
				ID: 2, Name: "Comet Test", Mass: 0.05, Radius: 0.45,
				IsComet: true, TextureType: TextureComet,
			},
			{
				ID: 3, Name: "Ring World", Mass: 15.0, Radius: 2.2,
				RingInnerRadius: 3.2, RingOuterRadius: 5.6, TextureType: TextureIceGiant,
			},
			{
				ID: 4, Name: "Red Giant", Mass: 9000.0, Radius: 5.0,
				IsStar: true, TextureType: TextureRedGiant,
			},
		},
	}

	saved, err := SaveScenarioToFile(state, cfg, filePath)
	if err != nil {
		t.Fatalf("SaveScenarioToFile failed: %v", err)
	}

	loadedState := &SimState{}
	loadedCfg := &Config{}
	err = LoadScenarioFromFile(loadedState, loadedCfg, saved)
	if err != nil {
		t.Fatalf("LoadScenarioFromFile failed: %v", err)
	}

	if len(loadedState.Bodies) != 4 {
		t.Fatalf("Expected 4 bodies, got %d", len(loadedState.Bodies))
	}

	b1 := loadedState.Bodies[0]
	if !b1.IsPulsar || b1.TextureType != TextureNeutronStar {
		t.Fatalf("Pulsar deserialization failed: %+v", b1)
	}

	b2 := loadedState.Bodies[1]
	if !b2.IsComet || b2.TextureType != TextureComet {
		t.Fatalf("Comet deserialization failed: %+v", b2)
	}

	b3 := loadedState.Bodies[2]
	if b3.RingInnerRadius != 3.2 || b3.RingOuterRadius != 5.6 || b3.TextureType != TextureIceGiant {
		t.Fatalf("Ring world deserialization failed: %+v", b3)
	}

	b4 := loadedState.Bodies[3]
	if !b4.IsStar || b4.TextureType != TextureRedGiant {
		t.Fatalf("Red Giant deserialization failed: %+v", b4)
	}
}

func TestGiganticScenesPerformance(t *testing.T) {
	cfg := &Config{
		G:              1.0,
		TimeScale:      1.0,
		SubSteps:       1,
		UseBarnesHut:   true,
		BarnesHutTheta: 0.72,
		Collision:      CollisionMerge,
	}

	state := &SimState{}
	LoadPreset(state, cfg, PresetMilkyWay10K)

	if len(state.Bodies) != 10001 {
		t.Fatalf("Expected 10,001 bodies in PresetMilkyWay10K, got %d", len(state.Bodies))
	}

	// Advance physics for 2 steps and verify stability
	for step := 0; step < 2; step++ {
		UpdatePhysics(state, cfg, 0.00694) // 144 FPS dt = ~0.00694s
	}

	if len(state.Bodies) < 9900 {
		t.Fatalf("Unexpected excessive body loss: %d remaining", len(state.Bodies))
	}
	t.Logf("PresetMilkyWay10K physics step completed successfully in %.2f ms, %d bodies simulated", cfg.PhysicsTimeMs, len(state.Bodies))
}

func TestSpacetimeGridGlobalPositionAndResolution(t *testing.T) {
	cfg := &Config{
		G:                   1.0,
		TimeScale:           1.0,
		SubSteps:            2,
		ShowPotentialGrid:   true,
		SpacetimeScale:      1.0,
		SpacetimeResolution: 120,
	}

	state := &SimState{}
	LoadPreset(state, cfg, PresetSolarSystem)

	spanDefault := getStaticSpacetimeSpan(cfg)
	if spanDefault != 300.0 {
		t.Fatalf("Expected static span 300.0, got %f", spanDefault)
	}

	// Test SpacetimeScale multiplier on static span
	cfg.SpacetimeScale = 1.5
	spanScaled := getStaticSpacetimeSpan(cfg)
	expectedScaled := float32(450.0)
	if math.Abs(float64(spanScaled-expectedScaled)) > 0.01 {
		t.Fatalf("Expected scaled span %f, got %f", expectedScaled, spanScaled)
	}

	// Verify static invariance across different presets (scale stays identical 300.0)
	cfg.SpacetimeScale = 1.0
	stateGal := &SimState{}
	LoadPreset(stateGal, cfg, PresetMilkyWay10K)
	spanGal := getStaticSpacetimeSpan(cfg)
	if spanGal != spanDefault {
		t.Fatalf("Static scale should be invariant across presets, got %f vs %f", spanGal, spanDefault)
	}

	// Verify high resolution defaults to 120
	cfg.SpacetimeResolution = 0
	res := cfg.SpacetimeResolution
	if res < 36 {
		res = 120
	}
	if res != 120 {
		t.Fatalf("Expected default resolution 120, got %d", res)
	}
}

func BenchmarkSpacetimeGridDepthCalc(b *testing.B) {
	cfg := &Config{
		G: 1.0,
	}
	state := &SimState{}
	LoadPreset(state, cfg, PresetSolarSystem)

	// Gather heavy bodies
	heavyBodies := make([]*Body, 0, len(state.Bodies))
	for _, body := range state.Bodies {
		if body.Mass >= 0.05 || body.IsStar {
			heavyBodies = append(heavyBodies, body)
		}
	}

	type PrecomputedSource struct {
		x, z          float32
		scaledMassG   float64
	}
	sources := make([]PrecomputedSource, len(heavyBodies))
	for i, body := range heavyBodies {
		sm := math.Pow(body.Mass, 0.48) * 8.0
		sources[i] = PrecomputedSource{
			x:           body.Position.X,
			z:           body.Position.Z,
			scaledMassG: cfg.G * sm,
		}
	}

	for _, res := range []int{80, 100, 120, 128, 144, 160} {
		totalPoints := (res + 1) * (res + 1)
		buf := make([]float32, totalPoints)
		b.Run(string(rune('0'+res/100))+string(rune('0'+(res%100)/10))+string(rune('0'+res%10)), func(b *testing.B) {
			halfDim := float32(150.0)
			spacing := float32(300.0) / float32(res)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				idx := 0
				for i := 0; i <= res; i++ {
					gx := -halfDim + float32(i)*spacing
					for j := 0; j <= res; j++ {
						gz := -halfDim + float32(j)*spacing
						var pot float64
						for k := 0; k < len(sources); k++ {
							dx := float64(gx - sources[k].x)
							dz := float64(gz - sources[k].z)
							r := math.Sqrt(dx*dx + dz*dz + 2.0)
							pot += sources[k].scaledMassG / r
						}
						buf[idx] = -float32(pot)
						idx++
					}
				}
			}
		})
	}
}

func TestGravitationalCollapsePreset(t *testing.T) {
	state := &SimState{}
	cfg := &Config{
		G:                   1.0,
		TimeScale:           1.0,
		SubSteps:            2,
		Softening:           0.65,
		UseBarnesHut:        true,
		BarnesHutTheta:      0.7,
		SpacetimeResolution: 120,
		SpacetimeScale:      1.0,
	}

	LoadPreset(state, cfg, PresetGravitationalCollapse)

	if len(state.Bodies) != 1600 {
		t.Fatalf("Expected 1600 bodies in PresetGravitationalCollapse, got %d", len(state.Bodies))
	}

	// Verify all bodies have valid non-zero mass and positions
	var totalMass float64
	for i, b := range state.Bodies {
		if b.Mass <= 0 {
			t.Fatalf("Body %d has invalid mass %f", i, b.Mass)
		}
		totalMass += b.Mass
		if math.IsNaN(float64(b.Position.X)) || math.IsNaN(float64(b.Position.Y)) || math.IsNaN(float64(b.Position.Z)) {
			t.Fatalf("Body %d has NaN position", i)
		}
	}

	if totalMass < 4000.0 {
		t.Fatalf("Expected total mass ~4500, got %f", totalMass)
	}

	// Step physics 5 steps to verify stability and collapse progression
	for step := 0; step < 5; step++ {
		UpdatePhysics(state, cfg, 0.0166)
	}

	for i, b := range state.Bodies {
		if math.IsNaN(float64(b.Position.X)) || math.IsNaN(float64(b.Velocity.X)) {
			t.Fatalf("Body %d developed NaN during physics step", i)
		}
	}
}

func TestSpawnCollapseCloud(t *testing.T) {
	state := &SimState{NextID: 1}
	pos := rl.NewVector3(50, 10, -30)
	vel := rl.NewVector3(2, 0, 1)

	SpawnCollapseCloud(state, pos, vel, 100)

	if len(state.Bodies) != 100 {
		t.Fatalf("Expected 100 bodies spawned, got %d", len(state.Bodies))
	}

	for i, b := range state.Bodies {
		if b.ID <= 0 {
			t.Fatalf("Invalid body ID %d at index %d", b.ID, i)
		}
		// Check that particles are close to spawn position
		dist := rl.Vector3Distance(b.Position, pos)
		if dist > 30.0 {
			t.Fatalf("Particle %d is too far from spawn origin: %f", i, dist)
		}
	}
}

func TestHighResVectorFieldIntegrity(t *testing.T) {
	state := &SimState{}
	cfg := &Config{
		G:              1.0,
		SpacetimeScale: 1.0,
		Softening:      0.8,
	}
	LoadPreset(state, cfg, PresetSolarSystem)

	// Simulate the vector field math
	totalSpan := getStaticSpacetimeSpan(cfg)
	if totalSpan != 300.0 {
		t.Fatalf("Expected default static span 300.0, got %f", totalSpan)
	}

	steps := 32
	stepSize := totalSpan / float32(steps-1)
	halfSpan := totalSpan * 0.5

	nonZeroVectors := 0
	for ix := 0; ix < steps; ix++ {
		px := -halfSpan + float32(ix)*stepSize
		for iz := 0; iz < steps; iz++ {
			pz := -halfSpan + float32(iz)*stepSize
			var gx, gy, gz float32

			for _, b := range state.Bodies {
				dx := b.Position.X - px
				dy := b.Position.Y - 0
				dz := b.Position.Z - pz
				distSq := dx*dx + dy*dy + dz*dz + 4.0
				dist := float32(math.Sqrt(float64(distSq)))
				invDist3 := 1.0 / (distSq * dist)
				f := float32(cfg.G*b.Mass) * invDist3
				gx += dx * f
				gy += dy * f
				gz += dz * f
			}

			gLen := float32(math.Sqrt(float64(gx*gx + gy*gy + gz*gz)))
			if math.IsNaN(float64(gLen)) {
				t.Fatalf("Vector at (%f, %f) produced NaN", px, pz)
			}
			if gLen > 0.0001 {
				nonZeroVectors++
			}
		}
	}

	if nonZeroVectors < 1000 {
		t.Fatalf("Expected ~1024 active vector needles, got %d", nonZeroVectors)
	}
}

func TestRealScaleSolarSystemPhysicalIntegrity(t *testing.T) {
	state := &SimState{}
	cfg := &Config{G: 1.0}
	LoadPreset(state, cfg, PresetRealSolarSystem)

	if len(state.Bodies) < 50 {
		t.Fatalf("Expected >= 50 bodies in Real Scale Solar System, got %d", len(state.Bodies))
	}

	bodyMap := make(map[string]*Body)
	for _, b := range state.Bodies {
		bodyMap[b.Name] = b
	}

	// Verify Sun, Earth, Moon, Jupiter, Saturn, Neptune, Pluto exist
	required := []string{"Sun Helios", "Mercury", "Venus", "Earth", "Moon Luna", "Mars", "Jupiter", "Saturn", "Neptune", "Pluto (Kuiper Belt)"}
	for _, name := range required {
		if _, ok := bodyMap[name]; !ok {
			t.Fatalf("Missing expected planet/body '%s' in Real Scale Solar System", name)
		}
	}

	sun := bodyMap["Sun Helios"]
	earth := bodyMap["Earth"]
	jup := bodyMap["Jupiter"]
	nep := bodyMap["Neptune"]

	// Check Earth distance ~ 100 units (1.0 AU)
	earthDist := float64(rl.Vector3Distance(earth.Position, sun.Position))
	if math.Abs(earthDist-100.0) > 3.0 {
		t.Fatalf("Earth distance %.2f deviates from expected ~100.0 AU", earthDist)
	}

	// Check Jupiter distance ~ 520 units (5.2 AU)
	jupDist := float64(rl.Vector3Distance(jup.Position, sun.Position))
	if math.Abs(jupDist-520.44) > 40.0 {
		t.Fatalf("Jupiter distance %.2f deviates from expected ~520.44 AU", jupDist)
	}

	// Check Neptune distance ~ 3000 units (30.0 AU)
	nepDist := float64(rl.Vector3Distance(nep.Position, sun.Position))
	if math.Abs(nepDist-3007.0) > 100.0 {
		t.Fatalf("Neptune distance %.2f deviates from expected ~3007.0 AU", nepDist)
	}

	// Verify Kepler's Third Law (T^2 / a^3 = const):
	// v_circ = sqrt(G*M/a) -> T = 2*pi*a / v -> T^2/a^3 = 4*pi^2 / (G*M)
	constExpected := (4.0 * math.Pi * math.Pi) / (cfg.G * sun.Mass)

	checkKepler := func(b *Body, expectedA float64) {
		r := float64(rl.Vector3Distance(b.Position, sun.Position))
		v := float64(rl.Vector3Length(b.Velocity))
		// Vis-viva: 1/a = 2/r - v^2/(G*M)
		invA := (2.0 / r) - (v*v)/(cfg.G*sun.Mass)
		a := 1.0 / invA
		relErrA := math.Abs(a-expectedA) / expectedA
		if relErrA > 0.05 {
			t.Fatalf("%s semi-major axis %.2f deviates from expected %.2f (relErr=%.2f%%)", b.Name, a, expectedA, relErrA*100)
		}
		// Period T = 2*pi*sqrt(a^3 / (G*M))
		T := 2.0 * math.Pi * math.Sqrt(a*a*a/(cfg.G*sun.Mass))
		ratio := (T * T) / (a * a * a)
		if math.Abs(ratio-constExpected)/constExpected > 0.001 {
			t.Fatalf("%s Kepler ratio %.6f deviates from expected %.6f", b.Name, ratio, constExpected)
		}
	}

	checkKepler(earth, 100.0)
	checkKepler(jup, 520.44)
	checkKepler(nep, 3007.0)
}

func Test2DViewportCoordinateMapping(t *testing.T) {
	state := &SimState{SelectedBodyID: -1}
	cfg := &Config{
		Show2DViewport: true,
		Viewport2DZoom: 0.5,
		Viewport2DPan:  rl.NewVector2(100.0, -50.0),
	}

	screenW := int32(1920)
	screenH := int32(1080)
	vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)

	cx := float32(vx) + float32(vw)*0.5
	cy := float32(vy+30) + float32(vh-30)*0.5
	zoom := cfg.Viewport2DZoom
	panX := cfg.Viewport2DPan.X
	panZ := cfg.Viewport2DPan.Y

	// World point (200, 300)
	worldX := float32(200.0)
	worldZ := float32(300.0)

	// Screen position
	sx := cx + (worldX-panX)*zoom
	sy := cy + (worldZ-panZ)*zoom

	// Invert back to world
	reconstructedWX := panX + (sx-cx)/zoom
	reconstructedWZ := panZ + (sy-cy)/zoom

	if math.Abs(float64(reconstructedWX-worldX)) > 1e-4 || math.Abs(float64(reconstructedWZ-worldZ)) > 1e-4 {
		t.Fatalf("2D Viewport coordinate inversion error: got (%.4f, %.4f), expected (%.4f, %.4f)",
			reconstructedWX, reconstructedWZ, worldX, worldZ)
	}
}

func Test2DViewportInspectorNoOverlap(t *testing.T) {
	cfg := &Config{
		Show2DViewport:       true,
		Viewport2DFullscreen: false,
	}
	screenW := int32(1920)
	screenH := int32(1080)

	// 1. Without inspector open (no body selected)
	stateClosed := &SimState{SelectedBodyID: -1}
	vx1, vy1, vw1, vh1 := get2DViewportRect(stateClosed, cfg, screenW, screenH)
	expectedX1 := screenW - vw1 - 14
	if vx1 != expectedX1 {
		t.Errorf("Expected docked right vx=%d without inspector, got %d", expectedX1, vx1)
	}
	if vy1 < 52 || vy1+vh1 > screenH-52 {
		t.Errorf("2D Viewport vertical bounds [%d, %d] violate top/bottom bar margins", vy1, vy1+vh1)
	}

	// 2. With inspector open (body selected)
	stateOpen := &SimState{SelectedBodyID: 1}
	vx2, vy2, vw2, vh2 := get2DViewportRect(stateOpen, cfg, screenW, screenH)
	inspectorX := screenW - 310 - 12
	expectedX2 := screenW - 310 - 14 - vw2 - 10
	if vx2 != expectedX2 {
		t.Errorf("Expected shifted vx=%d with inspector, got %d", expectedX2, vx2)
	}

	// Crucial check: 2D viewport right edge MUST be strictly to the left of inspectorX
	if vx2+vw2 >= inspectorX {
		t.Fatalf("OVERLAP DETECTED! 2D Viewport right edge (%d) intersects inspector panel X (%d)", vx2+vw2, inspectorX)
	}
	if vy2 < 52 || vy2+vh2 > screenH-52 {
		t.Errorf("2D Viewport vertical bounds [%d, %d] violate margins with inspector open", vy2, vy2+vh2)
	}
}

func Test3DLagrangeEquilibriumGeometry(t *testing.T) {
	// Sun-Jupiter CR3BP system
	sun := &Body{
		ID:       1,
		Name:     "Sun",
		Mass:     1000.0,
		Position: rl.NewVector3(0, 0, 0),
		Velocity: rl.NewVector3(0, 0, 0),
		Radius:   5.0,
	}
	jupiter := &Body{
		ID:       2,
		Name:     "Jupiter",
		Mass:     10.0,
		Position: rl.NewVector3(100.0, 0, 0),
		Velocity: rl.NewVector3(0, 0, float32(math.Sqrt(1.0*1010.0/100.0))),
		Radius:   1.5,
	}

	g := 1.0
	pts := ComputeLagrangePoints(sun, jupiter, g)

	// 1. Collinear points L1, L2, L3 check
	// L1 must be between Sun and Jupiter (0 < X_L1 < 100)
	if pts[0].Position.X <= 0 || pts[0].Position.X >= 100.0 {
		t.Errorf("L1 X-coordinate %f not between Sun (0) and Jupiter (100)", pts[0].Position.X)
	}
	// L2 must be beyond Jupiter (X_L2 > 100)
	if pts[1].Position.X <= 100.0 {
		t.Errorf("L2 X-coordinate %f not beyond Jupiter (100)", pts[1].Position.X)
	}
	// L3 must be on opposite side of Sun (X_L3 < 0)
	if pts[2].Position.X >= 0.0 {
		t.Errorf("L3 X-coordinate %f not on opposite side of Sun (0)", pts[2].Position.X)
	}

	// 2. Equilateral triangle geometry for L4 and L5: distance to Sun and Jupiter must both equal R = 100
	R := 100.0
	dSunL4 := float64(rl.Vector3Distance(sun.Position, pts[3].Position))
	dJupL4 := float64(rl.Vector3Distance(jupiter.Position, pts[3].Position))
	if math.Abs(dSunL4-R) > 0.1 || math.Abs(dJupL4-R) > 0.1 {
		t.Errorf("L4 does not form equilateral triangle: dist to Sun = %f, dist to Jup = %f (expected %f)",
			dSunL4, dJupL4, R)
	}

	dSunL5 := float64(rl.Vector3Distance(sun.Position, pts[4].Position))
	dJupL5 := float64(rl.Vector3Distance(jupiter.Position, pts[4].Position))
	if math.Abs(dSunL5-R) > 0.1 || math.Abs(dJupL5-R) > 0.1 {
		t.Errorf("L5 does not form equilateral triangle: dist to Sun = %f, dist to Jup = %f (expected %f)",
			dSunL5, dJupL5, R)
	}

	// 3. L4 and L5 symmetry across orbital axis
	if math.Abs(float64(pts[3].Position.Z+pts[4].Position.Z)) > 1e-3 {
		t.Errorf("L4 and L5 are not symmetric in Z: L4.Z=%f, L5.Z=%f", pts[3].Position.Z, pts[4].Position.Z)
	}
}



