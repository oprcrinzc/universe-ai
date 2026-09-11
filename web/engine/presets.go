package main

import (
	"fmt"
	"math"
	"math/rand"
)

func LoadPreset(state *SimState, cfg *Config, preset PresetType) {
	state.Bodies = make([]*Body, 0)
	state.NextID = 1
	state.Time = 0
	state.CollisionCount = 0
	state.MergeCount = 0
	state.TidalBreakupCount = 0
	state.SelectedBodyID = -1

	g := cfg.G

	switch preset {
	case PresetSolarSystem:
		loadSolarSystem(state, g)
	case PresetSolarSystemGrand:
		loadSolarSystemGrand(state, g)
	case PresetBinaryStars:
		loadBinaryStars(state, g)
	case PresetThreeBody:
		loadThreeBody(state, g)
	case PresetLagrangeTrojans:
		loadLagrangeTrojans(state, cfg, g)
	case PresetGalaxyCollision:
		loadGalaxyCollision(state, cfg, g)
	case PresetPulsarAccretion:
		loadPulsarAccretion(state, cfg, g)
	case PresetMilkyWay:
		loadMilkyWay(state, cfg, g)
	case PresetGlobularCluster:
		loadGlobularCluster(state, cfg, g)
	case PresetTrappist1:
		loadTrappist1(state, cfg, g)
	case PresetFiveBodyChoreography:
		loadFiveBodyChoreography(state, g)
	case PresetRelativisticRosette:
		loadRelativisticRosette(state, cfg, g)
	case PresetRocheDisruption:
		loadRocheDisruption(state, cfg, g)
	case PresetEmpty:
		// Clean canvas
	case PresetMilkyWay10K:
		loadMilkyWay10K(state, cfg, g)
	case PresetAsteroidBelt2K:
		loadAsteroidBelt2K(state, cfg, g)
	case PresetGalaxyCollision5K:
		loadGalaxyCollision5K(state, cfg, g)
	case PresetBlackHoleSwarm3K:
		loadBlackHoleSwarm3K(state, cfg, g)
	case PresetGravitationalCollapse:
		loadGravitationalCollapse(state, cfg, g)
	case PresetRealSolarSystem:
		loadRealScaleSolarSystem(state, cfg, g)
	default:
		loadSolarSystem(state, g)
	}
}

func addBody(state *SimState, name string, pos, vel Vector3, mass float64, radius float32, col Color, isStationary, isStar bool, texType int) *Body {
	b := &Body{
		ID:           state.NextID,
		Name:         name,
		Position:     pos,
		Velocity:     vel,
		Mass:         mass,
		Radius:       radius,
		Color:        col,
		IsStationary: isStationary,
		IsStar:       isStar,
		TextureType:  texType,
		Trail:        make([]Vector3, 0, MaxTrailPoints),
	}
	state.NextID++
	state.Bodies = append(state.Bodies, b)
	return b
}

func loadSolarSystem(state *SimState, g float64) {
	sunM := 1000.0
	addBody(state, "Sun", Vector3{0, 0, 0}, Vector3{0, 0, 0}, sunM, 4.0, Color{255, 220, 60, 255}, true, true, 6)

	planets := []struct {
		name   string
		dist   float32
		mass   float64
		radius float32
		color  Color
		tex    int
	}{
		{"Mercury", 9.0, 0.08, 0.45, Color{190, 185, 180, 255}, 5},
		{"Venus", 14.0, 0.65, 0.75, Color{235, 195, 130, 255}, 2},
		{"Earth", 20.0, 8.0, 0.95, Color{90, 160, 255, 255}, 1},
		{"Mars", 27.0, 0.15, 0.55, Color{230, 100, 60, 255}, 2},
		{"Jupiter", 40.0, 18.0, 2.2, Color{230, 190, 140, 255}, 3},
		{"Saturn", 56.0, 10.0, 1.8, Color{235, 215, 165, 255}, 3},
		{"Uranus", 72.0, 4.0, 1.3, Color{165, 225, 240, 255}, 4},
		{"Neptune", 88.0, 4.2, 1.3, Color{80, 120, 245, 255}, 4},
	}

	for _, p := range planets {
		vOrb := float32(math.Sqrt((g * sunM) / float64(p.dist)))
		addBody(state, p.name,
			Vector3{X: p.dist, Y: 0, Z: 0},
			Vector3{X: 0, Y: 0, Z: vOrb},
			p.mass, p.radius, p.color, false, false, p.tex)
	}

	// Earth Moon (stable inside Hill sphere)
	earthDist := float32(20.0)
	vEarth := float32(math.Sqrt((g * sunM) / float64(earthDist)))
	moonDist := float32(1.4)
	vMoonRel := float32(math.Sqrt((g * 8.0) / float64(moonDist)))
	addBody(state, "Moon",
		Vector3{X: earthDist + moonDist, Y: 0, Z: 0},
		Vector3{X: 0, Y: 0, Z: vEarth + vMoonRel},
		0.012, 0.28, Color{210, 210, 215, 255}, false, false, 5)
}

func loadSolarSystemGrand(state *SimState, g float64) {
	loadSolarSystem(state, g)

	// Add Asteroid Belt (250 bodies)
	sunM := 1000.0
	rSource := rand.New(rand.NewSource(42))
	for i := 0; i < 250; i++ {
		r := 31.0 + rSource.Float64()*7.0
		theta := rSource.Float64() * 2.0 * math.Pi
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rSource.Float64() - 0.5) * 1.2)

		vCirc := float32(math.Sqrt((g * sunM) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))

		addBody(state, fmt.Sprintf("Asteroid #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, 0, vz},
			0.0005, 0.2, Color{170, 165, 160, 255}, false, false, 5)
	}

	// Comet Halley
	addBody(state, "Halley's Comet",
		Vector3{X: 110.0, Y: 12.0, Z: 40.0},
		Vector3{X: -1.8, Y: -0.3, Z: 0.6},
		0.0001, 0.25, Color{180, 240, 255, 255}, false, false, 5)
	state.Bodies[len(state.Bodies)-1].IsComet = true
}

func loadBinaryStars(state *SimState, g float64) {
	starMass := 650.0
	dist := float32(16.0)
	vOrb := float32(math.Sqrt((g * starMass) / (4.0 * float64(dist))))

	addBody(state, "Alpha Centauri A", Vector3{-dist, 0, 0}, Vector3{0, 0, -vOrb}, starMass, 3.2, Color{255, 215, 80, 255}, false, true, 6)
	addBody(state, "Alpha Centauri B", Vector3{dist, 0, 0}, Vector3{0, 0, vOrb}, starMass, 2.9, Color{255, 140, 60, 255}, false, true, 6)

	// Circumbinary planets orbiting around the binary pair (safely outside chaotic resonance boundary r >= 74)
	planets := []struct {
		name string
		dist float32
		mass float64
		rad  float32
		col  Color
	}{
		{"Tatooine I", 75.0, 1.2, 0.75, Color{240, 200, 120, 255}},
		{"Tatooine II", 105.0, 1.8, 0.90, Color{120, 220, 255, 255}},
		{"Arrakis", 140.0, 0.9, 0.70, Color{235, 150, 75, 255}},
		{"Oricon", 185.0, 3.5, 1.30, Color{185, 130, 245, 255}},
	}

	totalM := starMass * 2.0
	for _, p := range planets {
		vp := float32(math.Sqrt((g * totalM) / float64(p.dist)))
		addBody(state, p.name, Vector3{p.dist, 0, 0}, Vector3{0, 0, vp}, p.mass, p.rad, p.col, false, false, 1)
	}
}

func loadThreeBody(state *SimState, g float64) {
	// Stable Lagrange equilateral 3-body rotating choreography
	mass := 700.0
	r := float32(24.0)
	vMag := float32(math.Sqrt((g * mass * (1.0 / math.Sqrt(3.0))) / float64(r)))

	for i := 0; i < 3; i++ {
		angle := float64(i) * (2.0 * math.Pi / 3.0)
		x := r * float32(math.Cos(angle))
		z := r * float32(math.Sin(angle))
		vx := -vMag * float32(math.Sin(angle))
		vz := vMag * float32(math.Cos(angle))
		cols := []Color{
			{255, 100, 100, 255},
			{100, 255, 120, 255},
			{100, 180, 255, 255},
		}
		addBody(state, fmt.Sprintf("Trisolaris %c", 'A'+i), Vector3{x, 0, z}, Vector3{vx, 0, vz}, mass, 3.0, cols[i], false, true, 6)
	}
}

func loadLagrangeTrojans(state *SimState, cfg *Config, g float64) {
	sunM := 1200.0
	jupM := 35.0
	jupDist := float32(45.0)
	vJup := float32(math.Sqrt((g * sunM) / float64(jupDist)))

	addBody(state, "Sun", Vector3{0, 0, 0}, Vector3{0, 0, 0}, sunM, 4.0, Color{255, 220, 60, 255}, true, true, 6)
	addBody(state, "Jupiter", Vector3{jupDist, 0, 0}, Vector3{0, 0, vJup}, jupM, 2.5, Color{230, 190, 140, 255}, false, false, 3)

	rSource := rand.New(rand.NewSource(1234))
	// L4 (+60 deg ahead) and L5 (-60 deg behind) swarms
	for sIdx, baseAngle := range []float64{math.Pi / 3.0, -math.Pi / 3.0} {
		swarmName := "Trojan (L4)"
		if sIdx == 1 {
			swarmName = "Greek (L5)"
		}
		for i := 0; i < 40; i++ {
			ang := baseAngle + (rSource.Float64()-0.5)*0.16
			r := float64(jupDist) + (rSource.Float64()-0.5)*2.2
			x := float32(r * math.Cos(ang))
			z := float32(r * math.Sin(ang))
			v := float32(math.Sqrt((g * sunM) / r))
			vx := -v * float32(math.Sin(ang))
			vz := v * float32(math.Cos(ang))

			addBody(state, fmt.Sprintf("%s #%d", swarmName, i+1),
				Vector3{x, float32((rSource.Float64() - 0.5) * 1.0), z},
				Vector3{vx, 0, vz},
				0.001, 0.22, Color{190, 190, 195, 255}, false, false, 5)
		}
	}
}

func loadGalaxyCollision(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	g1Pos := Vector3{-45, 0, -25}
	g1Vel := Vector3{1.4, 0, 0.7}
	g2Pos := Vector3{45, 10, 25}
	g2Vel := Vector3{-1.4, -0.3, -0.7}

	addGalaxyDisk(state, g1Pos, g1Vel, 1800.0, 120, 32.0, g, Color{140, 200, 255, 255})
	addGalaxyDisk(state, g2Pos, g2Vel, 1400.0, 100, 26.0, g, Color{255, 190, 140, 255})
}

func addGalaxyDisk(state *SimState, center, vel Vector3, coreMass float64, starCount int, radius float32, g float64, starCol Color) {
	addBody(state, "Galactic Core", center, vel, coreMass, 3.2, Color{255, 245, 230, 255}, false, true, 6)
	rSource := rand.New(rand.NewSource(int64(coreMass)))

	for i := 0; i < starCount; i++ {
		r := 4.0 + math.Pow(rSource.Float64(), 1.4)*float64(radius)
		theta := rSource.Float64() * 2.0 * math.Pi
		x := center.X + float32(r*math.Cos(theta))
		z := center.Z + float32(r*math.Sin(theta))
		y := center.Y + float32((rSource.Float64()-0.5)*1.8)

		vCirc := float32(math.Sqrt((g * coreMass) / r))
		vx := vel.X - vCirc*float32(math.Sin(theta))
		vz := vel.Z + vCirc*float32(math.Cos(theta))
		vy := vel.Y

		addBody(state, fmt.Sprintf("Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			0.2, 0.35, starCol, false, true, 6)
	}
}

func loadPulsarAccretion(state *SimState, cfg *Config, g float64) {
	pulsarM := 1800.0
	pulsarPos := Vector3{0, 0, 0}
	p := addBody(state, "PSR B1919+21", pulsarPos, Vector3{0, 0, 0}, pulsarM, 1.8, Color{160, 230, 255, 255}, true, false, 7)
	p.IsPulsar = true

	// Red Supergiant companion
	donorDist := float32(38.0)
	vDonor := float32(math.Sqrt((g * pulsarM) / float64(donorDist)))
	addBody(state, "Betelgeuse Secondary",
		Vector3{donorDist, 0, 0}, Vector3{0, 0, vDonor},
		350.0, 5.2, Color{255, 80, 40, 255}, false, true, 6)

	// Accretion stream
	rSource := rand.New(rand.NewSource(999))
	for i := 0; i < 90; i++ {
		r := 4.0 + rSource.Float64()*26.0
		theta := rSource.Float64() * 2.0 * math.Pi
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		v := float32(math.Sqrt((g * pulsarM) / r))
		vx := -v * float32(math.Sin(theta))
		vz := v * float32(math.Cos(theta))

		addBody(state, fmt.Sprintf("Plasma #%d", i+1),
			Vector3{x, float32((rSource.Float64() - 0.5) * 0.8), z},
			Vector3{vx, 0, vz},
			0.005, 0.22, Color{200, 120, 255, 255}, false, false, 5)
	}
}

func loadMilkyWay(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	smbhMass := 15000.0
	bh := addBody(state, "Sagittarius A*", Vector3{0, 0, 0}, Vector3{0, 0, 0}, smbhMass, 4.2, Color{10, 8, 18, 255}, true, false, 7)
	bh.IsBlackHole = true

	rSource := rand.New(rand.NewSource(777))
	const numStars = 650
	for i := 0; i < numStars; i++ {
		r := 8.0 + math.Pow(rSource.Float64(), 1.3)*140.0
		armOffset := float64(i%4) * (math.Pi / 2.0)
		theta := armOffset + 2.8*math.Log(r/8.0) + (rSource.Float64()-0.5)*0.5
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rSource.Float64() - 0.5) * (1.5 + r*0.02))

		vCirc := float32(math.Sqrt((g * smbhMass) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))

		addBody(state, fmt.Sprintf("Milky Way Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, 0, vz},
			0.3, 0.32, Color{220, 235, 255, 255}, false, true, 6)
	}
}

func loadGlobularCluster(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	rSource := rand.New(rand.NewSource(555))
	const n = 180
	for i := 0; i < n; i++ {
		r := math.Pow(rSource.Float64(), 2.0) * 35.0
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := math.Acos(2.0*rSource.Float64() - 1.0)
		x := float32(r * math.Sin(phi) * math.Cos(theta))
		y := float32(r * math.Sin(phi) * math.Sin(theta))
		z := float32(r * math.Cos(phi))

		vMag := float32(1.2 * math.Sqrt(float64(n)/(r+4.0)))
		vx := float32(rSource.Float64()-0.5) * vMag
		vy := float32(rSource.Float64()-0.5) * vMag
		vz := float32(rSource.Float64()-0.5) * vMag

		addBody(state, fmt.Sprintf("Cluster Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			8.0, 0.45, Color{255, 230, 180, 255}, false, true, 6)
	}
}

func loadTrappist1(state *SimState, cfg *Config, g float64) {
	starM := 1000.0
	addBody(state, "TRAPPIST-1 Red Dwarf", Vector3{0, 0, 0}, Vector3{0, 0, 0}, starM, 2.4, Color{255, 90, 50, 255}, true, true, 6)

	planets := []struct {
		name string
		dist float32
		mass float64
		col  Color
	}{
		{"TRAPPIST-1b", 6.5, 0.85, Color{190, 170, 160, 255}},
		{"TRAPPIST-1c", 8.8, 1.15, Color{210, 185, 140, 255}},
		{"TRAPPIST-1d", 11.6, 0.30, Color{140, 195, 230, 255}},
		{"TRAPPIST-1e", 15.2, 0.77, Color{110, 175, 245, 255}},
		{"TRAPPIST-1f", 19.8, 0.93, Color{165, 215, 240, 255}},
		{"TRAPPIST-1g", 25.2, 1.15, Color{130, 160, 220, 255}},
		{"TRAPPIST-1h", 32.0, 0.33, Color{200, 220, 245, 255}},
	}

	for i, p := range planets {
		ang := float64(i) * (2.0 * math.Pi / float64(len(planets)))
		x := p.dist * float32(math.Cos(ang))
		z := p.dist * float32(math.Sin(ang))
		v := float32(math.Sqrt((g * starM) / float64(p.dist)))
		vx := -v * float32(math.Sin(ang))
		vz := v * float32(math.Cos(ang))
		addBody(state, p.name, Vector3{x, 0, z}, Vector3{vx, 0, vz}, p.mass, 0.55, p.col, false, false, 1)
	}
}

func loadFiveBodyChoreography(state *SimState, g float64) {
	m := 250.0
	r := float32(22.0)
	// Exact regular 5-body Lagrange equilibrium velocity factor:
	// S_5 = 1/4 * sum_{k=1..4} 1 / sin(k * pi / 5) = 1.37638192
	const s5 = 1.37638192
	v := float32(math.Sqrt((g * m * s5) / float64(r)))
	for i := 0; i < 5; i++ {
		ang := float64(i) * (2.0 * math.Pi / 5.0)
		x := r * float32(math.Cos(ang))
		z := r * float32(math.Sin(ang))
		vx := -v * float32(math.Sin(ang))
		vz := v * float32(math.Cos(ang))
		addBody(state, fmt.Sprintf("Pentagon Star #%d", i+1),
			Vector3{x, 0, z}, Vector3{vx, 0, vz},
			m, 1.8, Color{120, 230, 255, 255}, false, true, 6)
	}
}

func loadRelativisticRosette(state *SimState, cfg *Config, g float64) {
	cfg.EnableRelativity = true
	cfg.SpeedOfLight = 120.0
	bhM := 3500.0
	bh := addBody(state, "Intermediate Black Hole", Vector3{0, 0, 0}, Vector3{0, 0, 0}, bhM, 2.8, Color{12, 10, 20, 255}, true, false, 7)
	bh.IsBlackHole = true

	// High-eccentricity precessing exoplanet
	rApo := float32(32.0)
	vApo := float32(math.Sqrt((g*bhM)/float64(rApo))) * 0.65
	addBody(state, "Rosette Planet", Vector3{rApo, 0, 0}, Vector3{0, 0, vApo}, 0.5, 0.7, Color{255, 140, 80, 255}, false, false, 2)
}

func loadRocheDisruption(state *SimState, cfg *Config, g float64) {
	cfg.EnableRocheLimit = true
	giantM := 2200.0
	addBody(state, "Kronos Gas Giant", Vector3{0, 0, 0}, Vector3{0, 0, 0}, giantM, 4.5, Color{235, 195, 145, 255}, true, false, 3)

	// Infalling fragile icy moon heading inside Roche limit
	addBody(state, "Fragile Moon Pandora", Vector3{35.0, 0, 0}, Vector3{-1.8, 0, 4.2}, 0.8, 0.75, Color{200, 230, 255, 255}, false, false, 5)
}

func loadMilkyWay10K(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	cfg.BarnesHutTheta = 0.72
	cfg.SubSteps = 1
	smbhMass := 50000.0
	bh := addBody(state, "Sagittarius A* Singularity", Vector3{0, 0, 0}, Vector3{0, 0, 0}, smbhMass, 4.5, Color{15, 10, 25, 255}, true, false, 7)
	bh.IsBlackHole = true

	rSource := rand.New(rand.NewSource(10001))
	// Bulge (2,000 stars)
	for i := 0; i < 2000; i++ {
		r := 6.5 + math.Pow(rSource.Float64(), 1.8)*32.0
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := (rSource.Float64() - 0.5) * 0.45
		x := float32(r * math.Cos(theta) * math.Cos(phi))
		z := float32(r * math.Sin(theta) * math.Cos(phi))
		y := float32(r * math.Sin(phi))

		vCirc := float32(math.Sqrt((g * smbhMass) / r))
		vx := -vCirc*float32(math.Sin(theta)) + float32(rSource.Float64()-0.5)*0.6
		vz := vCirc*float32(math.Cos(theta)) + float32(rSource.Float64()-0.5)*0.6
		vy := float32(rSource.Float64()-0.5) * 0.3

		addBody(state, fmt.Sprintf("Bulge Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			0.5, 0.35, Color{255, 210, 140, 255}, false, true, 6)
	}

	// 4 Spiral Arms (8,000 stars)
	for i := 0; i < 8000; i++ {
		armIdx := i % 4
		armOffset := float64(armIdx) * (math.Pi / 2.0)
		r := 25.0 + math.Pow(rSource.Float64(), 1.3)*230.0
		theta := armOffset + 3.2*math.Log(r/25.0) + (rSource.Float64()-0.5)*0.65
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rSource.Float64() - 0.5) * (2.0 + r*0.02))

		haloMass := smbhMass * 0.8 * (r / (r + 40.0))
		vCirc := float32(math.Sqrt((g * (smbhMass + haloMass)) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))
		vy := float32((rSource.Float64() - 0.5) * 0.3)

		addBody(state, fmt.Sprintf("Arm Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			0.4, 0.32, Color{180, 220, 255, 255}, false, true, 6)
	}
}

func loadAsteroidBelt2K(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	loadSolarSystem(state, g)

	sunM := 1000.0
	rSource := rand.New(rand.NewSource(2000))
	for i := 0; i < 1400; i++ {
		r := 30.0 + rSource.Float64()*9.0
		// Kirkwood gaps: 3:1 (r~32.5), 5:2 (r~35.2), 2:1 (r~37.8)
		if (r > 32.2 && r < 32.8) || (r > 34.9 && r < 35.5) || (r > 37.5 && r < 38.1) {
			if rSource.Float64() > 0.08 {
				r += 1.2
			}
		}
		theta := rSource.Float64() * 2.0 * math.Pi
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rSource.Float64() - 0.5) * 1.5)

		vCirc := float32(math.Sqrt((g * sunM) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))

		addBody(state, fmt.Sprintf("Belt Asteroid #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, 0, vz},
			0.0004, 0.18, Color{165, 160, 155, 255}, false, false, 5)
	}
}

func loadGalaxyCollision5K(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	g1Pos := Vector3{-65, 0, -35}
	g1Vel := Vector3{1.6, 0, 0.8}
	g2Pos := Vector3{65, 18, 35}
	g2Vel := Vector3{-1.6, -0.4, -0.8}

	addGalaxyDisk(state, g1Pos, g1Vel, 4500.0, 2700, 48.0, g, Color{135, 205, 255, 255})
	addGalaxyDisk(state, g2Pos, g2Vel, 3800.0, 2300, 42.0, g, Color{255, 185, 130, 255})
}

func loadBlackHoleSwarm3K(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	cfg.EnableRelativity = true
	cfg.SpeedOfLight = 160.0
	gargantuaM := 45000.0
	bh := addBody(state, "Gargantua Singularity", Vector3{0, 0, 0}, Vector3{0, 0, 0}, gargantuaM, 4.8, Color{10, 8, 20, 255}, true, false, 7)
	bh.IsBlackHole = true
	bh.IsPulsar = true // relativistic jet beams

	rSource := rand.New(rand.NewSource(3001))
	// 1,500 Accretion Disk Particles
	for i := 0; i < 1500; i++ {
		r := 6.5 + math.Pow(rSource.Float64(), 1.5)*38.0
		theta := rSource.Float64() * 2.0 * math.Pi
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rSource.Float64() - 0.5) * 0.45)

		vCirc := float32(math.Sqrt((g * gargantuaM) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))

		addBody(state, fmt.Sprintf("Accretion Plasma #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, 0, vz},
			0.002, 0.22, Color{255, 170, 70, 255}, false, false, 5)
	}

	// 1,500 Relativistic Precessing Stars
	for i := 0; i < 1500; i++ {
		r := 35.0 + math.Pow(rSource.Float64(), 1.2)*90.0
		theta := rSource.Float64() * 2.0 * math.Pi
		inc := (rSource.Float64() - 0.5) * 0.7
		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32(r * math.Sin(inc))

		vCirc := float32(math.Sqrt((g * gargantuaM) / r))
		vx := -vCirc*float32(math.Sin(theta)) + float32(rSource.Float64()-0.5)*0.5
		vz := vCirc*float32(math.Cos(theta)) + float32(rSource.Float64()-0.5)*0.5
		vy := float32((rSource.Float64() - 0.5) * 0.5)

		addBody(state, fmt.Sprintf("Rosette Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			0.5, 0.32, Color{190, 220, 255, 255}, false, true, 6)
	}
}

// loadGravitationalCollapse loads 1,600 cold gas particles that collapse under mutual gravity into a sphere
func loadGravitationalCollapse(state *SimState, cfg *Config, g float64) {
	cfg.EnableBarnesHut = true
	cfg.BarnesHutTheta = 0.7
	cfg.Softening = 0.65

	count := 1600
	cloudRadius := 55.0
	totalMass := 4500.0
	partMass := totalMass / float64(count)

	rSource := rand.New(rand.NewSource(1337))

	for i := 0; i < count; i++ {
		u := rSource.Float64()
		r := cloudRadius * math.Pow(u, 0.45)
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := math.Acos(2.0*rSource.Float64() - 1.0)

		clump := 1.0 + 0.15*math.Sin(theta*3.0)*math.Cos(phi*2.0)
		x := float32(r * math.Sin(phi) * math.Cos(theta) * 1.35 * clump)
		y := float32(r * math.Sin(phi) * math.Sin(theta) * 0.85 * clump)
		z := float32(r * math.Cos(phi) * 1.15 * clump)

		vDisp := float32(0.35)
		vx := float32(rSource.Float64()-0.5) * vDisp
		vy := float32(rSource.Float64()-0.5) * vDisp * 0.6
		vz := float32(rSource.Float64()-0.5) * vDisp

		mass := partMass * (0.7 + rSource.Float64()*0.6)
		rad := float32(0.25 + mass*0.04)

		var col Color
		roll := rSource.Float64()
		if roll < 0.25 {
			col = Color{100, 200, 255, 255}
		} else if roll < 0.65 {
			col = Color{255, 220, 110, 255}
		} else {
			col = Color{255, 120, 50, 255}
		}

		addBody(state, fmt.Sprintf("Cloud Star #%d", i+1),
			Vector3{x, y, z}, Vector3{vx, vy, vz},
			mass, rad, col, false, true, 1)
	}
}

func loadRealScaleSolarSystem(state *SimState, cfg *Config, g float64) {
	sunMass := 10000.0
	addBody(state, "Sun Helios", Vector3{0, 0, 0}, Vector3{0, 0, 0}, sunMass, 4.5, Color{255, 215, 0, 255}, true, true, 6)

	spawnPlanet := func(name string, aAU, e, iDeg float64, mass float64, baseRadius float32, col Color, tex int, trueAnomaly float64) *Body {
		a := aAU * 100.0
		theta := trueAnomaly
		r := (a * (1.0 - e*e)) / (1.0 + e*math.Cos(theta))
		h := math.Sqrt(g * sunMass * a * (1.0 - e*e))
		vr := (g * sunMass / h) * e * math.Sin(theta)
		vtheta := (g * sunMass / h) * (1.0 + e*math.Cos(theta))

		vxPlane := vr*math.Cos(theta) - vtheta*math.Sin(theta)
		vzPlane := vr*math.Sin(theta) + vtheta*math.Cos(theta)

		iRad := iDeg * math.Pi / 180.0
		x := float32(r * math.Cos(theta))
		y := float32(r * math.Sin(theta) * math.Sin(iRad))
		z := float32(r * math.Sin(theta) * math.Cos(iRad))

		vx := float32(vxPlane)
		vy := float32(vzPlane * math.Sin(iRad))
		vz := float32(vzPlane * math.Cos(iRad))

		return addBody(state, name, Vector3{x, y, z}, Vector3{vx, vy, vz}, mass, baseRadius, col, false, false, tex)
	}

	spawnPlanet("Mercury", 0.3871, 0.2056, 7.00, 0.00166, 0.65, Color{180, 180, 180, 255}, 5, 0.8)
	spawnPlanet("Venus", 0.7233, 0.0068, 3.39, 0.0245, 1.1, Color{230, 205, 150, 255}, 2, 2.1)
	earth := spawnPlanet("Earth", 1.0000, 0.0167, 0.00, 0.0300, 1.2, Color{60, 150, 245, 255}, 1, 0.0)

	moonDist := float32(2.5)
	moonV := float32(math.Sqrt((g * earth.Mass) / float64(moonDist)))
	addBody(state, "Moon Luna",
		Vector3{earth.Position.X + moonDist, earth.Position.Y, earth.Position.Z},
		Vector3{earth.Velocity.X, earth.Velocity.Y, earth.Velocity.Z - moonV},
		0.000369, 0.45, Color{200, 200, 200, 255}, false, false, 5)

	spawnPlanet("Mars", 1.5237, 0.0934, 1.85, 0.00323, 0.75, Color{235, 95, 55, 255}, 2, 3.4)
	spawnPlanet("Ceres (Dwarf Planet)", 2.767, 0.0758, 10.59, 0.000047, 0.5, Color{160, 160, 170, 255}, 5, 1.2)
	spawnPlanet("Vesta", 2.362, 0.0887, 7.14, 0.000013, 0.4, Color{200, 190, 180, 255}, 5, 4.5)
	spawnPlanet("Pallas", 2.772, 0.2307, 34.84, 0.000011, 0.4, Color{150, 180, 190, 255}, 5, 2.9)

	rSource := rand.New(rand.NewSource(42))
	for i := 0; i < 30; i++ {
		aAst := 2.15 + rSource.Float64()*1.15
		eAst := rSource.Float64() * 0.18
		iAst := (rSource.Float64() - 0.5) * 16.0
		thetaAst := rSource.Float64() * 2.0 * math.Pi
		spawnPlanet(fmt.Sprintf("Belt Asteroid #%d", i+1), aAst, eAst, iAst, 0.000005, 0.28, Color{150, 150, 150, 255}, 5, thetaAst)
	}

	jup := spawnPlanet("Jupiter", 5.2044, 0.0484, 1.30, 9.548, 3.2, Color{230, 175, 120, 255}, 3, 5.1)
	galMoons := []struct {
		name string
		dist float32
		mass float64
		col  Color
	}{
		{"Io", 5.2, 0.00045, Color{255, 230, 80, 255}},
		{"Europa", 7.8, 0.00024, Color{210, 230, 255, 255}},
		{"Ganymede", 11.5, 0.00075, Color{180, 170, 160, 255}},
		{"Callisto", 18.0, 0.00054, Color{130, 130, 140, 255}},
	}
	for _, gm := range galMoons {
		spd := float32(math.Sqrt((g * jup.Mass) / float64(gm.dist)))
		addBody(state, gm.name,
			Vector3{jup.Position.X + gm.dist, jup.Position.Y, jup.Position.Z},
			Vector3{jup.Velocity.X, jup.Velocity.Y, jup.Velocity.Z - spd},
			gm.mass, 0.45, gm.col, false, false, 5)
	}

	sat := spawnPlanet("Saturn", 9.5826, 0.0541, 2.48, 2.858, 2.6, Color{240, 215, 160, 255}, 3, 1.7)
	sat.RingInnerRadius = 3.6
	sat.RingOuterRadius = 6.8
	titanSpd := float32(math.Sqrt((g * sat.Mass) / 12.0))
	addBody(state, "Titan",
		Vector3{sat.Position.X + 12.0, sat.Position.Y, sat.Position.Z},
		Vector3{sat.Velocity.X, sat.Velocity.Y, sat.Velocity.Z - titanSpd},
		0.00067, 0.6, Color{245, 190, 80, 255}, false, false, 1)

	spawnPlanet("Uranus", 19.2184, 0.0472, 0.77, 0.436, 1.8, Color{140, 230, 240, 255}, 4, 3.8)
	nep := spawnPlanet("Neptune", 30.0700, 0.0086, 1.77, 0.515, 1.8, Color{70, 120, 255, 255}, 4, 0.4)
	tritonSpd := float32(math.Sqrt((g * nep.Mass) / 7.5))
	addBody(state, "Triton (Retrograde Moon)",
		Vector3{nep.Position.X + 7.5, nep.Position.Y, nep.Position.Z},
		Vector3{nep.Velocity.X, nep.Velocity.Y, nep.Velocity.Z + tritonSpd},
		0.00011, 0.45, Color{200, 220, 235, 255}, false, false, 5)

	spawnPlanet("Pluto (Kuiper Belt)", 39.482, 0.2488, 17.16, 0.000065, 0.5, Color{210, 185, 160, 255}, 5, 4.9)
	comet := spawnPlanet("Halley's Comet (1P/Halley)", 17.83, 0.9671, 17.8, 0.000001, 0.4, Color{255, 255, 255, 255}, 5, 0.15)
	comet.IsComet = true
}
