package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

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
	state.FollowBarycenter = false

	if cfg.SpacetimeResolution <= 0 {
		cfg.SpacetimeResolution = 120
	}
	cfg.SpacetimeScale = 1.0
	if cfg.TimeScaleStep <= 0 {
		cfg.TimeScaleStep = 0.5
	}

	g := cfg.G

	switch preset {
	case PresetSolarSystem:
		loadSolarSystem(state, g)
	case PresetSolarSystemGrand:
		loadSolarSystemGrand(state, g)
	case PresetMilkyWay10K:
		loadMilkyWay10K(state, cfg, g)
	case PresetAsteroidBelt2K:
		loadAsteroidBelt2K(state, cfg, g)
	case PresetGalaxyCollision5K:
		loadGalaxyCollision5K(state, cfg, g)
	case PresetBlackHoleSwarm3K:
		loadBlackHoleSwarm3K(state, cfg, g)
	case PresetBinaryStars:
		loadBinaryStars(state, g)
	case PresetThreeBody:
		loadThreeBody(state, g)
	case PresetGalaxyDisk:
		loadGalaxyDisk(state, g)
	case PresetMilkyWayCluster:
		loadMilkyWayCluster(state, cfg, g)
	case PresetLagrangeTrojans:
		loadLagrangeTrojans(state, cfg, g)
	case PresetGalaxyCollision:
		loadGalaxyCollision(state, cfg, g)
	case PresetPulsarAccretion:
		loadPulsarAccretion(state, cfg, g)
	case PresetGlobularCluster:
		loadGlobularCluster(state, cfg, g)
	case PresetTrappistResonance:
		loadTrappistResonance(state, cfg, g)
	case PresetFiveBody:
		loadFiveBody(state, g)
	case PresetRelativityRosette:
		loadRelativityRosette(state, cfg, g)
	case PresetRocheDisruption:
		loadRocheDisruption(state, cfg, g)
	case PresetGravitationalCollapse:
		loadGravitationalCollapse(state, cfg, g)
	case PresetRealSolarSystem:
		loadRealScaleSolarSystem(state, cfg, g)
	case PresetEmpty:
		// Clean slate
	}
}

func addBody(state *SimState, name string, pos, vel rl.Vector3, mass float64, radius float32, col rl.Color, isStationary, isStar bool) *Body {
	return addBodyTex(state, name, pos, vel, mass, radius, col, isStationary, isStar, TextureNone, 1.0)
}

func addBodyTex(state *SimState, name string, pos, vel rl.Vector3, mass float64, radius float32, col rl.Color, isStationary, isStar bool, texType BodyTextureType, rotSpeed float32) *Body {
	b := &Body{
		ID:            state.NextID,
		Name:          name,
		Position:      pos,
		Velocity:      vel,
		Mass:          mass,
		Radius:        radius,
		Color:         col,
		IsStationary:  isStationary,
		IsStar:        isStar,
		Trail:         make([]rl.Vector3, 0, 100),
		TextureType:   texType,
		RotationAngle: rand.Float32() * 360.0,
		RotationSpeed: rotSpeed,
	}
	state.NextID++
	state.Bodies = append(state.Bodies, b)
	return b
}

func loadSolarSystem(state *SimState, g float64) {
	// Central Sun (Massive, Stationary anchor or movable)
	sunMass := 12000.0
	sun := addBodyTex(state, "Sun",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		sunMass, 4.2, rl.Gold, true, true, TextureSun, 0.4)

	// Helper to calculate circular velocity around Sun
	orbitSun := func(dist float32, mass float64, radius float32, col rl.Color, name string, texType BodyTextureType, rotSpeed float32) *Body {
		speed := float32(math.Sqrt((g * sunMass) / float64(dist)))
		return addBodyTex(state, name,
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -speed),
			mass, radius, col, false, false, texType, rotSpeed)
	}

	// Inner Planets
	orbitSun(12.0, 0.4, 0.7, rl.NewColor(170, 170, 170, 255), "Mercury", TextureMoon, 0.5)
	orbitSun(19.0, 1.2, 1.0, rl.NewColor(225, 200, 140, 255), "Venus", TextureDesert, -0.4)

	// Earth & Moon (Hierarchical two-body system, stable inside Hill sphere)
	earthDist := float32(28.0)
	earthSpeed := float32(math.Sqrt((g * sunMass) / float64(earthDist)))
	earthMass := 24.0
	earth := addBodyTex(state, "Earth",
		rl.NewVector3(earthDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed),
		earthMass, 1.2, rl.NewColor(65, 140, 240, 255), false, false, TextureTerrestrial, 1.2)

	// Moon orbiting Earth stably inside Hill sphere (rH ~ 2.44)
	moonDist := float32(1.6)
	moonSpeed := float32(math.Sqrt((g * earth.Mass) / float64(moonDist)))
	addBodyTex(state, "Moon",
		rl.NewVector3(earthDist+moonDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed-moonSpeed),
		0.08, 0.38, rl.NewColor(200, 200, 200, 255), false, false, TextureMoon, 0.8)

	// Mars
	orbitSun(38.0, 1.2, 0.85, rl.NewColor(230, 80, 45, 255), "Mars", TextureDesert, 1.1)

	// Jupiter and Galilean Moons (stable inside Hill sphere rH ~ 8.36)
	jupDist := float32(56.0)
	jupSpeed := float32(math.Sqrt((g * sunMass) / float64(jupDist)))
	jupMass := 120.0
	jup := addBodyTex(state, "Jupiter",
		rl.NewVector3(jupDist, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed),
		jupMass, 2.5, rl.NewColor(220, 175, 125, 255), false, false, TextureGasGiant, 2.4)

	// Io and Europa orbiting Jupiter stably
	addBodyTex(state, "Io",
		rl.NewVector3(jupDist+2.8, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed-float32(math.Sqrt((g*jup.Mass)/2.8))),
		0.05, 0.35, rl.NewColor(240, 220, 80, 255), false, false, TextureMoon, 1.0)

	addBodyTex(state, "Europa",
		rl.NewVector3(jupDist+4.2, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed-float32(math.Sqrt((g*jup.Mass)/4.2))),
		0.04, 0.35, rl.NewColor(200, 230, 255, 255), false, false, TextureMoon, 1.0)

	// Saturn and Ring Particles
	satDist := float32(82.0)
	satSpeed := float32(math.Sqrt((g * sunMass) / float64(satDist)))
	satMass := 45.0
	sat := addBodyTex(state, "Saturn",
		rl.NewVector3(satDist, 0, 0),
		rl.NewVector3(0, 0, -satSpeed),
		satMass, 2.1, rl.NewColor(235, 210, 160, 255), false, false, TextureGasGiant, 2.0)

	// Saturn Rings (small asteroids in ring formation)
	for i := 0; i < 16; i++ {
		angle := float64(i) * (2 * math.Pi / 16.0)
		ringR := float32(3.2 + float64(i%3)*0.4)
		ringSpeed := float32(math.Sqrt((g * sat.Mass) / float64(ringR)))

		rx := satDist + ringR*float32(math.Cos(angle))
		rz := ringR * float32(math.Sin(angle))
		vx := sat.Velocity.X - ringSpeed*float32(math.Sin(angle))
		vz := sat.Velocity.Z + ringSpeed*float32(math.Cos(angle))

		addBodyTex(state, "Ring Particle",
			rl.NewVector3(rx, 0.1*float32(math.Sin(angle*2)), rz),
			rl.NewVector3(vx, 0, vz),
			0.001, 0.18, rl.NewColor(210, 195, 170, 200), false, false, TextureMoon, 1.0)
	}

	// Uranus & Neptune
	orbitSun(112.0, 8.0, 1.6, rl.NewColor(140, 220, 235, 255), "Uranus", TextureIceGiant, 1.4)
	orbitSun(145.0, 9.0, 1.6, rl.NewColor(75, 110, 245, 255), "Neptune", TextureIceGiant, 1.4)

	// Eccentric Inclined Comet
	cometDist := float32(110.0)
	addBodyTex(state, "Comet Halley",
		rl.NewVector3(cometDist, 20.0, 40.0),
		rl.NewVector3(-3.2, -1.8, -4.5),
		0.01, 0.45, rl.NewColor(180, 240, 255, 255), false, false, TextureMoon, 0.5)

	_ = sun
}

func loadBinaryStars(state *SimState, g float64) {
	// Two massive stars orbiting their common center of mass
	mStar := 5000.0
	sep := float32(16.0)
	orbitV := float32(math.Sqrt((g * mStar) / (4.0 * float64(sep))))

	addBodyTex(state, "Star Alpha",
		rl.NewVector3(-sep, 0, 0),
		rl.NewVector3(0, 0, orbitV),
		mStar, 3.2, rl.NewColor(255, 190, 50, 255), false, true, TextureSun, 0.8)

	addBodyTex(state, "Star Beta",
		rl.NewVector3(sep, 0, 0),
		rl.NewVector3(0, 0, -orbitV),
		mStar, 3.2, rl.NewColor(100, 200, 255, 255), false, true, TextureSun, -0.8)

	// Circumbinary planets orbiting around the binary pair (safely outside chaotic resonance boundary r >= 72)
	totalM := mStar * 2
	distances := []float32{72.0, 105.0, 145.0, 195.0}
	colors := []rl.Color{rl.Lime, rl.SkyBlue, rl.Pink, rl.Beige}
	names := []string{"Planet Tatooine I", "Planet Tatooine II", "Planet Oricon", "Gas Giant Arrakis"}
	texTypes := []BodyTextureType{TextureTerrestrial, TextureDesert, TextureIceGiant, TextureGasGiant}

	for idx, dist := range distances {
		speed := float32(math.Sqrt((g * totalM) / float64(dist)))
		addBodyTex(state, names[idx],
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -speed),
			2.5, 1.2+float32(idx)*0.25, colors[idx], false, false, texTypes[idx], 1.0)
	}
}

func loadThreeBody(state *SimState, g float64) {
	// Stable Lagrange equilateral 3-body rotating choreography
	m := 3000.0
	r := float32(24.0)

	// Equilateral triangle layout with exact centripetal orbital velocity
	angles := []float64{0, 2 * math.Pi / 3, 4 * math.Pi / 3}
	names := []string{"Trisolaris A", "Trisolaris B", "Trisolaris C"}
	colors := []rl.Color{rl.Gold, rl.NewColor(0, 230, 255, 255), rl.Magenta}

	// Exact Lagrange 3-body circular speed: v = sqrt(G * m / (sqrt(3) * r))
	vMag := float32(math.Sqrt((g * m * (1.0 / math.Sqrt(3.0))) / float64(r)))

	for i := 0; i < 3; i++ {
		x := r * float32(math.Cos(angles[i]))
		z := r * float32(math.Sin(angles[i]))

		vx := -vMag * float32(math.Sin(angles[i]))
		vz := vMag * float32(math.Cos(angles[i]))

		addBodyTex(state, names[i],
			rl.NewVector3(x, 0, z),
			rl.NewVector3(vx, 0, vz),
			m, 2.6, colors[i], false, true, TextureSun, 0.6)
	}

	// Stable circumbinary/circumsystem planet outside the chaotic inner region (R = 75)
	pDist := float32(75.0)
	totM := m * 3.0
	pSpeed := float32(math.Sqrt((g * totM) / float64(pDist)))
	addBodyTex(state, "Trisolaran Exoplanet",
		rl.NewVector3(pDist, 0, 0),
		rl.NewVector3(0, 0, -pSpeed),
		1.0, 0.9, rl.SkyBlue, false, false, TextureTerrestrial, 1.0)
}

func loadGalaxyDisk(state *SimState, g float64) {
	// Supermassive Central Black Hole
	smbhMass := 25000.0
	addBodyTex(state, "Supermassive Singularity",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		smbhMass, 3.5, rl.NewColor(30, 20, 50, 255), true, true, TextureBlackHole, 0.5)

	// Swirling accretion disk of 60 orbital bodies
	particleCount := 60
	for i := 0; i < particleCount; i++ {
		dist := float32(16.0 + rand.Float64()*120.0)
		angle := rand.Float64() * 2 * math.Pi

		speed := float32(math.Sqrt((g*smbhMass)/float64(dist))) * (0.96 + rand.Float32()*0.08)

		x := dist * float32(math.Cos(angle))
		z := dist * float32(math.Sin(angle))
		y := (rand.Float32() - 0.5) * (dist * 0.08)

		vx := -speed * float32(math.Sin(angle))
		vz := speed * float32(math.Cos(angle))
		vy := (rand.Float32() - 0.5) * 0.4

		mass := 0.2 + rand.Float64()*1.8
		rad := float32(0.35 + mass*0.25)

		cVal := uint8(140 + rand.Intn(115))
		col := rl.NewColor(cVal, uint8(120+rand.Intn(100)), uint8(200+rand.Intn(55)), 255)

		addBodyTex(state, "Disk Mass",
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, TextureMoon, 1.0)
	}
}

// loadMilkyWayCluster demonstrates the Barnes-Hut O(N log N) octree with 650 stars
func loadMilkyWayCluster(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.7
	cfg.NotificationText = "Barnes-Hut Octree active: O(N log N) acceleration for 650+ stars!"
	cfg.NotificationTimer = 4.0

	// Central Supermassive Singularity
	smbhMass := 50000.0
	addBodyTex(state, "Sagittarius A* Singularity",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		smbhMass, 4.5, rl.NewColor(25, 15, 35, 255), true, true, TextureBlackHole, 0.4)

	// 650 Stars distributed in galactic bulge and spiral arms
	totalStars := 650
	for i := 0; i < totalStars; i++ {
		// Logarithmic spiral distribution with 2 main arms
		arm := float64(i % 2)
		r := 15.0 + math.Pow(rand.Float64(), 1.5)*180.0
		theta := arm*math.Pi + 2.5*math.Log(r/15.0) + (rand.Float64()-0.5)*0.75

		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		y := float32((rand.Float64() - 0.5) * (r * 0.08))

		// Keplerian circular speed
		vCirc := float32(math.Sqrt((g * smbhMass) / r))
		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))
		vy := float32((rand.Float64() - 0.5) * 0.3)

		mass := 0.5 + rand.Float64()*3.0
		rad := float32(0.3 + mass*0.15)

		// Temperature-based stellar color
		var col rl.Color
		roll := rand.Float64()
		var texType BodyTextureType = TextureMoon
		if roll < 0.20 {
			col = rl.NewColor(130, 190, 255, 255) // Blue giant
			texType = TextureSun
		} else if roll < 0.65 {
			col = rl.NewColor(255, 230, 160, 255) // Yellow-white star
			texType = TextureSun
		} else {
			col = rl.NewColor(255, 120, 80, 255)  // Red dwarf
			texType = TextureSun
		}

		addBodyTex(state, fmt.Sprintf("Star #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, texType, float32(rand.Float64()*2.0))
	}
}

// loadRelativityRosette demonstrates 1PN General Relativity perihelion precession
func loadRelativityRosette(state *SimState, cfg *Config, g float64) {
	cfg.EnableRelativity = true
	cfg.SpeedOfLight = 160.0
	cfg.UseBarnesHut = false
	cfg.NotificationText = "1PN General Relativity active: Rosette perihelion precession!"
	cfg.NotificationTimer = 4.0

	// Dense Central Singularity
	bhMass := 15000.0
	addBodyTex(state, "Kerr Black Hole",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		bhMass, 3.2, rl.NewColor(20, 10, 30, 255), true, true, TextureBlackHole, 0.6)

	// Relativistic Planet on high-eccentricity orbit (periapsis rp = 12.0, apoapsis ra = 42.0)
	// a = (rp + ra)/2 = 27.0, e = (ra - rp)/(ra + rp) = 30/54 = 0.555
	rp := float32(12.0)
	ra := float32(42.0)
	a := (rp + ra) * 0.5
	vPeri := float32(math.Sqrt(g * bhMass * float64((2.0/rp) - (1.0/a))))

	addBodyTex(state, "Relativistic Planet S2",
		rl.NewVector3(rp, 0, 0),
		rl.NewVector3(0, 0, -vPeri),
		1.5, 1.1, rl.NewColor(80, 180, 255, 255), false, false, TextureTerrestrial, 1.5)

	// Second relativistic companion with different inclination to demonstrate 3D precession
	rp2 := float32(18.0)
	ra2 := float32(52.0)
	a2 := (rp2 + ra2) * 0.5
	vPeri2 := float32(math.Sqrt(g * bhMass * float64((2.0/rp2) - (1.0/a2))))

	addBodyTex(state, "Inner Precessor",
		rl.NewVector3(0, 3.0, rp2),
		rl.NewVector3(vPeri2, 0.8, 0),
		0.8, 0.9, rl.NewColor(255, 140, 60, 255), false, false, TextureDesert, 1.0)
}

// loadRocheDisruption demonstrates Roche limit tidal tearing of an incoming moon
func loadRocheDisruption(state *SimState, cfg *Config, g float64) {
	cfg.EnableRocheLimit = true
	cfg.NotificationText = "Roche Disruption active: Incoming moon will shatter at Roche limit!"
	cfg.NotificationTimer = 4.5

	// Gas Giant Kronos
	giantMass := 18000.0
	giant := addBodyTex(state, "Gas Giant Kronos",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		giantMass, 5.2, rl.NewColor(230, 180, 120, 255), true, false, TextureGasGiant, 1.8)

	// Stable distant moon
	dist := float32(65.0)
	speed := float32(math.Sqrt((g * giantMass) / float64(dist)))
	addBodyTex(state, "Stable Moon Titanis",
		rl.NewVector3(dist, 0, 0),
		rl.NewVector3(0, 0, -speed),
		1.5, 0.9, rl.NewColor(200, 210, 230, 255), false, false, TextureMoon, 1.0)

	// Doomed moon on inward plunge heading into Roche limit without hitting surface
	addBodyTex(state, "Doomed Moon Peregrine",
		rl.NewVector3(50.0, 0, 35.0),
		rl.NewVector3(-12.5, 0, 1.5),
		2.2, 1.2, rl.NewColor(220, 240, 255, 255), false, false, TextureMoon, 1.0)

	_ = giant
}

// CreateBodyTemplate creates a ready-to-launch body based on SpawnType
func CreateBodyTemplate(spawnType SpawnType, pos, vel rl.Vector3, id int64) *Body {
	switch spawnType {
	case SpawnEarth:
		return &Body{
			ID:            id,
			Name:          "Terrestrial Planet",
			Position:      pos,
			Velocity:      vel,
			Mass:          2.0,
			Radius:        1.2,
			Color:         rl.NewColor(70, 160, 245, 255),
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureTerrestrial,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 1.2,
		}
	case SpawnMoon:
		return &Body{
			ID:            id,
			Name:          "Moon / Asteroid",
			Position:      pos,
			Velocity:      vel,
			Mass:          0.1,
			Radius:        0.5,
			Color:         rl.NewColor(190, 190, 190, 255),
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureMoon,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 0.8,
		}
	case SpawnGasGiant:
		return &Body{
			ID:            id,
			Name:          "Gas Giant",
			Position:      pos,
			Velocity:      vel,
			Mass:          20.0,
			Radius:        2.3,
			Color:         rl.NewColor(230, 170, 110, 255),
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureGasGiant,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 2.2,
		}
	case SpawnIceGiant:
		return &Body{
			ID:              id,
			Name:            "Ice Giant",
			Position:        pos,
			Velocity:        vel,
			Mass:            9.0,
			Radius:          1.7,
			Color:           rl.NewColor(100, 210, 240, 255),
			IsStationary:    false,
			IsStar:          false,
			RingInnerRadius: 2.4,
			RingOuterRadius: 3.8,
			Trail:           make([]rl.Vector3, 0, 100),
			TextureType:     TextureIceGiant,
			RotationAngle:   rand.Float32() * 360.0,
			RotationSpeed:   1.6,
		}
	case SpawnStar:
		return &Body{
			ID:            id,
			Name:          "Protostar",
			Position:      pos,
			Velocity:      vel,
			Mass:          5000.0,
			Radius:        3.6,
			Color:         rl.Gold,
			IsStationary:  false,
			IsStar:        true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureSun,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 0.5,
		}
	case SpawnRedGiant:
		return &Body{
			ID:            id,
			Name:          "Red Supergiant",
			Position:      pos,
			Velocity:      vel,
			Mass:          10000.0,
			Radius:        5.5,
			Color:         rl.NewColor(255, 75, 40, 255),
			IsStationary:  false,
			IsStar:        true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureRedGiant,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 0.3,
		}
	case SpawnWhiteDwarf:
		return &Body{
			ID:            id,
			Name:          "White Dwarf",
			Position:      pos,
			Velocity:      vel,
			Mass:          4500.0,
			Radius:        0.9,
			Color:         rl.NewColor(230, 245, 255, 255),
			IsStationary:  false,
			IsStar:        true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureWhiteDwarf,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 3.0,
		}
	case SpawnNeutronStar:
		return &Body{
			ID:            id,
			Name:          "Relativistic Pulsar",
			Position:      pos,
			Velocity:      vel,
			Mass:          6000.0,
			Radius:        1.1,
			Color:         rl.NewColor(150, 210, 255, 255),
			IsStationary:  false,
			IsStar:        true,
			IsPulsar:      true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureNeutronStar,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 6.5,
		}
	case SpawnBlackHole:
		return &Body{
			ID:            id,
			Name:          "Micro Black Hole",
			Position:      pos,
			Velocity:      vel,
			Mass:          15000.0,
			Radius:        2.0,
			Color:         rl.NewColor(40, 20, 60, 255),
			IsStationary:  true,
			IsStar:        true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureBlackHole,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 0.8,
		}
	case SpawnDwarfPlanet:
		return &Body{
			ID:            id,
			Name:          "Dwarf Planet",
			Position:      pos,
			Velocity:      vel,
			Mass:          0.06,
			Radius:        0.55,
			Color:         rl.NewColor(215, 185, 160, 255),
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureDwarfPlanet,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 0.9,
		}
	case SpawnComet:
		return &Body{
			ID:            id,
			Name:          "Icy Comet",
			Position:      pos,
			Velocity:      vel,
			Mass:          0.02,
			Radius:        0.45,
			Color:         rl.NewColor(200, 240, 255, 255),
			IsStationary:  false,
			IsStar:        false,
			IsComet:       true,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureComet,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 1.1,
		}
	case SpawnAsteroidRing:
		return &Body{
			ID:            id,
			Name:          "Asteroid Swarm",
			Position:      pos,
			Velocity:      vel,
			Mass:          0.05,
			Radius:        0.4,
			Color:         rl.NewColor(180, 175, 165, 255),
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureMoon,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 1.0,
		}
	default:
		return &Body{
			ID:            id,
			Name:          "New Body",
			Position:      pos,
			Velocity:      vel,
			Mass:          1.0,
			Radius:        1.0,
			Color:         rl.White,
			IsStationary:  false,
			IsStar:        false,
			Trail:         make([]rl.Vector3, 0, 100),
			TextureType:   TextureNone,
			RotationAngle: rand.Float32() * 360.0,
			RotationSpeed: 1.0,
		}
	}
}

// loadSolarSystemGrand populates full Solar System with planets, asteroid belt, moons & comets
func loadSolarSystemGrand(state *SimState, g float64) {
	sunMass := 14000.0
	addBodyTex(state, "Sun",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		sunMass, 4.6, rl.Gold, true, true, TextureSun, 0.4)

	orbitSun := func(dist float32, mass float64, radius float32, col rl.Color, name string, texType BodyTextureType, rotSpeed float32) *Body {
		speed := float32(math.Sqrt((g * sunMass) / float64(dist)))
		return addBodyTex(state, name,
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -speed),
			mass, radius, col, false, false, texType, rotSpeed)
	}

	// Terrestrial Planets
	orbitSun(12.0, 0.4, 0.65, rl.NewColor(170, 170, 170, 255), "Mercury", TextureMoon, 0.5)
	orbitSun(19.0, 1.2, 0.95, rl.NewColor(225, 200, 140, 255), "Venus", TextureDesert, -0.4)

	// Earth & Moon (Hierarchical two-body system, stable inside Hill sphere)
	earthDist := float32(28.0)
	earthSpeed := float32(math.Sqrt((g * sunMass) / float64(earthDist)))
	earthMass := 28.0
	earth := addBodyTex(state, "Earth",
		rl.NewVector3(earthDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed),
		earthMass, 1.15, rl.NewColor(70, 150, 255, 255), false, false, TextureTerrestrial, 1.2)

	moonDist := float32(1.6)
	moonSpeed := float32(math.Sqrt((g * earth.Mass) / float64(moonDist)))
	addBodyTex(state, "Moon",
		rl.NewVector3(earthDist+moonDist, 0, 0),
		rl.NewVector3(0, 0, -earthSpeed-moonSpeed),
		0.08, 0.38, rl.NewColor(200, 200, 200, 255), false, false, TextureMoon, 0.8)

	// Mars & Phobos (stable inside Hill sphere rH ~ 2.5)
	marsDist := float32(38.0)
	marsSpeed := float32(math.Sqrt((g * sunMass) / float64(marsDist)))
	marsMass := 12.0
	mars := addBodyTex(state, "Mars",
		rl.NewVector3(marsDist, 0, 0),
		rl.NewVector3(0, 0, -marsSpeed),
		marsMass, 0.8, rl.NewColor(230, 80, 45, 255), false, false, TextureDesert, 1.1)

	addBodyTex(state, "Phobos",
		rl.NewVector3(marsDist+1.1, 0, 0),
		rl.NewVector3(0, 0, -marsSpeed-float32(math.Sqrt((g*mars.Mass)/1.1))),
		0.005, 0.22, rl.NewColor(180, 170, 160, 255), false, false, TextureMoon, 1.5)

	// Main Asteroid Belt (25 belt asteroids + Ceres & Vesta in PROGRADE coplanar orbits)
	orbitSun(43.5, 0.05, 0.45, rl.NewColor(190, 185, 175, 255), "Ceres (Dwarf Planet)", TextureDwarfPlanet, 0.9)
	orbitSun(46.0, 0.03, 0.38, rl.NewColor(175, 165, 150, 255), "Vesta", TextureMoon, 0.8)

	for i := 0; i < 22; i++ {
		ang := float64(i) * (2.0 * math.Pi / 22.0)
		r := float32(42.0 + float64(i%5)*1.8)
		spd := float32(math.Sqrt((g * sunMass) / float64(r)))
		x := r * float32(math.Cos(ang))
		z := r * float32(math.Sin(ang))
		vx := spd * float32(math.Sin(ang))
		vz := -spd * float32(math.Cos(ang))

		addBodyTex(state, fmt.Sprintf("Belt Asteroid %d", i+1),
			rl.NewVector3(x, float32((i%3)-1)*0.2, z),
			rl.NewVector3(vx, 0, vz),
			0.002, 0.2, rl.NewColor(180, 175, 170, 220), false, false, TextureMoon, 1.0)
	}

	// Jupiter & 4 Galilean Moons (stable inside Hill sphere rH ~ 9.17)
	jupDist := float32(60.0)
	jupSpeed := float32(math.Sqrt((g * sunMass) / float64(jupDist)))
	jupMass := 150.0
	jup := addBodyTex(state, "Jupiter",
		rl.NewVector3(jupDist, 0, 0),
		rl.NewVector3(0, 0, -jupSpeed),
		jupMass, 2.6, rl.NewColor(220, 175, 125, 255), false, false, TextureGasGiant, 2.5)

	moons := []struct {
		name string
		dist float32
		rad  float32
		col  rl.Color
	}{
		{"Io", 2.8, 0.32, rl.NewColor(240, 220, 70, 255)},
		{"Europa", 4.0, 0.30, rl.NewColor(210, 235, 255, 255)},
		{"Ganymede", 5.5, 0.42, rl.NewColor(190, 200, 215, 255)},
		{"Callisto", 7.5, 0.40, rl.NewColor(160, 170, 185, 255)},
	}
	for _, m := range moons {
		mSpd := float32(math.Sqrt((g * jup.Mass) / float64(m.dist)))
		addBodyTex(state, m.name,
			rl.NewVector3(jupDist+m.dist, 0, 0),
			rl.NewVector3(0, 0, -jupSpeed-mSpd),
			0.02, m.rad, m.col, false, false, TextureMoon, 1.0)
	}

	// Saturn & Rings + Titan (stable inside Hill sphere rH ~ 9.0)
	satDist := float32(88.0)
	satSpeed := float32(math.Sqrt((g * sunMass) / float64(satDist)))
	satMass := 45.0
	sat := addBodyTex(state, "Saturn",
		rl.NewVector3(satDist, 0, 0),
		rl.NewVector3(0, 0, -satSpeed),
		satMass, 2.2, rl.NewColor(235, 210, 160, 255), false, false, TextureGasGiant, 2.0)
	sat.RingInnerRadius = 3.0
	sat.RingOuterRadius = 5.4

	titanDist := float32(5.0)
	titanSpeed := float32(math.Sqrt((g * sat.Mass) / float64(titanDist)))
	addBodyTex(state, "Titan",
		rl.NewVector3(satDist+titanDist, 0, 0),
		rl.NewVector3(0, 0, -satSpeed-titanSpeed),
		0.04, 0.42, rl.NewColor(240, 200, 120, 255), false, false, TextureMoon, 0.9)

	// Uranus & Neptune
	orbitSun(120.0, 8.5, 1.7, rl.NewColor(140, 220, 235, 255), "Uranus", TextureIceGiant, 1.5)

	nepDist := float32(150.0)
	nepSpeed := float32(math.Sqrt((g * sunMass) / float64(nepDist)))
	nep := addBodyTex(state, "Neptune",
		rl.NewVector3(nepDist, 0, 0),
		rl.NewVector3(0, 0, -nepSpeed),
		9.0, 1.7, rl.NewColor(75, 110, 245, 255), false, false, TextureIceGiant, 1.5)

	// Triton (Retrograde orbit)
	tritDist := float32(4.5)
	tritSpeed := float32(math.Sqrt((g * nep.Mass) / float64(tritDist)))
	addBodyTex(state, "Triton (Retrograde)",
		rl.NewVector3(nepDist+tritDist, 0, 0),
		rl.NewVector3(0, 0, -nepSpeed+tritSpeed), // + retrograde
		0.03, 0.38, rl.NewColor(200, 220, 240, 255), false, false, TextureMoon, -1.0)

	// Pluto & Charon
	plutoDist := float32(185.0)
	plutoSpeed := float32(math.Sqrt((g * sunMass) / float64(plutoDist)))
	pluto := addBodyTex(state, "Pluto (Dwarf Planet)",
		rl.NewVector3(plutoDist, 5.0, 0),
		rl.NewVector3(0, 0, -plutoSpeed),
		0.05, 0.5, rl.NewColor(220, 190, 165, 255), false, false, TextureDwarfPlanet, 0.8)

	charDist := float32(1.8)
	charSpeed := float32(math.Sqrt((g * pluto.Mass) / float64(charDist)))
	addBodyTex(state, "Charon",
		rl.NewVector3(plutoDist+charDist, 5.0, 0),
		rl.NewVector3(0, 0, -plutoSpeed-charSpeed),
		0.01, 0.28, rl.NewColor(170, 165, 160, 255), false, false, TextureMoon, 0.8)

	// Comet Halley
	comet := addBodyTex(state, "Comet Halley",
		rl.NewVector3(120.0, 20.0, 45.0),
		rl.NewVector3(-3.8, -1.5, -4.8),
		0.015, 0.45, rl.NewColor(190, 240, 255, 255), false, false, TextureComet, 0.6)
	comet.IsComet = true
}

// loadLagrangeTrojans demonstrates stable 3-body circular restricted equilibrium (CR3BP)
// Features exact barycentric orbits, analytical L1-L5 points, and authentic tadpole/horseshoe libration swarms
func loadLagrangeTrojans(state *SimState, cfg *Config, g float64) {
	cfg.NotificationText = "Lagrange Equilibrium (CR3BP): All 5 points (L1-L5) with Trojan & Greek libration!"
	cfg.NotificationTimer = 5.5
	cfg.Softening = 0.005 // Minimal softening for true 1/r^2 Newtonian physics
	cfg.SubSteps = 6
	cfg.Integrator = IntegratorYoshida4
	cfg.ShowLagrangePoints = true

	m1 := 10000.0 // Sun
	m2 := 100.0   // Zeus Gas Giant (mu = 100 / 10100 ~= 0.0099 < 0.03852 Routh limit)
	totalM := m1 + m2
	mu := m2 / totalM
	R := float64(75.0)

	// Barycentric positions and velocities
	omega := math.Sqrt((g * totalM) / (R * R * R))
	x1 := float32(-mu * R)
	x2 := float32((1.0 - mu) * R)
	v1z := float32(float64(x1) * omega)
	v2z := float32(float64(x2) * omega)

	sun := addBodyTex(state, "Sun Helios",
		rl.NewVector3(x1, 0, 0),
		rl.NewVector3(0, 0, -v1z), // Momentum balanced at barycenter (0,0,0)
		m1, 4.0, rl.Gold, false, true, TextureSun, 0.4)

	zeus := addBodyTex(state, "Gas Giant Zeus",
		rl.NewVector3(x2, 0, 0),
		rl.NewVector3(0, 0, -v2z),
		m2, 2.4, rl.NewColor(225, 165, 110, 255), false, false, TextureGasGiant, 2.0)

	// Calculate exact Lagrange points
	pts := ComputeLagrangePoints(sun, zeus, g)

	// Place probe satellites at all 5 equilibrium points
	addBodyTex(state, "L1 SOHO Solar Sentinel",
		pts[0].Position, pts[0].Velocity,
		0.001, 0.4, rl.NewColor(255, 220, 80, 255), false, false, TextureMoon, 1.0)

	addBodyTex(state, "L2 JWST Deep Observatory",
		pts[1].Position, pts[1].Velocity,
		0.001, 0.4, rl.NewColor(100, 220, 255, 255), false, false, TextureMoon, 1.0)

	addBodyTex(state, "L3 Vulcan Counter-Orbital",
		pts[2].Position, pts[2].Velocity,
		0.001, 0.4, rl.NewColor(255, 120, 100, 255), false, false, TextureMoon, 1.0)

	addBodyTex(state, "L4 Achilles Prime (Trojan Lead)",
		pts[3].Position, pts[3].Velocity,
		0.01, 0.45, rl.NewColor(130, 210, 255, 255), false, false, TextureMoon, 1.0)

	addBodyTex(state, "L5 Patroclus Prime (Greek Trail)",
		pts[4].Position, pts[4].Velocity,
		0.01, 0.45, rl.NewColor(255, 180, 120, 255), false, false, TextureMoon, 1.0)

	// L4 Trojan Swarm: 16 asteroids with varied libration amplitudes (stable tadpole orbits)
	for i := 0; i < 16; i++ {
		offsetAngle := (float64(i%4) - 1.5) * 0.04
		offsetR := (float64(i/4) - 1.5) * 1.4
		ang := -math.Pi/3.0 + offsetAngle
		r := R + offsetR
		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		vx := float32(-float64(z) * omega)
		vz := float32(float64(x) * omega)

		addBodyTex(state, fmt.Sprintf("Trojan #%d (L4 Tadpole)", i+1),
			rl.NewVector3(x, float32((i%3)-1)*0.2, z),
			rl.NewVector3(vx, 0, vz),
			0.005, 0.35, rl.NewColor(140, 200, 255, 255), false, false, TextureMoon, 1.0)
	}

	// L5 Greek Swarm: 16 asteroids with varied libration amplitudes (stable tadpole orbits)
	for i := 0; i < 16; i++ {
		offsetAngle := (float64(i%4) - 1.5) * 0.04
		offsetR := (float64(i/4) - 1.5) * 1.4
		ang := math.Pi/3.0 + offsetAngle
		r := R + offsetR
		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		vx := float32(-float64(z) * omega)
		vz := float32(float64(x) * omega)

		addBodyTex(state, fmt.Sprintf("Greek #%d (L5 Tadpole)", i+1),
			rl.NewVector3(x, float32((i%3)-1)*0.2, z),
			rl.NewVector3(vx, 0, vz),
			0.005, 0.35, rl.NewColor(255, 190, 130, 255), false, false, TextureMoon, 1.0)
	}

	// Horseshoe Libration Asteroid (Cruithne-type co-orbital transitioning around L4, L3, L5)
	horseR := R + 1.2
	horseSpd := float32(math.Sqrt((g * m1) / horseR))
	addBodyTex(state, "Cruithne (Horseshoe Libration)",
		rl.NewVector3(float32(horseR*math.Cos(math.Pi*0.9)), 0, float32(horseR*math.Sin(math.Pi*0.9))),
		rl.NewVector3(-horseSpd*float32(math.Sin(math.Pi*0.9)), 0, horseSpd*float32(math.Cos(math.Pi*0.9))),
		0.002, 0.38, rl.NewColor(210, 255, 180, 255), false, false, TextureMoon, 1.0)
}

// loadRealScaleSolarSystem initializes the solar system using true astronomical Astronomical Unit (AU) scale
// 1 AU = 100.0 units, G = 1.0, M_Sun = 10,000.0, v_earth = 10.0 units/s, T_earth = 62.83 s
func loadRealScaleSolarSystem(state *SimState, cfg *Config, g float64) {
	cfg.NotificationText = "Real Scale Solar System (Astronomical AU): 1 AU = 100 units. True Keplerian distances & periods!"
	cfg.NotificationTimer = 6.0
	cfg.Softening = 0.01
	cfg.SubSteps = 5
	cfg.Integrator = IntegratorYoshida4
	cfg.Show2DViewport = true // Enable 2D tactical map by default for large-scale solar system

	sunMass := 10000.0
	sunRad := float32(4.5)
	if cfg.RealScaleVisualMode {
		sunRad = 0.465 // True physical 1:1 radius (0.00465 AU * 100)
	}

	sun := addBodyTex(state, "Sun Helios",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		sunMass, sunRad, rl.Gold, true, true, TextureSun, 0.4)
	_ = sun

	// Helper to spawn Keplerian elliptical orbit around Sun
	spawnPlanet := func(name string, aAU, e, iDeg float64, mass float64, baseRadius float32, col rl.Color, tex BodyTextureType, rotSpd float32, trueAnomaly float64) *Body {
		a := aAU * 100.0
		theta := trueAnomaly
		r := (a * (1.0 - e*e)) / (1.0 + e*math.Cos(theta))
		h := math.Sqrt(g * sunMass * a * (1.0 - e*e))
		vr := (g * sunMass / h) * e * math.Sin(theta)
		vtheta := (g * sunMass / h) * (1.0 + e*math.Cos(theta))

		// In-plane velocity
		vxPlane := vr*math.Cos(theta) - vtheta*math.Sin(theta)
		vzPlane := vr*math.Sin(theta) + vtheta*math.Cos(theta)

		// Apply inclination iDeg
		iRad := iDeg * math.Pi / 180.0
		x := float32(r * math.Cos(theta))
		y := float32(r * math.Sin(theta) * math.Sin(iRad))
		z := float32(r * math.Sin(theta) * math.Cos(iRad))

		vx := float32(vxPlane)
		vy := float32(vzPlane * math.Sin(iRad))
		vz := float32(vzPlane * math.Cos(iRad))

		rad := baseRadius
		if cfg.RealScaleVisualMode {
			rad = baseRadius * 0.08
			if rad < 0.05 {
				rad = 0.05
			}
		}

		return addBodyTex(state, name,
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, tex, rotSpd)
	}

	// 1. Mercury (a = 0.3871 AU, e = 0.2056, i = 7.00 deg)
	spawnPlanet("Mercury", 0.3871, 0.2056, 7.00, 0.00166, 0.65, rl.NewColor(180, 180, 180, 255), TextureMoon, 0.3, 0.8)

	// 2. Venus (a = 0.7233 AU, e = 0.0068, i = 3.39 deg)
	spawnPlanet("Venus", 0.7233, 0.0068, 3.39, 0.0245, 1.1, rl.NewColor(230, 205, 150, 255), TextureDesert, -0.2, 2.1)

	// 3. Earth (a = 1.0000 AU, e = 0.0167, i = 0.00 deg)
	earth := spawnPlanet("Earth", 1.0000, 0.0167, 0.00, 0.0300, 1.2, rl.NewColor(60, 150, 245, 255), TextureTerrestrial, 1.2, 0.0)

	// Earth-Moon system (Luna at realistic Hill sphere orbit)
	moonDist := float32(2.5) // Enhanced visible separation (true is 0.257 AU*100 = 0.257)
	if cfg.RealScaleVisualMode {
		moonDist = 0.35
	}
	moonV := float32(math.Sqrt((g * earth.Mass) / float64(moonDist)))
	addBodyTex(state, "Moon Luna",
		rl.NewVector3(earth.Position.X+moonDist, earth.Position.Y, earth.Position.Z),
		rl.NewVector3(earth.Velocity.X, earth.Velocity.Y, earth.Velocity.Z-moonV),
		0.000369, 0.45, rl.LightGray, false, false, TextureMoon, 1.0)

	// Sun-Earth L1 (SOHO) and L2 (JWST)
	rL1Offset := float32(1.0)
	addBodyTex(state, "JWST (Earth L2 Halo)",
		rl.NewVector3(earth.Position.X+rL1Offset*1.8, earth.Position.Y, earth.Position.Z),
		rl.NewVector3(earth.Velocity.X, earth.Velocity.Y, earth.Velocity.Z-moonV*0.25),
		0.00001, 0.3, rl.NewColor(120, 240, 255, 255), false, false, TextureMoon, 1.0)

	// 4. Mars (a = 1.5237 AU, e = 0.0934, i = 1.85 deg)
	spawnPlanet("Mars", 1.5237, 0.0934, 1.85, 0.00323, 0.75, rl.NewColor(235, 95, 55, 255), TextureDesert, 1.0, 3.4)

	// 5. Main Asteroid Belt (2.1 to 3.3 AU) with dwarf planets Ceres, Vesta, Pallas
	spawnPlanet("Ceres (Dwarf Planet)", 2.767, 0.0758, 10.59, 0.000047, 0.5, rl.NewColor(160, 160, 170, 255), TextureMoon, 1.5, 1.2)
	spawnPlanet("Vesta", 2.362, 0.0887, 7.14, 0.000013, 0.4, rl.NewColor(200, 190, 180, 255), TextureMoon, 2.0, 4.5)
	spawnPlanet("Pallas", 2.772, 0.2307, 34.84, 0.000011, 0.4, rl.NewColor(150, 180, 190, 255), TextureMoon, 1.8, 2.9)

	rSource := rand.New(rand.NewSource(42))
	for i := 0; i < 45; i++ {
		aAst := 2.15 + rSource.Float64()*1.15
		if (aAst > 2.47 && aAst < 2.53) || (aAst > 2.80 && aAst < 2.84) {
			aAst += 0.07
		}
		eAst := rSource.Float64() * 0.18
		iAst := (rSource.Float64() - 0.5) * 16.0
		thetaAst := rSource.Float64() * 2.0 * math.Pi
		spawnPlanet(fmt.Sprintf("Belt Asteroid #%d", i+1), aAst, eAst, iAst, 0.000005, 0.28, rl.NewColor(uint8(140+rSource.Intn(60)), uint8(130+rSource.Intn(50)), uint8(120+rSource.Intn(40)), 255), TextureMoon, 1.0, thetaAst)
	}

	// 6. Jupiter (a = 5.2044 AU -> 520.44, e = 0.0484, i = 1.30 deg, Mass = 9.548)
	jup := spawnPlanet("Jupiter", 5.2044, 0.0484, 1.30, 9.548, 3.2, rl.NewColor(230, 175, 120, 255), TextureGasGiant, 2.4, 5.1)

	// Jupiter Galilean Moons (Io, Europa, Ganymede, Callisto)
	galMoons := []struct {
		name string
		dist float32
		mass float64
		col  rl.Color
	}{
		{"Io", 5.2, 0.00045, rl.NewColor(255, 230, 80, 255)},
		{"Europa", 7.8, 0.00024, rl.NewColor(210, 230, 255, 255)},
		{"Ganymede", 11.5, 0.00075, rl.NewColor(180, 170, 160, 255)},
		{"Callisto", 18.0, 0.00054, rl.NewColor(130, 130, 140, 255)},
	}
	for _, gm := range galMoons {
		spd := float32(math.Sqrt((g * jup.Mass) / float64(gm.dist)))
		addBodyTex(state, gm.name,
			rl.NewVector3(jup.Position.X+gm.dist, jup.Position.Y, jup.Position.Z),
			rl.NewVector3(jup.Velocity.X, jup.Velocity.Y, jup.Velocity.Z-spd),
			gm.mass, 0.45, gm.col, false, false, TextureMoon, 1.0)
	}

	// Jupiter Trojans at L4 and L5 (5.2044 AU at +/- 60 deg)
	jupPTS := ComputeLagrangePoints(sun, jup, g)
	addBodyTex(state, "Hector (Jupiter L4 Trojan)",
		jupPTS[3].Position, jupPTS[3].Velocity,
		0.0001, 0.4, rl.NewColor(130, 210, 255, 255), false, false, TextureMoon, 1.0)
	addBodyTex(state, "Nestor (Jupiter L5 Greek)",
		jupPTS[4].Position, jupPTS[4].Velocity,
		0.0001, 0.4, rl.NewColor(255, 180, 130, 255), false, false, TextureMoon, 1.0)

	// 7. Saturn (a = 9.5826 AU -> 958.26, e = 0.0541, i = 2.48 deg, Mass = 2.858)
	sat := spawnPlanet("Saturn", 9.5826, 0.0541, 2.48, 2.858, 2.6, rl.NewColor(240, 215, 160, 255), TextureGasGiant, 2.2, 1.7)
	sat.RingInnerRadius = 3.6
	sat.RingOuterRadius = 6.8
	// Titan moon
	titanSpd := float32(math.Sqrt((g * sat.Mass) / 12.0))
	addBodyTex(state, "Titan",
		rl.NewVector3(sat.Position.X+12.0, sat.Position.Y, sat.Position.Z),
		rl.NewVector3(sat.Velocity.X, sat.Velocity.Y, sat.Velocity.Z-titanSpd),
		0.00067, 0.6, rl.NewColor(245, 190, 80, 255), false, false, TextureTerrestrial, 1.0)

	// 8. Uranus (a = 19.2184 AU -> 1921.84, e = 0.0472, i = 0.77 deg, Mass = 0.436)
	spawnPlanet("Uranus", 19.2184, 0.0472, 0.77, 0.436, 1.8, rl.NewColor(140, 230, 240, 255), TextureIceGiant, -1.8, 3.8)

	// 9. Neptune (a = 30.0700 AU -> 3007.00, e = 0.0086, i = 1.77 deg, Mass = 0.515)
	nep := spawnPlanet("Neptune", 30.0700, 0.0086, 1.77, 0.515, 1.8, rl.NewColor(70, 120, 255, 255), TextureIceGiant, 1.7, 0.4)
	// Triton moon (retrograde orbit)
	tritonSpd := float32(math.Sqrt((g * nep.Mass) / 7.5))
	addBodyTex(state, "Triton (Retrograde Moon)",
		rl.NewVector3(nep.Position.X+7.5, nep.Position.Y, nep.Position.Z),
		rl.NewVector3(nep.Velocity.X, nep.Velocity.Y, nep.Velocity.Z+tritonSpd),
		0.00011, 0.45, rl.NewColor(200, 220, 235, 255), false, false, TextureMoon, 1.0)

	// 10. Pluto (a = 39.482 AU -> 3948.20, e = 0.2488, i = 17.16 deg, Mass = 0.000065)
	spawnPlanet("Pluto (Kuiper Belt)", 39.482, 0.2488, 17.16, 0.000065, 0.5, rl.NewColor(210, 185, 160, 255), TextureMoon, 0.5, 4.9)

	// 11. Halley's Comet (a = 17.83 AU -> 1783.0, e = 0.9671, i = 17.8 deg)
	comet := spawnPlanet("Halley's Comet (1P/Halley)", 17.83, 0.9671, 17.8, 0.000001, 0.4, rl.White, TextureMoon, 1.0, 0.15)
	comet.IsComet = true

	// 12. Interstellar Voyager 1 Probe
	vEsc := float32(math.Sqrt((2.0 * g * sunMass) / 4500.0) * 1.35)
	addBodyTex(state, "Voyager 1 (Interstellar Probe)",
		rl.NewVector3(4500.0, 180.0, 1200.0),
		rl.NewVector3(vEsc*0.9, vEsc*0.35, vEsc*0.25),
		0.000001, 0.35, rl.Gold, false, false, TextureMoon, 1.0)
}

// loadGalaxyCollision demonstrates two colliding spiral galaxies with tidal tails
func loadGalaxyCollision(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.7
	cfg.NotificationText = "Galaxy Collision: Mutual gravity rips tidal bridges and tails!"
	cfg.NotificationTimer = 4.5

	// Galaxy 1: Andromeda analog (starts as clean isolated rotating disk)
	smbh1Mass := 22000.0
	g1Center := rl.NewVector3(-65.0, 0, -35.0)
	g1Vel := rl.NewVector3(1.6, 0.2, 0.9)

	addBodyTex(state, "SMBH Andromeda",
		g1Center, g1Vel,
		smbh1Mass, 3.8, rl.NewColor(30, 20, 55, 255), false, true, TextureBlackHole, 0.5)

	for i := 0; i < 50; i++ {
		r := 6.0 + float64(i)*0.55
		ang := float64(i) * 1.3
		spd := float32(math.Sqrt((g * smbh1Mass) / r))

		x := g1Center.X + float32(r*math.Cos(ang))
		z := g1Center.Z + float32(r*math.Sin(ang))
		y := g1Center.Y + float32((i%3)-1)*0.3

		vx := g1Vel.X - spd*float32(math.Sin(ang))
		vz := g1Vel.Z + spd*float32(math.Cos(ang))

		addBodyTex(state, fmt.Sprintf("Andromeda Star %d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, g1Vel.Y, vz),
			0.8, 0.45, rl.NewColor(150, 210, 255, 255), false, false, TextureMoon, 1.0)
	}

	// Galaxy 2: Milky Way analog (inclined, starts as clean isolated rotating disk)
	smbh2Mass := 18000.0
	g2Center := rl.NewVector3(65.0, 12.0, 35.0)
	g2Vel := rl.NewVector3(-1.6, -0.2, -0.9)

	addBodyTex(state, "SMBH Milky Way",
		g2Center, g2Vel,
		smbh2Mass, 3.5, rl.NewColor(20, 30, 65, 255), false, true, TextureBlackHole, 0.5)

	for i := 0; i < 50; i++ {
		r := 6.0 + float64(i)*0.55
		ang := float64(i) * 1.4
		spd := float32(math.Sqrt((g * smbh2Mass) / r))

		// 30 degree inclination
		tilt := 0.5
		x := g2Center.X + float32(r*math.Cos(ang))
		z := g2Center.Z + float32(r*math.Sin(ang)*math.Cos(tilt))
		y := g2Center.Y + float32(r*math.Sin(ang)*math.Sin(tilt))

		vx := g2Vel.X + spd*float32(math.Sin(ang))
		vz := g2Vel.Z - spd*float32(math.Cos(ang)*math.Cos(tilt))
		vy := g2Vel.Y - spd*float32(math.Cos(ang)*math.Sin(tilt))

		addBodyTex(state, fmt.Sprintf("Milky Way Star %d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			0.7, 0.42, rl.NewColor(255, 220, 160, 255), false, false, TextureMoon, 1.0)
	}
}

// loadPulsarAccretion demonstrates a millisecond pulsar with polar jets in binary orbit with a red giant
func loadPulsarAccretion(state *SimState, cfg *Config, g float64) {
	cfg.EnableRelativity = true
	cfg.SpeedOfLight = 160.0
	cfg.NotificationText = "Relativistic Pulsar Binary: Polar jet beams & mass-transfer accretion stream!"
	cfg.NotificationTimer = 4.5

	pulsarMass := 6000.0
	redGiantMass := 8000.0
	sep := float32(32.0)

	// Mutual circular orbit
	totalM := pulsarMass + redGiantMass
	omega := float32(math.Sqrt((g * totalM) / math.Pow(float64(sep), 3)))

	r1 := sep * float32(redGiantMass/totalM)
	r2 := sep * float32(pulsarMass/totalM)

	v1 := r1 * omega
	v2 := r2 * omega

	pulsar := addBodyTex(state, "Pulsar PSR-B1919",
		rl.NewVector3(-r1, 0, 0),
		rl.NewVector3(0, 0, v1),
		pulsarMass, 1.4, rl.NewColor(140, 210, 255, 255), false, true, TextureNeutronStar, 8.0)
	pulsar.IsPulsar = true

	addBodyTex(state, "Red Supergiant Mira",
		rl.NewVector3(r2, 0, 0),
		rl.NewVector3(0, 0, -v2),
		redGiantMass, 4.6, rl.NewColor(255, 70, 35, 255), false, true, TextureRedGiant, 0.4)

	// Accretion gas stream pulled toward pulsar through Roche lobe overflow (L1)
	for i := 1; i <= 8; i++ {
		t := 0.22 + float32(i)*0.075
		posX := r2 - t*(r1+r2)
		posZ := float32(math.Sin(float64(t)*math.Pi) * 3.8)
		velZ := -v2 + t*(v1+v2)

		addBodyTex(state, fmt.Sprintf("Accretion Clump %d", i),
			rl.NewVector3(posX, 0, posZ),
			rl.NewVector3(-1.5, 0, velZ),
			0.08, 0.32, rl.NewColor(255, 180, 90, 255), false, false, TextureMoon, 1.0)
	}
}

// loadGlobularCluster demonstrates 180 stars dynamically relaxing and core collapse
func loadGlobularCluster(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.7
	cfg.NotificationText = "Globular Cluster: 180 self-gravitating stars showing dynamic core relaxation!"
	cfg.NotificationTimer = 4.5

	starCount := 180
	clusterRadius := 45.0

	for i := 0; i < starCount; i++ {
		// Plummer sphere density distribution
		r := clusterRadius * math.Pow(rand.Float64(), 0.75) / math.Sqrt(1.0-math.Pow(rand.Float64(), 0.66)+0.08)
		if r > clusterRadius*1.4 {
			r = clusterRadius * 1.4
		}
		theta := rand.Float64() * 2.0 * math.Pi
		phi := math.Acos(2.0*rand.Float64() - 1.0)

		x := float32(r * math.Sin(phi) * math.Cos(theta))
		y := float32(r * math.Sin(phi) * math.Sin(theta))
		z := float32(r * math.Cos(phi))

		// Virialized isotropic velocity dispersion
		vDisp := float32(math.Sqrt(g*float64(starCount)*1.5/(r+10.0))) * 0.75
		vx := (rand.Float32() - 0.5) * vDisp * 2.0
		vy := (rand.Float32() - 0.5) * vDisp * 2.0
		vz := (rand.Float32() - 0.5) * vDisp * 2.0

		mass := 1.0 + rand.Float64()*4.0
		rad := float32(0.35 + mass*0.1)

		var col rl.Color
		var tex BodyTextureType = TextureSun
		roll := rand.Float64()
		if roll < 0.15 {
			col = rl.NewColor(130, 200, 255, 255)
			tex = TextureWhiteDwarf
		} else if roll < 0.35 {
			col = rl.NewColor(255, 100, 50, 255)
			tex = TextureRedGiant
		} else {
			col = rl.NewColor(255, 230, 160, 255)
			tex = TextureSun
		}

		addBodyTex(state, fmt.Sprintf("Cluster Star %d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, true, tex, float32(rand.Float64()*2.0))
	}
}

// loadTrappistResonance populates the TRAPPIST-1 red dwarf system with 7 resonant planets
func loadTrappistResonance(state *SimState, cfg *Config, g float64) {
	cfg.NotificationText = "TRAPPIST-1: 7 resonant terrestrial exoplanets in harmonic resonant chain!"
	cfg.NotificationTimer = 4.5

	starMass := 8000.0
	addBodyTex(state, "TRAPPIST-1 (Ultra-Cool Dwarf)",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		starMass, 3.6, rl.NewColor(255, 85, 35, 255), true, true, TextureSun, 0.4)

	planets := []struct {
		name string
		dist float32
		mass float64
		rad  float32
		col  rl.Color
		tex  BodyTextureType
	}{
		{"TRAPPIST-1b", 11.0, 1.02, 0.95, rl.NewColor(210, 160, 120, 255), TextureDesert},
		{"TRAPPIST-1c", 15.2, 1.16, 0.98, rl.NewColor(180, 190, 210, 255), TextureTerrestrial},
		{"TRAPPIST-1d", 21.5, 0.30, 0.72, rl.NewColor(160, 210, 240, 255), TextureTerrestrial},
		{"TRAPPIST-1e (Habitable)", 28.0, 0.77, 0.88, rl.NewColor(70, 170, 255, 255), TextureTerrestrial},
		{"TRAPPIST-1f", 36.0, 0.93, 0.94, rl.NewColor(140, 200, 220, 255), TextureTerrestrial},
		{"TRAPPIST-1g", 45.0, 1.15, 1.02, rl.NewColor(130, 160, 200, 255), TextureIceGiant},
		{"TRAPPIST-1h", 58.0, 0.33, 0.74, rl.NewColor(190, 200, 220, 255), TextureMoon},
	}

	for i, p := range planets {
		ang := float64(i) * (2.0 * math.Pi / float64(len(planets)))
		spd := float32(math.Sqrt((g * starMass) / float64(p.dist)))
		x := p.dist * float32(math.Cos(ang))
		z := p.dist * float32(math.Sin(ang))
		vx := spd * float32(math.Sin(ang))
		vz := -spd * float32(math.Cos(ang))

		addBodyTex(state, p.name,
			rl.NewVector3(x, 0, z),
			rl.NewVector3(vx, 0, vz),
			p.mass, p.rad, p.col, false, false, p.tex, 1.0)
	}
}

// loadFiveBody demonstrates 5-body symmetric choreographic dance
func loadFiveBody(state *SimState, g float64) {
	m := 2500.0
	r := float32(26.0)
	const n = 5
	colors := []rl.Color{rl.Gold, rl.SkyBlue, rl.Lime, rl.Magenta, rl.Orange}

	// Exact regular 5-body Lagrange equilibrium velocity factor:
	// S_5 = 1/4 * sum_{k=1..4} 1 / sin(k * pi / 5) = 1.37638192
	const s5 = 1.37638192
	spd := float32(math.Sqrt((g * m * s5) / float64(r)))

	for i := 0; i < n; i++ {
		ang := float64(i) * (2.0 * math.Pi / float64(n))
		x := r * float32(math.Cos(ang))
		z := r * float32(math.Sin(ang))

		vx := -spd * float32(math.Sin(ang))
		vz := spd * float32(math.Cos(ang))

		addBodyTex(state, fmt.Sprintf("Choreography Body %c", 'A'+rune(i)),
			rl.NewVector3(x, 0, z),
			rl.NewVector3(vx, 0, vz),
			m, 2.2, colors[i], false, true, TextureSun, 0.8)
	}
}

// loadMilkyWay10K generates a full-scale 10,000-particle galaxy with Sagittarius A*, galactic bulge, and 4 spiral arms
func loadMilkyWay10K(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.72
	cfg.SubSteps = 1
	cfg.NotificationText = "Milky Way Extreme: 10,000 Stellar Bodies active at 144 FPS!"
	cfg.NotificationTimer = 4.0

	// Sagittarius A* Supermassive Black Hole
	smbhMass := 85000.0
	addBodyTex(state, "Sagittarius A* Singularity",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		smbhMass, 5.2, rl.NewColor(15, 10, 25, 255), true, true, TextureBlackHole, 0.4)

	rSource := rand.New(rand.NewSource(10001))

	// 1. Dense Galactic Bulge (2,000 stars)
	bulgeStars := 2000
	for i := 0; i < bulgeStars; i++ {
		r := 6.5 + math.Pow(rSource.Float64(), 1.8)*32.0
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := (rSource.Float64() - 0.5) * 0.45

		x := float32(r * math.Cos(theta) * math.Cos(phi))
		z := float32(r * math.Sin(theta) * math.Cos(phi))
		y := float32(r * math.Sin(phi))

		vCirc := float32(math.Sqrt((g * smbhMass) / r))
		vx := -vCirc*float32(math.Sin(theta)) + float32(rSource.Float64()-0.5)*0.8
		vz := vCirc*float32(math.Cos(theta)) + float32(rSource.Float64()-0.5)*0.8
		vy := float32(rSource.Float64()-0.5) * 0.4

		mass := 0.2 + rSource.Float64()*1.8
		rad := float32(0.35)

		col := rl.NewColor(255, uint8(180+rSource.Intn(75)), uint8(120+rSource.Intn(100)), 255)
		addBodyTex(state, fmt.Sprintf("Bulge Star #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, TextureSun, float32(rSource.Float64()*2.0))
	}

	// 2. Four Logarithmic Spiral Arms (8,000 stars)
	diskStars := 8000
	const numArms = 4
	armNames := []string{"Scutum-Centaurus", "Perseus", "Sagittarius", "Norma"}

	for i := 0; i < diskStars; i++ {
		armIdx := i % numArms
		armOffset := float64(armIdx) * (2.0 * math.Pi / float64(numArms))

		r := 25.0 + math.Pow(rSource.Float64(), 1.3)*230.0
		theta := armOffset + 3.2*math.Log(r/25.0) + (rSource.Float64()-0.5)*0.65

		x := float32(r * math.Cos(theta))
		z := float32(r * math.Sin(theta))
		diskThickness := float32(2.0 + r*0.02)
		y := float32((rSource.Float64() - 0.5) * float64(diskThickness))

		// Flat rotation curve incorporating dark matter halo potential
		haloMass := smbhMass * 0.8 * (r / (r + 40.0))
		effectiveMass := smbhMass + haloMass
		vCirc := float32(math.Sqrt((g * effectiveMass) / r))

		vx := -vCirc * float32(math.Sin(theta))
		vz := vCirc * float32(math.Cos(theta))
		vy := float32((rSource.Float64() - 0.5) * 0.3)

		mass := 0.1 + rSource.Float64()*2.5
		rad := float32(0.28 + mass*0.08)

		roll := rSource.Float64()
		var col rl.Color
		var texType BodyTextureType = TextureMoon
		if roll < 0.18 {
			col = rl.NewColor(135, 205, 255, 255) // Blue supergiant
			texType = TextureSun
			rad *= 1.3
		} else if roll < 0.60 {
			col = rl.NewColor(255, 240, 180, 255) // Yellow-white star
			texType = TextureSun
		} else if roll < 0.90 {
			col = rl.NewColor(255, 140, 90, 255) // Red dwarf
			texType = TextureSun
		} else {
			col = rl.NewColor(220, 210, 255, 255) // White dwarf / neutron star
			texType = TextureWhiteDwarf
		}

		addBodyTex(state, fmt.Sprintf("%s Star #%d", armNames[armIdx], i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, texType, float32(rSource.Float64()*2.0))
	}
}

// loadAsteroidBelt2K creates the outer Solar System with 2,000 asteroids exhibiting Kirkwood resonance gaps and Trojan swarms
func loadAsteroidBelt2K(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.SubSteps = 1
	cfg.NotificationText = "Grand Solar System: 2,000 Asteroids with Kirkwood Gaps & Trojans!"
	cfg.NotificationTimer = 4.0

	rSource := rand.New(rand.NewSource(20002))

	// Central Sun
	sunMass := 22000.0
	addBodyTex(state, "Sun",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		sunMass, 4.5, rl.Gold, true, true, TextureSun, 0.4)

	orbitSun := func(dist float32, mass float64, radius float32, col rl.Color, name string, texType BodyTextureType) *Body {
		spd := float32(math.Sqrt((g * sunMass) / float64(dist)))
		return addBodyTex(state, name,
			rl.NewVector3(dist, 0, 0),
			rl.NewVector3(0, 0, -spd),
			mass, radius, col, false, false, texType, 1.0)
	}

	// Rocky inner planets
	orbitSun(12.0, 0.33, 0.55, rl.NewColor(170, 165, 160, 255), "Mercury", TextureDesert)
	orbitSun(18.0, 4.87, 0.88, rl.NewColor(230, 195, 140, 255), "Venus", TextureDesert)
	orbitSun(26.0, 5.97, 1.05, rl.NewColor(70, 150, 255, 255), "Earth", TextureTerrestrial)
	orbitSun(36.0, 0.64, 0.72, rl.NewColor(225, 95, 65, 255), "Mars", TextureDesert)

	// Jupiter and Galilean Moons at r = 100.0
	jupDist := float32(100.0)
	jupMass := 350.0
	jupSpd := float32(math.Sqrt((g * sunMass) / float64(jupDist)))
	jupiter := addBodyTex(state, "Jupiter",
		rl.NewVector3(jupDist, 0, 0),
		rl.NewVector3(0, 0, -jupSpd),
		jupMass, 2.8, rl.NewColor(220, 175, 130, 255), false, false, TextureGasGiant, 2.0)

	// Galilean Moons
	addBodyTex(state, "Io", rl.NewVector3(jupDist+4.2, 0, 0), rl.NewVector3(0, 0, -jupSpd-4.2), 0.05, 0.32, rl.Yellow, false, false, TextureMoon, 1.0)
	addBodyTex(state, "Europa", rl.NewVector3(jupDist+6.5, 0, 0), rl.NewVector3(0, 0, -jupSpd-3.4), 0.04, 0.30, rl.SkyBlue, false, false, TextureMoon, 1.0)
	addBodyTex(state, "Ganymede", rl.NewVector3(jupDist+9.5, 0, 0), rl.NewVector3(0, 0, -jupSpd-2.8), 0.08, 0.40, rl.LightGray, false, false, TextureMoon, 1.0)
	addBodyTex(state, "Callisto", rl.NewVector3(jupDist+13.0, 0, 0), rl.NewVector3(0, 0, -jupSpd-2.3), 0.06, 0.38, rl.DarkGray, false, false, TextureMoon, 1.0)

	// Saturn with Rings at r = 155.0
	satDist := float32(155.0)
	satMass := 140.0
	satSpd := float32(math.Sqrt((g * sunMass) / float64(satDist)))
	saturn := addBodyTex(state, "Saturn",
		rl.NewVector3(0, 0, satDist),
		rl.NewVector3(satSpd, 0, 0),
		satMass, 2.3, rl.NewColor(235, 215, 170, 255), false, false, TextureGasGiant, 1.8)
	saturn.RingInnerRadius = 3.4
	saturn.RingOuterRadius = 6.8

	// Main Asteroid Belt: 1,400 Asteroids between Mars (36) and Jupiter (100)
	// Realistic Kirkwood resonance gaps at 3:1 (r=50.0), 5:2 (r=58.5), 2:1 (r=65.5)
	beltCount := 1400
	ceresAdded := false

	for i := 0; i < beltCount; i++ {
		var r float64
		for attempts := 0; attempts < 10; attempts++ {
			r = 42.0 + rSource.Float64()*46.0
			if math.Abs(r-50.0) < 1.4 || math.Abs(r-58.5) < 1.2 || math.Abs(r-65.5) < 1.5 {
				continue
			}
			break
		}

		ang := rSource.Float64() * 2.0 * math.Pi
		inclination := (rSource.Float64() - 0.5) * 0.12

		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		y := float32(r * math.Sin(inclination))

		vCirc := float32(math.Sqrt((g * sunMass) / r))
		vx := vCirc * float32(math.Sin(ang))
		vz := -vCirc * float32(math.Cos(ang))
		vy := float32((rSource.Float64() - 0.5) * 0.25)

		name := fmt.Sprintf("Asteroid #%d", i+1)
		mass := 0.002 + rSource.Float64()*0.02
		rad := float32(0.25)

		if !ceresAdded && i == 0 {
			name = "Ceres (Dwarf Planet)"
			mass = 12.0
			rad = 0.65
			ceresAdded = true
		} else if i == 1 {
			name = "Vesta"
			mass = 6.0
			rad = 0.52
		} else if i == 2 {
			name = "Pallas"
			mass = 5.0
			rad = 0.48
		}

		col := rl.NewColor(uint8(150+rSource.Intn(60)), uint8(140+rSource.Intn(50)), uint8(130+rSource.Intn(50)), 255)
		addBodyTex(state, name,
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, TextureMoon, float32(rSource.Float64()))
	}

	// Jupiter's Trojan (L4) and Greek (L5) Asteroid Swarms: 400 Asteroids
	trojanCount := 200
	greekCount := 200

	l4Angle := -math.Pi / 3.0 // 60 degrees ahead in clockwise orbit
	for i := 0; i < trojanCount; i++ {
		ang := l4Angle + (rSource.Float64()-0.5)*0.35
		r := float64(jupDist) + (rSource.Float64()-0.5)*12.0

		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		y := float32((rSource.Float64() - 0.5) * 4.0)

		vCirc := float32(math.Sqrt((g * sunMass) / r))
		vx := vCirc * float32(math.Sin(ang))
		vz := -vCirc * float32(math.Cos(ang))
		vy := float32((rSource.Float64() - 0.5) * 0.2)

		col := rl.NewColor(185, 175, 140, 255)
		addBodyTex(state, fmt.Sprintf("Trojan Asteroid (L4) #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			0.01, 0.28, col, false, false, TextureMoon, 1.0)
	}

	l5Angle := math.Pi / 3.0 // 60 degrees behind in clockwise orbit
	for i := 0; i < greekCount; i++ {
		ang := l5Angle + (rSource.Float64()-0.5)*0.35
		r := float64(jupDist) + (rSource.Float64()-0.5)*12.0

		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		y := float32((rSource.Float64() - 0.5) * 4.0)

		vCirc := float32(math.Sqrt((g * sunMass) / r))
		vx := vCirc * float32(math.Sin(ang))
		vz := -vCirc * float32(math.Cos(ang))
		vy := float32((rSource.Float64() - 0.5) * 0.2)

		col := rl.NewColor(160, 180, 170, 255)
		addBodyTex(state, fmt.Sprintf("Greek Asteroid (L5) #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			0.01, 0.28, col, false, false, TextureMoon, 1.0)
	}
	_ = jupiter
}

// loadGalaxyCollision5K generates two interacting spiral galaxies (5,000 stars total) colliding in 3D
func loadGalaxyCollision5K(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.72
	cfg.SubSteps = 1
	cfg.NotificationText = "Milky Way & Andromeda Collision: 5,000 Interacting Stars at 144 FPS!"
	cfg.NotificationTimer = 4.0

	rSource := rand.New(rand.NewSource(50005))

	generateGalaxy := func(centerX, centerY, centerZ float32, velX, velY, velZ float32, smbhMass float64, starCount int, name string, tiltAngle float64, smbhCol rl.Color, baseStarCol rl.Color) {
		addBodyTex(state, fmt.Sprintf("%s Core SMBH", name),
			rl.NewVector3(centerX, centerY, centerZ),
			rl.NewVector3(velX, velY, velZ),
			smbhMass, 4.2, smbhCol, false, true, TextureBlackHole, 0.5)

		cosT := float32(math.Cos(tiltAngle))
		sinT := float32(math.Sin(tiltAngle))

		for i := 0; i < starCount; i++ {
			arm := float64(i % 2)
			r := 8.0 + math.Pow(rSource.Float64(), 1.4)*95.0
			theta := arm*math.Pi + 2.8*math.Log(r/8.0) + (rSource.Float64()-0.5)*0.6

			px := float32(r * math.Cos(theta))
			pz := float32(r * math.Sin(theta))
			py := float32((rSource.Float64() - 0.5) * (r * 0.06))

			vCirc := float32(math.Sqrt((g * smbhMass) / r))
			pvx := -vCirc * float32(math.Sin(theta))
			pvz := vCirc * float32(math.Cos(theta))
			pvy := float32((rSource.Float64() - 0.5) * 0.2)

			rotatedY := py*cosT - pz*sinT
			rotatedZ := py*sinT + pz*cosT
			rotatedVy := pvy*cosT - pvz*sinT
			rotatedVz := pvy*sinT + pvz*cosT

			mass := 0.2 + rSource.Float64()*2.0
			rad := float32(0.28)

			col := baseStarCol
			if rSource.Float64() < 0.25 {
				col = rl.NewColor(160, 220, 255, 255)
			}

			addBodyTex(state, fmt.Sprintf("%s Star #%d", name, i+1),
				rl.NewVector3(centerX+px, centerY+rotatedY, centerZ+rotatedZ),
				rl.NewVector3(velX+pvx, velY+rotatedVy, velZ+rotatedVz),
				mass, rad, col, false, false, TextureSun, float32(rSource.Float64()*2.0))
		}
	}

	// Galaxy 1: Milky Way (2,500 stars)
	generateGalaxy(-85.0, -8.0, -30.0, 1.6, 0.0, 0.55, 48000.0, 2500, "Milky Way", 0.25, rl.NewColor(30, 20, 45, 255), rl.NewColor(255, 230, 180, 255))

	// Galaxy 2: Andromeda M31 (2,500 stars)
	generateGalaxy(85.0, 18.0, 30.0, -1.6, -0.25, -0.55, 56000.0, 2500, "Andromeda", 0.78, rl.NewColor(20, 35, 55, 255), rl.NewColor(200, 225, 255, 255))
}

// loadBlackHoleSwarm3K creates a gargantuan Kerr Black Hole with 3,000 relativistic accretion and cluster bodies
func loadBlackHoleSwarm3K(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.EnableRelativity = true
	cfg.SpeedOfLight = 190.0
	cfg.SubSteps = 1
	cfg.NotificationText = "Gargantua Kerr Black Hole: 3,000 Relativistic Particles at 144 FPS!"
	cfg.NotificationTimer = 4.0

	rSource := rand.New(rand.NewSource(30003))

	// Central Gargantua Kerr Black Hole
	bhMass := 120000.0
	bh := addBodyTex(state, "Gargantua Kerr Singularity",
		rl.NewVector3(0, 0, 0),
		rl.NewVector3(0, 0, 0),
		bhMass, 6.2, rl.NewColor(10, 5, 20, 255), true, true, TextureBlackHole, 1.2)
	bh.IsPulsar = true // Relativistic polar jet beams

	// 1. High-Velocity Inner Accretion Disk (1,500 plasma particles)
	innerCount := 1500
	for i := 0; i < innerCount; i++ {
		r := 10.0 + math.Pow(rSource.Float64(), 1.6)*40.0
		ang := rSource.Float64() * 2.0 * math.Pi

		x := float32(r * math.Cos(ang))
		z := float32(r * math.Sin(ang))
		y := float32((rSource.Float64() - 0.5) * 1.2)

		vCirc := float32(math.Sqrt((g * bhMass) / r))
		vx := -vCirc * float32(math.Sin(ang))
		vz := vCirc * float32(math.Cos(ang))
		vy := float32((rSource.Float64() - 0.5) * 0.2)

		col := rl.NewColor(uint8(255), uint8(140+rSource.Intn(115)), uint8(40+rSource.Intn(80)), 255)
		if rSource.Float64() < 0.3 {
			col = rl.NewColor(120, 220, 255, 255)
		}

		addBodyTex(state, fmt.Sprintf("Accretion Ion #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			0.05, 0.25, col, false, false, TextureSun, 2.0)
	}

	// 2. Relativistic Swarm on Inclined Rosette Orbits (1,500 stars)
	outerCount := 1500
	for i := 0; i < outerCount; i++ {
		rp := 18.0 + rSource.Float64()*40.0
		ecc := 0.35 + rSource.Float64()*0.45
		ra := rp * (1.0 + ecc) / (1.0 - ecc)
		a := (rp + ra) * 0.5
		vPeri := float32(math.Sqrt(g * bhMass * ((2.0 / rp) - (1.0 / a))))

		ang := rSource.Float64() * 2.0 * math.Pi
		incl := (rSource.Float64() - 0.5) * 1.1

		px := float32(rp * math.Cos(ang))
		pz := float32(rp * math.Sin(ang))
		py := float32(math.Sin(incl) * float64(rp))

		vx := -vPeri * float32(math.Sin(ang))
		vz := vPeri * float32(math.Cos(ang))
		vy := float32((rSource.Float64() - 0.5) * 0.6)

		mass := 0.2 + rSource.Float64()*3.0
		rad := float32(0.32)

		col := rl.NewColor(uint8(180+rSource.Intn(75)), uint8(180+rSource.Intn(75)), 255, 255)
		addBodyTex(state, fmt.Sprintf("Relativistic Star #%d", i+1),
			rl.NewVector3(px, py, pz),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, false, TextureMoon, float32(rSource.Float64()*2.0))
	}
}

// SpawnStarCluster spawns 50 stars clustered around position with launch velocity
func SpawnStarCluster(state *SimState, pos, vel rl.Vector3, count int) {
	for i := 0; i < count; i++ {
		ang := rand.Float64() * 2.0 * math.Pi
		r := 1.5 + math.Pow(rand.Float64(), 1.5)*8.0
		offset := rl.NewVector3(
			float32(r*math.Cos(ang)),
			float32((rand.Float64()-0.5)*2.0),
			float32(r*math.Sin(ang)),
		)
		bodyPos := rl.Vector3Add(pos, offset)
		spreadVel := rl.NewVector3(
			vel.X+float32(rand.Float64()-0.5)*1.2,
			vel.Y+float32(rand.Float64()-0.5)*0.6,
			vel.Z+float32(rand.Float64()-0.5)*1.2,
		)
		b := CreateBodyTemplate(SpawnMoon, bodyPos, spreadVel, state.NextID)
		b.Name = fmt.Sprintf("Cluster Star #%d", state.NextID)
		b.Radius = 0.32
		b.Mass = 0.2 + rand.Float64()*1.0
		b.Color = rl.NewColor(uint8(200+rand.Intn(55)), uint8(200+rand.Intn(55)), 255, 255)
		state.NextID++
		state.Bodies = append(state.Bodies, b)
	}
}

// SpawnMiniGalaxy spawns a 150-star rotating mini spiral disk
func SpawnMiniGalaxy(state *SimState, pos, vel rl.Vector3, g float64) {
	coreMass := 1500.0
	addBodyTex(state, fmt.Sprintf("Dwarf Core #%d", state.NextID),
		pos, vel, coreMass, 1.8, rl.Gold, false, true, TextureBlackHole, 1.0)

	for i := 0; i < 150; i++ {
		arm := float64(i % 2)
		r := 2.5 + math.Pow(rand.Float64(), 1.3)*22.0
		theta := arm*math.Pi + 2.5*math.Log(r/2.5) + (rand.Float64()-0.5)*0.5

		ox := float32(r * math.Cos(theta))
		oz := float32(r * math.Sin(theta))
		oy := float32((rand.Float64() - 0.5) * 0.8)

		vCirc := float32(math.Sqrt((g * coreMass) / r))
		vx := vel.X - vCirc*float32(math.Sin(theta))
		vz := vel.Z + vCirc*float32(math.Cos(theta))
		vy := vel.Y

		addBodyTex(state, fmt.Sprintf("Dwarf Star #%d", state.NextID),
			rl.NewVector3(pos.X+ox, pos.Y+oy, pos.Z+oz),
			rl.NewVector3(vx, vy, vz),
			0.1, 0.26, rl.SkyBlue, false, false, TextureSun, 1.0)
	}
}

// loadGravitationalCollapse creates a cold turbulent particle cloud of 1,600 bodies that collapses under self-gravity into a sphere
func loadGravitationalCollapse(state *SimState, cfg *Config, g float64) {
	cfg.UseBarnesHut = true
	cfg.BarnesHutTheta = 0.7
	cfg.Softening = 0.65
	cfg.NotificationText = "Gravitational Collapse: 1,600 cold particles collapsing into a spherical cluster under self-gravity!"
	cfg.NotificationTimer = 4.5

	count := 1600
	cloudRadius := 55.0
	totalMass := 4500.0
	partMass := totalMass / float64(count)

	rSource := rand.New(rand.NewSource(1337))

	for i := 0; i < count; i++ {
		// Non-spherical triaxial turbulent ellipsoid with clumpy perturbations
		u := rSource.Float64()
		r := cloudRadius * math.Pow(u, 0.45)
		theta := rSource.Float64() * 2.0 * math.Pi
		phi := math.Acos(2.0*rSource.Float64() - 1.0)

		clump := 1.0 + 0.15*math.Sin(theta*3.0)*math.Cos(phi*2.0)
		x := float32(r * math.Sin(phi) * math.Cos(theta) * 1.35 * clump)
		y := float32(r * math.Sin(phi) * math.Sin(theta) * 0.85 * clump)
		z := float32(r * math.Cos(phi) * 1.15 * clump)

		// Sub-virial cold velocity (Q ~ 0.06): particles fall directly inward to form a sphere
		vDisp := float32(0.35)
		vx := float32(rSource.Float64()-0.5) * vDisp
		vy := float32(rSource.Float64()-0.5) * vDisp * 0.6
		vz := float32(rSource.Float64()-0.5) * vDisp

		mass := partMass * (0.7 + rSource.Float64()*0.6)
		rad := float32(0.25 + mass*0.04)

		var col rl.Color
		var tex BodyTextureType = TextureSun
		roll := rSource.Float64()
		if roll < 0.25 {
			col = rl.NewColor(100, 200, 255, 255) // Hot proto-star blue
			tex = TextureWhiteDwarf
		} else if roll < 0.65 {
			col = rl.NewColor(255, 220, 110, 255) // Golden star
			tex = TextureSun
		} else {
			col = rl.NewColor(255, 120, 50, 255) // Red dwarf
			tex = TextureRedGiant
		}

		addBodyTex(state, fmt.Sprintf("Cloud Star #%d", i+1),
			rl.NewVector3(x, y, z),
			rl.NewVector3(vx, vy, vz),
			mass, rad, col, false, true, tex, float32(rSource.Float64()*2.0))
	}
}

// SpawnCollapseCloud spawns 100 cold self-gravitating particles in an irregular cloud that collapses into a sphere under gravity
func SpawnCollapseCloud(state *SimState, pos, vel rl.Vector3, count int) {
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

		bodyPos := rl.Vector3Add(pos, rl.NewVector3(ox, oy, oz))

		// Sub-virial cold velocity (Q ~ 0.08)
		vDisp := float32(0.12)
		spreadVel := rl.NewVector3(
			vel.X+float32(rSource.Float64()-0.5)*vDisp,
			vel.Y+float32(rSource.Float64()-0.5)*vDisp*0.5,
			vel.Z+float32(rSource.Float64()-0.5)*vDisp,
		)

		b := CreateBodyTemplate(SpawnMoon, bodyPos, spreadVel, state.NextID)
		b.Name = fmt.Sprintf("Collapse Particle #%d", state.NextID)
		b.Radius = 0.28
		b.Mass = partMass * (0.8 + rSource.Float64()*0.4)
		roll := rSource.Float64()
		if roll < 0.3 {
			b.Color = rl.NewColor(100, 210, 255, 255)
		} else if roll < 0.7 {
			b.Color = rl.NewColor(255, 210, 100, 255)
		} else {
			b.Color = rl.NewColor(255, 140, 60, 255)
		}
		b.TextureType = TextureSun
		state.NextID++
		state.Bodies = append(state.Bodies, b)
	}
}

