package main

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
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Position      Vector3   `json:"pos"`
	Velocity      Vector3   `json:"vel"`
	Acceleration  Vector3   `json:"acc"`
	Mass          float64   `json:"mass"`
	Radius        float32   `json:"radius"`
	Color         Color     `json:"color"`
	Trail         []Vector3 `json:"trail,omitempty"`
	TrailHead     int       `json:"-"`
	TrailLen      int       `json:"-"`
	IsStationary  bool      `json:"is_stationary"`
	IsStar        bool      `json:"is_star"`
	IsPulsar      bool      `json:"is_pulsar"`
	IsBlackHole   bool      `json:"is_blackhole"`
	IsComet       bool      `json:"is_comet"`
	RingInnerRadius float32   `json:"ring_inner_radius,omitempty"`
	RingOuterRadius float32   `json:"ring_outer_radius,omitempty"`
	TextureType   int       `json:"texture_type"`
	RotationAngle float32   `json:"rotation_angle"`
}

type Config struct {
	G                  float64       `json:"g"`
	TimeScale          float64       `json:"time_scale"`
	TimeScaleStep      float64       `json:"time_scale_step"`
	SubSteps           int           `json:"sub_steps"`
	Softening          float64       `json:"softening"`
	Collision          CollisionMode `json:"collision"`
	Paused             bool          `json:"paused"`
	EnableBarnesHut    bool          `json:"enable_barnes_hut"`
	BarnesHutTheta     float64       `json:"barnes_hut_theta"`
	EnableRelativity   bool          `json:"enable_relativity"`
	SpeedOfLight       float64       `json:"speed_of_light"`
	EnableRocheLimit   bool          `json:"enable_roche_limit"`
	ParticleGlowMode   bool          `json:"particle_glow_mode"`
	ShowPotentialGrid  bool          `json:"show_potential_grid"`
	SpacetimeScale     float32       `json:"spacetime_scale"`
	SpacetimeResolution int          `json:"spacetime_resolution"`
	ShowVectorField    bool          `json:"show_vector_field"`
	Show2DLabels       bool          `json:"show_2d_labels"`
	Show3DHeatmapPlane bool          `json:"show_3d_heatmap_plane"`
	ShowLagrangePoints bool          `json:"show_lagrange_points"`
}

type SimState struct {
	Bodies             []*Body `json:"bodies"`
	NextID             int     `json:"-"`
	Time               float64 `json:"time"`
	SelectedBodyID     int     `json:"selected_id"`
	CollisionCount     int     `json:"collisions"`
	MergeCount         int     `json:"merges"`
	TidalBreakupCount  int     `json:"tidal_breakups"`
	OctreeDepth        int     `json:"octree_depth"`
	InteractionsPerSec int64   `json:"interactions_per_sec"`
	PhysicsTimeMs      float64 `json:"physics_time_ms"`
}
