package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// CollisionMode defines how bodies interact when colliding
type CollisionMode int

const (
	CollisionMerge CollisionMode = iota
	CollisionBounce
	CollisionNone
)

// Body represents a celestial body / mass in 3D space
type Body struct {
	ID           int64
	Name         string
	Position     rl.Vector3
	Velocity     rl.Vector3
	Acceleration rl.Vector3
	Mass         float64
	Radius       float32
	Color        rl.Color
	IsStationary bool // If true, fixed in space (infinite mass approximation / anchor)
	IsStar       bool // Emits glow
	Trail        []rl.Vector3
	TrailTimer   float32
	NetForce     rl.Vector3 // For visualization
}

// Config stores global simulation parameters
type Config struct {
	G             float64       // Gravitational constant
	TimeScale     float64       // Simulation speed multiplier (0.1x to 10x)
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
}

// PresetType identifies available galaxy / solar system presets
type PresetType int

const (
	PresetSolarSystem PresetType = iota
	PresetBinaryStars
	PresetThreeBody
	PresetGalaxyDisk
	PresetEmpty
)

// SpawnType defines template for creating new celestial bodies
type SpawnType int

const (
	SpawnEarth SpawnType = iota
	SpawnMoon
	SpawnGasGiant
	SpawnStar
	SpawnBlackHole
)

// SimState holds runtime world state and UI interactions
type SimState struct {
	Bodies          []*Body
	NextID          int64
	SelectedBodyID  int64
	FollowSelected  bool
	CurrentPreset   PresetType
	
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
	Stars           []rl.Vector3
	StarColors      []rl.Color
	StarSizes       []float32
}
