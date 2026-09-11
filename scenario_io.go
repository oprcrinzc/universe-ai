package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ScenarioConfigJSON holds serialized config fields
type ScenarioConfigJSON struct {
	G                float64 `json:"g"`
	TimeScale        float64 `json:"time_scale"`
	SubSteps         int     `json:"sub_steps"`
	Softening        float64 `json:"softening"`
	CollisionMode    string  `json:"collision_mode"`
	UseBarnesHut     bool    `json:"use_barnes_hut"`
	BarnesHutTheta   float32 `json:"barnes_hut_theta"`
	EnableRelativity bool    `json:"enable_relativity"`
	SpeedOfLight     float64 `json:"speed_of_light"`
	EnableRocheLimit bool    `json:"enable_roche_limit"`
	ShowTextures     bool    `json:"show_textures"`
}

// ScenarioBodyJSON holds serialized body parameters
type ScenarioBodyJSON struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Position      [3]float32 `json:"position"`
	Velocity      [3]float32 `json:"velocity"`
	Mass          float64    `json:"mass"`
	Radius        float32    `json:"radius"`
	Color         [4]uint8   `json:"color"`
	IsStationary  bool       `json:"is_stationary"`
	IsStar        bool       `json:"is_star"`
	IsPulsar      bool       `json:"is_pulsar,omitempty"`
	IsComet       bool       `json:"is_comet,omitempty"`
	RingInner     float32    `json:"ring_inner,omitempty"`
	RingOuter     float32    `json:"ring_outer,omitempty"`
	TextureType   string     `json:"texture_type"`
	RotationSpeed float32    `json:"rotation_speed"`
}

// ScenarioFileJSON defines the complete JSON format for exported universes
type ScenarioFileJSON struct {
	Version     string             `json:"version"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Timestamp   int64              `json:"timestamp"`
	Config      ScenarioConfigJSON `json:"config"`
	Bodies      []ScenarioBodyJSON `json:"bodies"`
}

func textureTypeToString(t BodyTextureType) string {
	switch t {
	case TextureTerrestrial:
		return "terrestrial"
	case TextureDesert:
		return "desert"
	case TextureGasGiant:
		return "gas_giant"
	case TextureIceGiant:
		return "ice_giant"
	case TextureMoon:
		return "moon"
	case TextureSun:
		return "sun"
	case TextureBlackHole:
		return "black_hole"
	case TextureRedGiant:
		return "red_giant"
	case TextureWhiteDwarf:
		return "white_dwarf"
	case TextureNeutronStar:
		return "neutron_star"
	case TextureDwarfPlanet:
		return "dwarf_planet"
	case TextureComet:
		return "comet"
	default:
		return "none"
	}
}

func stringToTextureType(s string) BodyTextureType {
	switch strings.ToLower(s) {
	case "terrestrial":
		return TextureTerrestrial
	case "desert":
		return TextureDesert
	case "gas_giant":
		return TextureGasGiant
	case "ice_giant":
		return TextureIceGiant
	case "moon":
		return TextureMoon
	case "sun":
		return TextureSun
	case "black_hole":
		return TextureBlackHole
	case "red_giant":
		return TextureRedGiant
	case "white_dwarf":
		return TextureWhiteDwarf
	case "neutron_star", "pulsar":
		return TextureNeutronStar
	case "dwarf_planet":
		return TextureDwarfPlanet
	case "comet":
		return TextureComet
	default:
		return TextureNone
	}
}

func collisionModeToString(m CollisionMode) string {
	switch m {
	case CollisionMerge:
		return "merge"
	case CollisionBounce:
		return "bounce"
	default:
		return "ghost"
	}
}

func stringToCollisionMode(s string) CollisionMode {
	switch strings.ToLower(s) {
	case "merge":
		return CollisionMerge
	case "bounce":
		return CollisionBounce
	default:
		return CollisionNone
	}
}

// SaveScenarioToFile exports the active simulation state to a formatted JSON file
func SaveScenarioToFile(state *SimState, cfg *Config, filePath string) (string, error) {
	if filePath == "" {
		_ = os.MkdirAll("scenarios", 0755)
		filePath = fmt.Sprintf("scenarios/scenario_%d.json", time.Now().Unix())
	}

	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	bodiesJSON := make([]ScenarioBodyJSON, len(state.Bodies))
	for i, b := range state.Bodies {
		bodiesJSON[i] = ScenarioBodyJSON{
			ID:            b.ID,
			Name:          b.Name,
			Position:      [3]float32{b.Position.X, b.Position.Y, b.Position.Z},
			Velocity:      [3]float32{b.Velocity.X, b.Velocity.Y, b.Velocity.Z},
			Mass:          b.Mass,
			Radius:        b.Radius,
			Color:         [4]uint8{b.Color.R, b.Color.G, b.Color.B, b.Color.A},
			IsStationary:  b.IsStationary,
			IsStar:        b.IsStar,
			IsPulsar:      b.IsPulsar,
			IsComet:       b.IsComet,
			RingInner:     b.RingInnerRadius,
			RingOuter:     b.RingOuterRadius,
			TextureType:   textureTypeToString(b.TextureType),
			RotationSpeed: b.RotationSpeed,
		}
	}

	fileData := ScenarioFileJSON{
		Version:     "1.0",
		Name:        "Custom Universe Scenario",
		Description: fmt.Sprintf("Exported with %d celestial bodies", len(state.Bodies)),
		Timestamp:   time.Now().Unix(),
		Config: ScenarioConfigJSON{
			G:                cfg.G,
			TimeScale:        cfg.TimeScale,
			SubSteps:         cfg.SubSteps,
			Softening:        cfg.Softening,
			CollisionMode:    collisionModeToString(cfg.Collision),
			UseBarnesHut:     cfg.UseBarnesHut,
			BarnesHutTheta:   cfg.BarnesHutTheta,
			EnableRelativity: cfg.EnableRelativity,
			SpeedOfLight:     cfg.SpeedOfLight,
			EnableRocheLimit: cfg.EnableRocheLimit,
			ShowTextures:     cfg.ShowTextures,
		},
		Bodies: bodiesJSON,
	}

	bytes, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal scenario: %w", err)
	}

	err = os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write scenario file: %w", err)
	}

	return filePath, nil
}

// LoadScenarioFromFile imports and initializes universe state from a JSON scenario file
func LoadScenarioFromFile(state *SimState, cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read scenario file %s: %w", filePath, err)
	}

	var scenario ScenarioFileJSON
	if err := json.Unmarshal(data, &scenario); err != nil {
		return fmt.Errorf("failed to parse scenario JSON: %w", err)
	}

	// Apply configuration if valid
	if scenario.Config.G > 0 {
		cfg.G = scenario.Config.G
	}
	if scenario.Config.TimeScale > 0 {
		cfg.TimeScale = scenario.Config.TimeScale
	}
	if scenario.Config.SubSteps > 0 {
		cfg.SubSteps = scenario.Config.SubSteps
	}
	if scenario.Config.Softening > 0 {
		cfg.Softening = scenario.Config.Softening
	}
	cfg.Collision = stringToCollisionMode(scenario.Config.CollisionMode)
	cfg.UseBarnesHut = scenario.Config.UseBarnesHut
	if scenario.Config.BarnesHutTheta > 0.1 {
		cfg.BarnesHutTheta = scenario.Config.BarnesHutTheta
	}
	cfg.EnableRelativity = scenario.Config.EnableRelativity
	if scenario.Config.SpeedOfLight > 10.0 {
		cfg.SpeedOfLight = scenario.Config.SpeedOfLight
	}
	cfg.EnableRocheLimit = scenario.Config.EnableRocheLimit

	// Clear existing state
	state.Bodies = make([]*Body, 0, len(scenario.Bodies))
	state.SelectedBodyID = -1
	state.FollowSelected = false
	var maxID int64 = 0

	for _, b := range scenario.Bodies {
		if b.ID > maxID {
			maxID = b.ID
		}
		newBody := &Body{
			ID:              b.ID,
			Name:            b.Name,
			Position:        rl.NewVector3(b.Position[0], b.Position[1], b.Position[2]),
			Velocity:        rl.NewVector3(b.Velocity[0], b.Velocity[1], b.Velocity[2]),
			Mass:            b.Mass,
			Radius:          b.Radius,
			Color:           rl.NewColor(b.Color[0], b.Color[1], b.Color[2], b.Color[3]),
			IsStationary:    b.IsStationary,
			IsStar:          b.IsStar,
			IsPulsar:        b.IsPulsar,
			IsComet:         b.IsComet,
			RingInnerRadius: b.RingInner,
			RingOuterRadius: b.RingOuter,
			Trail:           make([]rl.Vector3, 0, 100),
			TextureType:     stringToTextureType(b.TextureType),
			RotationAngle:   0,
			RotationSpeed:   b.RotationSpeed,
		}
		state.Bodies = append(state.Bodies, newBody)
	}

	state.NextID = maxID + 1
	cfg.NotificationText = fmt.Sprintf("Loaded scenario: %s (%d bodies)", filepath.Base(filePath), len(state.Bodies))
	cfg.NotificationTimer = 4.0

	return nil
}

// ScanScenariosDirectory returns a sorted list of all available JSON scenario files in scenarios/
func ScanScenariosDirectory() []string {
	_ = os.MkdirAll("scenarios", 0755)
	entries, err := os.ReadDir("scenarios")
	if err != nil {
		return nil
	}

	var results []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			results = append(results, filepath.Join("scenarios", e.Name()))
		}
	}
	sort.Strings(results)
	return results
}

// InitBuiltinScenarios writes pre-configured iconic astrophysics scenarios into scenarios/
func InitBuiltinScenarios() {
	_ = os.MkdirAll("scenarios", 0755)

	// 1. TRAPPIST-1 System
	trappistPath := "scenarios/trappist_1.json"
	if _, err := os.Stat(trappistPath); os.IsNotExist(err) {
		trappist := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "TRAPPIST-1 Resonant Exoplanet System",
			Description: "Ultra-cool red dwarf star with 7 resonant terrestrial planets in tight orbits",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G:                1.0,
				TimeScale:        1.0,
				SubSteps:         5,
				Softening:        0.8,
				CollisionMode:    "merge",
				UseBarnesHut:     false,
				BarnesHutTheta:   0.7,
				EnableRelativity: false,
				SpeedOfLight:     200.0,
				EnableRocheLimit: true,
				ShowTextures:     true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "TRAPPIST-1 (Red Dwarf)", Position: [3]float32{0, 0, 0}, Velocity: [3]float32{0, 0, 0}, Mass: 8000.0, Radius: 3.5, Color: [4]uint8{255, 90, 40, 255}, IsStationary: true, IsStar: true, TextureType: "sun"},
				{ID: 2, Name: "TRAPPIST-1b", Position: [3]float32{11.0, 0, 0}, Velocity: [3]float32{0, 0, -26.96}, Mass: 1.02, Radius: 0.95, Color: [4]uint8{210, 160, 120, 255}, TextureType: "desert"},
				{ID: 3, Name: "TRAPPIST-1c", Position: [3]float32{15.2, 0, 0}, Velocity: [3]float32{0, 0, -22.94}, Mass: 1.16, Radius: 0.98, Color: [4]uint8{180, 190, 210, 255}, TextureType: "terrestrial"},
				{ID: 4, Name: "TRAPPIST-1d", Position: [3]float32{21.5, 0, 0}, Velocity: [3]float32{0, 0, -19.29}, Mass: 0.30, Radius: 0.72, Color: [4]uint8{160, 210, 240, 255}, TextureType: "terrestrial"},
				{ID: 5, Name: "TRAPPIST-1e (Habitable)", Position: [3]float32{28.0, 0, 0}, Velocity: [3]float32{0, 0, -16.90}, Mass: 0.77, Radius: 0.88, Color: [4]uint8{70, 170, 255, 255}, TextureType: "terrestrial"},
				{ID: 6, Name: "TRAPPIST-1f", Position: [3]float32{36.0, 0, 0}, Velocity: [3]float32{0, 0, -14.90}, Mass: 0.93, Radius: 0.94, Color: [4]uint8{140, 200, 220, 255}, TextureType: "terrestrial"},
				{ID: 7, Name: "TRAPPIST-1g", Position: [3]float32{45.0, 0, 0}, Velocity: [3]float32{0, 0, -13.33}, Mass: 1.15, Radius: 1.02, Color: [4]uint8{130, 160, 200, 255}, TextureType: "ice_giant"},
				{ID: 8, Name: "TRAPPIST-1h", Position: [3]float32{58.0, 0, 0}, Velocity: [3]float32{0, 0, -11.74}, Mass: 0.33, Radius: 0.74, Color: [4]uint8{190, 200, 220, 255}, TextureType: "moon"},
			},
		}
		bytes, _ := json.MarshalIndent(trappist, "", "  ")
		_ = os.WriteFile(trappistPath, bytes, 0644)
	}

	// 2. Stable Figure-8 Three-Body Orbit (Chenciner & Montgomery)
	fig8Path := "scenarios/figure8_threebody.json"
	if _, err := os.Stat(fig8Path); os.IsNotExist(err) {
		m := 2500.0
		// Exact normalized velocities and positions scaled for our engine
		x1, y1 := float32(0.97000436*25.0), float32(-0.24308753*25.0)
		vx3, vy3 := float32(-2.0 * -0.46620531 * 5.0), float32(-2.0 * -0.43236573 * 5.0)
		vx1, vy1 := float32(-0.46620531*5.0), float32(-0.43236573*5.0)

		fig8 := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Stable Figure-8 Three-Body Choreography",
			Description: "Three equal masses tracing a single continuous figure-eight planar trajectory",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G:                1.0,
				TimeScale:        1.0,
				SubSteps:         8,
				Softening:        0.1,
				CollisionMode:    "bounce",
				UseBarnesHut:     false,
				BarnesHutTheta:   0.7,
				EnableRelativity: false,
				SpeedOfLight:     200.0,
				EnableRocheLimit: false,
				ShowTextures:     true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "Choreography Star Alpha", Position: [3]float32{x1, 0, y1}, Velocity: [3]float32{vx1, 0, vy1}, Mass: m, Radius: 2.2, Color: [4]uint8{255, 200, 50, 255}, IsStar: true, TextureType: "sun"},
				{ID: 2, Name: "Choreography Star Beta", Position: [3]float32{-x1, 0, -y1}, Velocity: [3]float32{vx1, 0, vy1}, Mass: m, Radius: 2.2, Color: [4]uint8{100, 220, 255, 255}, IsStar: true, TextureType: "sun"},
				{ID: 3, Name: "Choreography Star Gamma", Position: [3]float32{0, 0, 0}, Velocity: [3]float32{vx3, 0, vy3}, Mass: m, Radius: 2.2, Color: [4]uint8{255, 100, 200, 255}, IsStar: true, TextureType: "sun"},
			},
		}
		bytes, _ := json.MarshalIndent(fig8, "", "  ")
		_ = os.WriteFile(fig8Path, bytes, 0644)
	}

	// 3. Roche Limit Breakup Scenario
		rochePath := "scenarios/roche_breakup.json"
	if _, err := os.Stat(rochePath); os.IsNotExist(err) {
		rocheScen := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Roche Limit Moon Tidal Breakup",
			Description: "An incoming icy moon plunging past a massive gas giant into its Roche disruption threshold",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G:                1.0,
				TimeScale:        1.0,
				SubSteps:         5,
				Softening:        0.8,
				CollisionMode:    "merge",
				UseBarnesHut:     false,
				BarnesHutTheta:   0.7,
				EnableRelativity: false,
				SpeedOfLight:     200.0,
				EnableRocheLimit: true,
				ShowTextures:     true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "Gas Giant Kronos", Position: [3]float32{0, 0, 0}, Velocity: [3]float32{0, 0, 0}, Mass: 16000.0, Radius: 5.5, Color: [4]uint8{225, 175, 110, 255}, IsStationary: true, TextureType: "gas_giant"},
				{ID: 2, Name: "Outer Moon Rhea", Position: [3]float32{65.0, 0, 0}, Velocity: [3]float32{0, 0, -15.68}, Mass: 1.5, Radius: 0.9, Color: [4]uint8{200, 210, 230, 255}, TextureType: "moon"},
				{ID: 3, Name: "Doomed Moon Boreas", Position: [3]float32{48.0, 0, 32.0}, Velocity: [3]float32{-14.2, 0, -1.5}, Mass: 2.0, Radius: 1.1, Color: [4]uint8{220, 235, 255, 255}, TextureType: "moon"},
			},
		}
		bytes, _ := json.MarshalIndent(rocheScen, "", "  ")
		_ = os.WriteFile(rochePath, bytes, 0644)
	}

	// 4. Lagrange Points L4 & L5 Trojan Asteroids
	trojanPath := "scenarios/lagrange_trojans.json"
	if _, err := os.Stat(trojanPath); os.IsNotExist(err) {
		sunMass := 15000.0
		jupDist := float32(50.0)
		jupSpeed := float32(math.Sqrt(sunMass / float64(jupDist)))

		trojanBodies := []ScenarioBodyJSON{
			{ID: 1, Name: "Helios (Central Star)", Position: [3]float32{0, 0, 0}, Velocity: [3]float32{0, 0, 0}, Mass: sunMass, Radius: 4.0, Color: [4]uint8{255, 200, 60, 255}, IsStationary: true, IsStar: true, TextureType: "sun"},
			{ID: 2, Name: "Gas Giant Zeus", Position: [3]float32{jupDist, 0, 0}, Velocity: [3]float32{0, 0, -jupSpeed}, Mass: 45.0, Radius: 2.4, Color: [4]uint8{220, 160, 100, 255}, TextureType: "gas_giant"},
		}

		// L4 Trojan cluster (+60 degrees ahead)
		idCounter := int64(3)
		for i := 0; i < 6; i++ {
			ang := math.Pi/3.0 + (float64(i)-2.5)*0.06
			r := float64(jupDist) + (float64(i%3)-1.0)*1.8
			spd := math.Sqrt(sunMass / r)
			x := float32(r * math.Cos(ang))
			z := float32(r * math.Sin(ang))
			vx := float32(-spd * math.Sin(ang))
			vz := float32(spd * math.Cos(ang))

			trojanBodies = append(trojanBodies, ScenarioBodyJSON{
				ID: idCounter, Name: fmt.Sprintf("Trojan #%d (L4)", i+1),
				Position: [3]float32{x, 0, z}, Velocity: [3]float32{vx, 0, vz},
				Mass: 0.05, Radius: 0.45, Color: [4]uint8{160, 200, 240, 255}, TextureType: "moon",
			})
			idCounter++
		}

		// L5 Greek cluster (-60 degrees behind)
		for i := 0; i < 6; i++ {
			ang := -math.Pi/3.0 + (float64(i)-2.5)*0.06
			r := float64(jupDist) + (float64(i%3)-1.0)*1.8
			spd := math.Sqrt(sunMass / r)
			x := float32(r * math.Cos(ang))
			z := float32(r * math.Sin(ang))
			vx := float32(-spd * math.Sin(ang))
			vz := float32(spd * math.Cos(ang))

			trojanBodies = append(trojanBodies, ScenarioBodyJSON{
				ID: idCounter, Name: fmt.Sprintf("Greek #%d (L5)", i+1),
				Position: [3]float32{x, 0, z}, Velocity: [3]float32{vx, 0, vz},
				Mass: 0.05, Radius: 0.45, Color: [4]uint8{240, 190, 140, 255}, TextureType: "moon",
			})
			idCounter++
		}

		trojanScen := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Lagrange L4 & L5 Trojan Asteroid Swarms",
			Description: "Stable gravitational equilateral Lagrangian equilibrium points with Greek and Trojan asteroid swarms",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G: 1.0, TimeScale: 1.0, SubSteps: 6, Softening: 0.5,
				CollisionMode: "merge", ShowTextures: true,
			},
			Bodies: trojanBodies,
		}
		bytes, _ := json.MarshalIndent(trojanScen, "", "  ")
		_ = os.WriteFile(trojanPath, bytes, 0644)
	}

	// 5. Galaxy Hyperbolic Collision & Tidal Bridges
	collPath := "scenarios/galaxy_collision.json"
	if _, err := os.Stat(collPath); os.IsNotExist(err) {
		collScen := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Spiral Galaxy Hyperbolic Collision",
			Description: "Two colliding disk galaxies generating dramatic tidal bridges and stellar ejections",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G: 1.0, TimeScale: 1.2, SubSteps: 5, Softening: 0.8,
				CollisionMode: "merge", UseBarnesHut: true, BarnesHutTheta: 0.7,
				ShowTextures: true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "SMBH Andromeda", Position: [3]float32{-45.0, 0, -20.0}, Velocity: [3]float32{1.8, 0.2, 1.2}, Mass: 18000.0, Radius: 3.8, Color: [4]uint8{30, 20, 50, 255}, IsStar: true, TextureType: "black_hole"},
				{ID: 2, Name: "SMBH Milky Way", Position: [3]float32{45.0, 0, 20.0}, Velocity: [3]float32{-1.8, -0.2, -1.2}, Mass: 14000.0, Radius: 3.5, Color: [4]uint8{20, 30, 60, 255}, IsStar: true, TextureType: "black_hole"},
			},
		}

		// Add orbital stars for Galaxy 1
		idCnt := int64(3)
		for i := 0; i < 18; i++ {
			r := 12.0 + float64(i)*2.2
			ang := float64(i) * 1.3
			spd := math.Sqrt(18000.0 / r)
			collScen.Bodies = append(collScen.Bodies, ScenarioBodyJSON{
				ID: idCnt, Name: fmt.Sprintf("Andromeda Star %d", i+1),
				Position: [3]float32{float32(-45.0 + r*math.Cos(ang)), float32((i%3)*2 - 2), float32(-20.0 + r*math.Sin(ang))},
				Velocity: [3]float32{float32(1.8 - spd*math.Sin(ang)), 0.2, float32(1.2 + spd*math.Cos(ang))},
				Mass: 1.0, Radius: 0.5, Color: [4]uint8{160, 210, 255, 255}, TextureType: "moon",
			})
			idCnt++
		}

		// Add orbital stars for Galaxy 2 (inclined plane)
		for i := 0; i < 18; i++ {
			r := 10.0 + float64(i)*2.0
			ang := float64(i) * 1.4
			spd := math.Sqrt(14000.0 / r)
			collScen.Bodies = append(collScen.Bodies, ScenarioBodyJSON{
				ID: idCnt, Name: fmt.Sprintf("Milky Way Star %d", i+1),
				Position: [3]float32{float32(45.0 + r*math.Cos(ang)), float32(r * 0.3 * math.Sin(ang)), float32(20.0 + r*math.Sin(ang))},
				Velocity: [3]float32{float32(-1.8 + spd*math.Sin(ang)), -0.2, float32(-1.2 - spd*math.Cos(ang))},
				Mass: 0.8, Radius: 0.45, Color: [4]uint8{255, 220, 160, 255}, TextureType: "moon",
			})
			idCnt++
		}

		bytes, _ := json.MarshalIndent(collScen, "", "  ")
		_ = os.WriteFile(collPath, bytes, 0644)
	}

	// 6. Relativistic Pulsar & Red Giant Accretion Binary
	pulsarPath := "scenarios/pulsar_accretion.json"
	if _, err := os.Stat(pulsarPath); os.IsNotExist(err) {
		pulsarScen := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Relativistic Pulsar & Red Giant Binary",
			Description: "Millisecond pulsar with relativistic polar jets accreting matter from a swelling Red Supergiant",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G: 1.0, TimeScale: 1.0, SubSteps: 6, Softening: 0.6,
				CollisionMode: "merge", EnableRelativity: true, SpeedOfLight: 170.0,
				ShowTextures: true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "Pulsar PSR-J0737", Position: [3]float32{-14.0, 0, 0}, Velocity: [3]float32{0, 0, 11.2}, Mass: 5500.0, Radius: 1.2, Color: [4]uint8{140, 200, 255, 255}, IsStar: true, IsPulsar: true, TextureType: "neutron_star", RotationSpeed: 6.0},
				{ID: 2, Name: "Red Giant Betelgeuse II", Position: [3]float32{11.0, 0, 0}, Velocity: [3]float32{0, 0, -8.8}, Mass: 7000.0, Radius: 4.8, Color: [4]uint8{255, 80, 40, 255}, IsStar: true, TextureType: "red_giant", RotationSpeed: 0.3},
				{ID: 3, Name: "Accretion Clump 1", Position: [3]float32{-2.0, 0, 2.0}, Velocity: [3]float32{-6.5, 0, 4.0}, Mass: 0.1, Radius: 0.35, Color: [4]uint8{255, 180, 100, 255}, TextureType: "moon"},
				{ID: 4, Name: "Accretion Clump 2", Position: [3]float32{-5.0, 0, -3.0}, Velocity: [3]float32{5.0, 0, -7.0}, Mass: 0.1, Radius: 0.35, Color: [4]uint8{255, 180, 100, 255}, TextureType: "moon"},
				{ID: 5, Name: "Accretion Clump 3", Position: [3]float32{-8.0, 0, 1.5}, Velocity: [3]float32{-8.0, 0, -3.5}, Mass: 0.1, Radius: 0.35, Color: [4]uint8{255, 180, 100, 255}, TextureType: "moon"},
			},
		}
		bytes, _ := json.MarshalIndent(pulsarScen, "", "  ")
		_ = os.WriteFile(pulsarPath, bytes, 0644)
	}

	// 7. Grand Solar System
	grandPath := "scenarios/solar_system_grand.json"
	if _, err := os.Stat(grandPath); os.IsNotExist(err) {
		sunM := 12000.0
		grandScen := ScenarioFileJSON{
			Version:     "1.0",
			Name:        "Grand Solar System (All Planets + Belt + Moons)",
			Description: "Comprehensive solar system simulation including all 8 planets, Main Asteroid Belt, major moons, and Halley's comet",
			Timestamp:   time.Now().Unix(),
			Config: ScenarioConfigJSON{
				G: 1.0, TimeScale: 1.0, SubSteps: 6, Softening: 0.8,
				CollisionMode: "merge", ShowTextures: true,
			},
			Bodies: []ScenarioBodyJSON{
				{ID: 1, Name: "Sun", Position: [3]float32{0, 0, 0}, Velocity: [3]float32{0, 0, 0}, Mass: sunM, Radius: 4.5, Color: [4]uint8{255, 215, 0, 255}, IsStationary: true, IsStar: true, TextureType: "sun"},
				{ID: 2, Name: "Mercury", Position: [3]float32{12.0, 0, 0}, Velocity: [3]float32{0, 0, -31.62}, Mass: 0.4, Radius: 0.65, Color: [4]uint8{170, 170, 170, 255}, TextureType: "moon"},
				{ID: 3, Name: "Venus", Position: [3]float32{19.0, 0, 0}, Velocity: [3]float32{0, 0, -25.13}, Mass: 1.2, Radius: 0.95, Color: [4]uint8{225, 200, 140, 255}, TextureType: "desert"},
				{ID: 4, Name: "Earth", Position: [3]float32{28.0, 0, 0}, Velocity: [3]float32{0, 0, -20.70}, Mass: 2.0, Radius: 1.15, Color: [4]uint8{70, 150, 255, 255}, TextureType: "terrestrial"},
				{ID: 5, Name: "Moon", Position: [3]float32{30.5, 0, 0}, Velocity: [3]float32{0, 0, -21.60}, Mass: 0.05, Radius: 0.35, Color: [4]uint8{190, 190, 190, 255}, TextureType: "moon"},
				{ID: 6, Name: "Mars", Position: [3]float32{38.0, 0, 0}, Velocity: [3]float32{0, 0, -17.77}, Mass: 0.7, Radius: 0.82, Color: [4]uint8{230, 80, 45, 255}, TextureType: "desert"},
				{ID: 7, Name: "Ceres (Dwarf Planet)", Position: [3]float32{44.0, 0, 0}, Velocity: [3]float32{0, 0, -16.51}, Mass: 0.05, Radius: 0.45, Color: [4]uint8{180, 170, 160, 255}, TextureType: "dwarf_planet"},
				{ID: 8, Name: "Jupiter", Position: [3]float32{58.0, 0, 0}, Velocity: [3]float32{0, 0, -14.38}, Mass: 25.0, Radius: 2.6, Color: [4]uint8{220, 175, 125, 255}, TextureType: "gas_giant"},
				{ID: 9, Name: "Ganymede", Position: [3]float32{62.5, 0, 0}, Velocity: [3]float32{0, 0, -16.74}, Mass: 0.03, Radius: 0.38, Color: [4]uint8{200, 210, 220, 255}, TextureType: "moon"},
				{ID: 10, Name: "Saturn", Position: [3]float32{85.0, 0, 0}, Velocity: [3]float32{0, 0, -11.88}, Mass: 18.0, Radius: 2.2, Color: [4]uint8{235, 210, 160, 255}, RingInner: 3.0, RingOuter: 5.2, TextureType: "gas_giant"},
				{ID: 11, Name: "Titan", Position: [3]float32{89.8, 0, 0}, Velocity: [3]float32{0, 0, -13.82}, Mass: 0.03, Radius: 0.38, Color: [4]uint8{240, 200, 120, 255}, TextureType: "moon"},
				{ID: 12, Name: "Uranus", Position: [3]float32{115.0, 0, 0}, Velocity: [3]float32{0, 0, -10.22}, Mass: 8.0, Radius: 1.6, Color: [4]uint8{140, 220, 235, 255}, TextureType: "ice_giant"},
				{ID: 13, Name: "Neptune", Position: [3]float32{145.0, 0, 0}, Velocity: [3]float32{0, 0, -9.10}, Mass: 9.0, Radius: 1.6, Color: [4]uint8{75, 110, 245, 255}, TextureType: "ice_giant"},
				{ID: 14, Name: "Pluto (Dwarf Planet)", Position: [3]float32{180.0, 0, 0}, Velocity: [3]float32{0, 0, -8.16}, Mass: 0.04, Radius: 0.45, Color: [4]uint8{215, 185, 155, 255}, TextureType: "dwarf_planet"},
				{ID: 15, Name: "Comet Halley", Position: [3]float32{110.0, 16.0, 35.0}, Velocity: [3]float32{-3.5, -1.2, -4.5}, Mass: 0.01, Radius: 0.4, Color: [4]uint8{190, 240, 255, 255}, IsComet: true, TextureType: "comet"},
			},
		}

		// Add 10 Asteroid Belt particles
		astID := int64(16)
		for i := 0; i < 10; i++ {
			ang := float64(i) * (2.0 * math.Pi / 10.0)
			r := 42.0 + float64(i%4)*1.5
			spd := math.Sqrt(sunM / r)
			grandScen.Bodies = append(grandScen.Bodies, ScenarioBodyJSON{
				ID: astID, Name: fmt.Sprintf("Asteroid #%d", i+1),
				Position: [3]float32{float32(r * math.Cos(ang)), float32((i%3)-1) * 0.4, float32(r * math.Sin(ang))},
				Velocity: [3]float32{float32(-spd * math.Sin(ang)), 0, float32(spd * math.Cos(ang))},
				Mass: 0.002, Radius: 0.22, Color: [4]uint8{180, 175, 165, 255}, TextureType: "moon",
			})
			astID++
		}

		bytes, _ := json.MarshalIndent(grandScen, "", "  ")
		_ = os.WriteFile(grandPath, bytes, 0644)
	}

	// 8. Milky Way Extreme (10,000 Particles)
	mwPath := "scenarios/milky_way_10000.json"
	if _, err := os.Stat(mwPath); os.IsNotExist(err) {
		tempState := &SimState{}
		tempCfg := &Config{G: 1.0, TimeScale: 1.0, SubSteps: 1, Softening: 0.8, UseBarnesHut: true, BarnesHutTheta: 0.72}
		loadMilkyWay10K(tempState, tempCfg, 1.0)
		_, _ = SaveScenarioToFile(tempState, tempCfg, mwPath)
	}

	// 9. Asteroid Belt & Trojan Swarm (2,000 Asteroids)
	beltPath := "scenarios/asteroid_belt_2000.json"
	if _, err := os.Stat(beltPath); os.IsNotExist(err) {
		tempState := &SimState{}
		tempCfg := &Config{G: 1.0, TimeScale: 1.0, SubSteps: 1, Softening: 0.8, UseBarnesHut: true, BarnesHutTheta: 0.72}
		loadAsteroidBelt2K(tempState, tempCfg, 1.0)
		_, _ = SaveScenarioToFile(tempState, tempCfg, beltPath)
	}

	// 10. Galaxy Collision (5,000 Stars)
	coll5kPath := "scenarios/galaxy_collision_5000.json"
	if _, err := os.Stat(coll5kPath); os.IsNotExist(err) {
		tempState := &SimState{}
		tempCfg := &Config{G: 1.0, TimeScale: 1.0, SubSteps: 1, Softening: 0.8, UseBarnesHut: true, BarnesHutTheta: 0.72}
		loadGalaxyCollision5K(tempState, tempCfg, 1.0)
		_, _ = SaveScenarioToFile(tempState, tempCfg, coll5kPath)
	}

	// 11. Gargantua Black Hole Swarm (3,000 Relativistic Bodies)
	bhSwarmPath := "scenarios/blackhole_swarm_3000.json"
	if _, err := os.Stat(bhSwarmPath); os.IsNotExist(err) {
		tempState := &SimState{}
		tempCfg := &Config{G: 1.0, TimeScale: 1.0, SubSteps: 1, Softening: 0.8, UseBarnesHut: true, BarnesHutTheta: 0.72, EnableRelativity: true, SpeedOfLight: 190.0}
		loadBlackHoleSwarm3K(tempState, tempCfg, 1.0)
		_, _ = SaveScenarioToFile(tempState, tempCfg, bhSwarmPath)
	}
}
