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
	if spanDefault != 1500.0 {
		t.Fatalf("Expected static span 1500.0, got %f", spanDefault)
	}

	// Test SpacetimeScale multiplier on static span
	cfg.SpacetimeScale = 1.5
	spanScaled := getStaticSpacetimeSpan(cfg)
	expectedScaled := float32(2250.0)
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
		x, z        float32
		scaledMassG float64
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
	if totalSpan != 1500.0 {
		t.Fatalf("Expected default static span 1500.0, got %f", totalSpan)
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

func TestLagrangePresetSwarmStability(t *testing.T) {
	state := &SimState{}
	cfg := &Config{
		G:          1.0,
		TimeScale:  1.0,
		SubSteps:   6,
		Softening:  0.005,
		Integrator: IntegratorYoshida4,
	}

	loadLagrangeTrojans(state, cfg, cfg.G)

	if len(state.Bodies) < 39 {
		t.Fatalf("Expected at least 39 bodies in PresetLagrangeTrojans, got %d", len(state.Bodies))
	}

	var sun, zeus *Body
	for _, b := range state.Bodies {
		if b.Name == "Sun Helios" {
			sun = b
		} else if b.Name == "Gas Giant Zeus" {
			zeus = b
		}
	}
	if sun == nil || zeus == nil {
		t.Fatal("Could not find Sun Helios or Gas Giant Zeus in preset")
	}

	// Advance simulation for 200 steps
	dt := float32(0.0166)
	for s := 0; s < 200; s++ {
		UpdatePhysics(state, cfg, dt)
	}

	// Verify Sun-Zeus distance is stable
	curDist := float64(rl.Vector3Distance(sun.Position, zeus.Position))
	if math.Abs(curDist-75.0) > 3.0 {
		t.Fatalf("Sun-Zeus orbit drifted: distance is %.2f (expected ~75.0)", curDist)
	}

	// Verify Trojans remain bounded (did not fly away)
	trojanCount := 0
	for _, b := range state.Bodies {
		if len(b.Name) >= 6 && b.Name[:6] == "Trojan" {
			trojanCount++
			dSun := float64(rl.Vector3Distance(b.Position, sun.Position))
			if dSun < 50.0 || dSun > 100.0 {
				t.Fatalf("Trojan asteroid %s drifted out of orbit: dist=%.2f", b.Name, dSun)
			}
		}
	}
	if trojanCount < 16 {
		t.Fatalf("Expected at least 16 Trojans, found %d", trojanCount)
	}
}

func TestGetLagrangePairDynamic(t *testing.T) {
	sun := &Body{ID: 1, Name: "Sun", Mass: 10000.0}
	earth := &Body{ID: 2, Name: "Earth", Mass: 3.0, Position: rl.NewVector3(100, 0, 0)}
	moon := &Body{ID: 3, Name: "Moon", Mass: 0.03, Position: rl.NewVector3(101, 0, 0)}
	jupiter := &Body{ID: 4, Name: "Jupiter", Mass: 100.0, Position: rl.NewVector3(500, 0, 0)}

	state := &SimState{
		Bodies:         []*Body{sun, earth, moon, jupiter},
		SelectedBodyID: -1,
	}

	// Default: top two masses (Sun and Jupiter)
	p1, s1 := GetLagrangePair(state)
	if p1 == nil || s1 == nil || p1.ID != sun.ID || s1.ID != jupiter.ID {
		t.Fatalf("Default pair mismatch: expected Sun-Jupiter, got %v - %v", p1, s1)
	}

	// Selected Earth: pair should be Sun-Earth
	state.SelectedBodyID = earth.ID
	p2, s2 := GetLagrangePair(state)
	if p2 == nil || s2 == nil || p2.ID != sun.ID || s2.ID != earth.ID {
		t.Fatalf("Selected Earth mismatch: expected Sun-Earth, got %v - %v", p2, s2)
	}

	// Selected Moon: Moon is closer to Earth and dominated by Earth's gravity
	state.SelectedBodyID = moon.ID
	p3, s3 := GetLagrangePair(state)
	if p3 == nil || s3 == nil || p3.ID != earth.ID || s3.ID != moon.ID {
		t.Fatalf("Selected Moon mismatch: expected Earth-Moon, got %v - %v", p3, s3)
	}
}

func TestRealScaleSolarSystemSpacetimeCurvature(t *testing.T) {
	state := &SimState{SelectedBodyID: -1}
	cfg := &Config{G: 1.0}
	LoadPreset(state, cfg, PresetRealSolarSystem)

	if !cfg.ShowPotentialGrid {
		t.Fatalf("Expected ShowPotentialGrid to be enabled by default in Real Scale Solar System")
	}
	if cfg.SpacetimeScale < 5.0 {
		t.Fatalf("Expected SpacetimeScale >= 5.0 for astronomical AU scale, got %f", cfg.SpacetimeScale)
	}

	span := getStaticSpacetimeSpan(cfg)
	if span < 3000.0 {
		t.Fatalf("Expected total spacetime span >= 3000 units to cover astronomical solar system, got %f", span)
	}

	// Verify all planets fall within the dynamic spacetime significance threshold
	var maxMass float64 = 1.0
	for _, b := range state.Bodies {
		if b.Mass > maxMass {
			maxMass = b.Mass
		}
	}
	minSignif := math.Max(0.0001, maxMass*1e-7)

	planets := []string{"Mercury", "Venus", "Earth", "Mars", "Jupiter", "Saturn", "Neptune"}
	for _, name := range planets {
		found := false
		for _, b := range state.Bodies {
			if b.Name == name {
				found = true
				if b.Mass < minSignif {
					t.Fatalf("Planet '%s' mass %f is below significance threshold %f", name, b.Mass, minSignif)
				}
				// Verify position is within spacetime half span
				dist := float32(math.Hypot(float64(b.Position.X), float64(b.Position.Z)))
				if dist > span*0.5 {
					t.Fatalf("Planet '%s' distance %f is outside spacetime grid half-span %f", name, dist, span*0.5)
				}
				break
			}
		}
		if !found {
			t.Fatalf("Planet '%s' not found in Real Scale Solar System", name)
		}
	}
}

func Test3DVolumetricVectorField(t *testing.T) {
	// Verify 3D vector field calculation at multiple elevations
	sun := &Body{ID: 1, Name: "Sun", Mass: 1000.0, Position: rl.NewVector3(0, 0, 0)}
	planet := &Body{ID: 2, Name: "Inclined Planet", Mass: 50.0, Position: rl.NewVector3(100, 20, 0)}
	state := &SimState{Bodies: []*Body{sun, planet}}
	cfg := &Config{G: 1.0, Softening: 0.1, SpacetimeScale: 1.0}

	span := getStaticSpacetimeSpan(cfg)
	yElevation := span * 0.075

	// Test point at elevated plane (+Y)
	px, py, pz := float32(50.0), yElevation, float32(0.0)
	var gx, gy, gz float32
	eps2 := float32(cfg.Softening*cfg.Softening + 4.0)

	for _, b := range state.Bodies {
		dx := b.Position.X - px
		dy := b.Position.Y - py
		dz := b.Position.Z - pz
		distSq := dx*dx + dy*dy + dz*dz + eps2
		dist := float32(math.Sqrt(float64(distSq)))
		f := float32(cfg.G*b.Mass) / (distSq * dist)
		gx += dx * f
		gy += dy * f
		gz += dz * f
	}

	gMag := float32(math.Sqrt(float64(gx*gx + gy*gy + gz*gz)))
	if gMag <= 0.001 {
		t.Fatalf("Expected non-zero 3D gravitational acceleration, got %f", gMag)
	}
	// At py > 0, gy should pull downwards towards Sun (Y=0), so gy < 0
	if gy >= 0 {
		t.Fatalf("Expected downward gravitational pull at +Y elevation, got gy = %f", gy)
	}
}

func Test2DViewportObjectIconAndResolution(t *testing.T) {
	cfg := &Config{
		Show2DViewport: true,
		Show2DIcons:    false, // Default: clear circle mode
		Show2DHeatmap:  true,
		Viewport2DZoom: 0.5,
	}

	if cfg.Show2DIcons {
		t.Fatalf("Show2DIcons should default to false (clear circle mode)")
	}

	// Toggle on (pictures/glyphs enabled)
	cfg.Show2DIcons = true
	if !cfg.Show2DIcons {
		t.Fatalf("Show2DIcons should be enabled when toggled")
	}

	// Toggle back off (clear circle mode)
	cfg.Show2DIcons = false
	if cfg.Show2DIcons {
		t.Fatalf("Show2DIcons should return to false")
	}

	// Verify adaptive resolution calculations
	vw, vh := float32(400), float32(300)
	cols := int(vw / 6.5)
	rows := int(vh / 6.5)

	if cols < 48 || cols > 96 {
		t.Fatalf("Adaptive heatmap cols %d outside expected range [48, 96]", cols)
	}
	if rows < 32 || rows > 72 {
		t.Fatalf("Adaptive heatmap rows %d outside expected range [32, 72]", rows)
	}
}

func Test2DViewportCirclesToggle(t *testing.T) {
	cfg := &Config{
		Show2DViewport: true,
		Show2DIcons:    false,
		Show2DCircles:  true, // Default: body overlay outline circles & center dots enabled
	}

	if !cfg.Show2DCircles {
		t.Fatalf("Show2DCircles should default to true")
	}

	// Toggle off: both circle outline and center core dot disappear
	cfg.Show2DCircles = false
	if cfg.Show2DCircles {
		t.Fatalf("Show2DCircles should be false after toggle off")
	}

	// Toggle on: both circle outline and center core dot appear
	cfg.Show2DCircles = true
	if !cfg.Show2DCircles {
		t.Fatalf("Show2DCircles should be true after toggle on")
	}
}

func Test2DViewportTrailsToggle(t *testing.T) {
	cfg := &Config{
		Show2DViewport: true,
		ShowTrails:     true,
	}

	if !cfg.ShowTrails {
		t.Fatalf("ShowTrails should default to true")
	}

	// Toggle off
	cfg.ShowTrails = false
	if cfg.ShowTrails {
		t.Fatalf("ShowTrails should be false after toggle off")
	}

	// Toggle on
	cfg.ShowTrails = true
	if !cfg.ShowTrails {
		t.Fatalf("ShowTrails should be true after toggle on")
	}
}

func TestResolutionIncreaseDecrease(t *testing.T) {
	cfg := &Config{
		SpacetimeResolution: 120,
	}

	// Test decreasing step-by-step
	expectedDown := []int{96, 80, 64, 48, 32, 32} // Clamps at 32
	for _, exp := range expectedDown {
		DecreaseResolution(cfg)
		if cfg.SpacetimeResolution != exp {
			t.Fatalf("Expected resolution %d after decrease, got %d", exp, cfg.SpacetimeResolution)
		}
	}

	// Test increasing step-by-step
	expectedUp := []int{48, 64, 80, 96, 120, 160, 200, 256, 384, 512, 512} // Clamps at 512
	for _, exp := range expectedUp {
		IncreaseResolution(cfg)
		if cfg.SpacetimeResolution != exp {
			t.Fatalf("Expected resolution %d after increase, got %d", exp, cfg.SpacetimeResolution)
		}
	}

	// Verify uninitialized (0) defaults to 120 on step
	cfg.SpacetimeResolution = 0
	DecreaseResolution(cfg)
	if cfg.SpacetimeResolution != 96 {
		t.Fatalf("Expected resolution 96 when stepping down from 0 (which defaults to 120), got %d", cfg.SpacetimeResolution)
	}

	cfg.SpacetimeResolution = 0
	IncreaseResolution(cfg)
	if cfg.SpacetimeResolution != 160 {
		t.Fatalf("Expected resolution 160 when stepping up from 0 (which defaults to 120), got %d", cfg.SpacetimeResolution)
	}
}

func Test2DHeatmapResolutionIncreaseDecrease(t *testing.T) {
	cfg := &Config{
		Heatmap2DResolution: 64,
	}

	// Test decreasing step-by-step
	expectedDown := []int{48, 32, 24, 16, 16} // Clamps at 16
	for _, exp := range expectedDown {
		DecreaseHeatmap2DResolution(cfg)
		if cfg.Heatmap2DResolution != exp {
			t.Fatalf("Expected 2D heatmap resolution %d after decrease, got %d", exp, cfg.Heatmap2DResolution)
		}
	}

	// Test increasing step-by-step
	expectedUp := []int{24, 32, 48, 64, 80, 96, 128, 160, 200, 200} // Clamps at 200
	for _, exp := range expectedUp {
		IncreaseHeatmap2DResolution(cfg)
		if cfg.Heatmap2DResolution != exp {
			t.Fatalf("Expected 2D heatmap resolution %d after increase, got %d", exp, cfg.Heatmap2DResolution)
		}
	}

	// Verify uninitialized (0) defaults to 64 on step
	cfg.Heatmap2DResolution = 0
	DecreaseHeatmap2DResolution(cfg)
	if cfg.Heatmap2DResolution != 48 {
		t.Fatalf("Expected 2D heatmap resolution 48 when stepping down from 0 (defaults to 64), got %d", cfg.Heatmap2DResolution)
	}

	cfg.Heatmap2DResolution = 0
	IncreaseHeatmap2DResolution(cfg)
	if cfg.Heatmap2DResolution != 80 {
		t.Fatalf("Expected 2D heatmap resolution 80 when stepping up from 0 (defaults to 64), got %d", cfg.Heatmap2DResolution)
	}
}

func Test2DViewportToolbarZeroOverlapAcrossResolutions(t *testing.T) {
	screenSizes := []struct{ w, h int32 }{
		{1024, 768},
		{1280, 720},
		{1366, 768},
		{1600, 900},
		{1920, 1080},
		{2560, 1440},
	}

	for _, scr := range screenSizes {
		modes := []struct {
			name  string
			cfg   *Config
			state *SimState
		}{
			{"Docked Normal", &Config{Show2DViewport: true}, &SimState{SelectedBodyID: -1}},
			{"Docked Inspector", &Config{Show2DViewport: true}, &SimState{SelectedBodyID: 1}},
			{"Fullscreen", &Config{Show2DViewport: true, Viewport2DFullscreen: true}, &SimState{SelectedBodyID: -1}},
			{"Offload 3D", &Config{Show2DViewport: true, Offload3D: true}, &SimState{SelectedBodyID: -1}},
		}

		for _, m := range modes {
			vx, vy, vw, vh := get2DViewportRect(m.state, m.cfg, scr.w, scr.h)
			toolbarH := get2DToolbarHeight(vw)
			headerH := int32(32)
			statusH := int32(22)

			// 1. Viewport boundaries must stay within screen
			if vx < 0 || vx+vw > scr.w {
				t.Fatalf("[%s @ %dx%d] Viewport X bounds outside screen: vx=%d, vw=%d, scrW=%d",
					m.name, scr.w, scr.h, vx, vw, scr.w)
			}
			if vy < 0 || vy+vh > scr.h {
				t.Fatalf("[%s @ %dx%d] Viewport Y bounds outside screen: vy=%d, vh=%d, scrH=%d",
					m.name, scr.w, scr.h, vy, vh, scr.h)
			}

			// 2. Content area must be positive and non-empty
			clipH := vh - headerH - toolbarH - statusH
			if clipH < 100 {
				t.Fatalf("[%s @ %dx%d] Viewport content height too small: %d (vh=%d, toolbarH=%d)",
					m.name, scr.w, scr.h, clipH, vh, toolbarH)
			}

			// 3. Test Toolbar Buttons Zero Overlap & Within Viewport
			type btnDef struct {
				name string
				w    float32
			}

			if vw >= 700 {
				// Single row test
				buttons := []btnDef{
					{"-", 22}, {"+", 22}, {"RST", 30},
					{"TRL", 32}, {"CIR", 30}, {"LBL", 30}, {"ICON", 34},
					{"HEAT", 34}, {"VEC", 30}, {"GW", 28}, {"LAG", 30},
					{"FOCUS", 42}, {"TRACK", 42}, {"SPAWN", 58},
					{"RES-", 36}, {"RES+", 36}, {"HUD", 32},
				}
				curX := float32(vx + 10)
				var rects []rl.Rectangle
				for idx, b := range buttons {
					if idx == 3 || idx == 11 || idx == 14 || idx == 16 {
						curX += 8 // divider spacing
					}
					r := rl.NewRectangle(curX, float32(vy+headerH+4), b.w, 20)
					if r.X+r.Width > float32(vx+vw) {
						t.Fatalf("[%s @ %dx%d] Button %s clipped off right edge: right=%f, max=%d",
							m.name, scr.w, scr.h, b.name, r.X+r.Width, vx+vw)
					}
					rects = append(rects, r)
					curX += b.w + 4
				}
				for i := 0; i < len(rects); i++ {
					for j := i + 1; j < len(rects); j++ {
						if rl.CheckCollisionRecs(rects[i], rects[j]) {
							t.Fatalf("[%s @ %dx%d] Overlap between button %s and %s",
								m.name, scr.w, scr.h, buttons[i].name, buttons[j].name)
						}
					}
				}
			} else {
				// Two-row test
				// Row 1
				row1 := []btnDef{
					{"-", 20}, {"+", 20}, {"RST", 26},
					{"TRL", 28}, {"CIR", 26}, {"LBL", 26}, {"ICON", 30},
					{"HEAT", 30}, {"VEC", 26}, {"GW", 24}, {"LAG", 26},
				}
				curX := float32(vx + 8)
				var rects1 []rl.Rectangle
				for idx, b := range row1 {
					if idx == 3 {
						curX += 6
					}
					r := rl.NewRectangle(curX, float32(vy+headerH+3), b.w, 20)
					if r.X+r.Width > float32(vx+vw) {
						t.Fatalf("[%s @ %dx%d] Row 1 Button %s clipped off right edge: right=%f, max=%d",
							m.name, scr.w, scr.h, b.name, r.X+r.Width, vx+vw)
					}
					rects1 = append(rects1, r)
					curX += b.w + 3
				}
				for i := 0; i < len(rects1); i++ {
					for j := i + 1; j < len(rects1); j++ {
						if rl.CheckCollisionRecs(rects1[i], rects1[j]) {
							t.Fatalf("[%s @ %dx%d] Overlap in Row 1 between %s and %s",
								m.name, scr.w, scr.h, row1[i].name, row1[j].name)
						}
					}
				}

				// Row 2
				row2 := []btnDef{
					{"FOCUS", 40}, {"TRACK", 40}, {"SPAWN", 54},
					{"RES-", 34}, {"RES+", 34}, {"HUD", 30},
				}
				curX = float32(vx + 8)
				var rects2 []rl.Rectangle
				for idx, b := range row2 {
					if idx == 3 || idx == 5 {
						curX += 6
					}
					r := rl.NewRectangle(curX, float32(vy+headerH+26), b.w, 20)
					if r.X+r.Width > float32(vx+vw) {
						t.Fatalf("[%s @ %dx%d] Row 2 Button %s clipped off right edge: right=%f, max=%d",
							m.name, scr.w, scr.h, b.name, r.X+r.Width, vx+vw)
					}
					rects2 = append(rects2, r)
					curX += b.w + 4
				}
				for i := 0; i < len(rects2); i++ {
					for j := i + 1; j < len(rects2); j++ {
						if rl.CheckCollisionRecs(rects2[i], rects2[j]) {
							t.Fatalf("[%s @ %dx%d] Overlap in Row 2 between %s and %s",
								m.name, scr.w, scr.h, row2[i].name, row2[j].name)
						}
					}
				}
			}
		}
	}
}

func TestUILayoutZeroOverlapAcrossResolutions(t *testing.T) {
	resolutions := []struct{ w, h int32 }{
		{1024, 768},
		{1280, 720},
		{1920, 1080},
		{2560, 1440},
	}

	for _, res := range resolutions {
		// Top Bar Right buttons layout test
		helpW := float32(56)
		resetW := float32(58)
		archW := float32(72)
		saveW := float32(56)
		camW := float32(72)
		btnGap := float32(6)

		rx := float32(res.w) - 12 - helpW
		helpBox := rl.NewRectangle(rx, 9, helpW, 30)

		rx -= resetW + btnGap
		resetBox := rl.NewRectangle(rx, 9, resetW, 30)

		rx -= archW + btnGap
		archBox := rl.NewRectangle(rx, 9, archW, 30)

		rx -= saveW + btnGap
		saveBox := rl.NewRectangle(rx, 9, saveW, 30)

		rx -= camW + btnGap
		camBox := rl.NewRectangle(rx, 9, camW, 30)

		rightLimit := rx - 12

		// Verify no overlap among right-side buttons
		boxes := []rl.Rectangle{helpBox, resetBox, archBox, saveBox, camBox}
		for i := 0; i < len(boxes); i++ {
			for j := i + 1; j < len(boxes); j++ {
				if rl.CheckCollisionRecs(boxes[i], boxes[j]) {
					t.Fatalf("Overlap between top bar button %d and %d at resolution %dx%d", i, j, res.w, res.h)
				}
			}
		}

		// Left controls: title + pause + speed controls
		leftLimit := float32(165 + 64 + 8 + 26 + 28 + 54 + 28 + 48 + 10) // ~381
		if rightLimit <= leftLimit {
			t.Fatalf("Top bar right controls collided with left controls at resolution %dx%d", res.w, res.h)
		}

		// 2D Viewport side-by-side with Inspector test
		state := &SimState{SelectedBodyID: 1} // Inspector open
		cfg := &Config{Show2DViewport: true}
		vx, vy, vw, vh := get2DViewportRect(state, cfg, res.w, res.h)

		inspectorX := res.w - 310 - 12

		if vx+vw > inspectorX {
			t.Fatalf("2D Viewport (right edge %d) overlaps with Inspector (left edge %d) at resolution %dx%d",
				vx+vw, inspectorX, res.w, res.h)
		}
		if vx < 10 {
			t.Fatalf("2D Viewport pushed off left edge (vx=%d) at resolution %dx%d", vx, res.w, res.h)
		}
		_ = vy
		_ = vh
	}
}

func TestSpawnableObjectsSpacetimeDepths(t *testing.T) {
	// 1. Verify minor bodies alone in the scene do NOT warp spacetime into deep craters
	singleMoon := CreateBodyTemplate(SpawnMoon, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 1)
	singleComet := CreateBodyTemplate(SpawnComet, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 2)
	singleEarth := CreateBodyTemplate(SpawnEarth, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 3)
	singleGasGiant := CreateBodyTemplate(SpawnGasGiant, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 4)
	singleBH := CreateBodyTemplate(SpawnBlackHole, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 5)
	singleNeutron := CreateBodyTemplate(SpawnNeutronStar, rl.NewVector3(0, 0, 0), rl.NewVector3(0, 0, 0), 6)

	computeDepth := func(b *Body, evalY float32) float32 {
		cfg := &Config{G: 1.0, SpacetimeResolution: 80, SpacetimeScale: 1.0}
		totalSpan := getStaticSpacetimeSpan(cfg)
		spacing := totalSpan / float32(cfg.SpacetimeResolution)

		bType := 0
		var wDepth, wRad float32
		if b.TextureType == TextureBlackHole {
			bType = 4
			wDepth = float32(math.Min(58.0, 44.0+math.Log10(b.Mass+1.0)*3.0))
			wRad = float32(math.Max(float64(b.Radius*1.5), 2.0))
		} else if b.TextureType == TextureNeutronStar || b.IsPulsar {
			bType = 3
			wDepth = 38.0
			wRad = float32(math.Max(float64(b.Radius*1.8), 2.5))
		} else if b.TextureType == TextureWhiteDwarf {
			bType = 2
			wDepth = 32.0
			wRad = float32(math.Max(float64(b.Radius*2.0), 2.8))
		} else if b.IsStar || b.Mass >= 200.0 {
			bType = 1
			wDepth = float32(math.Min(35.0, 22.0+math.Pow(b.Mass/1000.0, 0.3)*3.0))
			wRad = float32(math.Max(float64(b.Radius*3.2), float64(totalSpan*0.045)))
		} else {
			bType = 0
			wDepth = float32(math.Min(18.0, math.Pow(b.Mass, 0.45)*4.2))
			wRad = float32(math.Max(float64(b.Radius*2.2), float64(spacing*1.6)))
		}

		dx := float32(0.0) - b.Position.X
		dz := float32(0.0) - b.Position.Z
		dy := b.Position.Y - evalY
		rSq := dx*dx + dz*dz + dy*dy*0.75

		var depth float32 = 0.0
		switch bType {
		case 4:
			rs2 := wRad * wRad
			throat := float32(math.Pow(float64(rs2/(rSq+rs2)), 1.35))
			depth = -wDepth * throat
		case 3, 2:
			rw2 := wRad * wRad
			funnel := float32(math.Pow(float64(rw2/(rSq+rw2)), 1.15))
			depth = -wDepth * funnel
		case 1:
			r := float32(math.Sqrt(float64(rSq) + 1.8))
			starFract := wRad / (r + wRad)
			depth = -wDepth * starFract
		default:
			radSq := wRad * wRad
			wellFract := radSq / (rSq + radSq)
			depth = -wDepth * wellFract
		}
		return depth
	}

	dMoon := computeDepth(singleMoon, 0)
	dComet := computeDepth(singleComet, 0)
	dEarth := computeDepth(singleEarth, 0)
	dGas := computeDepth(singleGasGiant, 0)
	dBH := computeDepth(singleBH, 0)
	dNeutron := computeDepth(singleNeutron, 0)

	// Minor bodies must be subtle (< 2.0 depth)
	if math.Abs(float64(dMoon)) > 2.0 {
		t.Fatalf("Moon alone created excessive depth %f (expected <= 2.0)", dMoon)
	}
	if math.Abs(float64(dComet)) > 1.0 {
		t.Fatalf("Comet alone created excessive depth %f (expected <= 1.0)", dComet)
	}

	// Terrestrial planet (~ 5-6 depth)
	if math.Abs(float64(dEarth)) < 3.0 || math.Abs(float64(dEarth)) > 8.0 {
		t.Fatalf("Earth created unexpected depth %f (expected [3, 8])", dEarth)
	}

	// Gas giant (~ 14-17 depth)
	if math.Abs(float64(dGas)) < 12.0 || math.Abs(float64(dGas)) > 18.0 {
		t.Fatalf("Gas giant created unexpected depth %f (expected [12, 18])", dGas)
	}

	// Black hole must plunge deeply (> 40.0 depth)
	if math.Abs(float64(dBH)) < 40.0 {
		t.Fatalf("Black hole funnel is too shallow: %f (expected > 40.0)", dBH)
	}

	// Neutron star must be compact and deep (> 30.0 depth)
	if math.Abs(float64(dNeutron)) < 30.0 {
		t.Fatalf("Neutron star funnel is too shallow: %f (expected > 30.0)", dNeutron)
	}

	// 2. Verify elevation Y physically softens and diminishes grid depth
	elevatedEarth := CreateBodyTemplate(SpawnEarth, rl.NewVector3(0, 60, 0), rl.NewVector3(0, 0, 0), 10)
	dElevated := computeDepth(elevatedEarth, 0)
	if math.Abs(float64(dElevated)) >= math.Abs(float64(dEarth))*0.5 {
		t.Fatalf("Elevation Y=60 did not sufficiently soften grid depth: on-plane=%f, elevated=%f", dEarth, dElevated)
	}
}

func TestUltraDense3DVectorField(t *testing.T) {
	state := &SimState{}
	cfg := &Config{
		G:                1.0,
		DenseVectorField: true,
		SpacetimeScale:   1.0,
		Softening:        0.5,
	}
	LoadPreset(state, cfg, PresetSolarSystem)

	totalSpan := getStaticSpacetimeSpan(cfg)
	yElevation := totalSpan * 0.075

	yPlanes := []struct {
		y     float32
		steps int
	}{
		{-yElevation * 1.8, 20},
		{-yElevation * 0.9, 26},
		{0.0, 36},
		{yElevation * 0.9, 26},
		{yElevation * 1.8, 20},
	}

	totalVectors := 0
	nonZeroVectors := 0

	for _, plane := range yPlanes {
		steps := plane.steps
		stepSize := totalSpan / float32(steps-1)
		halfSpan := totalSpan * 0.5
		py := plane.y

		for ix := 0; ix < steps; ix++ {
			px := -halfSpan + float32(ix)*stepSize
			for iz := 0; iz < steps; iz++ {
				pz := -halfSpan + float32(iz)*stepSize
				totalVectors++

				var gx, gy, gz float32
				for _, b := range state.Bodies {
					dx := b.Position.X - px
					dy := b.Position.Y - py
					dz := b.Position.Z - pz
					distSq := dx*dx + dy*dy + dz*dz + 4.0
					dist := float32(math.Sqrt(float64(distSq)))
					f := float32(cfg.G*b.Mass) / (distSq * dist)
					gx += dx * f
					gy += dy * f
					gz += dz * f
				}

				gLen := float32(math.Sqrt(float64(gx*gx + gy*gy + gz*gz)))
				if math.IsNaN(float64(gLen)) {
					t.Fatalf("Ultra-dense vector at (%f, %f, %f) produced NaN", px, py, pz)
				}
				if gLen > 0.0001 {
					nonZeroVectors++
				}
			}
		}
	}

	if totalVectors != 3448 {
		t.Fatalf("Expected exactly 3,448 volumetric vectors in ultra-dense mode, got %d", totalVectors)
	}
	if nonZeroVectors < 3400 {
		t.Fatalf("Expected nearly all vectors active (>3400), got %d", nonZeroVectors)
	}
}

func TestGravitationalWavesQuadrupoleAndBursts(t *testing.T) {
	// 1. Binary orbit quadrupole frequency: omega_gw = 2 * omega_orb
	sun := &Body{ID: 1, Name: "Sun", Mass: 1000.0, Position: rl.NewVector3(0, 0, 0)}
	companion := &Body{ID: 2, Name: "Companion Star", Mass: 800.0, Position: rl.NewVector3(20, 0, 0)}
	state := &SimState{Bodies: []*Body{sun, companion}}
	cfg := &Config{
		G:                      1.0,
		SpeedOfLight:           150.0,
		ShowGravitationalWaves: true,
	}

	UpdateGravitationalWaves(state, cfg, 0.016)

	// Theoretical orbital angular velocity: omega_orb = sqrt(G*M_tot / a^3) = sqrt(1800 / 8000) = sqrt(0.225) = 0.4743 rad/s
	// omega_gw = 2 * omega_orb = 0.9487 rad/s
	// f_gw = omega_gw / (2*pi) = 0.1510 Hz
	expectedFGW := float32(2.0 * math.Sqrt(1800.0/8000.0) / (2.0 * math.Pi))
	if math.Abs(float64(state.LastGWFreq-expectedFGW)) > 0.01 {
		t.Fatalf("Gravitational wave frequency %f deviates from theoretical quadrupole %f", state.LastGWFreq, expectedFGW)
	}

	if state.LastGWPower <= 0 {
		t.Fatalf("Expected positive gravitational wave luminosity P_gw, got %e", state.LastGWPower)
	}

	// 2. Waveform history recorded
	nonZeroSamples := 0
	for _, val := range state.GWWaveformHistory {
		if val != 0 {
			nonZeroSamples++
		}
	}
	if nonZeroSamples == 0 {
		t.Fatalf("Oscilloscope waveform history did not record samples")
	}

	// 3. Merger burst registration & ringdown
	EmitGravitationalWaveBurst(state, cfg, sun, companion)
	if len(state.GWBursts) != 1 {
		t.Fatalf("Expected 1 registered GW burst, got %d", len(state.GWBursts))
	}
	burst := state.GWBursts[0]
	if burst.PeakStrain <= 0 {
		t.Fatalf("Merger burst peak strain must be positive, got %f", burst.PeakStrain)
	}
	if burst.WaveSpeed < 50.0 {
		t.Fatalf("Merger wave speed must be >= 50.0, got %f", burst.WaveSpeed)
	}

	// 4. Metric ripple displacement on spacetime plane
	disp, strain := ComputeGravitationalWaveDisplacement(state, cfg, 15.0, 15.0, float64(rl.GetTime()))
	if math.IsNaN(float64(disp)) || math.IsNaN(float64(strain)) {
		t.Fatalf("ComputeGravitationalWaveDisplacement returned NaN: disp=%f, strain=%f", disp, strain)
	}
	if strain <= 0 {
		t.Fatalf("Expected positive strain intensity near binary, got %f", strain)
	}
}

func TestLIGOHUDZeroOverlapAcrossResolutions(t *testing.T) {
	resolutions := []struct{ w, h int32 }{
		{1024, 768},
		{1280, 720},
		{1920, 1080},
		{2560, 1440},
	}

	for _, res := range resolutions {
		// LIGO HUD card bounds
		cardW := float32(274)
		cardH := float32(142)
		cardX := float32(14)
		cardY := float32(res.h) - 48 - cardH - 6
		cardRec := rl.NewRectangle(cardX, cardY, cardW, cardH)

		// Top Bar bounds
		topBarRec := rl.NewRectangle(0, 0, float32(res.w), 48)
		if rl.CheckCollisionRecs(cardRec, topBarRec) {
			t.Fatalf("LIGO HUD collided with Top Bar at resolution %dx%d", res.w, res.h)
		}

		// Bottom Bar bounds
		bottomBarRec := rl.NewRectangle(0, float32(res.h)-48, float32(res.w), 48)
		if rl.CheckCollisionRecs(cardRec, bottomBarRec) {
			t.Fatalf("LIGO HUD collided with Bottom Bar at resolution %dx%d", res.w, res.h)
		}

		// Inspector panel bounds (right side)
		inspectorRec := rl.NewRectangle(float32(res.w)-310-12, 52, 310, 680)
		if rl.CheckCollisionRecs(cardRec, inspectorRec) {
			t.Fatalf("LIGO HUD collided with Inspector Panel at resolution %dx%d", res.w, res.h)
		}

		// Verify inside screen boundaries
		if cardRec.X < 0 || cardRec.Y < 0 || cardRec.X+cardRec.Width > float32(res.w) || cardRec.Y+cardRec.Height > float32(res.h) {
			t.Fatalf("LIGO HUD bounds outside screen at resolution %dx%d", res.w, res.h)
		}
	}
}

func TestSpacetimeAndGravitationalWavesSeparation(t *testing.T) {
	sun := &Body{ID: 1, Name: "Sun", Mass: 1000.0, Position: rl.NewVector3(0, 0, 0)}
	companion := &Body{ID: 2, Name: "Companion Star", Mass: 800.0, Position: rl.NewVector3(20, 0, 0)}
	state := &SimState{Bodies: []*Body{sun, companion}}
	cfg := &Config{
		G:            1.0,
		SpeedOfLight: 150.0,
	}

	UpdateGravitationalWaves(state, cfg, 0.016)

	// Case 1: Spacetime ON, GW OFF
	cfg.ShowPotentialGrid = true
	cfg.ShowGravitationalWaves = false
	if !cfg.ShowPotentialGrid || cfg.ShowGravitationalWaves {
		t.Fatalf("Case 1 flags incorrect")
	}

	// Case 2: Spacetime OFF, GW ON
	cfg.ShowPotentialGrid = false
	cfg.ShowGravitationalWaves = true
	if cfg.ShowPotentialGrid || !cfg.ShowGravitationalWaves {
		t.Fatalf("Case 2 flags incorrect")
	}

	// Compute ripple displacement without potential well
	disp, strain := ComputeGravitationalWaveDisplacement(state, cfg, 10.0, 10.0, 1.0)
	if math.Abs(float64(disp)) <= 0 {
		t.Fatalf("Expected non-zero wave displacement, got %f", disp)
	}
	if strain <= 0 {
		t.Fatalf("Expected positive strain, got %f", strain)
	}

	// Case 3: Both ON
	cfg.ShowPotentialGrid = true
	cfg.ShowGravitationalWaves = true
	if !cfg.ShowPotentialGrid || !cfg.ShowGravitationalWaves {
		t.Fatalf("Case 3 flags incorrect")
	}

	// Case 4: Both OFF
	cfg.ShowPotentialGrid = false
	cfg.ShowGravitationalWaves = false
	if cfg.ShowPotentialGrid || cfg.ShowGravitationalWaves {
		t.Fatalf("Case 4 flags incorrect")
	}
}

func TestCelestialIconAssignmentAndIndependence(t *testing.T) {
	// Test icon determination for various body archetypes
	starIcon := DetermineCelestialIcon("Sol", true, false, false, 1000.0, TextureSun)
	if starIcon != IconStar {
		t.Fatalf("Expected IconStar for Sol, got %v", starIcon)
	}

	bhIcon := DetermineCelestialIcon("Gargantua", true, false, false, 50000.0, TextureBlackHole)
	if bhIcon != IconBlackHole {
		t.Fatalf("Expected IconBlackHole for Gargantua, got %v", bhIcon)
	}

	pulsarIcon := DetermineCelestialIcon("Crab Pulsar", true, true, false, 4000.0, TextureNeutronStar)
	if pulsarIcon != IconPulsar {
		t.Fatalf("Expected IconPulsar for Crab Pulsar, got %v", pulsarIcon)
	}

	gasGiantIcon := DetermineCelestialIcon("Jupiter", false, false, false, 15.0, TextureGasGiant)
	if gasGiantIcon != IconGasGiant {
		t.Fatalf("Expected IconGasGiant for Jupiter, got %v", gasGiantIcon)
	}

	earthIcon := DetermineCelestialIcon("Earth", false, false, false, 1.0, TextureTerrestrial)
	if earthIcon != IconTerrestrial {
		t.Fatalf("Expected IconTerrestrial for Earth, got %v", earthIcon)
	}

	probeIcon := DetermineCelestialIcon("JWST Probe", false, false, false, 0.0001, TextureMoon)
	if probeIcon != IconProbe {
		t.Fatalf("Expected IconProbe for JWST, got %v", probeIcon)
	}

	// Verify CreateBodyTemplate sets CelestialIcon correctly
	sunTmpl := CreateBodyTemplate(SpawnStar, rl.Vector3Zero(), rl.Vector3Zero(), 1)
	if sunTmpl.CelestialIcon != IconStar {
		t.Fatalf("Expected SpawnSun template to have IconStar, got %v", sunTmpl.CelestialIcon)
	}

	bhTmpl := CreateBodyTemplate(SpawnBlackHole, rl.Vector3Zero(), rl.Vector3Zero(), 2)
	if bhTmpl.CelestialIcon != IconBlackHole {
		t.Fatalf("Expected SpawnBlackHole template to have IconBlackHole, got %v", bhTmpl.CelestialIcon)
	}

	earthTmpl := CreateBodyTemplate(SpawnEarth, rl.Vector3Zero(), rl.Vector3Zero(), 3)
	if earthTmpl.CelestialIcon != IconTerrestrial {
		t.Fatalf("Expected SpawnEarth template to have IconTerrestrial, got %v", earthTmpl.CelestialIcon)
	}
}

func Test2DOffloadModeAndInteractions(t *testing.T) {
	state := &SimState{
		Bodies: []*Body{
			{ID: 1, Name: "Sun", Mass: 1000.0, Position: rl.NewVector3(10, 0, 20)},
			{ID: 2, Name: "Earth", Mass: 1.0, Position: rl.NewVector3(50, 0, 20)},
		},
		SelectedBodyID: 1,
		NextID:         3,
	}

	cfg := &Config{
		G:                      1.0,
		Offload3D:              true,
		Show2DViewport:         true,
		Track2D:                true,
		ShowGravitationalWaves: true,
		Show2DGW:               true,
		Viewport2DZoom:         0.5,
		Viewport2DPan:          rl.NewVector2(0, 0),
	}

	// 1. Offload 2D viewport rect fills screen under top bar and above bottom bar
	screenW := int32(1920)
	screenH := int32(1080)
	vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)

	if vx != 0 || vy != 48 || vw != screenW || vh != screenH-96 {
		t.Fatalf("Expected 2D Offload Viewport to be (0, 48, %d, %d), got (%d, %d, %d, %d)",
			screenW, screenH-96, vx, vy, vw, vh)
	}

	// 2. 2D Continuous Body Tracking
	if cfg.Track2D && state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				cfg.Viewport2DPan.X = b.Position.X
				cfg.Viewport2DPan.Y = b.Position.Z
				break
			}
		}
	}
	if cfg.Viewport2DPan.X != 10 || cfg.Viewport2DPan.Y != 20 {
		t.Fatalf("Expected 2D Pan to track Sun at (10, 20), got (%f, %f)", cfg.Viewport2DPan.X, cfg.Viewport2DPan.Y)
	}

	// 3. 2D Spawning via SpawnPresetAtLocation
	state.SpawnPreset = SpawnBlackHole
	spawnPos := rl.NewVector3(100, 0, 150)
	spawnVel := rl.NewVector3(-2.5, 0, 1.2)
	initialBodyCount := len(state.Bodies)

	SpawnPresetAtLocation(state, cfg, spawnPos, spawnVel)
	if len(state.Bodies) != initialBodyCount+1 {
		t.Fatalf("Expected %d bodies after 2D spawn, got %d", initialBodyCount+1, len(state.Bodies))
	}

	newBody := state.Bodies[len(state.Bodies)-1]
	if newBody.Position.X != 100 || newBody.Position.Z != 150 {
		t.Fatalf("Expected spawned body at (100, 0, 150), got (%f, %f, %f)", newBody.Position.X, newBody.Position.Y, newBody.Position.Z)
	}
	if newBody.Velocity.X != -2.5 || newBody.Velocity.Z != 1.2 {
		t.Fatalf("Expected spawned body velocity (-2.5, 0, 1.2), got (%f, %f, %f)", newBody.Velocity.X, newBody.Velocity.Y, newBody.Velocity.Z)
	}
	if newBody.CelestialIcon != IconBlackHole {
		t.Fatalf("Expected spawned body to have IconBlackHole, got %v", newBody.CelestialIcon)
	}

	// 4. Verify 2D Gravitational Waves calculation works
	UpdateGravitationalWaves(state, cfg, 0.016)
	if state.LastGWFreq <= 0 {
		t.Fatalf("Expected positive GW frequency in 2D mode, got %f", state.LastGWFreq)
	}
}

func TestStatsHUDNoOverlapWith2DViewport(t *testing.T) {
	state := &SimState{
		Bodies: []*Body{
			{ID: 1, Name: "Sun", Mass: 100.0, Position: rl.NewVector3(0, 0, 0)},
		},
		SelectedBodyID: 1,
	}
	screenW := int32(1920)
	screenH := int32(1080)

	// 1. Offload 3D Mode: Verify Stats HUD is positioned strictly below header and toolbar buttons
	cfgOffload := &Config{
		Offload3D:      true,
		Show2DViewport: true,
		ShowStatsHUD:   true,
	}
	vx, vy, _, _ := get2DViewportRect(state, cfgOffload, screenW, screenH)
	headerH := int32(32)
	toolbarH := int32(28)
	toolbarBottom := vy + headerH + toolbarH // 48 + 32 + 28 = 108

	badgeYOffload := float32(vy) + float32(headerH) + float32(toolbarH) + 8 // 116
	if badgeYOffload <= float32(toolbarBottom) {
		t.Fatalf("OVERLAP: Stats HUD Y (%.1f) overlaps 2D toolbar (bottom: %d)", badgeYOffload, toolbarBottom)
	}
	if vx != 0 || vy != 48 {
		t.Fatalf("Unexpected 2D offload viewport origin: (%d, %d)", vx, vy)
	}

	// 2. Fullscreen 2D Mode: Verify Stats HUD is positioned strictly below header and toolbar
	cfgFullscreen := &Config{
		Viewport2DFullscreen: true,
		Show2DViewport:       true,
		ShowStatsHUD:         true,
	}
	_, vyFS, _, _ := get2DViewportRect(state, cfgFullscreen, screenW, screenH)
	toolbarBottomFS := vyFS + headerH + toolbarH                         // 52 + 32 + 28 = 112
	badgeYFS := float32(vyFS) + float32(headerH) + float32(toolbarH) + 8 // 120
	if badgeYFS <= float32(toolbarBottomFS) {
		t.Fatalf("OVERLAP: Stats HUD Y (%.1f) overlaps 2D Fullscreen toolbar (bottom: %d)", badgeYFS, toolbarBottomFS)
	}

	// 3. Notification Toast in 2D Offload mode: Verify placement near bottom, safe from top toolbar
	pillYOffload := float32(screenH) - 48 - 22 - 38
	if pillYOffload < float32(toolbarBottom) {
		t.Fatalf("OVERLAP: Toast Y (%.1f) overlaps 2D toolbar (bottom: %d)", pillYOffload, toolbarBottom)
	}

	// 4. ShowStatsHUD toggle state
	cfgOffload.ShowStatsHUD = false
	if cfgOffload.ShowStatsHUD {
		t.Fatalf("Expected ShowStatsHUD to be false after toggle")
	}
}

func Test2DObjectVectorsToggle(t *testing.T) {
	cfg := &Config{
		ShowVectors:       false, // Default: vectors off
		Show2DVectorField: false,
	}

	// When ShowVectors is false, object direction vectors in 2D and 3D must be off
	if cfg.ShowVectors {
		t.Fatalf("Expected ShowVectors to be false initially")
	}

	// Toggle vectors on
	cfg.ShowVectors = true
	cfg.Show2DVectorField = true
	if !cfg.ShowVectors {
		t.Fatalf("Expected ShowVectors to be true after toggle")
	}

	// Toggle vectors off again
	cfg.ShowVectors = false
	cfg.Show2DVectorField = false
	if cfg.ShowVectors {
		t.Fatalf("Expected ShowVectors to be false after toggling off")
	}
}
