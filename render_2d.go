package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Handle2DViewportInput processes mouse clicks, dragging, and wheel zooming on the 2D viewport
func Handle2DViewportInput(state *SimState, cfg *Config, screenW, screenH int32) bool {
	if !cfg.Show2DViewport {
		return false
	}

	vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)
	mPos := rl.GetMousePosition()
	mouseInside := mPos.X >= float32(vx) && mPos.X <= float32(vx+vw) &&
		mPos.Y >= float32(vy) && mPos.Y <= float32(vy+vh)

	if !mouseInside {
		return false
	}

	// Initialize zoom if zero
	if cfg.Viewport2DZoom <= 0.0001 {
		cfg.Viewport2DZoom = 0.45
	}

	headerH := int32(32)
	toolbarH := get2DToolbarHeight(vw)
	statusH := int32(22)

	checkBtn := func(bx, by, bw, bh float32) bool {
		return mPos.X >= bx && mPos.X <= bx+bw && mPos.Y >= by && mPos.Y <= by+bh
	}

	// 1. Header Bar Window Action Buttons ([OFFLOAD 3D], [MAX/MIN] Fullscreen, [X] Close)
	if mPos.Y <= float32(vy+headerH) {
		btnY := float32(vy + 5)
		btnH := float32(22)

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			// [X] Close (24px)
			if checkBtn(float32(vx+vw-28), btnY, 24, btnH) {
				cfg.Show2DViewport = false
				cfg.Offload3D = false
				return true
			}
			// [MAX/MIN] Fullscreen (34px)
			if checkBtn(float32(vx+vw-66), btnY, 34, btnH) {
				cfg.Viewport2DFullscreen = !cfg.Viewport2DFullscreen
				return true
			}
			// [OFFLOAD 3D] button (104px)
			offBtnW := float32(104)
			if checkBtn(float32(vx+vw-66)-offBtnW-6, btnY, offBtnW, btnH) {
				cfg.Offload3D = !cfg.Offload3D
				if cfg.Offload3D {
					cfg.Show2DViewport = true
					cfg.NotificationText = "3D Render Offloaded: 2D Main Viewport Active"
				} else {
					cfg.NotificationText = "3D Render Restored: Standard Mode"
				}
				cfg.NotificationTimer = 2.0
				return true
			}
		}
		return true // Mouse is over header bar; consume event so it doesn't leak to 3D scene
	}

	// 2. Tactical Toolbar Bar Buttons (Zoom controls, layer toggles, actions, and resolution)
	if mPos.Y > float32(vy+headerH) && mPos.Y <= float32(vy+headerH+toolbarH) {
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			if vw >= 700 {
				// Single Row Layout (Fullscreen / 3D Offload / Wide Window)
				btnY := float32(vy + headerH + 4)
				btnH := float32(20)
				curX := float32(vx + 10)

				// [-] Zoom Out
				if checkBtn(curX, btnY, 22, btnH) {
					cfg.Viewport2DZoom = float32(math.Max(0.005, float64(cfg.Viewport2DZoom*0.75)))
					return true
				}
				curX += 25

				// [+] Zoom In
				if checkBtn(curX, btnY, 22, btnH) {
					cfg.Viewport2DZoom = float32(math.Min(80.0, float64(cfg.Viewport2DZoom*1.35)))
					return true
				}
				curX += 25

				// [RST] Reset Pan/Zoom
				if checkBtn(curX, btnY, 30, btnH) {
					cfg.Viewport2DZoom = 0.45
					cfg.Viewport2DPan = rl.NewVector2(0, 0)
					cfg.Track2D = false
					cfg.NotificationText = "2D Map View: RESET"
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 30 + 6

				curX += 8 // Divider

				// [TRL] Orbit Trails toggle
				if checkBtn(curX, btnY, 32, btnH) {
					cfg.ShowTrails = !cfg.ShowTrails
					status := "OFF"
					if cfg.ShowTrails {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("Orbit Reference Trails: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 36

				// [CIR] Body Overlay Circles toggle
				if checkBtn(curX, btnY, 30, btnH) {
					cfg.Show2DCircles = !cfg.Show2DCircles
					status := "OFF"
					if cfg.Show2DCircles {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("2D Circles & Dots: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 34

				// [LBL] Labels toggle
				if checkBtn(curX, btnY, 30, btnH) {
					cfg.Show2DLabels = !cfg.Show2DLabels
					status := "OFF"
					if cfg.Show2DLabels {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("2D Map Labels: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 34

				// [ICON] Celestial Object Icons toggle
				if checkBtn(curX, btnY, 34, btnH) {
					cfg.Show2DIcons = !cfg.Show2DIcons
					status := "OFF"
					if cfg.Show2DIcons {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("2D Object Icons: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 38

				// [HEAT] Heatmap toggle
				if checkBtn(curX, btnY, 34, btnH) {
					cfg.Show2DHeatmap = !cfg.Show2DHeatmap
					return true
				}
				curX += 38

				// [VEC] Object Direction Vectors toggle
				if checkBtn(curX, btnY, 30, btnH) {
					cfg.ShowVectors = !cfg.ShowVectors
					cfg.Show2DVectorField = cfg.ShowVectors
					status := "OFF"
					if cfg.ShowVectors {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("Object Direction Vectors: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 34

				// [GW] Gravitational Waves toggle
				if checkBtn(curX, btnY, 28, btnH) {
					cfg.ShowGravitationalWaves = !cfg.ShowGravitationalWaves
					cfg.Show2DGW = cfg.ShowGravitationalWaves
					status := "OFF"
					if cfg.ShowGravitationalWaves {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("Gravitational Waves: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 32

				// [LAG] Lagrange points toggle
				if checkBtn(curX, btnY, 30, btnH) {
					cfg.ShowLagrangePoints = !cfg.ShowLagrangePoints
					status := "OFF"
					if cfg.ShowLagrangePoints {
						status = "ON"
					}
					cfg.NotificationText = fmt.Sprintf("Lagrange Points L1-L5: %s", status)
					cfg.NotificationTimer = 1.2
					return true
				}
				curX += 30 + 6

				curX += 8 // Divider

				// [FOCUS] Center on selected target
				if checkBtn(curX, btnY, 42, btnH) {
					handleFocusTarget(state, cfg)
					return true
				}
				curX += 46

				// [TRACK] Lock & follow selected target
				if checkBtn(curX, btnY, 42, btnH) {
					handleTrackTarget(state, cfg)
					return true
				}
				curX += 46

				// [SPAWN] Toggle Spawning mode
				if checkBtn(curX, btnY, 58, btnH) {
					handleSpawnToggle(state, cfg)
					return true
				}
				curX += 58 + 6

				curX += 8 // Divider

				// [RES-] Decrease Resolution (Heatmap when active, Spacetime when inactive or with Shift)
				if checkBtn(curX, btnY, 36, btnH) {
					if cfg.Show2DHeatmap && !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
						DecreaseHeatmap2DResolution(cfg)
					} else {
						DecreaseResolution(cfg)
					}
					return true
				}
				curX += 40

				// [RES+] Increase Resolution (Heatmap when active, Spacetime when inactive or with Shift)
				if checkBtn(curX, btnY, 36, btnH) {
					if cfg.Show2DHeatmap && !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
						IncreaseHeatmap2DResolution(cfg)
					} else {
						IncreaseResolution(cfg)
					}
					return true
				}
				curX += 40 + 6

				// [HUD] Toggle Stats HUD
				if curX+32 <= float32(vx+vw-4) {
					curX += 8 // Divider
					if checkBtn(curX, btnY, 32, btnH) {
						cfg.ShowStatsHUD = !cfg.ShowStatsHUD
						status := "OFF"
						if cfg.ShowStatsHUD {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("Stats HUD: %s (F11)", status)
						cfg.NotificationTimer = 1.5
						return true
					}
				}
			} else {
				// Two Row Layout (Docked mode: vw < 700)
				btnH := float32(20)

				if mPos.Y <= float32(vy+headerH+24) {
					// Row 1: Zoom + Layers
					btnY := float32(vy + headerH + 3)
					curX := float32(vx + 8)

					// [-]
					if checkBtn(curX, btnY, 20, btnH) {
						cfg.Viewport2DZoom = float32(math.Max(0.005, float64(cfg.Viewport2DZoom*0.75)))
						return true
					}
					curX += 23

					// [+]
					if checkBtn(curX, btnY, 20, btnH) {
						cfg.Viewport2DZoom = float32(math.Min(80.0, float64(cfg.Viewport2DZoom*1.35)))
						return true
					}
					curX += 23

					// [RST]
					if checkBtn(curX, btnY, 26, btnH) {
						cfg.Viewport2DZoom = 0.45
						cfg.Viewport2DPan = rl.NewVector2(0, 0)
						cfg.Track2D = false
						cfg.NotificationText = "2D Map View: RESET"
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 26 + 4

					curX += 6 // Divider

					// [TRL]
					if checkBtn(curX, btnY, 28, btnH) {
						cfg.ShowTrails = !cfg.ShowTrails
						status := "OFF"
						if cfg.ShowTrails {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("Orbit Reference Trails: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 31

					// [CIR]
					if checkBtn(curX, btnY, 26, btnH) {
						cfg.Show2DCircles = !cfg.Show2DCircles
						status := "OFF"
						if cfg.Show2DCircles {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("2D Circles & Dots: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 29

					// [LBL]
					if checkBtn(curX, btnY, 26, btnH) {
						cfg.Show2DLabels = !cfg.Show2DLabels
						status := "OFF"
						if cfg.Show2DLabels {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("2D Map Labels: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 29

					// [ICON]
					if checkBtn(curX, btnY, 30, btnH) {
						cfg.Show2DIcons = !cfg.Show2DIcons
						status := "OFF"
						if cfg.Show2DIcons {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("2D Object Icons: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 33

					// [HEAT]
					if checkBtn(curX, btnY, 30, btnH) {
						cfg.Show2DHeatmap = !cfg.Show2DHeatmap
						return true
					}
					curX += 33

					// [VEC]
					if checkBtn(curX, btnY, 26, btnH) {
						cfg.ShowVectors = !cfg.ShowVectors
						cfg.Show2DVectorField = cfg.ShowVectors
						status := "OFF"
						if cfg.ShowVectors {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("Object Direction Vectors: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 29

					// [GW]
					if checkBtn(curX, btnY, 24, btnH) {
						cfg.ShowGravitationalWaves = !cfg.ShowGravitationalWaves
						cfg.Show2DGW = cfg.ShowGravitationalWaves
						status := "OFF"
						if cfg.ShowGravitationalWaves {
							status = "ON"
						}
						cfg.NotificationText = fmt.Sprintf("Gravitational Waves: %s", status)
						cfg.NotificationTimer = 1.2
						return true
					}
					curX += 27

					// [LAG]
					if curX+26 <= float32(vx+vw-4) {
						if checkBtn(curX, btnY, 26, btnH) {
							cfg.ShowLagrangePoints = !cfg.ShowLagrangePoints
							status := "OFF"
							if cfg.ShowLagrangePoints {
								status = "ON"
							}
							cfg.NotificationText = fmt.Sprintf("Lagrange Points L1-L5: %s", status)
							cfg.NotificationTimer = 1.2
							return true
						}
					}
				} else {
					// Row 2: Actions + Resolution + HUD
					btnY := float32(vy + headerH + 26)
					curX := float32(vx + 8)

					// [FOCUS]
					if checkBtn(curX, btnY, 40, btnH) {
						handleFocusTarget(state, cfg)
						return true
					}
					curX += 44

					// [TRACK]
					if checkBtn(curX, btnY, 40, btnH) {
						handleTrackTarget(state, cfg)
						return true
					}
					curX += 44

					// [SPAWN]
					if checkBtn(curX, btnY, 54, btnH) {
						handleSpawnToggle(state, cfg)
						return true
					}
					curX += 54 + 4

					curX += 6 // Divider

					// [RES-]
					if checkBtn(curX, btnY, 34, btnH) {
						if cfg.Show2DHeatmap && !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
							DecreaseHeatmap2DResolution(cfg)
						} else {
							DecreaseResolution(cfg)
						}
						return true
					}
					curX += 38

					// [RES+]
					if checkBtn(curX, btnY, 34, btnH) {
						if cfg.Show2DHeatmap && !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
							IncreaseHeatmap2DResolution(cfg)
						} else {
							IncreaseResolution(cfg)
						}
						return true
					}
					curX += 34 + 4

					curX += 6 // Divider

					// [HUD]
					if curX+30 <= float32(vx+vw-4) {
						if checkBtn(curX, btnY, 30, btnH) {
							cfg.ShowStatsHUD = !cfg.ShowStatsHUD
							status := "OFF"
							if cfg.ShowStatsHUD {
								status = "ON"
							}
							cfg.NotificationText = fmt.Sprintf("Stats HUD: %s (F11)", status)
							cfg.NotificationTimer = 1.5
							return true
						}
					}
				}
			}
		}
		return true // Mouse is over toolbar bar; consume event
	}

	clipY := vy + headerH + toolbarH
	clipH := vh - headerH - toolbarH - statusH
	cx := float32(vx) + float32(vw)*0.5
	cy := float32(clipY) + float32(clipH)*0.5

	w2s := func(wx, wz float32) (float32, float32) {
		return cx + (wx-cfg.Viewport2DPan.X)*cfg.Viewport2DZoom, cy + (wz-cfg.Viewport2DPan.Y)*cfg.Viewport2DZoom
	}

	// 3. Floating Quick-Action Card for Selected Lagrange Point
	if state.SelectedLagrangeIndex > 0 && state.SelectedLagrangeIndex <= 5 {
		prim, sec := GetLagrangePair(state)
		if prim != nil && sec != nil {
			pts := ComputeLagrangePoints(prim, sec, cfg.G)
			idx := state.SelectedLagrangeIndex - 1
			pt := pts[idx]

			cardW := int32(310)
			if cardW > vw-20 {
				cardW = vw - 20
			}
			cardH := int32(76)
			cardX := vx + 10
			cardY := clipY + clipH - cardH - 8

			if checkBtn(float32(cardX), float32(cardY), float32(cardW), float32(cardH)) {
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					bY := float32(cardY + 44)
					bH := float32(22)
					bX := float32(cardX + 8)

					// [+Probe]
					if checkBtn(bX, bY, 84, bH) {
						SpawnProbeAtLagrange(state, cfg, pt, fmt.Sprintf("%s %s", sec.Name, pt.Name[:2]))
						return true
					}
					bX += 90

					// If L4 or L5: [+Swarm]
					if idx == 3 || idx == 4 {
						if checkBtn(bX, bY, 86, bH) {
							SpawnLagrangeSwarm(state, cfg, prim, sec, pt, 8)
							return true
						}
						bX += 92
					}

					// [Track]
					if checkBtn(bX, bY, 64, bH) {
						state.FollowSelected = !state.FollowSelected
						if state.FollowSelected {
							cfg.NotificationText = fmt.Sprintf("Tracking %s: ON", pt.Name)
						} else {
							cfg.NotificationText = fmt.Sprintf("Tracking %s: OFF", pt.Name)
						}
						cfg.NotificationTimer = 1.5
						return true
					}

					// [X] Close
					if checkBtn(float32(cardX+cardW-26), float32(cardY+6), 20, 20) {
						state.SelectedLagrangeIndex = 0
						return true
					}
				}
				return true // Mouse inside card; consume
			}
		}
	}

	// 4. Mouse wheel zoom centered on cursor
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		oldZoom := cfg.Viewport2DZoom
		zoomFactor := float32(math.Pow(1.18, float64(wheel)))
		newZoom := oldZoom * zoomFactor
		if newZoom < 0.005 {
			newZoom = 0.005
		}
		if newZoom > 80.0 {
			newZoom = 80.0
		}

		cursorWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/oldZoom
		cursorWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/oldZoom

		cfg.Viewport2DZoom = newZoom
		cfg.Viewport2DPan.X = cursorWorldX - (mPos.X-cx)/newZoom
		cfg.Viewport2DPan.Y = cursorWorldZ - (mPos.Y-cy)/newZoom
		return true
	}

	// 5. Right-click or Middle-click drag to pan 2D map
	if (rl.IsMouseButtonDown(rl.MouseRightButton) || rl.IsMouseButtonDown(rl.MouseMiddleButton)) &&
		(mPos.Y > float32(clipY) && mPos.Y < float32(vy+vh-statusH)) {
		delta := rl.GetMouseDelta()
		cfg.Viewport2DPan.X -= delta.X / cfg.Viewport2DZoom
		cfg.Viewport2DPan.Y -= delta.Y / cfg.Viewport2DZoom
		cfg.Track2D = false // User manual pan disengages auto-track
		return true
	}

	// 6. Spawner Mode: Click & Drag to slingshot launch bodies directly in 2D plane
	if state.IsSpawning {
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && mPos.Y > float32(clipY) && mPos.Y < float32(vy+vh-statusH) {
			clickWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/cfg.Viewport2DZoom
			clickWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/cfg.Viewport2DZoom
			state.SpawnWorldPos = rl.NewVector3(clickWorldX, 0, clickWorldZ)
			state.IsDraggingSpawn = true
			state.DragVelocity = rl.NewVector3(0, 0, 0)
			return true
		}

		if state.IsDraggingSpawn {
			if rl.IsMouseButtonDown(rl.MouseLeftButton) {
				curWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/cfg.Viewport2DZoom
				curWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/cfg.Viewport2DZoom
				dx := state.SpawnWorldPos.X - curWorldX
				dz := state.SpawnWorldPos.Z - curWorldZ
				state.DragVelocity = rl.NewVector3(dx*0.35, 0, dz*0.35)
				return true
			}

			if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
				SpawnPresetAtLocation(state, cfg, state.SpawnWorldPos, state.DragVelocity)
				state.IsDraggingSpawn = false
				cfg.NotificationText = fmt.Sprintf("2D Spawned: %s", state.SpawnPreset.String())
				cfg.NotificationTimer = 2.0
				return true
			}
		}
	}

	// 7. Left-click to select Lagrange points or bodies in 2D viewport (when not spawning)
	if !state.IsSpawning && rl.IsMouseButtonPressed(rl.MouseLeftButton) && mPos.Y > float32(clipY) && mPos.Y < float32(vy+vh-statusH) {
		// First check Lagrange points
		if cfg.ShowLagrangePoints {
			prim, sec := GetLagrangePair(state)
			if prim != nil && sec != nil {
				pts := ComputeLagrangePoints(prim, sec, cfg.G)
				for idx := 0; idx < 5; idx++ {
					lx, ly := w2s(pts[idx].Position.X, pts[idx].Position.Z)
					dist := float32(math.Hypot(float64(mPos.X-lx), float64(mPos.Y-ly)))
					if dist <= 14.0 {
						state.SelectedLagrangeIndex = idx + 1
						state.SelectedBodyID = -1
						cfg.NotificationText = fmt.Sprintf("Selected: %s", pts[idx].Name)
						cfg.NotificationTimer = 2.0
						return true
					}
				}
			}
		}

		// Next check celestial bodies
		clickWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/cfg.Viewport2DZoom
		clickWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/cfg.Viewport2DZoom

		var closestID int64 = -1
		var closestDist float32 = 25.0 / cfg.Viewport2DZoom // 25 px selection radius

		for _, b := range state.Bodies {
			dx := b.Position.X - clickWorldX
			dz := b.Position.Z - clickWorldZ
			d := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			if d < closestDist {
				closestDist = d
				closestID = b.ID
			}
		}

		if closestID != -1 {
			state.SelectedBodyID = closestID
			state.SelectedLagrangeIndex = 0
			return true
		}

		// Clicked empty space: deselect Lagrange point
		state.SelectedLagrangeIndex = 0
	}

	return true // Keep capturing mouse whenever inside the 2D viewport rect
}

// get2DToolbarHeight calculates responsive toolbar height (50px for two rows in docked mode, 28px for single row)
func get2DToolbarHeight(vw int32) int32 {
	if vw < 700 {
		return 50
	}
	return 28
}

func handleFocusTarget(state *SimState, cfg *Config) {
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				cfg.Viewport2DPan = rl.NewVector2(b.Position.X, b.Position.Z)
				cfg.NotificationText = fmt.Sprintf("2D Focused on: %s", b.Name)
				cfg.NotificationTimer = 1.5
				return
			}
		}
	} else if state.SelectedLagrangeIndex > 0 {
		prim, sec := GetLagrangePair(state)
		if prim != nil && sec != nil {
			pts := ComputeLagrangePoints(prim, sec, cfg.G)
			idx := state.SelectedLagrangeIndex - 1
			if idx >= 0 && idx < 5 {
				cfg.Viewport2DPan = rl.NewVector2(pts[idx].Position.X, pts[idx].Position.Z)
				cfg.NotificationText = fmt.Sprintf("2D Focused on: %s", pts[idx].Name)
				cfg.NotificationTimer = 1.5
				return
			}
		}
	} else if len(state.Bodies) > 0 {
		var heaviest *Body
		for _, b := range state.Bodies {
			if heaviest == nil || b.Mass > heaviest.Mass {
				heaviest = b
			}
		}
		if heaviest != nil {
			cfg.Viewport2DPan = rl.NewVector2(heaviest.Position.X, heaviest.Position.Z)
			cfg.NotificationText = fmt.Sprintf("2D Focused on: %s", heaviest.Name)
			cfg.NotificationTimer = 1.5
		}
	}
}

func handleTrackTarget(state *SimState, cfg *Config) {
	cfg.Track2D = !cfg.Track2D
	status := "OFF"
	if cfg.Track2D {
		status = "ON (Target Locked)"
		if state.SelectedBodyID != -1 {
			for _, b := range state.Bodies {
				if b.ID == state.SelectedBodyID {
					cfg.Viewport2DPan = rl.NewVector2(b.Position.X, b.Position.Z)
					break
				}
			}
		}
	}
	cfg.NotificationText = fmt.Sprintf("2D Body Tracking: %s", status)
	cfg.NotificationTimer = 1.5
}

func handleSpawnToggle(state *SimState, cfg *Config) {
	state.IsSpawning = !state.IsSpawning
	state.IsDraggingSpawn = false
	status := "OFF"
	if state.IsSpawning {
		status = "ON (Click & Drag to launch)"
	}
	cfg.NotificationText = fmt.Sprintf("2D Spawn Mode: %s", status)
	cfg.NotificationTimer = 1.8
}

// get2DViewportRect computes responsive dimensions and position on the right side,
// intelligently shifting left when the Inspector panel is open to guarantee ZERO overlap
func get2DViewportRect(state *SimState, cfg *Config, screenW, screenH int32) (int32, int32, int32, int32) {
	if cfg.Offload3D {
		return 0, 48, screenW, screenH - 96
	}
	if cfg.Viewport2DFullscreen {
		return 10, 52, screenW - 20, screenH - 104
	}

	w := int32(500)
	h := int32(390)
	if screenW < 1440 {
		w = int32(440)
		h = int32(350)
	}
	if screenW < 1200 {
		w = int32(380)
		h = int32(300)
	}
	if screenW < 1000 {
		w = int32(330)
		h = int32(260)
	}

	y := int32(52)
	maxH := screenH - y - 54
	if h > maxH {
		h = maxH
	}

	// RIGHT SIDE DOCKING:
	// If the Inspector panel is open on the right (width 310, margin 12),
	// shift the 2D map to the left of the Inspector panel so they sit side-by-side with zero overlap!
	inspectorOpen := (state != nil && (state.SelectedBodyID != -1 || state.SelectedLagrangeIndex > 0))
	var x int32
	if inspectorOpen {
		panelW := int32(310)
		x = screenW - panelW - 14 - w - 10
	} else {
		x = screenW - w - 14
	}

	if x < 14 {
		x = 14
	}
	return x, y, w, h
}

// drawCelestialObjectIcon2D renders rich tactical astronomical glyphs for stars, black holes,
// terrestrial worlds, gas giants with rings, desert planets, icy bodies, asteroids, and satellites
// using the dedicated b.CelestialIcon property.
func drawCelestialObjectIcon2D(b *Body, sx, sy, radPx float32, isSelected bool) {
	iconR := float32(math.Max(5.0, math.Min(13.0, float64(radPx))))

	switch b.CelestialIcon {
	case IconProbe:
		// Satellite craft: center diamond bus with solar panel wings
		bW := float32(3.5)
		rl.DrawRectangleLines(int32(sx-bW), int32(sy-bW), int32(bW*2), int32(bW*2), rl.White)
		pW := float32(6.5)
		pH := float32(3.5)
		panelCol := rl.NewColor(45, 150, 255, 230)
		rl.DrawRectangleLines(int32(sx-bW-pW-1), int32(sy-pH*0.5), int32(pW), int32(pH), panelCol)
		rl.DrawRectangleLines(int32(sx+bW+1), int32(sy-pH*0.5), int32(pW), int32(pH), panelCol)
		rl.DrawLine(int32(sx), int32(sy-bW), int32(sx), int32(sy-bW-4), rl.Gold)

	case IconBlackHole:
		// Black Hole: Tactical singularity vortex icon (no physical accretion disc)
		sR := iconR * 0.8
		rl.DrawCircleLines(int32(sx), int32(sy), 3.0, rl.NewColor(120, 230, 255, 255))
		rl.DrawCircleLines(int32(sx), int32(sy), sR, rl.NewColor(255, 140, 40, 230))
		rl.DrawLineEx(rl.NewVector2(sx-sR-2, sy), rl.NewVector2(sx-sR+2, sy-3), 1.5, rl.NewColor(255, 170, 60, 240))
		rl.DrawLineEx(rl.NewVector2(sx+sR+2, sy), rl.NewVector2(sx+sR-2, sy+3), 1.5, rl.NewColor(255, 170, 60, 240))
		rl.DrawLineEx(rl.NewVector2(sx, sy-sR-2), rl.NewVector2(sx+3, sy-sR+2), 1.5, rl.NewColor(255, 170, 60, 240))
		rl.DrawLineEx(rl.NewVector2(sx, sy+sR+2), rl.NewVector2(sx-3, sy+sR-2), 1.5, rl.NewColor(255, 170, 60, 240))

	case IconPulsar:
		// Pulsar: Tactical energy pulse icon (no body disc)
		pR := iconR * 0.75
		pulCol := rl.NewColor(160, 230, 255, 245)
		rl.DrawPolyLines(rl.NewVector2(sx, sy), 4, pR, 45.0, b.Color)
		jetH := pR * 2.2
		rl.DrawLineEx(rl.NewVector2(sx, sy-pR), rl.NewVector2(sx, sy-jetH), 1.8, pulCol)
		rl.DrawLineEx(rl.NewVector2(sx, sy+pR), rl.NewVector2(sx, sy+jetH), 1.8, pulCol)
		rl.DrawLine(int32(sx-2), int32(sy-jetH+3), int32(sx), int32(sy-jetH), pulCol)
		rl.DrawLine(int32(sx+2), int32(sy-jetH+3), int32(sx), int32(sy-jetH), pulCol)
		rl.DrawLine(int32(sx-2), int32(sy+jetH-3), int32(sx), int32(sy+jetH), pulCol)
		rl.DrawLine(int32(sx+2), int32(sy+jetH-3), int32(sx), int32(sy+jetH), pulCol)

	case IconStar:
		// Star: Pure radiant 8-point sunburst icon (no body disc)
		rayLen := iconR * 1.8
		col := b.Color
		rayCol := rl.NewColor(col.R, col.G, col.B, 235)
		rl.DrawLineEx(rl.NewVector2(sx-rayLen, sy), rl.NewVector2(sx+rayLen, sy), 2.0, rayCol)
		rl.DrawLineEx(rl.NewVector2(sx, sy-rayLen), rl.NewVector2(sx, sy+rayLen), 2.0, rayCol)
		dRay := rayLen * 0.65
		rl.DrawLineEx(rl.NewVector2(sx-dRay, sy-dRay), rl.NewVector2(sx+dRay, sy+dRay), 1.4, rayCol)
		rl.DrawLineEx(rl.NewVector2(sx-dRay, sy+dRay), rl.NewVector2(sx+dRay, sy-dRay), 1.4, rayCol)
		rl.DrawLineEx(rl.NewVector2(sx-2.5, sy), rl.NewVector2(sx+2.5, sy), 2.0, rl.White)
		rl.DrawLineEx(rl.NewVector2(sx, sy-2.5), rl.NewVector2(sx, sy+2.5), 2.0, rl.White)

	case IconGasGiant:
		// Gas Giant: Tactical ringed planet icon (no cloud bands or filled body)
		gR := iconR * 0.7
		ringCol := rl.NewColor(240, 210, 150, 230)
		rl.DrawEllipseLines(int32(sx), int32(sy), gR*2.1, gR*0.7, ringCol)
		rl.DrawCircleLines(int32(sx), int32(sy), gR, b.Color)

	case IconTerrestrial:
		// Terrestrial: Tactical globe icon (no continents or landmass drawings)
		tR := iconR * 0.8
		gridCol := rl.NewColor(90, 210, 255, 210)
		rl.DrawCircleLines(int32(sx), int32(sy), tR, b.Color)
		rl.DrawLine(int32(sx-tR), int32(sy), int32(sx+tR), int32(sy), gridCol)
		rl.DrawEllipseLines(int32(sx), int32(sy), tR*0.38, tR, gridCol)

	case IconDesert:
		// Desert / Rocky: Tactical astronomical crosshair icon (no polar ice caps)
		dR := iconR * 0.75
		rl.DrawCircleLines(int32(sx), int32(sy), dR, b.Color)
		rl.DrawLine(int32(sx-dR*0.6), int32(sy), int32(sx+dR*0.6), int32(sy), rl.NewColor(255, 230, 200, 200))
		rl.DrawLine(int32(sx), int32(sy-dR*0.6), int32(sx), int32(sy+dR*0.6), rl.NewColor(255, 230, 200, 200))

	case IconIceGiant:
		// Ice Giant: Tactical crystalline snowflake icon (no body disc)
		iceR := iconR * 0.8
		iceCol := rl.NewColor(130, 235, 255, 240)
		for a := 0; a < 3; a++ {
			ang := float64(a) * math.Pi / 3.0
			dx := float32(math.Cos(ang)) * iceR
			dy := float32(math.Sin(ang)) * iceR
			rl.DrawLineEx(rl.NewVector2(sx-dx, sy-dy), rl.NewVector2(sx+dx, sy+dy), 1.6, iceCol)
		}
		rl.DrawPolyLines(rl.NewVector2(sx, sy), 4, iceR*0.35, 45.0, rl.White)

	case IconMoon:
		// Moon / Asteroid: Tactical crescent icon (no body disc, no craters)
		mR := iconR * 0.75
		moonCol := rl.NewColor(215, 225, 245, 240)
		var arcPts [8]rl.Vector2
		for i := 0; i < 5; i++ {
			ang := float64(i)*math.Pi/4.0 - math.Pi*0.5
			arcPts[i] = rl.NewVector2(sx+float32(math.Cos(ang))*mR, sy+float32(math.Sin(ang))*mR)
		}
		arcPts[5] = rl.NewVector2(sx+mR*0.35, sy+mR*0.4)
		arcPts[6] = rl.NewVector2(sx+mR*0.22, sy)
		arcPts[7] = rl.NewVector2(sx+mR*0.35, sy-mR*0.4)
		for i := 0; i < 7; i++ {
			rl.DrawLineEx(arcPts[i], arcPts[i+1], 1.6, moonCol)
		}
		rl.DrawLineEx(arcPts[7], arcPts[0], 1.6, moonCol)

	default:
		// Celestial marker: Tactical diamond icon
		cR := iconR * 0.75
		rl.DrawPolyLines(rl.NewVector2(sx, sy), 4, cR, 45.0, b.Color)
	}
}

// Draw2DViewport renders the top-down orthographic 2D orbital map, gravitational heatmap,
// color-coded vector field, projected masses, and Lagrange points
func Draw2DViewport(state *SimState, cfg *Config, screenW, screenH int32) {
	if !cfg.Show2DViewport {
		return
	}

	vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)

	// Update continuous tracking in 2D
	if cfg.Track2D && state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				cfg.Viewport2DPan.X = b.Position.X
				cfg.Viewport2DPan.Y = b.Position.Z
				break
			}
		}
	}

	// Ensure valid zoom & pan defaults
	if cfg.Viewport2DZoom <= 0.0001 {
		cfg.Viewport2DZoom = 0.45
	}

	headerH := int32(32)
	toolbarH := get2DToolbarHeight(vw)
	statusH := int32(22)

	// 1. Draw outer frame & background
	rl.DrawRectangle(vx, vy, vw, vh, rl.NewColor(7, 10, 20, 245))
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(vx), float32(vy), float32(vw), float32(vh)), 2.0, rl.NewColor(40, 140, 230, 220))

	// 2. Header Bar (Window Title, Telemetry Badge & Window Buttons)
	rl.DrawRectangle(vx, vy, vw, headerH, rl.NewColor(12, 22, 45, 252))
	rl.DrawLine(vx, vy+headerH, vx+vw, vy+headerH, rl.NewColor(45, 120, 200, 180))

	title := "2D TACTICAL MAP"
	if cfg.Offload3D {
		title = "2D MAIN COMMAND VIEWPORT · [3D RENDER OFFLOADED]"
	} else if cfg.Viewport2DFullscreen {
		title = "2D TACTICAL MAP (FULLSCREEN) [M to Minimize]"
	} else if vw < 380 {
		title = "2D MAP"
	}
	titleW := MeasureTextBoldUI(title, 14)
	rl.DrawText(title, vx+10, vy+8, 14, rl.RayWhite)

	offBtnW := int32(104)
	offBtnX := vx + vw - 66 - offBtnW - 6

	// Draw scale telemetry in a stylish rounded badge
	scaleSpan := float32(vw) / cfg.Viewport2DZoom
	scaleText := fmt.Sprintf("%.0f AU · %.2fx", scaleSpan/100.0, cfg.Viewport2DZoom)
	scaleW := int32(rl.MeasureText(scaleText, 11))
	badgeX := vx + int32(titleW) + 16
	badgeW := scaleW + 14

	if badgeX+badgeW < offBtnX-10 {
		badgeRec := rl.NewRectangle(float32(badgeX), float32(vy+5), float32(badgeW), 22)
		rl.DrawRectangleRounded(badgeRec, 0.3, 4, rl.NewColor(18, 32, 58, 240))
		rl.DrawRectangleRoundedLines(badgeRec, 0.3, 4, rl.NewColor(45, 95, 160, 200))
		rl.DrawText(scaleText, badgeX+7, vy+9, 11, rl.NewColor(140, 220, 255, 240))
	}

	mPos := rl.GetMousePosition()

	drawSmallBtn := func(label string, bx, by, bw, bh int32, active bool, baseCol rl.Color) {
		col := baseCol
		if active {
			col = rl.NewColor(col.R+40, col.G+40, col.B+40, 255)
		}
		hover := mPos.X >= float32(bx) && mPos.X <= float32(bx+bw) &&
			mPos.Y >= float32(by) && mPos.Y <= float32(by+bh)
		if hover {
			col = rl.NewColor(col.R+50, col.G+50, col.B+50, 255)
		}
		rl.DrawRectangle(bx, by, bw, bh, col)
		borderCol := rl.NewColor(80, 150, 220, 180)
		if active {
			borderCol = rl.Gold
		}
		rl.DrawRectangleLines(bx, by, bw, bh, borderCol)

		fSize := int32(11)
		tLen := rl.MeasureText(label, fSize)
		rl.DrawText(label, bx+(bw-tLen)/2, by+(bh-fSize)/2, fSize, rl.White)
	}

	// Window Controls (Right side of Header)
	maxBtnText := "MAX"
	if cfg.Viewport2DFullscreen {
		maxBtnText = "MIN"
	}
	drawSmallBtn(maxBtnText, vx+vw-66, vy+5, 34, 22, cfg.Viewport2DFullscreen, rl.NewColor(35, 60, 100, 240))
	drawSmallBtn("X", vx+vw-28, vy+5, 24, 22, false, rl.NewColor(90, 35, 35, 240))

	offBtnText := "OFFLOAD 3D"
	offBtnCol := rl.NewColor(35, 60, 100, 240)
	if cfg.Offload3D {
		offBtnText = "3D OFF (MAIN)"
		offBtnCol = rl.NewColor(180, 70, 20, 250)
	}
	drawSmallBtn(offBtnText, offBtnX, vy+5, offBtnW, 22, cfg.Offload3D, offBtnCol)

	// 3. Dedicated Tactical Toolbar Bar (Spacious responsive layout, zero overlap!)
	rl.DrawRectangle(vx, vy+headerH, vw, toolbarH, rl.NewColor(10, 16, 30, 245))
	rl.DrawLine(vx, vy+headerH+toolbarH, vx+vw, vy+headerH+toolbarH, rl.NewColor(35, 65, 110, 180))

	if vw >= 700 {
		// Single row layout (Fullscreen / 3D Offload / Wide Window)
		toolBtnY := vy + headerH + 4
		toolBtnH := int32(20)
		curX := vx + 10

		// Zoom Group
		drawSmallBtn("-", curX, toolBtnY, 22, toolBtnH, false, rl.NewColor(35, 55, 85, 240))
		curX += 25
		drawSmallBtn("+", curX, toolBtnY, 22, toolBtnH, false, rl.NewColor(35, 55, 85, 240))
		curX += 25
		drawSmallBtn("RST", curX, toolBtnY, 30, toolBtnH, false, rl.NewColor(40, 60, 90, 240))
		curX += 30 + 6

		// Divider
		rl.DrawLine(curX+2, toolBtnY+2, curX+2, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 8

		// Layers Group
		trlCol := rl.NewColor(35, 50, 75, 240)
		if cfg.ShowTrails {
			trlCol = rl.NewColor(25, 95, 140, 240)
		}
		drawSmallBtn("TRL", curX, toolBtnY, 32, toolBtnH, cfg.ShowTrails, trlCol)
		curX += 36

		cirCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DCircles {
			cirCol = rl.NewColor(20, 110, 130, 240)
		}
		drawSmallBtn("CIR", curX, toolBtnY, 30, toolBtnH, cfg.Show2DCircles, cirCol)
		curX += 34

		lblCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DLabels {
			lblCol = rl.NewColor(25, 85, 130, 240)
		}
		drawSmallBtn("LBL", curX, toolBtnY, 30, toolBtnH, cfg.Show2DLabels, lblCol)
		curX += 34

		iconCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DIcons {
			iconCol = rl.NewColor(20, 110, 130, 240)
		}
		drawSmallBtn("ICON", curX, toolBtnY, 34, toolBtnH, cfg.Show2DIcons, iconCol)
		curX += 38

		heatCol := rl.NewColor(55, 40, 25, 240)
		if cfg.Show2DHeatmap {
			heatCol = rl.NewColor(130, 70, 25, 240)
		}
		drawSmallBtn("HEAT", curX, toolBtnY, 34, toolBtnH, cfg.Show2DHeatmap, heatCol)
		curX += 38

		vecCol := rl.NewColor(30, 55, 45, 240)
		if cfg.ShowVectors {
			vecCol = rl.NewColor(25, 110, 75, 240)
		}
		drawSmallBtn("VEC", curX, toolBtnY, 30, toolBtnH, cfg.ShowVectors, vecCol)
		curX += 34

		gwCol := rl.NewColor(35, 50, 75, 240)
		if cfg.ShowGravitationalWaves {
			gwCol = rl.NewColor(135, 45, 175, 240)
		}
		drawSmallBtn("GW", curX, toolBtnY, 28, toolBtnH, cfg.ShowGravitationalWaves, gwCol)
		curX += 32

		lagCol := rl.NewColor(45, 40, 65, 240)
		if cfg.ShowLagrangePoints {
			lagCol = rl.NewColor(110, 85, 30, 240)
		}
		drawSmallBtn("LAG", curX, toolBtnY, 30, toolBtnH, cfg.ShowLagrangePoints, lagCol)
		curX += 30 + 6

		// Divider
		rl.DrawLine(curX+2, toolBtnY+2, curX+2, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 8

		// Actions Group
		focusCol := rl.NewColor(30, 65, 95, 240)
		drawSmallBtn("FOCUS", curX, toolBtnY, 42, toolBtnH, false, focusCol)
		curX += 46

		trackCol := rl.NewColor(30, 65, 95, 240)
		if cfg.Track2D {
			trackCol = rl.NewColor(20, 130, 110, 240)
		}
		drawSmallBtn("TRACK", curX, toolBtnY, 42, toolBtnH, cfg.Track2D, trackCol)
		curX += 46

		spawnCol := rl.NewColor(40, 60, 80, 240)
		spawnLabel := "+SPAWN"
		if state.IsSpawning {
			spawnCol = rl.NewColor(150, 50, 30, 240)
			spawnLabel = "SPAWN ON"
		}
		drawSmallBtn(spawnLabel, curX, toolBtnY, 58, toolBtnH, state.IsSpawning, spawnCol)
		curX += 58 + 6

		// Divider
		rl.DrawLine(curX+2, toolBtnY+2, curX+2, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 8

		// Resolution buttons (warm amber tint when 2D Heatmap is active)
		resCol := rl.NewColor(35, 55, 80, 240)
		if cfg.Show2DHeatmap {
			resCol = rl.NewColor(80, 50, 25, 240)
		}
		drawSmallBtn("RES-", curX, toolBtnY, 36, toolBtnH, false, resCol)
		curX += 40
		drawSmallBtn("RES+", curX, toolBtnY, 36, toolBtnH, false, resCol)
		curX += 40 + 6

		// HUD button
		if curX+32 <= vx+vw-4 {
			rl.DrawLine(curX+2, toolBtnY+2, curX+2, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
			curX += 8
			hudCol := rl.NewColor(35, 50, 75, 240)
			if cfg.ShowStatsHUD {
				hudCol = rl.NewColor(25, 95, 140, 240)
			}
			drawSmallBtn("HUD", curX, toolBtnY, 32, toolBtnH, cfg.ShowStatsHUD, hudCol)
		}
	} else {
		// Two-row responsive toolbar for docked mode
		rl.DrawLine(vx+8, vy+headerH+24, vx+vw-8, vy+headerH+24, rl.NewColor(30, 50, 80, 160))

		// Row 1: Zoom + Layers
		toolBtnY := vy + headerH + 3
		toolBtnH := int32(20)
		curX := vx + 8

		drawSmallBtn("-", curX, toolBtnY, 20, toolBtnH, false, rl.NewColor(35, 55, 85, 240))
		curX += 23
		drawSmallBtn("+", curX, toolBtnY, 20, toolBtnH, false, rl.NewColor(35, 55, 85, 240))
		curX += 23
		drawSmallBtn("RST", curX, toolBtnY, 26, toolBtnH, false, rl.NewColor(40, 60, 90, 240))
		curX += 26 + 4

		rl.DrawLine(curX+1, toolBtnY+2, curX+1, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 6

		trlCol := rl.NewColor(35, 50, 75, 240)
		if cfg.ShowTrails {
			trlCol = rl.NewColor(25, 95, 140, 240)
		}
		drawSmallBtn("TRL", curX, toolBtnY, 28, toolBtnH, cfg.ShowTrails, trlCol)
		curX += 31

		cirCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DCircles {
			cirCol = rl.NewColor(20, 110, 130, 240)
		}
		drawSmallBtn("CIR", curX, toolBtnY, 26, toolBtnH, cfg.Show2DCircles, cirCol)
		curX += 29

		lblCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DLabels {
			lblCol = rl.NewColor(25, 85, 130, 240)
		}
		drawSmallBtn("LBL", curX, toolBtnY, 26, toolBtnH, cfg.Show2DLabels, lblCol)
		curX += 29

		iconCol := rl.NewColor(35, 50, 75, 240)
		if cfg.Show2DIcons {
			iconCol = rl.NewColor(20, 110, 130, 240)
		}
		drawSmallBtn("ICON", curX, toolBtnY, 30, toolBtnH, cfg.Show2DIcons, iconCol)
		curX += 33

		heatCol := rl.NewColor(55, 40, 25, 240)
		if cfg.Show2DHeatmap {
			heatCol = rl.NewColor(130, 70, 25, 240)
		}
		drawSmallBtn("HEAT", curX, toolBtnY, 30, toolBtnH, cfg.Show2DHeatmap, heatCol)
		curX += 33

		vecCol := rl.NewColor(30, 55, 45, 240)
		if cfg.ShowVectors {
			vecCol = rl.NewColor(25, 110, 75, 240)
		}
		drawSmallBtn("VEC", curX, toolBtnY, 26, toolBtnH, cfg.ShowVectors, vecCol)
		curX += 29

		gwCol := rl.NewColor(35, 50, 75, 240)
		if cfg.ShowGravitationalWaves {
			gwCol = rl.NewColor(135, 45, 175, 240)
		}
		drawSmallBtn("GW", curX, toolBtnY, 24, toolBtnH, cfg.ShowGravitationalWaves, gwCol)
		curX += 27

		lagCol := rl.NewColor(45, 40, 65, 240)
		if cfg.ShowLagrangePoints {
			lagCol = rl.NewColor(110, 85, 30, 240)
		}
		if curX+26 <= vx+vw-4 {
			drawSmallBtn("LAG", curX, toolBtnY, 26, toolBtnH, cfg.ShowLagrangePoints, lagCol)
		}

		// Row 2: Actions + Resolution + HUD
		toolBtnY = vy + headerH + 26
		curX = vx + 8

		focusCol := rl.NewColor(30, 65, 95, 240)
		drawSmallBtn("FOCUS", curX, toolBtnY, 40, toolBtnH, false, focusCol)
		curX += 44

		trackCol := rl.NewColor(30, 65, 95, 240)
		if cfg.Track2D {
			trackCol = rl.NewColor(20, 130, 110, 240)
		}
		drawSmallBtn("TRACK", curX, toolBtnY, 40, toolBtnH, cfg.Track2D, trackCol)
		curX += 44

		spawnCol := rl.NewColor(40, 60, 80, 240)
		spawnLabel := "+SPAWN"
		if state.IsSpawning {
			spawnCol = rl.NewColor(150, 50, 30, 240)
			spawnLabel = "SPAWN ON"
		}
		drawSmallBtn(spawnLabel, curX, toolBtnY, 54, toolBtnH, state.IsSpawning, spawnCol)
		curX += 54 + 4

		rl.DrawLine(curX+1, toolBtnY+2, curX+1, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 6

		resCol := rl.NewColor(35, 55, 80, 240)
		if cfg.Show2DHeatmap {
			resCol = rl.NewColor(80, 50, 25, 240)
		}
		drawSmallBtn("RES-", curX, toolBtnY, 34, toolBtnH, false, resCol)
		curX += 38
		drawSmallBtn("RES+", curX, toolBtnY, 34, toolBtnH, false, resCol)
		curX += 34 + 4

		rl.DrawLine(curX+1, toolBtnY+2, curX+1, toolBtnY+toolBtnH-2, rl.NewColor(50, 80, 125, 180))
		curX += 6

		if curX+30 <= vx+vw-4 {
			hudCol := rl.NewColor(35, 50, 75, 240)
			if cfg.ShowStatsHUD {
				hudCol = rl.NewColor(25, 95, 140, 240)
			}
			drawSmallBtn("HUD", curX, toolBtnY, 30, toolBtnH, cfg.ShowStatsHUD, hudCol)
		}
	}

	// 4. Content Area Clipping
	clipY := vy + headerH + toolbarH
	clipH := vh - headerH - toolbarH - statusH
	rl.BeginScissorMode(vx, clipY, vw, clipH)

	cx := float32(vx) + float32(vw)*0.5
	cy := float32(clipY) + float32(clipH)*0.5
	zoom := cfg.Viewport2DZoom
	panX := cfg.Viewport2DPan.X
	panZ := cfg.Viewport2DPan.Y

	// Helper: World (X, Z) to Screen (sx, sy)
	w2s := func(wx, wz float32) (float32, float32) {
		return cx + (wx-panX)*zoom, cy + (wz-panZ)*zoom
	}

	// 5. Render 2D Gravitational Heatmap (Configurable Resolution Adaptive Grid)
	if cfg.Show2DHeatmap && len(state.Bodies) > 0 {
		res := cfg.Heatmap2DResolution
		if res <= 0 {
			res = 64 // Default: 64 columns
		} else if res < 16 {
			res = 16
		} else if res > 256 {
			res = 256
		}
		cols := res
		rows := int(float32(cols) * float32(clipH) / float32(vw))
		if rows < 8 {
			rows = 8
		}
		cellW := float32(vw) / float32(cols)
		cellH := float32(clipH) / float32(rows)

		type fast2DSource struct {
			x, z float32
			mass float32
		}
		sources := make([]fast2DSource, 0, 48)
		for _, b := range state.Bodies {
			if b.Mass >= 0.0001 || b.CelestialIcon == IconStar || b.ID == state.SelectedBodyID {
				sources = append(sources, fast2DSource{x: b.Position.X, z: b.Position.Z, mass: float32(b.Mass)})
				if len(sources) >= 48 {
					break
				}
			}
		}
		if len(sources) == 0 && len(state.Bodies) > 0 {
			sources = append(sources, fast2DSource{x: state.Bodies[0].Position.X, z: state.Bodies[0].Position.Z, mass: float32(state.Bodies[0].Mass)})
		}

		gVal := float32(cfg.G)
		eps2 := float32(cfg.Softening*cfg.Softening + 0.1)

		for iy := 0; iy < rows; iy++ {
			sy0 := float32(clipY) + float32(iy)*cellH
			wz := panZ + (sy0+cellH*0.5-cy)/zoom
			for ix := 0; ix < cols; ix++ {
				sx0 := float32(vx) + float32(ix)*cellW
				wx := panX + (sx0+cellW*0.5-cx)/zoom

				var gx, gz float32
				for k := range sources {
					dx := sources[k].x - wx
					dz := sources[k].z - wz
					distSq := dx*dx + dz*dz + eps2
					dist := float32(math.Sqrt(float64(distSq)))
					f := gVal * sources[k].mass / (distSq * dist)
					gx += dx * f
					gz += dz * f
				}
				gLen := float32(math.Sqrt(float64(gx*gx + gz*gz)))

				ratio := float32(math.Min(1.0, math.Max(0.0, (math.Log10(float64(gLen)+1e-6)+3.5)/4.5)))
				var cellCol rl.Color
				if ratio < 0.25 {
					t := ratio / 0.25
					cellCol = rl.NewColor(
						uint8(8+t*15),
						uint8(14+t*50),
						uint8(36+t*120),
						uint8(80+t*50),
					)
				} else if ratio < 0.50 {
					t := (ratio - 0.25) / 0.25
					cellCol = rl.NewColor(
						uint8(23+t*20),
						uint8(64+t*140),
						uint8(156+t*70),
						uint8(130+t*40),
					)
				} else if ratio < 0.75 {
					t := (ratio - 0.50) / 0.25
					cellCol = rl.NewColor(
						uint8(43+t*200),
						uint8(204+t*10),
						uint8(226-t*180),
						uint8(170+t*40),
					)
				} else {
					t := (ratio - 0.75) / 0.25
					cellCol = rl.NewColor(
						uint8(243+t*12),
						uint8(214-t*150),
						uint8(46+t*180),
						uint8(210+t*45),
					)
				}
				rl.DrawRectangle(int32(sx0), int32(sy0), int32(cellW+1), int32(cellH+1), cellCol)
			}
		}
	}

	// 6. Render 2D Vector Field (Crisp High Density Directional Lattice)
	if cfg.Show2DVectorField && len(state.Bodies) > 0 {
		vCols := int(float32(vw) / 13.5)
		if vCols < 24 {
			vCols = 24
		} else if vCols > 45 {
			vCols = 45
		}
		vRows := int(float32(clipH) / 13.5)
		if vRows < 18 {
			vRows = 18
		} else if vRows > 32 {
			vRows = 32
		}
		stepX := float32(vw) / float32(vCols)
		stepY := float32(clipH) / float32(vRows)

		type fast2DSource struct {
			x, z float32
			mass float32
		}
		sources := make([]fast2DSource, 0, 48)
		for _, b := range state.Bodies {
			if b.Mass >= 0.0001 || b.CelestialIcon == IconStar || b.ID == state.SelectedBodyID {
				sources = append(sources, fast2DSource{x: b.Position.X, z: b.Position.Z, mass: float32(b.Mass)})
				if len(sources) >= 48 {
					break
				}
			}
		}
		if len(sources) == 0 && len(state.Bodies) > 0 {
			sources = append(sources, fast2DSource{x: state.Bodies[0].Position.X, z: state.Bodies[0].Position.Z, mass: float32(state.Bodies[0].Mass)})
		}
		gVal := float32(cfg.G)
		eps2 := float32(cfg.Softening*cfg.Softening + 0.1)

		for iy := 0; iy < vRows; iy++ {
			sy0 := float32(clipY) + (float32(iy)+0.5)*stepY
			wz := panZ + (sy0-cy)/zoom
			for ix := 0; ix < vCols; ix++ {
				sx0 := float32(vx) + (float32(ix)+0.5)*stepX
				wx := panX + (sx0-cx)/zoom

				var gx, gz float32
				for k := range sources {
					dx := sources[k].x - wx
					dz := sources[k].z - wz
					distSq := dx*dx + dz*dz + eps2
					dist := float32(math.Sqrt(float64(distSq)))
					f := gVal * sources[k].mass / (distSq * dist)
					gx += dx * f
					gz += dz * f
				}
				gLen := float32(math.Sqrt(float64(gx*gx + gz*gz)))
				if gLen > 0.0001 {
					dirX := gx / gLen
					dirZ := gz / gLen

					arrowLen := float32(3.5 + 11.0*float64(gLen/(gLen+0.4)))
					sx1 := sx0 + dirX*arrowLen
					sy1 := sy0 + dirZ*arrowLen

					vecCol := rl.NewColor(120, 230, 255, 190)
					if gLen > 0.5 {
						vecCol = rl.NewColor(255, 210, 60, 210)
					}
					if gLen > 2.0 {
						vecCol = rl.NewColor(255, 80, 50, 230)
					}

					rl.DrawLine(int32(sx0), int32(sy0), int32(sx1), int32(sy1), vecCol)

					wingLen := arrowLen * 0.35
					w1x := sx1 - dirX*wingLen - dirZ*wingLen*0.6
					w1y := sy1 - dirZ*wingLen + dirX*wingLen*0.6
					w2x := sx1 - dirX*wingLen + dirZ*wingLen*0.6
					w2y := sy1 - dirZ*wingLen - dirX*wingLen*0.6
					rl.DrawLine(int32(sx1), int32(sy1), int32(w1x), int32(w1y), vecCol)
					rl.DrawLine(int32(sx1), int32(sy1), int32(w2x), int32(w2y), vecCol)
				}
			}
		}
	}

	// 6b. Render Gravitational Waves in 2D
	if cfg.ShowGravitationalWaves {
		curTime := rl.GetTime()

		// Continuous Binary Quadrupole Ripples
		if len(state.Bodies) >= 2 {
			var p1, p2 *Body
			for _, b := range state.Bodies {
				if p1 == nil || b.Mass > p1.Mass {
					p2 = p1
					p1 = b
				} else if p2 == nil || b.Mass > p2.Mass {
					p2 = b
				}
			}

			if p1 != nil && p2 != nil && p2.Mass > 0.0005 {
				baryX := float32((p1.Position.X*float32(p1.Mass) + p2.Position.X*float32(p2.Mass)) / float32(p1.Mass+p2.Mass))
				baryZ := float32((p1.Position.Z*float32(p1.Mass) + p2.Position.Z*float32(p2.Mass)) / float32(p1.Mass+p2.Mass))
				bsx, bsy := w2s(baryX, baryZ)

				sep := float32(math.Hypot(float64(p2.Position.X-p1.Position.X), float64(p2.Position.Z-p1.Position.Z)))
				if sep < 0.2 {
					sep = 0.2
				}
				mTot := p1.Mass + p2.Mass
				omegaOrb := float32(math.Sqrt((cfg.G * mTot) / float64(sep*sep*sep)))
				omegaGW := 2.0 * omegaOrb

				c := float32(cfg.SpeedOfLight)
				if c <= 0 {
					c = 150.0
				}
				gwWavelength := (2.0 * float32(math.Pi) * c) / omegaGW
				if gwWavelength < 4.0 {
					gwWavelength = 4.0
				}

				// Draw expanding quadrupole ripple rings
				numRings := 7
				phaseOffset := float32(math.Mod(float64(curTime)*float64(c), float64(gwWavelength)))
				for rIdx := 0; rIdx < numRings; rIdx++ {
					wRad := float32(rIdx)*gwWavelength + phaseOffset
					scrR := wRad * zoom
					if scrR > 6.0 && scrR < float32(vw)*1.2 {
						falloff := 1.0 - (float32(rIdx) / float32(numRings))
						alpha := uint8(falloff * 110.0)
						gwCol := rl.NewColor(190, 80, 255, alpha)
						rl.DrawCircleLines(int32(bsx), int32(bsy), scrR, gwCol)
					}
				}
			}
		}

		// Propagating Merger Burst Shockwaves
		for _, b := range state.GWBursts {
			age := float32(curTime - b.StartTime)
			if age > 0 && age <= b.DecayTime {
				bsx, bsy := w2s(b.Position.X, b.Position.Z)
				waveFrontR := b.WaveSpeed * age * zoom
				intensity := float32(math.Exp(float64(-age * 0.75)))
				alpha := uint8(intensity * 220.0)
				if alpha > 10 {
					burstCol := rl.NewColor(255, 100, 200, alpha)
					rl.DrawCircleLines(int32(bsx), int32(bsy), waveFrontR, burstCol)
					if waveFrontR > 6.0 {
						rl.DrawCircleLines(int32(bsx), int32(bsy), waveFrontR-4.0, rl.NewColor(100, 220, 255, uint8(float32(alpha)*0.7)))
						rl.DrawCircleLines(int32(bsx), int32(bsy), waveFrontR+4.0, rl.NewColor(255, 210, 80, uint8(float32(alpha)*0.7)))
					}
				}
			}
		}

		// Live 2D Gravitational Wave Telemetry Badge
		if state.LastGWStrain != 0 || state.LastGWFreq > 0 {
			gwText := fmt.Sprintf("GW · h: %.2e · f: %.1f Hz", state.LastGWStrain, state.LastGWFreq)
			tLen := int32(rl.MeasureText(gwText, 10))
			badgeX := vx + 12
			badgeY := clipY + 10
			badgeRec := rl.NewRectangle(float32(badgeX), float32(badgeY), float32(tLen+16), 18)
			rl.DrawRectangleRounded(badgeRec, 0.3, 4, rl.NewColor(24, 10, 36, 210))
			rl.DrawRectangleRoundedLines(badgeRec, 0.3, 4, rl.NewColor(180, 80, 240, 180))
			rl.DrawText(gwText, badgeX+8, badgeY+4, 10, rl.NewColor(230, 160, 255, 240))
		}
	}

	// 7. Render Orbit Reference Rings & Trails (Respects cfg.ShowTrails toggle)
	if cfg.ShowTrails {
		for _, b := range state.Bodies {
			if b.CelestialIcon == IconStar || b.Mass >= 0.01 {
				if len(b.Trail) > 1 {
					for t := 1; t < len(b.Trail); t++ {
						x0, y0 := w2s(b.Trail[t-1].X, b.Trail[t-1].Z)
						x1, y1 := w2s(b.Trail[t].X, b.Trail[t].Z)
						alpha := uint8(40 + (t*160)/len(b.Trail))
						trailCol := rl.NewColor(b.Color.R, b.Color.G, b.Color.B, alpha)
						rl.DrawLine(int32(x0), int32(y0), int32(x1), int32(y1), trailCol)
					}
				}
			}
		}
	}

	// 8. Render 2D Celestial Masses with Clean Decluttered Labels & Object Icons
	for _, b := range state.Bodies {
		sx, sy := w2s(b.Position.X, b.Position.Z)

		// Cull bodies far outside viewport
		if sx < float32(vx)-50 || sx > float32(vx+vw)+50 || sy < float32(clipY)-50 || sy > float32(clipY+clipH)+50 {
			continue
		}

		// Logarithmic pixel radius
		radPx := float32(2.5 + math.Log10(math.Max(1.0, b.Mass+1.0))*2.4)
		if b.ID == state.SelectedBodyID {
			radPx += 2.0
		}

		// Body circle or custom celestial object icon
		if cfg.Show2DIcons {
			drawCelestialObjectIcon2D(b, sx, sy, radPx, b.ID == state.SelectedBodyID)
		} else {
			// Clear circle & dot mode: when enabled, draw clean hollow outline circle and pinpoint center core dot
			if cfg.Show2DCircles {
				rl.DrawCircleLines(int32(sx), int32(sy), radPx, b.Color)
				if radPx >= 4.0 {
					rl.DrawCircleLines(int32(sx), int32(sy), radPx, rl.Fade(rl.White, 0.35))
				}
				rl.DrawCircle(int32(sx), int32(sy), 1.5, b.Color)
			}
		}

		// Glowing selection ring (only when selected)
		if b.ID == state.SelectedBodyID {
			haloCol := rl.NewColor(0, 255, 230, 200)
			rl.DrawCircleLines(int32(sx), int32(sy), radPx+3.0, haloCol)
		}

		// Velocity / direction vector indicator (respects cfg.ShowVectors like in 3D!)
		if cfg.ShowVectors && !b.IsStationary {
			vLen := float32(math.Sqrt(float64(b.Velocity.X*b.Velocity.X + b.Velocity.Z*b.Velocity.Z)))
			if vLen > 0.05 {
				vxEnd := sx + (b.Velocity.X/vLen)*float32(math.Min(24.0, float64(radPx+vLen*1.8)))
				vyEnd := sy + (b.Velocity.Z/vLen)*float32(math.Min(24.0, float64(radPx+vLen*1.8)))
				rl.DrawLine(int32(sx), int32(sy), int32(vxEnd), int32(vyEnd), rl.NewColor(255, 230, 100, 180))
			}
		}

		// Decluttered Labels with stylish semi-transparent dark rounded backing pill
		if cfg.Show2DLabels && (b.Mass >= 0.02 || b.CelestialIcon == IconStar || b.ID == state.SelectedBodyID || cfg.Viewport2DFullscreen) {
			tagCol := rl.NewColor(220, 235, 255, 230)
			tagText := b.Name
			if b.ID == state.SelectedBodyID {
				tagCol = rl.NewColor(0, 255, 230, 255)
				tagText = "▶ " + b.Name
			}
			tLen := int32(rl.MeasureText(tagText, 11))
			tx := int32(sx + radPx + 5)
			ty := int32(sy - 7)

			// Clean pill background so text is never washed out by heatmap or trails
			pillRec := rl.NewRectangle(float32(tx-4), float32(ty-2), float32(tLen+8), 15)
			rl.DrawRectangleRounded(pillRec, 0.3, 4, rl.NewColor(8, 12, 22, 210))
			if b.ID == state.SelectedBodyID {
				rl.DrawRectangleRoundedLines(pillRec, 0.3, 4, rl.NewColor(0, 255, 230, 220))
			}
			rl.DrawText(tagText, tx, ty, 11, tagCol)
		}
	}

	// 9. Render Lagrange Equilibrium Points (L1-L5) with Interactive Selection
	if cfg.ShowLagrangePoints && len(state.Bodies) >= 2 {
		primary, secondary := GetLagrangePair(state)

		if primary != nil && secondary != nil && secondary.Mass >= 0.0001 {
			pts := ComputeLagrangePoints(primary, secondary, cfg.G)
			labels := [5]string{"L1", "L2", "L3", "L4 Trojan", "L5 Greek"}
			colors := [5]rl.Color{
				rl.NewColor(255, 220, 50, 255),  // L1 Gold
				rl.NewColor(80, 220, 255, 255),  // L2 Sky Cyan
				rl.NewColor(255, 120, 80, 255),  // L3 Coral
				rl.NewColor(120, 255, 180, 255), // L4 Emerald
				rl.NewColor(255, 180, 120, 255), // L5 Amber
			}

			for idx := 0; idx < 5; idx++ {
				lx, ly := w2s(pts[idx].Position.X, pts[idx].Position.Z)
				if lx >= float32(vx)-20 && lx <= float32(vx+vw)+20 && ly >= float32(clipY)-20 && ly <= float32(clipY+clipH)+20 {
					dSize := float32(6.0)
					col := colors[idx]
					isSelected := state.SelectedLagrangeIndex == idx+1

					mDist := float32(math.Hypot(float64(mPos.X-lx), float64(mPos.Y-ly)))
					isHovered := mDist <= 14.0

					// Pulsating selection halo
					if isSelected {
						pulse := float32(math.Sin(float64(rl.GetTime()*5.0)))*0.5 + 0.5
						rl.DrawCircleLines(int32(lx), int32(ly), 12.0+pulse*4.0, rl.Gold)
						rl.DrawCircleLines(int32(lx), int32(ly), 8.0, col)
					} else if isHovered {
						rl.DrawCircleLines(int32(lx), int32(ly), 10.0, rl.White)
					}

					// Diamond crosshair
					rl.DrawLine(int32(lx-dSize), int32(ly), int32(lx+dSize), int32(ly), col)
					rl.DrawLine(int32(lx), int32(ly-dSize), int32(lx), int32(ly+dSize), col)
					rl.DrawPolyLines(rl.NewVector2(lx, ly), 4, dSize*1.3, 45.0, col)

					// Libration zone halo preview
					rl.DrawCircleLines(int32(lx), int32(ly), dSize*2.4, rl.Fade(col, 0.35))

					if cfg.Show2DLabels {
						tag := labels[idx]
						if isSelected {
							tag = "★ " + tag
						}
						tLen := int32(rl.MeasureText(tag, 11))
						tagX := int32(lx + 8)
						tagY := int32(ly - 7)

						// Rounded pill backing for crisp readability
						pillRec := rl.NewRectangle(float32(tagX-3), float32(tagY-2), float32(tLen+6), 15)
						rl.DrawRectangleRounded(pillRec, 0.3, 4, rl.NewColor(10, 14, 25, 225))
						borderCol := rl.Fade(col, 0.8)
						if isSelected {
							borderCol = rl.Gold
						}
						rl.DrawRectangleRoundedLines(pillRec, 0.3, 4, borderCol)
						rl.DrawText(tag, tagX, tagY, 11, col)
					}
				}
			}

			// 10. Floating Quick-Action Card when a Lagrange point is selected
			if state.SelectedLagrangeIndex > 0 && state.SelectedLagrangeIndex <= 5 {
				selIdx := state.SelectedLagrangeIndex - 1
				selPt := pts[selIdx]
				selCol := colors[selIdx]

				cardW := int32(310)
				if cardW > vw-20 {
					cardW = vw - 20
				}
				cardH := int32(76)
				cardX := vx + 10
				cardY := clipY + clipH - cardH - 8

				cardRec := rl.NewRectangle(float32(cardX), float32(cardY), float32(cardW), float32(cardH))
				rl.DrawRectangleRec(cardRec, rl.NewColor(10, 16, 28, 248))
				rl.DrawRectangleLinesEx(cardRec, 1.5, selCol)

				DrawTextBoldUI(fmt.Sprintf("POINT: %s", selPt.Name), cardX+8, cardY+6, 12, selCol)
				stabText := "UNSTABLE"
				stabCol := rl.NewColor(255, 140, 100, 255)
				if selPt.Stable {
					stabText = "STABLE"
					stabCol = rl.NewColor(100, 240, 160, 255)
				}
				rl.DrawText(stabText, cardX+cardW-int32(rl.MeasureText(stabText, 10))-26, cardY+7, 10, stabCol)

				velMag := float32(rl.Vector3Length(selPt.Velocity))
				tStr := fmt.Sprintf("Pos: (%.1f, %.1f) | Vel: (%.1f, %.1f) | |V|: %.2f", selPt.Position.X, selPt.Position.Z, selPt.Velocity.X, selPt.Velocity.Z, velMag)
				rl.DrawText(tStr, cardX+8, cardY+24, 10, rl.NewColor(170, 205, 240, 230))

				bY := cardY + 44
				bH := int32(22)
				bX := cardX + 8

				// [+Probe]
				drawSmallBtn("+Probe", bX, bY, 84, bH, false, rl.NewColor(35, 80, 130, 240))
				bX += 90

				// If L4 or L5, [+Swarm]
				if selIdx == 3 || selIdx == 4 {
					drawSmallBtn("+Swarm", bX, bY, 86, bH, false, rl.NewColor(30, 100, 70, 240))
					bX += 92
				}

				// [Track]
				trackCol := rl.NewColor(40, 60, 90, 240)
				if state.FollowSelected {
					trackCol = rl.NewColor(30, 110, 150, 240)
				}
				drawSmallBtn("Track", bX, bY, 64, bH, state.FollowSelected, trackCol)

				// [X]
				drawSmallBtn("X", cardX+cardW-26, cardY+6, 20, 20, false, rl.NewColor(80, 30, 30, 240))
			}
		}
	}

	// 10. Animated Target Reticle when Track2D is active
	if cfg.Track2D && state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				sx, sy := w2s(b.Position.X, b.Position.Z)
				radPx := float32(2.5+math.Log10(math.Max(1.0, b.Mass+1.0))*2.4) + 6.0
				pulse := float32(math.Sin(float64(rl.GetTime()*6.0)))*2.0 + 2.0
				rLock := radPx + pulse

				retCol := rl.NewColor(20, 240, 200, 230)
				bLen := float32(6.0)
				// Top-left
				rl.DrawLine(int32(sx-rLock), int32(sy-rLock), int32(sx-rLock+bLen), int32(sy-rLock), retCol)
				rl.DrawLine(int32(sx-rLock), int32(sy-rLock), int32(sx-rLock), int32(sy-rLock+bLen), retCol)
				// Top-right
				rl.DrawLine(int32(sx+rLock), int32(sy-rLock), int32(sx+rLock-bLen), int32(sy-rLock), retCol)
				rl.DrawLine(int32(sx+rLock), int32(sy-rLock), int32(sx+rLock), int32(sy-rLock+bLen), retCol)
				// Bottom-left
				rl.DrawLine(int32(sx-rLock), int32(sy+rLock), int32(sx-rLock+bLen), int32(sy+rLock), retCol)
				rl.DrawLine(int32(sx-rLock), int32(sy+rLock), int32(sx-rLock), int32(sy+rLock-bLen), retCol)
				// Bottom-right
				rl.DrawLine(int32(sx+rLock), int32(sy+rLock), int32(sx+rLock-bLen), int32(sy+rLock), retCol)
				rl.DrawLine(int32(sx+rLock), int32(sy+rLock), int32(sx+rLock), int32(sy+rLock-bLen), retCol)

				trackTag := "⛶ TRACK LOCKED"
				tLen := int32(rl.MeasureText(trackTag, 9))
				rl.DrawText(trackTag, int32(sx)-tLen/2, int32(sy+rLock+4), 9, retCol)
				break
			}
		}
	}

	// 11. 2D Spawning Trajectory Overlay
	if state.IsSpawning {
		// Active spawn banner at top of viewport content area
		bannerText := fmt.Sprintf("▶ 2D SPAWN ACTIVE: Click & Drag slingshot to launch %s", state.SpawnPreset.String())
		bLen := int32(rl.MeasureText(bannerText, 11))
		bX := vx + (vw-bLen)/2
		bY := clipY + 6
		bRec := rl.NewRectangle(float32(bX-8), float32(bY-2), float32(bLen+16), 18)
		rl.DrawRectangleRounded(bRec, 0.3, 4, rl.NewColor(18, 28, 48, 230))
		rl.DrawRectangleRoundedLines(bRec, 0.3, 4, rl.NewColor(255, 140, 40, 220))
		rl.DrawText(bannerText, bX, bY, 11, rl.NewColor(255, 210, 100, 255))

		// Draw slingshot velocity vector when dragging
		if state.IsDraggingSpawn {
			sx0, sy0 := w2s(state.SpawnWorldPos.X, state.SpawnWorldPos.Z)
			rl.DrawCircleLines(int32(sx0), int32(sy0), 8.0, rl.NewColor(255, 180, 40, 240))
			rl.DrawCircle(int32(sx0), int32(sy0), 3.0, rl.Gold)

			velLen := float32(math.Sqrt(float64(state.DragVelocity.X*state.DragVelocity.X + state.DragVelocity.Z*state.DragVelocity.Z)))
			if velLen > 0.01 {
				sx1, sy1 := w2s(state.SpawnWorldPos.X+state.DragVelocity.X*4.0, state.SpawnWorldPos.Z+state.DragVelocity.Z*4.0)
				rl.DrawLineEx(rl.NewVector2(sx0, sy0), rl.NewVector2(sx1, sy1), 2.5, rl.NewColor(50, 255, 150, 240))
				spdText := fmt.Sprintf("|V| = %.2f", velLen)
				rl.DrawText(spdText, int32(sx1+6), int32(sy1-6), 10, rl.NewColor(120, 255, 180, 255))
			}
		}
	}

	rl.EndScissorMode()

	// 12. Responsive Bottom Status Bar (Never cut off or overflowed!)
	rl.DrawRectangle(vx, vy+vh-statusH, vw, statusH, rl.NewColor(8, 14, 25, 252))
	rl.DrawLine(vx, vy+vh-statusH, vx+vw, vy+vh-statusH, rl.NewColor(35, 55, 90, 180))

	legText := "[R-Drag] Pan  ·  [Wheel] Zoom  ·  [L-Click] Select/Spawn  ·  [F] Focus  ·  [T] Track  ·  [Z/F5] 3D Offload"
	if vw < 540 {
		legText = "[R-Drag] Pan · [Wheel] Zoom · [L-Click] Select/Spawn · [F] Focus · [T] Track"
	}
	rl.DrawText(legText, vx+10, vy+vh-statusH+5, 10, rl.NewColor(165, 195, 235, 230))

	// Live Telemetry on right side of 2D status bar (Always visible, zero obstruction!)
	if vw >= 680 {
		fps := rl.GetFPS()
		method := "CPU"
		if cfg.UseGPUCompute && IsGPUComputeAvailable() {
			method = "GPU"
		} else if cfg.UseBarnesHut || len(state.Bodies) >= 350 {
			method = "B-H"
		}
		statStr := fmt.Sprintf("FPS: %d · Bodies: %d · Phys: %.1fms · %s · G: %.1f", fps, len(state.Bodies), cfg.PhysicsTimeMs, method, cfg.G)
		statW := int32(rl.MeasureText(statStr, 10))
		if vx+vw-statW-12 > vx+int32(rl.MeasureText(legText, 10))+24 {
			fpsCol := rl.NewColor(160, 210, 255, 230)
			if fps >= 135 {
				fpsCol = rl.Lime
			} else if fps < 60 {
				fpsCol = rl.Orange
			}
			rl.DrawText(statStr, vx+vw-statW-12, vy+vh-statusH+5, 10, fpsCol)
		}
	}
}
