package main

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CollisionMode defines how bodies interact when colliding
type CollisionMode int

const (
	CollisionMerge CollisionMode = iota
	CollisionBounce
	CollisionNone
)

// BodyTextureType defines procedural texture archetype for celestial spheres
type BodyTextureType int

const (
	TextureNone BodyTextureType = iota
	TextureTerrestrial
	TextureDesert
	TextureGasGiant
	TextureIceGiant
	TextureMoon
	TextureSun
	TextureBlackHole
	TextureRedGiant
	TextureWhiteDwarf
	TextureNeutronStar
	TextureDwarfPlanet
	TextureComet
)

// CelestialIcon defines 2D tactical map icon archetype for celestial bodies
type CelestialIcon int

const (
	IconCircle      CelestialIcon = iota // Clean default circular disc
	IconStar                             // Glowing radiant sunburst star glyph
	IconBlackHole                        // Schwarzschild black hole with accretion disc
	IconPulsar                           // Relativistic jet beam pulsar
	IconGasGiant                         // Striped planet with inclined rings
	IconTerrestrial                      // Habitable ocean/continent world
	IconDesert                           // Desert/rocky world with polar cap
	IconIceGiant                         // Crystalline ice giant
	IconMoon                             // Cratered satellite / asteroid
	IconProbe                            // Scientific satellite craft with solar wings
)

// Body represents a celestial body / mass in 3D space
type Body struct {
	ID              int64
	Name            string
	Position        rl.Vector3
	Velocity        rl.Vector3
	Acceleration    rl.Vector3
	Mass            float64
	Radius          float32
	Color           rl.Color
	IsStationary    bool    // If true, fixed in space (infinite mass approximation / anchor)
	IsStar          bool    // Emits glow
	IsPulsar        bool    // Emits rotating relativistic jet beams
	IsComet         bool    // Emits dynamic solar wind ion/dust tail
	RingInnerRadius float32 // If > 0, renders planar dust ring
	RingOuterRadius float32
	Trail           []rl.Vector3
	TrailTimer      float32
	NetForce        rl.Vector3 // For visualization
	TextureType     BodyTextureType
	CelestialIcon   CelestialIcon // Dedicated 2D tactical map icon
	RotationAngle   float32
	RotationSpeed   float32
	Disrupted       bool // Flag to prevent multiple Roche disruptions
}

// Config stores global simulation parameters
type Config struct {
	G             float64       // Gravitational constant
	TimeScale     float64       // Simulation speed multiplier (0.1x to 10x)
	TimeScaleStep float64       // Simulation speed step increment (0.1, 0.5, 1.0, 2.0; default 0.5)
	SubSteps      int           // Number of physics substeps per frame for high precision
	Softening     float64       // Gravitational softening epsilon^2 to prevent slingshot singularities
	Collision     CollisionMode // Collision behavior
	Paused        bool
	ShowTrails    bool
	ShowGrid      bool
	ShowVectors   bool
	ShowForces    bool
	ShowLabels    bool
	ShowStarfield bool
	ShowHelp      bool

	// Roadmap Features:
	UseBarnesHut        bool    // Enable Barnes-Hut O(N log N) octree for large N simulations
	BarnesHutTheta      float32 // Opening criterion angle theta (default ~0.7)
	UseGPUCompute       bool    // Enable GPU Compute Shader acceleration for direct N-body gravity
	GPUAvailable        bool    // Flag indicating if GPU Compute Shader is available on this system
	EnableRelativity    bool    // Enable post-Newtonian 1PN general relativity perihelion advance
	SpeedOfLight        float64 // Effective speed of light in simulation units (c)
	EnableRocheLimit    bool    // Enable tidal disruption when approaching massive bodies
	ShowTextures        bool    // Render procedural celestial textures and axial rotation
	ShowPotentialGrid   bool    // 3D Spacetime gravitational curvature potential grid
	SpacetimeScale      float32 // Global scale multiplier for spacetime grid (default 1.0)
	SpacetimeResolution int     // Grid subdivisions per axis (e.g. 64, 80, 96; default 80)
	ShowVectorField     bool    // Gravitational vector field in 3D
	CinematicCamera     bool    // Cinematic camera auto-tour mode
	ParticleGlowMode    bool    // Color particles based on orbital kinetic velocity
	AutoPerformanceMode bool    // Automatically optimize substeps and collision for large N
	PhysicsTimeMs       float32 // Physics step duration in ms
	RenderTimeMs        float32 // Render step duration in ms
	TreeDepth           int     // Barnes-Hut octree maximum depth
	InteractionsPerSec  int64   // Calculated interactions per second (for telemetry)
	// Roadmap & Advanced Astrophysical Simulation Features:
	Integrator             IntegratorType // Numerical integration algorithm (Verlet or Yoshida 4th order)
	Show2DViewport         bool           // Show 2D tactical orbital viewport / minimap
	Viewport2DFullscreen   bool           // Expand 2D viewport to full screen
	Show2DHeatmap          bool           // Show thermodynamic gravitational field heatmap in 2D
	Heatmap2DResolution    int            // 2D Gravitational Heatmap grid resolution columns (16 to 200, default 64)
	Show2DVectorField      bool           // Show 2D directional vector arrows
	Viewport2DZoom         float32        // Zoom scale for 2D viewport (default 1.0)
	Viewport2DPan          rl.Vector2     // Pan offset (X, Z) for 2D viewport
	RealScaleVisualMode    bool           // Toggle 1:1 true physical radius vs visible scaled radius
	ShowLagrangePoints     bool           // Render calculated Lagrange equilibrium points L1-L5
	Show2DLabels           bool           // Show celestial and Lagrange labels in 2D viewport (default true)
	Show2DIcons            bool           // Toggle celestial object icons in 2D viewport (stars, planets, moons, probes)
	Show2DCircles          bool           // Toggle body overlay outline circles in 2D viewport (default true)
	Show3DHeatmapPlane     bool           // Render continuous thermodynamic gravitational heatmap overlay on the 3D plane
	ShowGravitationalWaves bool           // Render dynamic gravitational wave metric ripples and LIGO detector HUD
	DenseVectorField       bool           // High-density 3D gravitational vector lattice
	Offload3D              bool           // Offload/stop 3D rendering and use 2D Viewport as primary main display
	Track2D                bool           // Continuously lock & track selected body in 2D viewport center
	Show2DGW               bool           // Render gravitational wave ripples & propagating bursts in 2D viewport
	ShowStatsHUD           bool           // Toggle FPS & simulation telemetry HUD overlay
	NotificationText       string         // Temporary HUD notification message
	NotificationTimer      float32        // Notification countdown timer
}

// GravitationalWaveBurst represents a transient gravitational wave pulse (e.g. merger ringdown)
type GravitationalWaveBurst struct {
	Position   rl.Vector3
	StartTime  float64
	PeakStrain float32
	Frequency  float32
	DecayTime  float32
	WaveSpeed  float32
	Source     string
}

// IntegratorType specifies the numerical integration algorithm
type IntegratorType int

const (
	IntegratorVerlet   IntegratorType = iota // 2nd-order symplectic Velocity Verlet
	IntegratorYoshida4                       // 4th-order symplectic Yoshida integrator (ultra-high precision)
)

func (it IntegratorType) String() string {
	switch it {
	case IntegratorVerlet:
		return "Velocity Verlet (2nd Order Symplectic)"
	case IntegratorYoshida4:
		return "Yoshida 4th-Order (Symplectic O(dt^4))"
	default:
		return "Velocity Verlet"
	}
}

// PresetType identifies available galaxy / solar system presets
type PresetType int

const (
	PresetSolarSystem       PresetType = iota
	PresetSolarSystemGrand             // Full Solar System with 40+ Asteroids & Moons
	PresetMilkyWay10K                  // Version 1: Gigantic Milky Way Galaxy with 10,000 particles & Sgr A*
	PresetAsteroidBelt2K               // Version 2: Massive Deep Asteroid Belt with 2,000 asteroids & Kirkwood gaps
	PresetGalaxyCollision5K            // Version 3: Gigantic Collision: Milky Way & Andromeda with 5,000 stars
	PresetBlackHoleSwarm3K             // Version 4: Supermassive Kerr Black Hole Accretion Swarm (3,000 particles)
	PresetBinaryStars
	PresetThreeBody
	PresetGalaxyDisk
	PresetMilkyWayCluster       // Large-scale Barnes-Hut showcase (650+ stars)
	PresetLagrangeTrojans       // L4 / L5 Lagrangian equilibrium Trojan asteroid swarms
	PresetGalaxyCollision       // Two interacting spiral galaxies with tidal tails
	PresetPulsarAccretion       // High-speed relativistic pulsar with jet beams + companion
	PresetGlobularCluster       // 180 stars dynamically relaxing & core collapse
	PresetTrappistResonance     // TRAPPIST-1 with 7 resonant terrestrial exoplanets
	PresetFiveBody              // 5-Body symmetric choreography
	PresetRelativityRosette     // 1PN General Relativity precession around black hole
	PresetRocheDisruption       // Tidal breakup of incoming moon into planetary rings
	PresetGravitationalCollapse // Cold gas/stellar cloud collapsing under mutual self-gravity into a sphere
	PresetRealSolarSystem       // Real Scale Solar System (Astronomical AU, authentic masses & Keplerian orbits)
	PresetEmpty
)

func (p PresetType) String() string {
	switch p {
	case PresetSolarSystem:
		return "Solar System"
	case PresetSolarSystemGrand:
		return "Solar System Grand"
	case PresetRealSolarSystem:
		return "Real Scale Solar System"
	case PresetMilkyWay10K:
		return "Milky Way 10K"
	case PresetAsteroidBelt2K:
		return "Asteroid Belt 2K"
	case PresetGalaxyCollision5K:
		return "Galaxy Collision 5K"
	case PresetBlackHoleSwarm3K:
		return "Black Hole Swarm 3K"
	case PresetGravitationalCollapse:
		return "Gravitational Collapse (Sphere)"
	case PresetBinaryStars:
		return "Binary Stars"
	case PresetThreeBody:
		return "3-Body"
	case PresetGalaxyDisk:
		return "Galaxy Disk"
	case PresetMilkyWayCluster:
		return "Milky Way Cluster"
	case PresetLagrangeTrojans:
		return "Lagrange Trojans"
	case PresetGalaxyCollision:
		return "Galaxy Collision"
	case PresetPulsarAccretion:
		return "Pulsar Accretion"
	case PresetGlobularCluster:
		return "Globular Cluster"
	case PresetTrappistResonance:
		return "TRAPPIST-1 Resonance"
	case PresetFiveBody:
		return "5-Body Choreography"
	case PresetRelativityRosette:
		return "1PN Relativity Rosette"
	case PresetRocheDisruption:
		return "Roche Tidal Disruption"
	case PresetEmpty:
		return "Empty Space"
	default:
		return "Custom Scenario"
	}
}

// SpawnType defines template for creating new celestial bodies
type SpawnType int

const (
	SpawnEarth SpawnType = iota
	SpawnMoon
	SpawnGasGiant
	SpawnIceGiant
	SpawnStar
	SpawnRedGiant
	SpawnWhiteDwarf
	SpawnNeutronStar
	SpawnBlackHole
	SpawnDwarfPlanet
	SpawnComet
	SpawnAsteroidRing
	SpawnStarCluster50
	SpawnMiniGalaxy150
	SpawnCollapseCloud100
)

func (s SpawnType) String() string {
	switch s {
	case SpawnEarth:
		return "Earth"
	case SpawnMoon:
		return "Moon"
	case SpawnGasGiant:
		return "Gas Giant"
	case SpawnIceGiant:
		return "Ice Giant"
	case SpawnStar:
		return "Sun"
	case SpawnRedGiant:
		return "Red Giant"
	case SpawnWhiteDwarf:
		return "White Dwarf"
	case SpawnNeutronStar:
		return "Pulsar"
	case SpawnBlackHole:
		return "Black Hole"
	case SpawnDwarfPlanet:
		return "Dwarf Planet"
	case SpawnComet:
		return "Comet"
	case SpawnAsteroidRing:
		return "Asteroid Ring"
	case SpawnStarCluster50:
		return "Star Cluster (50x)"
	case SpawnMiniGalaxy150:
		return "Mini Galaxy (150x)"
	case SpawnCollapseCloud100:
		return "Collapse Cloud (100x)"
	default:
		return "Body"
	}
}

// SimState holds runtime world state and UI interactions
type SimState struct {
	Bodies                []*Body
	NextID                int64
	SelectedBodyID        int64
	SelectedLagrangeIndex int // 1..5 if a Lagrange equilibrium point (L1-L5) is selected, 0 otherwise
	FollowSelected        bool
	CurrentPreset         PresetType

	// Spawning mode
	IsSpawning      bool
	SpawnPreset     SpawnType
	SpawnWorldPos   rl.Vector3
	IsDraggingSpawn bool
	DragVelocity    rl.Vector3

	// Force application mode (impulse on selected body)
	IsDraggingForce bool
	ForceStartPos   rl.Vector2
	ForceVector     rl.Vector3

	// Background stars
	Stars      []rl.Vector3
	StarColors []rl.Color
	StarSizes  []float32

	// Cinematic camera state
	CinematicAngle float32
	CinematicTime  float32

	// Scenario file selector modal
	ShowScenarioModal  bool
	AvailableScenarios []string
	ScenarioModalTimer float32

	// Preset overflow dropdown menu
	ShowMorePresetsDropdown bool

	// Camera control panel state
	ShowCameraPanel  bool
	FollowBarycenter bool

	// Gravitational Wave state & telemetry
	GWBursts          []GravitationalWaveBurst
	LastGWStrain      float32
	LastGWFreq        float32
	LastGWPower       float64
	GWWaveformHistory [128]float32
	GWWaveformHead    int
}

// DetermineCelestialIcon returns the proper 2D tactical glyph based on astronomical properties
func DetermineCelestialIcon(name string, isStar, isPulsar, isComet bool, mass float64, texType BodyTextureType) CelestialIcon {
	if strings.Contains(name, "Probe") || strings.Contains(name, "JWST") ||
		strings.Contains(name, "SOHO") || strings.Contains(name, "Satellite") || strings.Contains(name, "Voyager") || (mass < 0.0002 && !isStar) {
		return IconProbe
	}
	if texType == TextureBlackHole || strings.Contains(name, "Black Hole") || strings.Contains(name, "Singularity") || strings.Contains(name, "Gargantua") {
		return IconBlackHole
	}
	if isPulsar || texType == TextureNeutronStar || strings.Contains(name, "Pulsar") || strings.Contains(name, "Neutron") {
		return IconPulsar
	}
	if isStar || texType == TextureSun || texType == TextureRedGiant || texType == TextureWhiteDwarf || strings.Contains(name, "Star") || strings.Contains(name, "Sun") {
		return IconStar
	}
	if texType == TextureTerrestrial || strings.Contains(name, "Earth") {
		return IconTerrestrial
	}
	if texType == TextureDesert || strings.Contains(name, "Mars") || strings.Contains(name, "Mercury") {
		return IconDesert
	}
	if texType == TextureIceGiant || strings.Contains(name, "Neptune") || strings.Contains(name, "Uranus") {
		return IconIceGiant
	}
	if texType == TextureGasGiant || strings.Contains(name, "Jupiter") || strings.Contains(name, "Saturn") || (mass >= 15.0 && !isStar) {
		return IconGasGiant
	}
	if texType == TextureMoon || strings.Contains(name, "Moon") || strings.Contains(name, "Asteroid") || isComet || mass < 0.01 {
		return IconMoon
	}
	return IconCircle
}
