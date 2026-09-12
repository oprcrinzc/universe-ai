package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// IntersectRayPlane finds intersection point of a ray with the horizontal plane Y = planeY
func IntersectRayPlane(ray rl.Ray, planeY float32) (rl.Vector3, bool) {
	if math.Abs(float64(ray.Direction.Y)) < 0.0001 {
		return rl.NewVector3(0, 0, 0), false
	}
	t := (planeY - ray.Position.Y) / ray.Direction.Y
	if t < 0 {
		return rl.NewVector3(0, 0, 0), false
	}
	hit := rl.Vector3Add(ray.Position, rl.Vector3Scale(ray.Direction, t))
	return hit, true
}

func main() {
	// Handle headless / server flags before initializing OpenGL
	for i, arg := range os.Args[1:] {
		if arg == "--web" || arg == "-w" {
			port := 8080
			if i+2 < len(os.Args) {
				var p int
				if _, err := fmt.Sscanf(os.Args[i+2], "%d", &p); err == nil && p > 0 {
					port = p
				}
			}
			runWebServer(port)
			return
		}
	}

	screenWidth := int32(1920)
	screenHeight := int32(1080)

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "Gravity Mass Simulator 3D - 144 FPS Universe Engine")
	defer rl.CloseWindow()

	// Initialize modern font manager
	InitFontManager()
	defer UnloadFontManager()

	// Initialize procedural celestial textures and unit sphere model
	InitTextureManager()
	defer UnloadTextureManager()

	// Initialize GPU Compute Shader acceleration pipeline
	gpuOk := InitGPUCompute()
	defer CleanupGPUCompute()

	for _, arg := range os.Args[1:] {
		if arg == "--verify-gpu" || arg == "--test-gpu" {
			VerifyGPUComputeDirect()
			return
		}
	}

	// Ensure built-in scenario templates are written to scenarios/
	InitBuiltinScenarios()

	rl.SetTargetFPS(144)

	// Simulation configuration with roadmap features
	cfg := &Config{
		G:                   1.0,
		TimeScale:           1.0,
		TimeScaleStep:       0.5,
		SubSteps:            2,
		Softening:           0.8,
		Collision:           CollisionMerge,
		Paused:              false,
		ShowTrails:          true,
		ShowGrid:            false,
		ShowPotentialGrid:   false,
		SpacetimeScale:      3.0,
		SpacetimeResolution: 512,
		ShowVectorField:     false,
		CinematicCamera:     false,
		ParticleGlowMode:    false,
		AutoPerformanceMode: true,
		ShowVectors:         false,
		ShowForces:          false,
		ShowLabels:          false,
		ShowStarfield:       true,
		ShowHelp:            false,
		UseBarnesHut:        true,
		BarnesHutTheta:      0.72,
		UseGPUCompute:       gpuOk,
		GPUAvailable:        gpuOk,
		EnableRelativity:    false,
		SpeedOfLight:        180.0,
		EnableRocheLimit:    true,
		ShowTextures:        true,
		Integrator:          IntegratorYoshida4,
		Show2DViewport:      true,
		Show2DLabels:        true,
		Show2DIcons:         false, // Default: clean, clear circle mode (pictures/glyphs hidden until user toggles [ICON])
		Show2DCircles:       true,  // Default: body overlay outline circles enabled in 2D
		Show2DHeatmap:       true,
		Heatmap2DResolution: 64, // Default: 64 columns
		Show3DHeatmapPlane:  false,
		Show2DVectorField:   true,
		Viewport2DZoom:      0.45,
		RealScaleVisualMode: false,
		ShowLagrangePoints:  false,
		Offload3D:           false,
		Track2D:             false,
		Show2DGW:            true,
		ShowStatsHUD:        true,
		NotificationText:    "Gravity Simulator 3D (144 FPS / 1080p). F5/Z: 2D Main Offload | F1: Manual",
		NotificationTimer:   5.0,
	}

	// Simulation state
	state := &SimState{
		SelectedBodyID:     -1,
		SpawnPreset:        SpawnEarth,
		AvailableScenarios: ScanScenariosDirectory(),
	}

	// Initialize background stars
	InitStarfield(state, 500)

	// Load default preset: Solar System
	LoadPreset(state, cfg, PresetSolarSystem)

	// Orbit camera
	camera := NewOrbitCamera()

	// Handle command line flags
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--capture-demo" {
			RunDemoCapture(state, cfg, camera, screenWidth, screenHeight)
			return
		} else if arg == "--verify-gpu" || arg == "--test-gpu" {
			VerifyGPUComputeDirect()
			return
		} else if (arg == "--load" || arg == "-l") && i+1 < len(os.Args) {
			scenFile := os.Args[i+1]
			i++
			if err := LoadScenarioFromFile(state, cfg, scenFile); err != nil {
				fmt.Printf("Failed to load scenario %s: %v\n", scenFile, err)
			}
		}
	}

	for !rl.WindowShouldClose() {
		w := int32(rl.GetScreenWidth())
		h := int32(rl.GetScreenHeight())
		mousePos := rl.GetMousePosition()
		dt := rl.GetFrameTime()
		if dt > 0.05 {
			dt = 0.05
		}

		// Global Hotkeys
		if rl.IsKeyPressed(rl.KeySpace) {
			cfg.Paused = !cfg.Paused
		}
		if rl.IsKeyPressed(rl.KeyComma) {
			step := cfg.TimeScaleStep
			if step <= 0 {
				step = 0.5
			}
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				step = 0.1
			}
			newSpeed := math.Round((cfg.TimeScale-step)*10.0) / 10.0
			cfg.TimeScale = math.Max(0.1, newSpeed)
			cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (-%.1fx)", cfg.TimeScale, step)
			cfg.NotificationTimer = 1.2
		}
		if rl.IsKeyPressed(rl.KeyPeriod) {
			step := cfg.TimeScaleStep
			if step <= 0 {
				step = 0.5
			}
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				step = 0.1
			}
			newSpeed := math.Round((cfg.TimeScale+step)*10.0) / 10.0
			cfg.TimeScale = math.Min(10.0, newSpeed)
			cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (+%.1fx)", cfg.TimeScale, step)
			cfg.NotificationTimer = 1.2
		}
		if rl.IsKeyPressed(rl.KeyBackSlash) {
			cfg.TimeScale = 1.0
			cfg.NotificationText = "Speed reset to 1.0x (Normal)"
			cfg.NotificationTimer = 1.2
		}
		if rl.IsKeyPressed(rl.KeyF1) {
			cfg.ShowHelp = !cfg.ShowHelp
		}
		if rl.IsKeyPressed(rl.KeyF2) {
			savedPath, err := SaveScenarioToFile(state, cfg, "")
			if err == nil {
				cfg.NotificationText = fmt.Sprintf("Saved scenario: %s", filepath.Base(savedPath))
				cfg.NotificationTimer = 3.5
				state.AvailableScenarios = ScanScenariosDirectory()
			}
		}
		if rl.IsKeyPressed(rl.KeyF3) {
			state.AvailableScenarios = ScanScenariosDirectory()
			state.ShowScenarioModal = !state.ShowScenarioModal
		}
		if rl.IsKeyPressed(rl.KeyF4) {
			state.ShowCameraPanel = !state.ShowCameraPanel
		}
		if rl.IsKeyPressed(rl.KeyB) {
			cfg.UseBarnesHut = !cfg.UseBarnesHut
			if cfg.UseBarnesHut {
				cfg.NotificationText = "Barnes-Hut Octree: ENABLED"
			} else {
				cfg.NotificationText = "Direct O(N^2) Gravity: ENABLED"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyT) {
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				cfg.ShowTrails = !cfg.ShowTrails
				status := "OFF"
				if cfg.ShowTrails {
					status = "ON"
				}
				cfg.NotificationText = fmt.Sprintf("Orbit Reference Trails: %s", status)
				cfg.NotificationTimer = 1.8
			} else {
				cfg.ShowTextures = !cfg.ShowTextures
				if cfg.ShowTextures {
					cfg.NotificationText = "Procedural Textures: ON"
				} else {
					cfg.NotificationText = "Procedural Textures: OFF"
				}
				cfg.NotificationTimer = 2.0
			}
		}
		if rl.IsKeyPressed(rl.KeyG) {
			cfg.EnableRelativity = !cfg.EnableRelativity
			if cfg.EnableRelativity {
				cfg.NotificationText = "1PN General Relativity: ON"
			} else {
				cfg.NotificationText = "1PN General Relativity: OFF"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyV) {
			cfg.ShowVectors = !cfg.ShowVectors
			cfg.Show2DVectorField = cfg.ShowVectors
			status := "OFF"
			if cfg.ShowVectors {
				status = "ON"
			}
			cfg.NotificationText = fmt.Sprintf("Object Direction Vectors: %s", status)
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyP) {
			cfg.ShowPotentialGrid = !cfg.ShowPotentialGrid
			if cfg.ShowPotentialGrid {
				cfg.NotificationText = fmt.Sprintf("3D Spacetime Grid: ON (Scale: %.2fx, Res: %dx%d)", cfg.SpacetimeScale, cfg.SpacetimeResolution, cfg.SpacetimeResolution)
			} else {
				cfg.NotificationText = "3D Spacetime Curvature Grid: OFF"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyLeftBracket) {
			if rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt) {
				DecreaseHeatmap2DResolution(cfg)
			} else if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				DecreaseResolution(cfg)
			} else {
				cfg.SpacetimeScale = float32(math.Max(0.5, float64(cfg.SpacetimeScale-0.25)))
				span := getStaticSpacetimeSpan(cfg)
				cfg.NotificationText = fmt.Sprintf("Spacetime Grid Scale: %.2fx (Span: %.0f)", cfg.SpacetimeScale, span)
				cfg.NotificationTimer = 2.0
			}
		}
		if rl.IsKeyPressed(rl.KeyRightBracket) {
			if rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt) {
				IncreaseHeatmap2DResolution(cfg)
			} else if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				IncreaseResolution(cfg)
			} else {
				cfg.SpacetimeScale = float32(math.Min(3.5, float64(cfg.SpacetimeScale+0.25)))
				span := getStaticSpacetimeSpan(cfg)
				cfg.NotificationText = fmt.Sprintf("Spacetime Grid Scale: %.2fx (Span: %.0f)", cfg.SpacetimeScale, span)
				cfg.NotificationTimer = 2.0
			}
		}
		if rl.IsKeyPressed(rl.KeyL) {
			cfg.ParticleGlowMode = !cfg.ParticleGlowMode
			if cfg.ParticleGlowMode {
				cfg.NotificationText = "Particle Velocity Glow: ON"
			} else {
				cfg.NotificationText = "Particle Velocity Glow: OFF"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyU) {
			if cfg.GPUAvailable {
				cfg.UseGPUCompute = !cfg.UseGPUCompute
				if cfg.UseGPUCompute {
					cfg.NotificationText = "Engine: GPU Compute Shader (N² Exact) ON"
				} else {
					cfg.NotificationText = "Engine: GPU Compute OFF (Barnes-Hut/CPU Active)"
				}
			} else {
				cfg.NotificationText = "GPU Compute Shader not supported on this context"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyO) {
			if !cfg.ShowVectorField {
				cfg.ShowVectorField = true
				cfg.DenseVectorField = false
				cfg.NotificationText = "Gravitational Vector Field: ON (3D Volume)"
			} else if !cfg.DenseVectorField {
				cfg.DenseVectorField = true
				cfg.NotificationText = "Gravitational Vector Field: ULTRA-DENSE 3D (3,448 volumetric vectors)"
			} else {
				cfg.ShowVectorField = false
				cfg.DenseVectorField = false
				cfg.NotificationText = "Gravitational Vector Field: OFF"
			}
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyJ) {
			cfg.ShowGravitationalWaves = !cfg.ShowGravitationalWaves
			if cfg.ShowGravitationalWaves {
				cfg.NotificationText = "Gravitational Waves & LIGO Detector: ON (Dynamic Quadrupole Metric)"
			} else {
				cfg.NotificationText = "Gravitational Waves: OFF"
			}
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyM) {
			cfg.Show2DViewport = !cfg.Show2DViewport
			if cfg.Show2DViewport {
				cfg.NotificationText = "2D Viewport & Tactical Map: ON"
			} else {
				cfg.NotificationText = "2D Viewport & Tactical Map: OFF"
			}
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyK) {
			cfg.Show2DLabels = !cfg.Show2DLabels
			if cfg.Show2DLabels {
				cfg.NotificationText = "2D Map Celestial Labels: ON"
			} else {
				cfg.NotificationText = "2D Map Celestial Labels: OFF (Clean Mode)"
			}
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyH) {
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
				IncreaseHeatmap2DResolution(cfg)
			} else if rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt) || rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
				DecreaseHeatmap2DResolution(cfg)
			} else {
				cfg.Show2DHeatmap = !cfg.Show2DHeatmap
				cfg.Show3DHeatmapPlane = cfg.Show2DHeatmap
				if cfg.Show2DHeatmap {
					cfg.NotificationText = "Gravitational Heatmap: ON (2D Map & 3D Plane Grid)"
				} else {
					cfg.NotificationText = "Gravitational Heatmap: OFF"
				}
				cfg.NotificationTimer = 2.0
			}
		}
		if rl.IsKeyPressed(rl.KeyY) {
			cfg.Show2DIcons = !cfg.Show2DIcons
			if cfg.Show2DIcons {
				cfg.NotificationText = "2D Object Icons: ON (Celestial Pictures / Glyphs Active)"
			} else {
				cfg.NotificationText = "2D Object Icons: OFF (Clear Circles Mode)"
			}
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyI) {
			if cfg.Integrator == IntegratorVerlet {
				cfg.Integrator = IntegratorYoshida4
				cfg.NotificationText = "Integrator: Yoshida 4th-Order Symplectic O(dt^4)"
			} else {
				cfg.Integrator = IntegratorVerlet
				cfg.NotificationText = "Integrator: Velocity Verlet 2nd-Order Symplectic"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyF5) || rl.IsKeyPressed(rl.KeyZ) {
			cfg.Offload3D = !cfg.Offload3D
			if cfg.Offload3D {
				cfg.Show2DViewport = true
				cfg.NotificationText = "3D Render Offloaded: 2D Main Viewport Active"
			} else {
				cfg.NotificationText = "3D Render Restored: Standard Mode"
			}
			cfg.NotificationTimer = 2.5
		}
		if rl.IsKeyPressed(rl.KeyF) {
			if state.SelectedBodyID != -1 {
				for _, b := range state.Bodies {
					if b.ID == state.SelectedBodyID {
						cfg.Viewport2DPan = rl.NewVector2(b.Position.X, b.Position.Z)
						camera.Target = b.Position
						cfg.NotificationText = fmt.Sprintf("Focused on: %s", b.Name)
						cfg.NotificationTimer = 1.5
						break
					}
				}
			} else if state.SelectedLagrangeIndex > 0 {
				prim, sec := GetLagrangePair(state)
				if prim != nil && sec != nil {
					pts := ComputeLagrangePoints(prim, sec, cfg.G)
					idx := state.SelectedLagrangeIndex - 1
					if idx >= 0 && idx < 5 {
						cfg.Viewport2DPan = rl.NewVector2(pts[idx].Position.X, pts[idx].Position.Z)
						camera.Target = pts[idx].Position
						cfg.NotificationText = fmt.Sprintf("Focused on: %s", pts[idx].Name)
						cfg.NotificationTimer = 1.5
					}
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyT) {
			cfg.Track2D = !cfg.Track2D
			if cfg.Track2D {
				cfg.NotificationText = "2D Body Tracking: ON (Target Locked)"
			} else {
				cfg.NotificationText = "2D Body Tracking: OFF"
			}
			cfg.NotificationTimer = 1.5
		}

		// Preset shortcuts (1-9, 0)
		if rl.IsKeyPressed(rl.KeyOne) {
			LoadPreset(state, cfg, PresetSolarSystem)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			LoadPreset(state, cfg, PresetSolarSystemGrand)
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			LoadPreset(state, cfg, PresetBinaryStars)
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			LoadPreset(state, cfg, PresetThreeBody)
		}
		if rl.IsKeyPressed(rl.KeyFive) {
			LoadPreset(state, cfg, PresetLagrangeTrojans)
		}
		if rl.IsKeyPressed(rl.KeySix) {
			LoadPreset(state, cfg, PresetGalaxyCollision)
		}
		if rl.IsKeyPressed(rl.KeySeven) {
			LoadPreset(state, cfg, PresetPulsarAccretion)
		}
		if rl.IsKeyPressed(rl.KeyEight) {
			LoadPreset(state, cfg, PresetMilkyWayCluster)
		}
		if rl.IsKeyPressed(rl.KeyNine) {
			LoadPreset(state, cfg, PresetRelativityRosette)
		}
		if rl.IsKeyPressed(rl.KeyZero) {
			LoadPreset(state, cfg, PresetEmpty)
		}

		// Gigantic 1K-10K Scene Shortcuts (F5 - F8)
		if rl.IsKeyPressed(rl.KeyF5) {
			LoadPreset(state, cfg, PresetMilkyWay10K)
			camera.Distance = 350.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}
		if rl.IsKeyPressed(rl.KeyF6) {
			LoadPreset(state, cfg, PresetAsteroidBelt2K)
			camera.Distance = 180.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}
		if rl.IsKeyPressed(rl.KeyF7) {
			LoadPreset(state, cfg, PresetGalaxyCollision5K)
			camera.Distance = 320.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}
		if rl.IsKeyPressed(rl.KeyF8) {
			LoadPreset(state, cfg, PresetBlackHoleSwarm3K)
			camera.Distance = 140.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}
		if rl.IsKeyPressed(rl.KeyF9) {
			LoadPreset(state, cfg, PresetGravitationalCollapse)
			camera.Distance = 120.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}
		if rl.IsKeyPressed(rl.KeyF10) {
			LoadPreset(state, cfg, PresetRealSolarSystem)
			camera.Distance = 450.0
			camera.Target = rl.NewVector3(0, 0, 0)
		}

		if rl.IsKeyPressed(rl.KeyF11) {
			cfg.ShowStatsHUD = !cfg.ShowStatsHUD
			status := "OFF"
			if cfg.ShowStatsHUD {
				status = "ON"
			}
			cfg.NotificationText = fmt.Sprintf("Stats HUD: %s", status)
			cfg.NotificationTimer = 1.8
		}

		if rl.IsKeyPressed(rl.KeyF12) {
			_ = os.MkdirAll("screenshots", 0o755)
			snapName := fmt.Sprintf("screenshots/screenshot_%d.png", time.Now().Unix())
			rl.TakeScreenshot(snapName)
			cfg.NotificationText = fmt.Sprintf("Screenshot saved to %s", snapName)
			cfg.NotificationTimer = 3.0
		}
		if (rl.IsKeyPressed(rl.KeyDelete) || rl.IsKeyPressed(rl.KeyBackspace)) && state.SelectedBodyID != -1 {
			removeBody(state, state.SelectedBodyID)
			state.SelectedBodyID = -1
			state.FollowSelected = false
		}
		if rl.IsKeyPressed(rl.KeyC) {
			if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || rl.IsKeyDown(rl.KeyLeftAlt) {
				cfg.Show2DCircles = !cfg.Show2DCircles
				status := "OFF"
				if cfg.Show2DCircles {
					status = "ON"
				}
				cfg.NotificationText = fmt.Sprintf("2D Circles & Dots: %s", status)
				cfg.NotificationTimer = 1.8
			} else if state.SelectedBodyID != -1 || state.SelectedLagrangeIndex > 0 {
				state.FollowSelected = !state.FollowSelected
				if state.FollowSelected {
					state.FollowBarycenter = false
				}
			} else {
				camera.LockToNearest(state, cfg)
			}
		}
		if rl.IsKeyPressed(rl.KeyF) && state.SelectedLagrangeIndex > 0 {
			primary, secondary := GetLagrangePair(state)
			if primary != nil && secondary != nil {
				pts := ComputeLagrangePoints(primary, secondary, cfg.G)
				idx := state.SelectedLagrangeIndex - 1
				if idx >= 0 && idx < 5 {
					camera.Target = pts[idx].Position
					cfg.NotificationText = fmt.Sprintf("Camera centered on %s", pts[idx].Name)
					cfg.NotificationTimer = 1.5
				}
			}
		}

		// Determine if UI captures the mouse
		uiCapturedMouse := IsMouseOverUI(state, cfg, w, h, mousePos)
		if Handle2DViewportInput(state, cfg, w, h) {
			uiCapturedMouse = true
		}

		// Handle 3D Mouse Interactions when not interacting with UI and not in 2D Offload mode
		if !uiCapturedMouse && !cfg.Offload3D {
			ray := rl.GetScreenToWorldRay(mousePos, camera.Camera)

			// Spawner Mode: click & drag to place and give initial velocity
			if state.IsSpawning {
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					hitPos, ok := IntersectRayPlane(ray, 0)
					if ok {
						state.SpawnWorldPos = hitPos
						state.IsDraggingSpawn = true
						state.DragVelocity = rl.NewVector3(0, 0, 0)
					}
				}

				if state.IsDraggingSpawn {
					if rl.IsMouseButtonDown(rl.MouseLeftButton) {
						hitPos, ok := IntersectRayPlane(ray, 0)
						if ok {
							// Slingshot pull vector
							dx := state.SpawnWorldPos.X - hitPos.X
							dz := state.SpawnWorldPos.Z - hitPos.Z
							state.DragVelocity = rl.NewVector3(dx*0.35, 0, dz*0.35)
						}
					}

					if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
						SpawnPresetAtLocation(state, cfg, state.SpawnWorldPos, state.DragVelocity)
						state.IsDraggingSpawn = false
					}
				}
			} else {
				// Selection Mode: Click to select body or Lagrange point in 3D
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					var hitBodyID int64 = -1
					var hitLagrangeIdx int = 0
					var closestDist float32 = 1e9

					for _, b := range state.Bodies {
						coll := rl.GetRayCollisionSphere(ray, b.Position, b.Radius*1.2)
						if coll.Hit && coll.Distance < closestDist {
							closestDist = coll.Distance
							hitBodyID = b.ID
							hitLagrangeIdx = 0
						}
					}

					if cfg.ShowLagrangePoints {
						primary, secondary := GetLagrangePair(state)
						if primary != nil && secondary != nil {
							pts := ComputeLagrangePoints(primary, secondary, cfg.G)
							for i := 0; i < 5; i++ {
								coll := rl.GetRayCollisionSphere(ray, pts[i].Position, 2.5)
								if coll.Hit && coll.Distance < closestDist {
									closestDist = coll.Distance
									hitLagrangeIdx = i + 1
									hitBodyID = -1
								}
							}
						}
					}

					state.SelectedBodyID = hitBodyID
					state.SelectedLagrangeIndex = hitLagrangeIdx
					if hitBodyID == -1 && hitLagrangeIdx == 0 {
						state.FollowSelected = false
					} else if hitLagrangeIdx > 0 {
						primary, secondary := GetLagrangePair(state)
						if primary != nil && secondary != nil {
							pts := ComputeLagrangePoints(primary, secondary, cfg.G)
							cfg.NotificationText = fmt.Sprintf("Selected: %s", pts[hitLagrangeIdx-1].Name)
							cfg.NotificationTimer = 2.0
						}
					}
				}
			}
		}

		// Update Camera
		camera.Update(state, cfg, uiCapturedMouse)

		// Advance Physics & Gravitational Wave field
		UpdatePhysics(state, cfg, dt)
		UpdateGravitationalWaves(state, cfg, float64(dt))

		// Drawing
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(6, 8, 14, 255))

		// 3D Scene (completely offloaded/stopped when 2D Main Viewport mode is active!)
		if !cfg.Offload3D {
			Render3DScene(state, cfg, camera)
		}

		// 2D Tactical Viewport & Gravitational Heatmap
		Draw2DViewport(state, cfg, w, h)

		// 2D UI & HUD
		DrawUI(state, cfg, camera, w, h)

		rl.EndDrawing()
	}
}
