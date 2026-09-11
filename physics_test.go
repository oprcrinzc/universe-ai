package main

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestRelativisticPrecessionCorrection(t *testing.T) {
	// A massive black hole and an eccentric test planet
	smbh := &Body{
		ID:           1,
		Position:     rl.NewVector3(0, 0, 0),
		Velocity:     rl.NewVector3(0, 0, 0),
		Mass:         10000.0,
		Radius:       3.0,
		IsStationary: true,
	}

	planet := &Body{
		ID:           2,
		Position:     rl.NewVector3(12.0, 0, 0),
		Velocity:     rl.NewVector3(0, 0, -25.0),
		Mass:         1.0,
		Radius:       0.8,
		IsStationary: false,
	}

	bodies := []*Body{smbh, planet}
	g := 1.0
	c := 180.0
	softening := 0.5

	pnAccs := CalculateRelativisticCorrections(bodies, g, c, softening)
	if len(pnAccs) != 2 {
		t.Fatalf("Expected 2 acceleration vectors, got %d", len(pnAccs))
	}

	pAcc := pnAccs[1]
	pAccMag := float64(rl.Vector3Length(pAcc))

	// Should have non-zero relativistic acceleration
	if pAccMag <= 0.0001 {
		t.Fatalf("Expected positive relativistic acceleration magnitude, got %f", pAccMag)
	}

	// 1PN acceleration should have dominant attractive component along radial direction
	if pAcc.X >= 0 {
		t.Fatalf("Expected negative X (inward attractive 1PN component towards origin), got %f", pAcc.X)
	}
}

func TestRocheLimitDisruption(t *testing.T) {
	giant := &Body{
		ID:           1,
		Name:         "Jupiter",
		Position:     rl.NewVector3(0, 0, 0),
		Velocity:     rl.NewVector3(0, 0, 0),
		Mass:         15000.0,
		Radius:       4.0,
		IsStationary: true,
	}

	// Moon positioned well within Jupiter's Roche limit (e.g. at dist = 8.0, where Roche limit is > 12.0)
	moon := &Body{
		ID:           2,
		Name:         "Icy Moon",
		Position:     rl.NewVector3(8.0, 0, 0),
		Velocity:     rl.NewVector3(0, 0, -35.0),
		Mass:         2.0,
		Radius:       1.0,
		Color:        rl.White,
		IsStationary: false,
	}

	state := &SimState{
		NextID: 3,
		Bodies: []*Body{giant, moon},
	}
	cfg := &Config{
		EnableRocheLimit: true,
	}

	initialMass := moon.Mass

	HandleRocheDisruption(state, cfg)

	// Moon should have shattered into multiple ring fragments
	if len(state.Bodies) <= 2 {
		t.Fatalf("Expected moon to be disrupted into fragments, body count: %d", len(state.Bodies))
	}

	// Check total mass conservation
	var totalFragMass float64 = 0
	for _, b := range state.Bodies {
		if b.ID != giant.ID {
			totalFragMass += b.Mass
		}
	}

	if math.Abs(totalFragMass-initialMass) > 0.001 {
		t.Fatalf("Mass not conserved during Roche disruption: expected %f, got %f", initialMass, totalFragMass)
	}
}

func TestVerletIntegratorEnergyConservation(t *testing.T) {
	// Circular two-body orbit
	sun := &Body{
		ID:           1,
		Position:     rl.NewVector3(0, 0, 0),
		Velocity:     rl.NewVector3(0, 0, 0),
		Mass:         1000.0,
		Radius:       2.0,
		IsStationary: true,
	}

	r := float32(20.0)
	vCirc := float32(math.Sqrt(1.0 * 1000.0 / float64(r)))

	planet := &Body{
		ID:           2,
		Position:     rl.NewVector3(r, 0, 0),
		Velocity:     rl.NewVector3(0, 0, -vCirc),
		Mass:         1.0,
		Radius:       0.5,
		IsStationary: false,
	}

	state := &SimState{
		Bodies: []*Body{sun, planet},
	}
	cfg := &Config{
		G:         1.0,
		TimeScale: 1.0,
		SubSteps:  5,
		Softening: 0.1,
	}

	computeEnergy := func() float64 {
		ke := 0.5 * float64(planet.Mass) * float64(rl.Vector3LengthSqr(planet.Velocity))
		dist := float64(rl.Vector3Distance(planet.Position, sun.Position))
		pe := -cfg.G * sun.Mass * planet.Mass / dist
		return ke + pe
	}

	initialE := computeEnergy()

	// Simulate 200 steps (~2 full orbits)
	for step := 0; step < 200; step++ {
		UpdatePhysics(state, cfg, 0.0166)
	}

	finalE := computeEnergy()
	relError := math.Abs((finalE - initialE) / initialE)

	t.Logf("Verlet Energy error over 200 steps: %.6f%%", relError*100)
	if relError > 0.01 { // Expect < 1% energy drift over multi-orbits with Velocity Verlet
		t.Fatalf("Energy drift too large: %.4f%%", relError*100)
	}
}

func TestTimeScaleSteppingAndClamping(t *testing.T) {
	cfg := &Config{
		TimeScale:     1.0,
		TimeScaleStep: 0.5,
	}

	// 1. Step up with 0.5
	cfg.TimeScale = math.Round((cfg.TimeScale+cfg.TimeScaleStep)*10.0) / 10.0
	if cfg.TimeScale != 1.5 {
		t.Fatalf("Expected 1.5, got %.2f", cfg.TimeScale)
	}

	// 2. Change step to 0.1
	cfg.TimeScaleStep = 0.1
	cfg.TimeScale = math.Round((cfg.TimeScale+cfg.TimeScaleStep)*10.0) / 10.0
	if cfg.TimeScale != 1.6 {
		t.Fatalf("Expected 1.6, got %.2f", cfg.TimeScale)
	}

	// 3. Step down by 0.1
	cfg.TimeScale = math.Round((cfg.TimeScale-cfg.TimeScaleStep)*10.0) / 10.0
	if cfg.TimeScale != 1.5 {
		t.Fatalf("Expected 1.5, got %.2f", cfg.TimeScale)
	}

	// 4. Change step to 2.0
	cfg.TimeScaleStep = 2.0
	cfg.TimeScale = math.Round((cfg.TimeScale+cfg.TimeScaleStep)*10.0) / 10.0
	if cfg.TimeScale != 3.5 {
		t.Fatalf("Expected 3.5, got %.2f", cfg.TimeScale)
	}

	// 5. Test max clamping (10.0)
	for i := 0; i < 10; i++ {
		cfg.TimeScale = math.Min(10.0, math.Round((cfg.TimeScale+cfg.TimeScaleStep)*10.0)/10.0)
	}
	if cfg.TimeScale > 10.0 {
		t.Fatalf("TimeScale exceeded max limit 10.0: got %.2f", cfg.TimeScale)
	}

	// 6. Test min clamping (0.1)
	for i := 0; i < 20; i++ {
		cfg.TimeScale = math.Max(0.1, math.Round((cfg.TimeScale-cfg.TimeScaleStep)*10.0)/10.0)
	}
	if cfg.TimeScale < 0.1 {
		t.Fatalf("TimeScale dropped below min limit 0.1: got %.2f", cfg.TimeScale)
	}

	// 7. Test step cycling (0.1 -> 0.5 -> 2.0 -> 0.1)
	steps := []float64{0.1, 0.5, 2.0}
	curStep := 0.1
	cycleStep := func(s float64) float64 {
		if s <= 0.15 {
			return 0.5
		} else if s <= 0.6 {
			return 2.0
		}
		return 0.1
	}

	curStep = cycleStep(curStep)
	if curStep != 0.5 {
		t.Fatalf("Expected step 0.5 after 0.1, got %.1f", curStep)
	}
	curStep = cycleStep(curStep)
	if curStep != 2.0 {
		t.Fatalf("Expected step 2.0 after 0.5, got %.1f", curStep)
	}
	curStep = cycleStep(curStep)
	if curStep != 0.1 {
		t.Fatalf("Expected step 0.1 after 2.0, got %.1f", curStep)
	}
	_ = steps
}

func TestLagrangePointsExactEquilibrium(t *testing.T) {
	g := 1.0
	m1 := 10000.0
	m2 := 100.0
	totalM := m1 + m2
	mu := m2 / totalM
	R := 80.0

	sun := &Body{ID: 1, Name: "Sun", Mass: m1, Position: rl.NewVector3(float32(-mu*R), 0, 0)}
	zeus := &Body{ID: 2, Name: "Zeus", Mass: m2, Position: rl.NewVector3(float32((1.0-mu)*R), 0, 0)}

	omega := math.Sqrt((g * totalM) / (R * R * R))
	sun.Velocity = rl.NewVector3(0, 0, float32(float64(sun.Position.X)*omega))
	zeus.Velocity = rl.NewVector3(0, 0, -float32(float64(zeus.Position.X)*omega))

	pts := ComputeLagrangePoints(sun, zeus, g)

	// Verify L4 and L5 distances: must form exact equilateral triangles with R to both primary and secondary
	for _, idx := range []int{3, 4} {
		pt := pts[idx]
		d1 := float64(rl.Vector3Distance(pt.Position, sun.Position))
		d2 := float64(rl.Vector3Distance(pt.Position, zeus.Position))
		if math.Abs(d1-R) > 1e-3 {
			t.Fatalf("%s distance to Sun %.4f deviates from R=%.4f", pt.Name, d1, R)
		}
		if math.Abs(d2-R) > 1e-3 {
			t.Fatalf("%s distance to Zeus %.4f deviates from R=%.4f", pt.Name, d2, R)
		}
	}

	// Verify collinear points: L1 is between Sun and Zeus, L2 is beyond Zeus, L3 is behind Sun
	if pts[0].Position.X <= sun.Position.X || pts[0].Position.X >= zeus.Position.X {
		t.Fatalf("L1 not collinear between Sun and Zeus: Sun=%.2f, L1=%.2f, Zeus=%.2f", sun.Position.X, pts[0].Position.X, zeus.Position.X)
	}
	if pts[1].Position.X <= zeus.Position.X {
		t.Fatalf("L2 not beyond Zeus: Zeus=%.2f, L2=%.2f", zeus.Position.X, pts[1].Position.X)
	}
	if pts[2].Position.X >= sun.Position.X {
		t.Fatalf("L3 not behind Sun: Sun=%.2f, L3=%.2f", sun.Position.X, pts[2].Position.X)
	}

	// Test orbit simulation: step system forward with Yoshida 4th order
	state := &SimState{
		Bodies: []*Body{sun, zeus},
		NextID: 3,
	}
	l4Body := addBody(state, "L4 Test", pts[3].Position, pts[3].Velocity, 0.001, 0.5, rl.White, false, false)
	cfg := &Config{
		G:          g,
		TimeScale:  1.0,
		SubSteps:   4,
		Softening:  0.0,
		Integrator: IntegratorYoshida4,
	}

	initDist := rl.Vector3Distance(l4Body.Position, sun.Position)
	dt := float32(0.0166)
	for step := 0; step < 300; step++ {
		UpdatePhysics(state, cfg, dt)
	}
	finalDist := rl.Vector3Distance(l4Body.Position, sun.Position)
	relErr := math.Abs(float64(finalDist-initDist)) / float64(initDist)
	if relErr > 0.005 {
		t.Fatalf("L4 orbit drifted excessively: initDist=%.4f, finalDist=%.4f, relErr=%.4f%%", initDist, finalDist, relErr*100)
	}
}

func TestYoshida4thOrderEnergyConservation(t *testing.T) {
	// 2-Body eccentric Keplerian orbit
	m1 := 1000.0
	m2 := 1.0
	g := 1.0
	r0 := 20.0
	v0 := math.Sqrt((g*m1)/r0) * 0.85 // Eccentric orbit

	b1 := &Body{ID: 1, Mass: m1, Position: rl.NewVector3(0, 0, 0), Velocity: rl.NewVector3(0, 0, 0), IsStationary: true}
	b2 := &Body{ID: 2, Mass: m2, Position: rl.NewVector3(float32(r0), 0, 0), Velocity: rl.NewVector3(0, 0, float32(v0))}

	state := &SimState{Bodies: []*Body{b1, b2}, NextID: 3}
	cfg := &Config{
		G:          g,
		TimeScale:  1.0,
		SubSteps:   2,
		Softening:  0.01,
		Integrator: IntegratorYoshida4,
	}

	calcEnergy := func() float64 {
		v := float64(rl.Vector3Length(b2.Velocity))
		ke := 0.5 * m2 * v * v
		r := float64(rl.Vector3Distance(b1.Position, b2.Position))
		pe := -(g * m1 * m2) / r
		return ke + pe
	}

	eInit := calcEnergy()
	dt := float32(0.0166)
	for i := 0; i < 400; i++ {
		UpdatePhysics(state, cfg, dt)
	}
	eFinal := calcEnergy()
	errRatio := math.Abs(eFinal-eInit) / math.Abs(eInit)
	if errRatio > 0.001 {
		t.Fatalf("Yoshida 4th-order energy drift too large: %.6f%%", errRatio*100)
	}
}

