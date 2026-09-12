//go:build js && wasm

package main

import "strings"

type CollisionMode int

const (
	CollisionMerge CollisionMode = iota
	CollisionBounce
	CollisionGhost
)

type PresetType int

const (
	PresetSolarSystem PresetType = iota
	PresetSolarSystemGrand
	PresetBinaryStars
	PresetThreeBody
	PresetLagrangeTrojans
	PresetGalaxyCollision
	PresetPulsarAccretion
	PresetMilkyWay
	PresetGlobularCluster
	PresetTrappist1
	PresetFiveBodyChoreography
	PresetRelativisticRosette
	PresetRocheDisruption
	PresetEmpty
	PresetMilkyWay10K
	PresetAsteroidBelt2K
	PresetGalaxyCollision5K
	PresetBlackHoleSwarm3K
	PresetGravitationalCollapse
	PresetRealSolarSystem
)

const MaxTrailPoints = 80

type Body struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Position        Vector3   `json:"pos"`
	Velocity        Vector3   `json:"vel"`
	Acceleration    Vector3   `json:"acc"`
	Mass            float64   `json:"mass"`
	Radius          float32   `json:"radius"`
	Color           Color     `json:"color"`
	Trail           []Vector3 `json:"trail,omitempty"`
	TrailHead       int       `json:"-"`
	TrailLen        int       `json:"-"`
	IsStationary    bool      `json:"is_stationary"`
	IsStar          bool      `json:"is_star"`
	IsPulsar        bool      `json:"is_pulsar"`
	IsBlackHole     bool      `json:"is_blackhole"`
	IsComet         bool      `json:"is_comet"`
	RingInnerRadius float32   `json:"ring_inner_radius,omitempty"`
	RingOuterRadius float32   `json:"ring_outer_radius,omitempty"`
	TextureType     int       `json:"texture_type"`
	CelestialIcon   int       `json:"celestial_icon"`
	RotationAngle   float32   `json:"rotation_angle"`
}

type Config struct {
	G                      float64       `json:"g"`
	TimeScale              float64       `json:"time_scale"`
	TimeScaleStep          float64       `json:"time_scale_step"`
	SubSteps               int           `json:"sub_steps"`
	Softening              float64       `json:"softening"`
	Collision              CollisionMode `json:"collision"`
	Paused                 bool          `json:"paused"`
	EnableBarnesHut        bool          `json:"enable_barnes_hut"`
	BarnesHutTheta         float64       `json:"barnes_hut_theta"`
	EnableRelativity       bool          `json:"enable_relativity"`
	SpeedOfLight           float64       `json:"speed_of_light"`
	EnableRocheLimit       bool          `json:"enable_roche_limit"`
	ParticleGlowMode       bool          `json:"particle_glow_mode"`
	ShowPotentialGrid      bool          `json:"show_potential_grid"`
	ShowGravitationalWaves bool          `json:"show_gravitational_waves"`
	SpacetimeScale         float32       `json:"spacetime_scale"`
	SpacetimeResolution    int           `json:"spacetime_resolution"`
	ShowVectorField        bool          `json:"show_vector_field"`
	Show2DLabels           bool          `json:"show_2d_labels"`
	Show3DHeatmapPlane     bool          `json:"show_3d_heatmap_plane"`
	ShowLagrangePoints     bool          `json:"show_lagrange_points"`
	Offload3D              bool          `json:"offload_3d"`
	Track2D                bool          `json:"track_2d"`
	Show2DGW               bool          `json:"show_2d_gw"`
	ShowStatsHUD           bool          `json:"show_stats_hud"`
	ShowVectors            bool          `json:"show_vectors"`
	Show2DCircles          bool          `json:"show_2d_circles"`
	Heatmap2DResolution    int           `json:"heatmap_2d_resolution"`
}

type SimState struct {
	Bodies             []*Body                  `json:"bodies"`
	NextID             int                      `json:"-"`
	Time               float64                  `json:"time"`
	SelectedBodyID     int                      `json:"selected_id"`
	CollisionCount     int                      `json:"collisions"`
	MergeCount         int                      `json:"merges"`
	TidalBreakupCount  int                      `json:"tidal_breakups"`
	OctreeDepth        int                      `json:"octree_depth"`
	InteractionsPerSec int64                    `json:"interactions_per_sec"`
	PhysicsTimeMs      float64                  `json:"physics_time_ms"`
	GWBursts           []GravitationalWaveBurst `json:"-"`
	LastGWStrain       float32                  `json:"gw_strain"`
	LastGWFreq         float32                  `json:"gw_freq"`
	LastGWPower        float64                  `json:"gw_power"`
	GWWaveformHistory  [120]float32             `json:"-"`
	GWWaveformHead     int                      `json:"-"`
}

func DetermineCelestialIcon(name string, isStar, isPulsar, isComet bool, mass float64, texType int) int {
	// 0: default, 1: star, 2: blackhole, 3: pulsar, 4: gas giant, 5: terrestrial, 6: desert, 7: ice giant, 8: moon, 9: probe
	if strings.Contains(name, "Probe") || strings.Contains(name, "JWST") ||
		strings.Contains(name, "SOHO") || strings.Contains(name, "Satellite") || strings.Contains(name, "Voyager") || (mass < 0.0002 && !isStar) {
		return 9
	}
	if texType == 7 || strings.Contains(name, "Black Hole") || strings.Contains(name, "Singularity") || strings.Contains(name, "Gargantua") {
		return 2
	}
	if isPulsar || texType == 10 || strings.Contains(name, "Pulsar") || strings.Contains(name, "Neutron") {
		return 3
	}
	if isStar || texType == 6 || texType == 8 || texType == 9 || strings.Contains(name, "Star") || strings.Contains(name, "Sun") {
		return 1
	}
	if texType == 1 || strings.Contains(name, "Earth") {
		return 5
	}
	if texType == 2 || strings.Contains(name, "Mars") || strings.Contains(name, "Mercury") {
		return 6
	}
	if texType == 4 || strings.Contains(name, "Neptune") || strings.Contains(name, "Uranus") {
		return 7
	}
	if texType == 3 || strings.Contains(name, "Jupiter") || strings.Contains(name, "Saturn") || (mass >= 15.0 && !isStar) {
		return 4
	}
	if texType == 5 || strings.Contains(name, "Moon") || strings.Contains(name, "Asteroid") || isComet || mass < 0.01 {
		return 8
	}
	return 0
}
