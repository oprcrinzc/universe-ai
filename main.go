package main

import (
	"fmt"
	"math"
	"os"
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
	screenWidth := int32(1280)
	screenHeight := int32(720)

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "Gravity Mass Simulator 3D - Solar System Universe")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// Simulation configuration
	cfg := &Config{
		G:             1.0,
		TimeScale:     1.0,
		SubSteps:      5,
		Softening:     0.8,
		Collision:     CollisionMerge,
		Paused:        false,
		ShowTrails:    true,
		ShowGrid:      true,
		ShowVectors:   false,
		ShowForces:    false,
		ShowLabels:    true,
		ShowStarfield: true,
		ShowHelp:      false,
	}

	// Simulation state
	state := &SimState{
		SelectedBodyID: -1,
		SpawnPreset:    SpawnEarth,
	}

	// Initialize background stars
	InitStarfield(state, 500)

	// Load default preset: Solar System
	LoadPreset(state, cfg, PresetSolarSystem)

	// Orbit camera
	camera := NewOrbitCamera()

	for _, arg := range os.Args[1:] {
		if arg == "--capture-demo" {
			RunDemoCapture(state, cfg, camera, screenWidth, screenHeight)
			return
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
		if rl.IsKeyPressed(rl.KeyF1) {
			cfg.ShowHelp = !cfg.ShowHelp
		}
		if rl.IsKeyPressed(rl.KeyOne) {
			LoadPreset(state, cfg, PresetSolarSystem)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			LoadPreset(state, cfg, PresetBinaryStars)
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			LoadPreset(state, cfg, PresetThreeBody)
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			LoadPreset(state, cfg, PresetGalaxyDisk)
		}
		if rl.IsKeyPressed(rl.KeyFive) {
			LoadPreset(state, cfg, PresetEmpty)
		}
		if rl.IsKeyPressed(rl.KeyF12) {
			_ = os.MkdirAll("screenshots", 0755)
			snapName := fmt.Sprintf("screenshots/screenshot_%d.png", time.Now().Unix())
			rl.TakeScreenshot(snapName)
		}
		if (rl.IsKeyPressed(rl.KeyDelete) || rl.IsKeyPressed(rl.KeyBackspace)) && state.SelectedBodyID != -1 {
			removeBody(state, state.SelectedBodyID)
			state.SelectedBodyID = -1
			state.FollowSelected = false
		}
		if rl.IsKeyPressed(rl.KeyC) && state.SelectedBodyID != -1 {
			state.FollowSelected = !state.FollowSelected
		}

		// Determine if UI captures the mouse
		uiCapturedMouse := false

		// Top bar zone
		if mousePos.Y <= 48 {
			uiCapturedMouse = true
		}
		// Inspector panel zone
		if state.SelectedBodyID != -1 && mousePos.X >= float32(w)-300 && mousePos.Y <= 530 {
			uiCapturedMouse = true
		}
		// Bottom bar zone
		if mousePos.Y >= float32(h)-48 {
			uiCapturedMouse = true
		}
		// Help modal zone
		if cfg.ShowHelp {
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
						// Spawn new body!
						newBody := CreateBodyTemplate(state.SpawnPreset, state.SpawnWorldPos, state.DragVelocity, state.NextID)
						state.NextID++
						state.Bodies = append(state.Bodies, newBody)
						state.SelectedBodyID = newBody.ID
						state.IsDraggingSpawn = false
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
		camera.Update(state, uiCapturedMouse)

		// Advance Physics
		UpdatePhysics(state, cfg, dt)

		// Drawing
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(6, 8, 14, 255))

		// 3D Scene
		Render3DScene(state, cfg, camera)

		// 2D UI & HUD
		DrawUI(state, cfg, camera, w, h)

		rl.EndDrawing()
	}
}
