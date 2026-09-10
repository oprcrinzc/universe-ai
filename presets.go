package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// InitStarfield creates random 3D background stars
func InitStarfield(state *SimState, count int) {
	state.Stars = make([]rl.Vector3, count)
	state.StarColors = make([]rl.Color, count)
	state.StarSizes = make([]float32, count)

	palette := []rl.Color{
		rl.NewColor(255, 255, 255, 200),
		rl.NewColor(200, 220, 255, 220),
		rl.NewColor(255, 240, 200, 210),
		rl.NewColor(255, 200, 180, 190),
		rl.NewColor(180, 200, 255, 180),
	}

	for i := 0; i < count; i++ {
		// Random point on sphere
		theta := rand.Float64() * 2 * math.Pi
		phi := math.Acos(2*rand.Float64() - 1)
		dist := 600.0 + rand.Float64()*300.0

		x := float32(dist * math.Sin(phi) * math.Cos(theta))
		y := float32(dist * math.Cos(phi))
		z := float32(dist * math.Sin(phi) * math.Sin(theta))

		state.Stars[i] = rl.NewVector3(x, y, z)
		state.StarColors[i] = palette[rand.Intn(len(palette))]
		state.StarSizes[i] = float32(0.8 + rand.Float64()*1.2)
	}
}

// LoadPreset populates the simulation with a selected planetary system
func LoadPreset(state *SimState, cfg *Config, preset PresetType) {
	state.CurrentPreset = preset
	state.Bodies = make([]*Body, 0)
	state.NextID = 1
	state.SelectedBodyID = -1
	state.FollowSelected = false

	g := cfg.G

	switch preset {
	case PresetSolarSystem:
		loadSolarSystem(state, g)
	case PresetBinaryStars:
		loadBinaryStars(state, g)
	case PresetThreeBody:
		loadThreeBody(state, g)
	case PresetGalaxyDisk:
		loadGalaxyDisk(state, g)
	case PresetEmpty:
		// Clean slate
	}
}

func addBody(state *SimState, name string, pos, vel rl.Vector3, mass float64, radius float32, col rl.Color, isStationary, isStar bool) *Body {
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
		Trail:        make([]rl.Vector3, 0, 100),
	}
	state.NextID++
	state.Bodies = append(state.Bodies, b)
	return b
}

func loadSolarSystem(state *SimState, g float64) {
	// Central Sun (Massive, Stationary anchor or movable)
	sunMass := 12000.0
	sun := addBody(state, "Sun",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		sunMass, 4.2, rl.Gold, true, true)

	// Helper to calculate circular velocity around Sun
	orbitSun := func(dist float32, mass float64, radius float32, col rl.Color, name string) *Body {
		speed := float32(math.Sqrt((g * sunMass) / float64(dist)))
		return addBody(state, name,
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -speed),
			mass, radius, col, false, false)
	}

	// Inner Planets
	orbitSun(12.0, 0.4, 0.7, rl.NewColor(170, 170, 170, 255), "Mercury")
	orbitSun(19.0, 1.2, 1.0, rl.NewColor(225, 200, 140, 255), "Venus")

	// Earth & Moon (Hierarchical two-body system)
	earthDist := float32(28.0)
	earthSpeed := float32(math.Sqrt((g * sunMass) / float64(earthDist)))
	earth := addBody(state, "Earth",
		rl.NewVector3(earthDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed),
		2.0, 1.2, rl.NewColor(65, 140, 240, 255), false, false)

	// Moon orbiting Earth
	moonDist := float32(2.5)
	moonSpeed := float32(math.Sqrt((g * earth.Mass) / float64(moonDist)))
	addBody(state, "Moon",
		rl.NewVector3(earthDist+moonDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed-moonSpeed),
		0.05, 0.4, rl.NewColor(200, 200, 200, 255), false, false)

	// Mars
	orbitSun(38.0, 0.8, 0.85, rl.NewColor(230, 80, 45, 255), "Mars")

	// Jupiter and Galilean Moons
	jupDist := float32(56.0)
	jupSpeed := float32(math.Sqrt((g * sunMass) / float64(jupDist)))
	jup := addBody(state, "Jupiter",
		rl.NewVector3(jupDist, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed),
		25.0, 2.5, rl.NewColor(220, 175, 125, 255), false, false)

	// Io and Europa orbiting Jupiter
	addBody(state, "Io",
		rl.NewVector3(jupDist+3.6, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed-float32(math.Sqrt((g*jup.Mass)/3.6))),
		0.02, 0.35, rl.NewColor(240, 220, 80, 255), false, false)

	addBody(state, "Europa",
		rl.NewVector3(jupDist+5.2, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed-float32(math.Sqrt((g*jup.Mass)/5.2))),
		0.02, 0.35, rl.NewColor(200, 230, 255, 255), false, false)

	// Saturn and Ring Particles
	satDist := float32(82.0)
	satSpeed := float32(math.Sqrt((g * sunMass) / float64(satDist)))
	sat := addBody(state, "Saturn",
		rl.NewVector3(satDist, 0, 0),
		rl.NewVector3(0, 0, -satSpeed),
		18.0, 2.1, rl.NewColor(235, 210, 160, 255), false, false)

	// Saturn Rings (small asteroids in ring formation)
	for i := 0; i < 16; i++ {
		angle := float64(i) * (2 * math.Pi / 16.0)
		ringR := float32(3.6 + float64(i%3)*0.5)
		ringSpeed := float32(math.Sqrt((g * sat.Mass) / float64(ringR)))

		rx := satDist + ringR*float32(math.Cos(angle))
		rz := ringR * float32(math.Sin(angle))
		vx := sat.Velocity.X - ringSpeed*float32(math.Sin(angle))
		vz := sat.Velocity.Z + ringSpeed*float32(math.Cos(angle))

		addBody(state, "Ring Particle",
			rl.NewVector3(rx, 0.1*float32(math.Sin(angle*2)), rz),
			rl.NewVector3(vx, 0, vz),
			0.001, 0.18, rl.NewColor(210, 195, 170, 200), false, false)
	}

	// Uranus & Neptune
	orbitSun(112.0, 8.0, 1.6, rl.NewColor(140, 220, 235, 255), "Uranus")
	orbitSun(145.0, 9.0, 1.6, rl.NewColor(75, 110, 245, 255), "Neptune")

	// Eccentric Inclined Comet
	cometDist := float32(110.0)
	addBody(state, "Comet Halley",
		rl.NewVector3(cometDist, 20.0, 40.0),
		rl.NewVector3(-3.2, -1.8, -4.5),
		0.01, 0.45, rl.NewColor(180, 240, 255, 255), false, false)

	_ = sun
}

func loadBinaryStars(state *SimState, g float64) {
	// Two massive stars orbiting their common center of mass
	mStar := 5000.0
	sep := float32(16.0)
	orbitV := float32(math.Sqrt((g * mStar) / (4.0 * float64(sep))))

	addBody(state, "Star Alpha",
		rl.NewVector3(-sep, 0, 0),
		rl.NewVector3(0, 0, orbitV),
		mStar, 3.2, rl.NewColor(255, 190, 50, 255), false, true)

	addBody(state, "Star Beta",
		rl.NewVector3(sep, 0, 0),
		rl.NewVector3(0, 0, -orbitV),
		mStar, 3.2, rl.NewColor(100, 200, 255, 255), false, true)

	// Circumbinary planets orbiting around the binary pair
	totalM := mStar * 2
	distances := []float32{48.0, 72.0, 105.0, 140.0}
	colors := []rl.Color{rl.Lime, rl.SkyBlue, rl.Pink, rl.Beige}
	names := []string{"Planet Tatooine I", "Planet Tatooine II", "Planet Oricon", "Gas Giant Arrakis"}

	for idx, dist := range distances {
		speed := float32(math.Sqrt((g * totalM) / float64(dist)))
		addBody(state, names[idx],
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -speed),
			2.5, 1.2+float32(idx)*0.25, colors[idx], false, false)
	}
}

func loadThreeBody(state *SimState, g float64) {
	// Chaotic 3-body choreography
	m := 3000.0
	r := float32(24.0)

	// Equilateral triangle layout with initial tangential velocities
	angles := []float64{0, 2 * math.Pi / 3, 4 * math.Pi / 3}
	names := []string{"Trisolaris A", "Trisolaris B", "Trisolaris C"}
	colors := []rl.Color{rl.Gold, rl.NewColor(0, 230, 255, 255), rl.Magenta}

	for i := 0; i < 3; i++ {
		x := r * float32(math.Cos(angles[i]))
		z := r * float32(math.Sin(angles[i]))

		// Perpendicular velocity
		vMag := float32(math.Sqrt((g * m * 1.5) / float64(r)))
		vx := -vMag * float32(math.Sin(angles[i]))
		vz := vMag * float32(math.Cos(angles[i]))

		addBody(state, names[i],
			rl.NewVector3(x, 0, z),
			rl.NewVector3(vx*0.9, float32(i-1)*0.5, vz*0.9), // slight Y perturbation for full 3D motion!
			m, 2.6, colors[i], false, true)
	}

	// A few small rogue planets trapped in the chaotic gravity
	addBody(state, "Rogue Planet 1",
		rl.NewVector3(0, 15, 0),
		rl.NewVector3(2.5, 0, -1.8),
		0.5, 0.7, rl.White, false, false)

	addBody(state, "Rogue Planet 2",
		rl.NewVector3(35, -10, 20),
		rl.NewVector3(-1.8, 1.2, 3.2),
		0.5, 0.7, rl.LightGray, false, false)
}

func loadGalaxyDisk(state *SimState, g float64) {
	// Supermassive Central Black Hole
	smbhMass := 25000.0
	addBody(state, "Supermassive Singularity",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		smbhMass, 3.5, rl.NewColor(30, 20, 50, 255), true, true)

	// Swirling accretion disk of 60 orbital bodies
	particleCount := 60
	for i := 0; i < particleCount; i++ {
		dist := float32(16.0 + rand.Float64()*120.0)
		angle := rand.Float64() * 2 * math.Pi

		// Circular speed with tiny perturbation
		speed := float32(math.Sqrt((g*smbhMass)/float64(dist))) * (0.96 + rand.Float32()*0.08)

		x := dist * float32(math.Cos(angle))
		z := dist * float32(math.Sin(angle))
		y := (rand.Float32() - 0.5) * (dist * 0.08) // Disk thickness

		vx := -speed * float32(math.Sin(angle))
		vz := speed * float32(math.Cos(angle))
		vy := (rand.Float32() - 0.5) * 0.4

		mass := 0.2 + rand.Float64()*1.8
		rad := float32(0.35 + mass*0.25)

		cVal := uint8(140 + rand.Intn(115))
		col := rl.NewColor(cVal, uint8(120+rand.Intn(100)), uint8(200+rand.Intn(55)), 255)

		addBody(state, "Disk Mass",
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false)
	}
}

// CreateBodyTemplate creates a ready-to-launch body based on SpawnType
func CreateBodyTemplate(spawnType SpawnType, pos, vel rl.Vector3, id int64) *Body {
	switch spawnType {
	case SpawnEarth:
		return &Body{
			ID:           id,
			Name:         "Terrestrial Planet",
			Position:     pos,
			Velocity:     vel,
			Mass:         2.0,
			Radius:       1.2,
			Color:        rl.NewColor(70, 160, 245, 255),
			IsStationary: false,
			IsStar:       false,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	case SpawnMoon:
		return &Body{
			ID:           id,
			Name:         "Moon / Asteroid",
			Position:     pos,
			Velocity:     vel,
			Mass:         0.1,
			Radius:       0.5,
			Color:        rl.NewColor(190, 190, 190, 255),
			IsStationary: false,
			IsStar:       false,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	case SpawnGasGiant:
		return &Body{
			ID:           id,
			Name:         "Gas Giant",
			Position:     pos,
			Velocity:     vel,
			Mass:         20.0,
			Radius:       2.3,
			Color:        rl.NewColor(230, 170, 110, 255),
			IsStationary: false,
			IsStar:       false,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	case SpawnStar:
		return &Body{
			ID:           id,
			Name:         "Protostar",
			Position:     pos,
			Velocity:     vel,
			Mass:         5000.0,
			Radius:       3.6,
			Color:        rl.Gold,
			IsStationary: false,
			IsStar:       true,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	case SpawnBlackHole:
		return &Body{
			ID:           id,
			Name:         "Micro Black Hole",
			Position:     pos,
			Velocity:     vel,
			Mass:         15000.0,
			Radius:       2.0,
			Color:        rl.NewColor(40, 20, 60, 255),
			IsStationary: true,
			IsStar:       true,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	default:
		return &Body{
			ID:           id,
			Name:         "New Body",
			Position:     pos,
			Velocity:     vel,
			Mass:         1.0,
			Radius:       1.0,
			Color:        rl.White,
			IsStationary: false,
			IsStar:       false,
			Trail:        make([]rl.Vector3, 0, 100),
		}
	}
}
