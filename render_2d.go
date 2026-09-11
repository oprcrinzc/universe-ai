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

	// 1. Interactive Header Bar Buttons
	if mPos.Y <= float32(vy+headerH) {
		btnY := float32(vy + 5)
		btnH := float32(22)

		checkBtn := func(bx, bw float32) bool {
			return mPos.X >= bx && mPos.X <= bx+bw && mPos.Y >= btnY && mPos.Y <= btnY+btnH
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			// From right to left
			// [X] Close (24px)
			if checkBtn(float32(vx+vw-28), 24) {
				cfg.Show2DViewport = false
				return true
			}
			// [⛶] Fullscreen (24px)
			if checkBtn(float32(vx+vw-56), 24) {
				cfg.Viewport2DFullscreen = !cfg.Viewport2DFullscreen
				return true
			}
			// [RST] Reset Pan/Zoom (34px)
			if checkBtn(float32(vx+vw-94), 34) {
				cfg.Viewport2DZoom = 0.45
				cfg.Viewport2DPan = rl.NewVector2(0, 0)
				cfg.NotificationText = "2D Map View: RESET"
				cfg.NotificationTimer = 1.2
				return true
			}
			// [VEC] Vector Field (32px)
			if checkBtn(float32(vx+vw-130), 32) {
				cfg.Show2DVectorField = !cfg.Show2DVectorField
				return true
			}
			// [HEAT] Heatmap (38px)
			if checkBtn(float32(vx+vw-172), 38) {
				cfg.Show2DHeatmap = !cfg.Show2DHeatmap
				return true
			}
			// [LBL] Labels (34px)
			if checkBtn(float32(vx+vw-210), 34) {
				cfg.Show2DLabels = !cfg.Show2DLabels
				status := "OFF"
				if cfg.Show2DLabels {
					status = "ON"
				}
				cfg.NotificationText = fmt.Sprintf("2D Map Labels: %s", status)
				cfg.NotificationTimer = 1.2
				return true
			}
			// [+] Zoom In (20px)
			if checkBtn(float32(vx+vw-234), 20) {
				cfg.Viewport2DZoom = float32(math.Min(80.0, float64(cfg.Viewport2DZoom*1.35)))
				return true
			}
			// [-] Zoom Out (20px)
			if checkBtn(float32(vx+vw-258), 20) {
				cfg.Viewport2DZoom = float32(math.Max(0.005, float64(cfg.Viewport2DZoom*0.75)))
				return true
			}
		}
		return true // Mouse is over header bar; consume event so it doesn't leak to 3D scene
	}

	// 2. Mouse wheel zoom centered on cursor
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

		cx := float32(vx) + float32(vw)*0.5
		cy := float32(vy+headerH) + float32(vh-headerH)*0.5
		cursorWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/oldZoom
		cursorWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/oldZoom

		cfg.Viewport2DZoom = newZoom
		cfg.Viewport2DPan.X = cursorWorldX - (mPos.X-cx)/newZoom
		cfg.Viewport2DPan.Y = cursorWorldZ - (mPos.Y-cy)/newZoom
		return true
	}

	// 3. Right-click or Middle-click drag to pan 2D map
	if (rl.IsMouseButtonDown(rl.MouseRightButton) || rl.IsMouseButtonDown(rl.MouseMiddleButton)) &&
		(mPos.Y > float32(vy+headerH)) {
		delta := rl.GetMouseDelta()
		cfg.Viewport2DPan.X -= delta.X / cfg.Viewport2DZoom
		cfg.Viewport2DPan.Y -= delta.Y / cfg.Viewport2DZoom
		return true
	}

	// 4. Left-click to select body in 2D viewport
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && mPos.Y > float32(vy+headerH) {
		cx := float32(vx) + float32(vw)*0.5
		cy := float32(vy+headerH) + float32(vh-headerH)*0.5
		clickWorldX := cfg.Viewport2DPan.X + (mPos.X-cx)/cfg.Viewport2DZoom
		clickWorldZ := cfg.Viewport2DPan.Y + (mPos.Y-cy)/cfg.Viewport2DZoom

		var closestID int64 = -1
		var closestDist float32 = 25.0 / cfg.Viewport2DZoom // 25 px radius

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
			return true
		}
	}

	return true // Keep capturing mouse whenever inside the 2D viewport rect
}

// get2DViewportRect computes responsive dimensions and position on the right side,
// intelligently shifting left when the Inspector panel is open to guarantee ZERO overlap
func get2DViewportRect(state *SimState, cfg *Config, screenW, screenH int32) (int32, int32, int32, int32) {
	if cfg.Viewport2DFullscreen {
		return 10, 52, screenW - 20, screenH - 104
	}

	w := int32(480)
	h := int32(360)
	if screenW < 1366 {
		w = int32(380)
		h = int32(290)
	}
	if screenW < 1080 {
		w = int32(320)
		h = int32(240)
	}

	y := int32(52)
	maxH := screenH - y - 54
	if h > maxH {
		h = maxH
	}

	// RIGHT SIDE DOCKING:
	// If the Inspector panel is open on the right (width 310, margin 12),
	// shift the 2D map to the left of the Inspector panel so they sit side-by-side with zero overlap!
	inspectorOpen := (state != nil && state.SelectedBodyID != -1)
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

// Draw2DViewport renders the top-down orthographic 2D orbital map, gravitational heatmap,
// color-coded vector field, projected masses, and Lagrange points
func Draw2DViewport(state *SimState, cfg *Config, screenW, screenH int32) {
	if !cfg.Show2DViewport {
		return
	}

	vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)

	// Ensure valid zoom & pan defaults
	if cfg.Viewport2DZoom <= 0.0001 {
		cfg.Viewport2DZoom = 0.45
	}

	// 1. Draw outer frame & background
	rl.DrawRectangle(vx, vy, vw, vh, rl.NewColor(7, 10, 20, 240))
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(vx), float32(vy), float32(vw), float32(vh)), 2.0, rl.NewColor(40, 140, 230, 220))

	// Header Bar
	headerH := int32(32)
	rl.DrawRectangle(vx, vy, vw, headerH, rl.NewColor(12, 22, 45, 252))
	rl.DrawLine(vx, vy+headerH, vx+vw, vy+headerH, rl.NewColor(50, 160, 240, 180))

	title := "🛰️ 2D TACTICAL MAP"
	if cfg.Viewport2DFullscreen {
		title = "🛰️ 2D TACTICAL MAP (FULLSCREEN) [M to Minimize]"
	}
	rl.DrawText(title, vx+10, vy+8, 14, rl.RayWhite)

	// Draw scale telemetry
	scaleSpan := float32(vw) / cfg.Viewport2DZoom
	scaleText := fmt.Sprintf("%.0f AU (%.2fx)", scaleSpan/100.0, cfg.Viewport2DZoom)
	titleW := MeasureTextBoldUI(title, 14)
	rl.DrawText(scaleText, vx+int32(titleW)+18, vy+9, 12, rl.NewColor(160, 220, 255, 220))

	// Interactive Header Control Buttons
	btnY := vy + 5
	btnH := int32(22)
	mPos := rl.GetMousePosition()

	drawSmallBtn := func(label string, bx, bw int32, active bool, baseCol rl.Color) {
		col := baseCol
		if active {
			col = rl.NewColor(col.R+40, col.G+40, col.B+40, 255)
		}
		hover := mPos.X >= float32(bx) && mPos.X <= float32(bx+bw) &&
			mPos.Y >= float32(btnY) && mPos.Y <= float32(btnY+btnH)
		if hover {
			col = rl.NewColor(col.R+60, col.G+60, col.B+60, 255)
		}
		rl.DrawRectangle(bx, btnY, bw, btnH, col)
		borderCol := rl.NewColor(100, 180, 240, 180)
		if active {
			borderCol = rl.Gold
		}
		rl.DrawRectangleLines(bx, btnY, bw, btnH, borderCol)

		fSize := int32(11)
		tLen := rl.MeasureText(label, fSize)
		rl.DrawText(label, bx+(bw-tLen)/2, btnY+5, fSize, rl.White)
	}

	// Render Header buttons from right to left
	// [X] Close
	drawSmallBtn("X", vx+vw-28, 24, false, rl.NewColor(90, 35, 35, 240))
	// [⛶] Fullscreen
	drawSmallBtn("⛶", vx+vw-56, 24, cfg.Viewport2DFullscreen, rl.NewColor(35, 60, 100, 240))
	// [RST] Reset Pan/Zoom
	drawSmallBtn("RST", vx+vw-94, 34, false, rl.NewColor(40, 60, 90, 240))
	// [VEC] Vector field
	drawSmallBtn("VEC", vx+vw-130, 32, cfg.Show2DVectorField, rl.NewColor(30, 75, 55, 240))
	// [HEAT] Heatmap
	drawSmallBtn("HEAT", vx+vw-172, 38, cfg.Show2DHeatmap, rl.NewColor(85, 55, 25, 240))
	// [LBL] Labels toggle
	lblCol := rl.NewColor(35, 50, 75, 240)
	if cfg.Show2DLabels {
		lblCol = rl.NewColor(25, 80, 120, 240)
	}
	drawSmallBtn("LBL", vx+vw-210, 34, cfg.Show2DLabels, lblCol)
	// [+] Zoom In
	drawSmallBtn("+", vx+vw-234, 20, false, rl.NewColor(35, 55, 85, 240))
	// [-] Zoom Out
	drawSmallBtn("-", vx+vw-258, 20, false, rl.NewColor(35, 55, 85, 240))

	// Content area clipping
	clipY := vy + headerH
	clipH := vh - headerH
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

	// 2. Render 2D Gravitational Heatmap
	if cfg.Show2DHeatmap && len(state.Bodies) > 0 {
		// Grid resolution for heatmap quads
		cols := 36
		rows := 26
		cellW := float32(vw) / float32(cols)
		cellH := float32(clipH) / float32(rows)

		// Gather active source bodies
		sources := make([]*Body, 0, 32)
		for _, b := range state.Bodies {
			if b.Mass >= 0.01 || b.IsStar || b.ID == state.SelectedBodyID {
				sources = append(sources, b)
				if len(sources) >= 32 {
					break
				}
			}
		}
		if len(sources) == 0 && len(state.Bodies) > 0 {
			sources = append(sources, state.Bodies[0])
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
				for _, src := range sources {
					dx := src.Position.X - wx
					dz := src.Position.Z - wz
					distSq := dx*dx + dz*dz + eps2
					dist := float32(math.Sqrt(float64(distSq)))
					f := gVal * float32(src.Mass) / (distSq * dist)
					gx += dx * f
					gz += dz * f
				}
				gLen := float32(math.Sqrt(float64(gx*gx + gz*gz)))

				// Heatmap color spectrum
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

	// 3. Render 2D Vector Field
	if cfg.Show2DVectorField && len(state.Bodies) > 0 {
		vCols := 20
		vRows := 15
		stepX := float32(vw) / float32(vCols)
		stepY := float32(clipH) / float32(vRows)

		sources := state.Bodies
		if len(sources) > 32 {
			sources = sources[:32]
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
				for _, src := range sources {
					dx := src.Position.X - wx
					dz := src.Position.Z - wz
					distSq := dx*dx + dz*dz + eps2
					dist := float32(math.Sqrt(float64(distSq)))
					f := gVal * float32(src.Mass) / (distSq * dist)
					gx += dx * f
					gz += dz * f
				}
				gLen := float32(math.Sqrt(float64(gx*gx + gz*gz)))
				if gLen > 0.0001 {
					dirX := gx / gLen
					dirZ := gz / gLen

					arrowLen := float32(4.0 + 12.0*float64(gLen/(gLen+0.4)))
					sx1 := sx0 + dirX*arrowLen
					sy1 := sy0 + dirZ*arrowLen

					vecCol := rl.NewColor(120, 230, 255, 200)
					if gLen > 0.5 {
						vecCol = rl.NewColor(255, 210, 60, 220)
					}
					if gLen > 2.0 {
						vecCol = rl.NewColor(255, 80, 50, 240)
					}

					rl.DrawLine(int32(sx0), int32(sy0), int32(sx1), int32(sy1), vecCol)

					// Arrowhead wing
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

	// 4. Render Orbit Reference Rings & Trails
	for _, b := range state.Bodies {
		if b.IsStar || b.Mass >= 0.01 {
			// Draw recent 2D trail
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

	// 5. Render 2D Celestial Masses
	for _, b := range state.Bodies {
		sx, sy := w2s(b.Position.X, b.Position.Z)

		// Cull bodies far outside viewport
		if sx < float32(vx)-50 || sx > float32(vx+vw)+50 || sy < float32(clipY)-50 || sy > float32(clipY+clipH)+50 {
			continue
		}

		// Logarithmic pixel radius
		radPx := float32(2.5 + math.Log10(math.Max(1.0, b.Mass+1.0))*2.4)
		if b.IsStar {
			radPx = float32(math.Max(float64(radPx), 5.5))
		}
		if b.ID == state.SelectedBodyID {
			radPx += 2.0
		}

		// Body circle
		rl.DrawCircle(int32(sx), int32(sy), radPx, b.Color)

		// Glowing halo for stars or selected
		if b.IsStar || b.ID == state.SelectedBodyID {
			haloCol := rl.NewColor(b.Color.R, b.Color.G, b.Color.B, 120)
			if b.ID == state.SelectedBodyID {
				haloCol = rl.NewColor(0, 255, 230, 200)
			}
			rl.DrawCircleLines(int32(sx), int32(sy), radPx+3.0, haloCol)
		}

		// Velocity vector indicator
		vLen := float32(math.Sqrt(float64(b.Velocity.X*b.Velocity.X + b.Velocity.Z*b.Velocity.Z)))
		if vLen > 0.05 {
			vxEnd := sx + (b.Velocity.X/vLen)*float32(math.Min(24.0, float64(radPx+vLen*1.8)))
			vyEnd := sy + (b.Velocity.Z/vLen)*float32(math.Min(24.0, float64(radPx+vLen*1.8)))
			rl.DrawLine(int32(sx), int32(sy), int32(vxEnd), int32(vyEnd), rl.NewColor(255, 230, 100, 180))
		}

		// Labels for major masses (toggleable via cfg.Show2DLabels)
		if cfg.Show2DLabels && (b.Mass >= 0.02 || b.IsStar || b.ID == state.SelectedBodyID || cfg.Viewport2DFullscreen) {
			tagCol := rl.NewColor(220, 235, 255, 230)
			if b.ID == state.SelectedBodyID {
				tagCol = rl.NewColor(0, 255, 230, 255)
			}
			rl.DrawText(b.Name, int32(sx+radPx+3), int32(sy-6), 11, tagCol)
		}
	}

	// 6. Render Lagrange Equilibrium Points (L1-L5)
	if cfg.ShowLagrangePoints && len(state.Bodies) >= 2 {
		// Find two heaviest primary masses
		var primary, secondary *Body
		for _, b := range state.Bodies {
			if primary == nil || b.Mass > primary.Mass {
				secondary = primary
				primary = b
			} else if secondary == nil || b.Mass > secondary.Mass {
				secondary = b
			}
		}

		if primary != nil && secondary != nil && secondary.Mass >= 0.001 {
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
				if lx >= float32(vx) && lx <= float32(vx+vw) && ly >= float32(clipY) && ly <= float32(clipY+clipH) {
					// Draw diamond crosshair
					dSize := float32(5.0)
					rl.DrawLine(int32(lx-dSize), int32(ly), int32(lx+dSize), int32(ly), colors[idx])
					rl.DrawLine(int32(lx), int32(ly-dSize), int32(lx), int32(ly+dSize), colors[idx])
					rl.DrawPolyLines(rl.NewVector2(lx, ly), 4, dSize*1.2, 45.0, colors[idx])
					if cfg.Show2DLabels {
						rl.DrawText(labels[idx], int32(lx+7), int32(ly-6), 11, colors[idx])
					}
				}
			}
		}
	}

	// 7. Mini Legend at bottom of 2D Viewport
	legY := clipY + clipH - 22
	rl.DrawRectangle(vx+8, legY, vw-16, 18, rl.NewColor(10, 15, 30, 210))
	legendText := "[Right-Drag] Pan | [Wheel] Zoom | [K] Labels | [H] Heatmap | [M] Fullscreen"
	rl.DrawText(legendText, vx+14, legY+3, 11, rl.NewColor(180, 210, 240, 220))

	rl.EndScissorMode()
}
