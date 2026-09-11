package main

import (
	"os"
	"path/filepath"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestScenarioExportAndImport(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_scenario.json")

	cfg := &Config{
		G:                1.5,
		TimeScale:        2.0,
		SubSteps:         6,
		Softening:        0.75,
		Collision:        CollisionBounce,
		UseBarnesHut:     true,
		BarnesHutTheta:   0.65,
		EnableRelativity: true,
		SpeedOfLight:     220.0,
		EnableRocheLimit: true,
		ShowTextures:     true,
	}

	state := &SimState{
		NextID: 10,
		Bodies: []*Body{
			{
				ID:           1,
				Name:         "Central Star",
				Position:     rl.NewVector3(0, 0, 0),
				Velocity:     rl.NewVector3(0, 0, 0),
				Mass:         10000.0,
				Radius:       4.0,
				Color:        rl.Gold,
				IsStationary: true,
				IsStar:       true,
				TextureType:  TextureSun,
			},
			{
				ID:           2,
				Name:         "Habitable World",
				Position:     rl.NewVector3(30, 0, 0),
				Velocity:     rl.NewVector3(0, 0, -18.25),
				Mass:         2.5,
				Radius:       1.2,
				Color:        rl.Blue,
				IsStationary: false,
				IsStar:       false,
				TextureType:  TextureTerrestrial,
			},
		},
	}

	// 1. Export
	outPath, err := SaveScenarioToFile(state, cfg, filePath)
	if err != nil {
		t.Fatalf("SaveScenarioToFile failed: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist: %s", outPath)
	}

	// 2. Import into fresh state
	newState := &SimState{}
	newCfg := &Config{}

	err = LoadScenarioFromFile(newState, newCfg, outPath)
	if err != nil {
		t.Fatalf("LoadScenarioFromFile failed: %v", err)
	}

	// 3. Assertions
	if len(newState.Bodies) != 2 {
		t.Fatalf("Expected 2 bodies, got %d", len(newState.Bodies))
	}

	if newCfg.G != 1.5 || newCfg.TimeScale != 2.0 || !newCfg.UseBarnesHut || !newCfg.EnableRelativity {
		t.Fatalf("Config values did not deserialize properly: %+v", newCfg)
	}

	b1 := newState.Bodies[0]
	if b1.Name != "Central Star" || b1.Mass != 10000.0 || !b1.IsStationary || b1.TextureType != TextureSun {
		t.Fatalf("Body 1 mismatch: %+v", b1)
	}

	b2 := newState.Bodies[1]
	if b2.Name != "Habitable World" || b2.Position.X != 30 || b2.Velocity.Z != -18.25 || b2.TextureType != TextureTerrestrial {
		t.Fatalf("Body 2 mismatch: %+v", b2)
	}
}

func TestBuiltinScenarioFiles(t *testing.T) {
	InitBuiltinScenarios()
	scens := ScanScenariosDirectory()
	if len(scens) < 3 {
		t.Fatalf("Expected at least 3 built-in scenarios, got %d", len(scens))
	}

	// Verify all built-ins can be parsed cleanly
	for _, p := range scens {
		st := &SimState{}
		cf := &Config{}
		if err := LoadScenarioFromFile(st, cf, p); err != nil {
			t.Fatalf("Failed to parse built-in scenario %s: %v", p, err)
		}
		if len(st.Bodies) == 0 {
			t.Fatalf("Built-in scenario %s has 0 bodies", p)
		}
	}
}
