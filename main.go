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
		ShowGrid:            true,
		ShowPotentialGrid:   false,
		SpacetimeScale:      3.0,
		SpacetimeResolution: 512,
		ShowVectorField:     false,
		CinematicCamera:     false,
		ParticleGlowMode:    false,
		AutoPerformanceMode: true,
		ShowVectors:         false,
		ShowForces:          false,
		ShowLabels:          true,
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
		Show2DHeatmap:       true,
		Show3DHeatmapPlane:  false,
		Show2DVectorField:   true,
		Viewport2DZoom:      0.45,
		RealScaleVisualMode: false,
		ShowLagrangePoints:  true,
		NotificationText:    "Gravity Simulator 3D (144 FPS / 1080p). F5-F8: 1K-10K Gigantic Scenes | F1: Manual",
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
			cfg.ShowTextures = !cfg.ShowTextures
			if cfg.ShowTextures {
				cfg.NotificationText = "Procedural Textures: ON"
			} else {
				cfg.NotificationText = "Procedural Textures: OFF"
			}
			cfg.NotificationTimer = 2.0
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
			cfg.CinematicCamera = !cfg.CinematicCamera
			if cfg.CinematicCamera {
				cfg.NotificationText = "Cinematic Auto-Director Camera: ON"
			} else {
				cfg.NotificationText = "Cinematic Auto-Director Camera: OFF"
			}
			cfg.NotificationTimer = 2.5
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
			cfg.SpacetimeScale = float32(math.Max(0.5, float64(cfg.SpacetimeScale-0.25)))
			span := getStaticSpacetimeSpan(cfg)
			cfg.NotificationText = fmt.Sprintf("Spacetime Grid Scale: %.2fx (Span: %.0f)", cfg.SpacetimeScale, span)
			cfg.NotificationTimer = 2.0
		}
		if rl.IsKeyPressed(rl.KeyRightBracket) {
			cfg.SpacetimeScale = float32(math.Min(3.5, float64(cfg.SpacetimeScale+0.25)))
			span := getStaticSpacetimeSpan(cfg)
			cfg.NotificationText = fmt.Sprintf("Spacetime Grid Scale: %.2fx (Span: %.0f)", cfg.SpacetimeScale, span)
			cfg.NotificationTimer = 2.0
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
			cfg.ShowVectorField = !cfg.ShowVectorField
			if cfg.ShowVectorField {
				cfg.NotificationText = "Gravitational Vector Field: ON (32x32 needles)"
			} else {
				cfg.NotificationText = "Gravitational Vector Field: OFF"
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
			cfg.Show2DHeatmap = !cfg.Show2DHeatmap
			cfg.Show3DHeatmapPlane = cfg.Show2DHeatmap
			if cfg.Show2DHeatmap {
				cfg.NotificationText = "Gravitational Heatmap: ON (2D Map & 3D Plane Grid)"
			} else {
				cfg.NotificationText = "Gravitational Heatmap: OFF"
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
			if state.SelectedBodyID != -1 {
				state.FollowSelected = !state.FollowSelected
				if state.FollowSelected {
					state.FollowBarycenter = false
				}
			} else {
				camera.LockToNearest(state, cfg)
			}
		}

		// Determine if UI captures the mouse
		uiCapturedMouse := IsMouseOverUI(state, cfg, w, h, mousePos)
		if Handle2DViewportInput(state, cfg, w, h) {
			uiCapturedMouse = true
		}

		// Handle 3D Mouse Interactions when not interacting with UI
		if !uiCapturedMouse {
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
						if state.SpawnPreset == SpawnAsteroidRing {
							for a := 0; a < 16; a++ {
								ang := float64(a) * (2.0 * math.Pi / 16.0)
								r := float32(2.5 + float64(a%3)*0.8)
								p := rl.NewVector3(
									state.SpawnWorldPos.X+r*float32(math.Cos(ang)),
									state.SpawnWorldPos.Y+float32((a%3)-1)*0.2,
									state.SpawnWorldPos.Z+r*float32(math.Sin(ang)),
								)
								v := rl.NewVector3(
									state.DragVelocity.X-float32(math.Sin(ang))*2.5,
									state.DragVelocity.Y,
									state.DragVelocity.Z+float32(math.Cos(ang))*2.5,
								)
								newBody := CreateBodyTemplate(SpawnMoon, p, v, state.NextID)
								newBody.Name = fmt.Sprintf("Belt Asteroid #%d", state.NextID)
								newBody.Radius = 0.25
								newBody.Mass = 0.005
								state.NextID++
								state.Bodies = append(state.Bodies, newBody)
							}
							state.IsDraggingSpawn = false
						} else if state.SpawnPreset == SpawnStarCluster50 {
							SpawnStarCluster(state, state.SpawnWorldPos, state.DragVelocity, 50)
							state.IsDraggingSpawn = false
						} else if state.SpawnPreset == SpawnMiniGalaxy150 {
							SpawnMiniGalaxy(state, state.SpawnWorldPos, state.DragVelocity, cfg.G)
							state.IsDraggingSpawn = false
						} else if state.SpawnPreset == SpawnCollapseCloud100 {
							SpawnCollapseCloud(state, state.SpawnWorldPos, state.DragVelocity, 100)
							state.IsDraggingSpawn = false
						} else {
							// Spawn new body!
							newBody := CreateBodyTemplate(state.SpawnPreset, state.SpawnWorldPos, state.DragVelocity, state.NextID)
							state.NextID++
							state.Bodies = append(state.Bodies, newBody)
							state.SelectedBodyID = newBody.ID
							state.IsDraggingSpawn = false
						}
					}
				}
			} else {
				// Selection Mode: Click to select body in 3D
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					var hitBodyID int64 = -1
					var closestDist float32 = 1e9

					for _, b := range state.Bodies {
						coll := rl.GetRayCollisionSphere(ray, b.Position, b.Radius*1.2)
						if coll.Hit && coll.Distance < closestDist {
							closestDist = coll.Distance
							hitBodyID = b.ID
						}
					}

					state.SelectedBodyID = hitBodyID
					if hitBodyID == -1 {
						state.FollowSelected = false
					}
				}
			}
		}

		// Update Camera
		camera.Update(state, cfg, uiCapturedMouse)

		// Advance Physics
		UpdatePhysics(state, cfg, dt)

		// Drawing
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(6, 8, 14, 255))

		// 3D Scene
		Render3DScene(state, cfg, camera)

		// 2D Tactical Viewport & Gravitational Heatmap
		Draw2DViewport(state, cfg, w, h)

		// 2D UI & HUD
		DrawUI(state, cfg, camera, w, h)

		rl.EndDrawing()
	}
}
