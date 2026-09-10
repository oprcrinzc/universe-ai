package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// UIState handles UI interaction events
type UIState struct {
	HoveredButton bool
	MouseCaptured bool
}

// DrawUI renders all HUD elements, inspector panel, toolbars, and help screens
func DrawUI(state *SimState, cfg *Config, camera *OrbitCamera, screenW, screenH int32) bool {
	mousePos := rl.GetMousePosition()
	mouseCaptured := false

	// 1. Top Navigation Bar
	if drawTopBar(state, cfg, camera, screenW, mousePos) {
		mouseCaptured = true
	}

	// 2. Selected Body Inspector Panel (Right side)
	if state.SelectedBodyID != -1 {
		if drawInspectorPanel(state, cfg, camera, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 3. Bottom Toolbar (Spawner & View toggles)
	if drawBottomBar(state, cfg, screenW, screenH, mousePos) {
		mouseCaptured = true
	}

	// 4. Help Overlay
	if cfg.ShowHelp {
		if drawHelpModal(cfg, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 5. FPS & Stats HUD (Top Right)
	drawStatsHUD(state, cfg, screenW)

	return mouseCaptured
}

func drawButton(rec rl.Rectangle, text string, active, hovered bool, baseCol rl.Color) bool {
	clicked := false
	col := baseCol

	if hovered {
		col = rl.ColorAlpha(col, 0.9)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			clicked = true
		}
	} else {
		col = rl.ColorAlpha(col, 0.65)
	}

	if active {
		col = rl.NewColor(50, 120, 220, 220)
	}

	rl.DrawRectangleRec(rec, col)
	borderCol := rl.NewColor(120, 160, 220, 200)
	if active {
		borderCol = rl.Gold
	}
	rl.DrawRectangleLinesEx(rec, 1.5, borderCol)

	textLen := rl.MeasureText(text, 12)
	textX := int32(rec.X + (rec.Width-float32(textLen))/2)
	textY := int32(rec.Y + (rec.Height-12)/2)

	textColor := rl.White
	if active {
		textColor = rl.Yellow
	}
	rl.DrawText(text, textX, textY, 12, textColor)

	return clicked
}

func drawTopBar(state *SimState, cfg *Config, camera *OrbitCamera, screenW int32, mousePos rl.Vector2) bool {
	barHeight := float32(46)
	topBarRec := rl.NewRectangle(0, 0, float32(screenW), barHeight)
	rl.DrawRectangleRec(topBarRec, rl.NewColor(18, 22, 32, 230))
	rl.DrawRectangleLinesEx(topBarRec, 1, rl.NewColor(45, 55, 80, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, topBarRec)

	// Title
	rl.DrawText("GRAVITY MASS SIMULATOR", 16, 15, 16, rl.Gold)

	x := float32(230)
	btnW := float32(64)
	btnH := float32(28)
	btnY := float32(9)

	// Play / Pause
	playText := "PAUSE"
	if cfg.Paused {
		playText = "PLAY"
	}
	playRec := rl.NewRectangle(x, btnY, btnW, btnH)
	if drawButton(playRec, playText, !cfg.Paused, rl.CheckCollisionPointRec(mousePos, playRec), rl.NewColor(30, 70, 40, 255)) {
		cfg.Paused = !cfg.Paused
	}
	x += btnW + 8

	// Speed controls
	speedText := fmt.Sprintf("%.1fx", cfg.TimeScale)
	slowRec := rl.NewRectangle(x, btnY, 32, btnH)
	if drawButton(slowRec, "<", false, rl.CheckCollisionPointRec(mousePos, slowRec), rl.NewColor(40, 45, 65, 255)) {
		cfg.TimeScale = math.Max(0.1, cfg.TimeScale*0.5)
	}
	x += 34

	spdRec := rl.NewRectangle(x, btnY, 50, btnH)
	rl.DrawRectangleRec(spdRec, rl.NewColor(25, 30, 45, 255))
	rl.DrawRectangleLinesEx(spdRec, 1, rl.NewColor(60, 70, 95, 255))
	rl.DrawText(speedText, int32(x+10), int32(btnY+8), 12, rl.SkyBlue)
	x += 52

	fastRec := rl.NewRectangle(x, btnY, 32, btnH)
	if drawButton(fastRec, ">", false, rl.CheckCollisionPointRec(mousePos, fastRec), rl.NewColor(40, 45, 65, 255)) {
		cfg.TimeScale = math.Min(10.0, cfg.TimeScale*2.0)
	}
	x += 42

	// Presets separator
	rl.DrawLineEx(rl.NewVector2(x, 10), rl.NewVector2(x, 36), 1, rl.NewColor(60, 70, 100, 255))
	x += 10

	presets := []struct {
		name string
		val  PresetType
	}{
		{"Solar System", PresetSolarSystem},
		{"Binary Stars", PresetBinaryStars},
		{"3-Body Problem", PresetThreeBody},
		{"Galaxy Disk", PresetGalaxyDisk},
		{"Empty", PresetEmpty},
	}

	for _, p := range presets {
		pW := float32(rl.MeasureText(p.name, 12) + 16)
		pRec := rl.NewRectangle(x, btnY, pW, btnH)
		isActive := (state.CurrentPreset == p.val)
		if drawButton(pRec, p.name, isActive, rl.CheckCollisionPointRec(mousePos, pRec), rl.NewColor(35, 45, 70, 255)) {
			LoadPreset(state, cfg, p.val)
		}
		x += pW + 6
	}

	// Reset button
	resetRec := rl.NewRectangle(x, btnY, 60, btnH)
	if drawButton(resetRec, "Reset", false, rl.CheckCollisionPointRec(mousePos, resetRec), rl.NewColor(75, 40, 35, 255)) {
		LoadPreset(state, cfg, state.CurrentPreset)
	}
	x += 66

	// Help Button (F1)
	helpRec := rl.NewRectangle(float32(screenW)-50, btnY, 36, btnH)
	if drawButton(helpRec, "?", cfg.ShowHelp, rl.CheckCollisionPointRec(mousePos, helpRec), rl.NewColor(60, 60, 80, 255)) {
		cfg.ShowHelp = !cfg.ShowHelp
	}

	return mouseInside
}

func drawInspectorPanel(state *SimState, cfg *Config, camera *OrbitCamera, screenW, screenH int32, mousePos rl.Vector2) bool {
	var body *Body
	for _, b := range state.Bodies {
		if b.ID == state.SelectedBodyID {
			body = b
			break
		}
	}
	if body == nil {
		state.SelectedBodyID = -1
		return false
	}

	panelW := float32(280)
	panelH := float32(460)
	panelX := float32(screenW) - panelW - 12
	panelY := float32(56)

	panelRec := rl.NewRectangle(panelX, panelY, panelW, panelH)
	rl.DrawRectangleRec(panelRec, rl.NewColor(20, 25, 38, 245))
	rl.DrawRectangleLinesEx(panelRec, 1.5, rl.NewColor(65, 85, 125, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, panelRec)

	// Panel Title
	rl.DrawText("BODY INSPECTOR & EDIT", int32(panelX+14), int32(panelY+12), 14, rl.Gold)

	// Close button [X]
	closeRec := rl.NewRectangle(panelX+panelW-30, panelY+8, 22, 22)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		state.SelectedBodyID = -1
		return true
	}

	y := panelY + 38

	// Name & Badge
	rl.DrawText(fmt.Sprintf("Name: %s", body.Name), int32(panelX+14), int32(y), 13, rl.White)
	y += 20

	// Status badge (Stationary vs Free)
	statusText := "Status: ORBITING (Dynamic)"
	statusCol := rl.Lime
	if body.IsStationary {
		statusText = "Status: STATIONARY (Fixed Anchor)"
		statusCol = rl.Red
	}
	rl.DrawText(statusText, int32(panelX+14), int32(y), 12, statusCol)
	y += 24

	// Stationary Toggle Button (Crucial feature from prompt!)
	statBtnText := "Make Stationary (Lock)"
	if body.IsStationary {
		statBtnText = "Unfreeze (Allow Orbit)"
	}
	statRec := rl.NewRectangle(panelX+14, y, panelW-28, 26)
	if drawButton(statRec, statBtnText, body.IsStationary, rl.CheckCollisionPointRec(mousePos, statRec), rl.NewColor(70, 40, 90, 255)) {
		body.IsStationary = !body.IsStationary
		if body.IsStationary {
			body.Velocity = rl.NewVector3(0, 0, 0)
		}
	}
	y += 34

	// Mass controls
	rl.DrawText(fmt.Sprintf("Mass: %.1f", body.Mass), int32(panelX+14), int32(y), 12, rl.White)
	y += 18

	btnW := float32(56)
	mRec1 := rl.NewRectangle(panelX+14, y, btnW, 22)
	if drawButton(mRec1, "-50%", false, rl.CheckCollisionPointRec(mousePos, mRec1), rl.NewColor(40, 50, 70, 255)) {
		body.Mass = math.Max(0.01, body.Mass*0.5)
	}
	mRec2 := rl.NewRectangle(panelX+14+btnW+6, y, btnW, 22)
	if drawButton(mRec2, "-10%", false, rl.CheckCollisionPointRec(mousePos, mRec2), rl.NewColor(40, 50, 70, 255)) {
		body.Mass = math.Max(0.01, body.Mass*0.9)
	}
	mRec3 := rl.NewRectangle(panelX+14+(btnW+6)*2, y, btnW, 22)
	if drawButton(mRec3, "+10%", false, rl.CheckCollisionPointRec(mousePos, mRec3), rl.NewColor(40, 50, 70, 255)) {
		body.Mass *= 1.1
	}
	mRec4 := rl.NewRectangle(panelX+14+(btnW+6)*3, y, btnW, 22)
	if drawButton(mRec4, "x2", false, rl.CheckCollisionPointRec(mousePos, mRec4), rl.NewColor(40, 50, 70, 255)) {
		body.Mass *= 2.0
	}
	y += 30

	// Radius controls
	rl.DrawText(fmt.Sprintf("Radius: %.2f", body.Radius), int32(panelX+14), int32(y), 12, rl.White)
	rSubRec := rl.NewRectangle(panelX+120, y-2, 30, 20)
	if drawButton(rSubRec, "-", false, rl.CheckCollisionPointRec(mousePos, rSubRec), rl.NewColor(40, 50, 70, 255)) {
		body.Radius = float32(math.Max(0.2, float64(body.Radius-0.2)))
	}
	rAddRec := rl.NewRectangle(panelX+156, y-2, 30, 20)
	if drawButton(rAddRec, "+", false, rl.CheckCollisionPointRec(mousePos, rAddRec), rl.NewColor(40, 50, 70, 255)) {
		body.Radius += 0.2
	}
	y += 24

	// Telemetry
	speed := rl.Vector3Length(body.Velocity)
	rl.DrawText(fmt.Sprintf("Speed: %.2f units/s", speed), int32(panelX+14), int32(y), 12, rl.SkyBlue)
	y += 18
	rl.DrawText(fmt.Sprintf("Pos: (%.1f, %.1f, %.1f)", body.Position.X, body.Position.Y, body.Position.Z), int32(panelX+14), int32(y), 11, rl.LightGray)
	y += 16
	rl.DrawText(fmt.Sprintf("Vel: (%.1f, %.1f, %.1f)", body.Velocity.X, body.Velocity.Y, body.Velocity.Z), int32(panelX+14), int32(y), 11, rl.LightGray)
	y += 22

	// ORBITAL MANEUVERS & FORCE SECTION
	rl.DrawText("MANEUVERS & FORCES:", int32(panelX+14), int32(y), 12, rl.Gold)
	y += 18

	// Circularize Orbit Button
	circRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	if drawButton(circRec, "Circularize Orbit (Auto-V)", false, rl.CheckCollisionPointRec(mousePos, circRec), rl.NewColor(30, 80, 110, 255)) {
		primary := FindPrimaryBody(state.Bodies, body)
		if primary != nil {
			CircularizeOrbit(body, primary, cfg.G, true)
		}
	}
	y += 28

	// Prograde & Retrograde thrust buttons
	halfW := (panelW - 34) / 2
	proRec := rl.NewRectangle(panelX+14, y, halfW, 24)
	if drawButton(proRec, "Prograde (+10%)", false, rl.CheckCollisionPointRec(mousePos, proRec), rl.NewColor(30, 90, 50, 255)) {
		BoostPrograde(body, 0.10)
	}
	retroRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 24)
	if drawButton(retroRec, "Retrograde (-10%)", false, rl.CheckCollisionPointRec(mousePos, retroRec), rl.NewColor(90, 45, 35, 255)) {
		BoostPrograde(body, -0.10)
	}
	y += 28

	// Vertical Nudge (+Y / -Y)
	nudgeUpRec := rl.NewRectangle(panelX+14, y, halfW, 22)
	if drawButton(nudgeUpRec, "Inclination +Y", false, rl.CheckCollisionPointRec(mousePos, nudgeUpRec), rl.NewColor(45, 55, 75, 255)) {
		ApplyImpulse(body, rl.NewVector3(0, 1.5, 0))
	}
	nudgeDownRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 22)
	if drawButton(nudgeDownRec, "Inclination -Y", false, rl.CheckCollisionPointRec(mousePos, nudgeDownRec), rl.NewColor(45, 55, 75, 255)) {
		ApplyImpulse(body, rl.NewVector3(0, -1.5, 0))
	}
	y += 26

	// Stop dead
	stopRec := rl.NewRectangle(panelX+14, y, panelW-28, 22)
	if drawButton(stopRec, "Zero Velocity (Stop Dead)", false, rl.CheckCollisionPointRec(mousePos, stopRec), rl.NewColor(75, 45, 60, 255)) {
		body.Velocity = rl.NewVector3(0, 0, 0)
	}
	y += 28

	// Camera Follow toggle
	followRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	followText := "Track with Camera: OFF"
	if state.FollowSelected {
		followText = "Track with Camera: ON"
	}
	if drawButton(followRec, followText, state.FollowSelected, rl.CheckCollisionPointRec(mousePos, followRec), rl.NewColor(50, 70, 95, 255)) {
		state.FollowSelected = !state.FollowSelected
	}
	y += 28

	// Delete Body
	delRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	if drawButton(delRec, "DELETE BODY", false, rl.CheckCollisionPointRec(mousePos, delRec), rl.NewColor(120, 30, 30, 255)) {
		removeBody(state, body.ID)
		state.SelectedBodyID = -1
		state.FollowSelected = false
	}

	return mouseInside
}

func removeBody(state *SimState, id int64) {
	newBodies := make([]*Body, 0, len(state.Bodies)-1)
	for _, b := range state.Bodies {
		if b.ID != id {
			newBodies = append(newBodies, b)
		}
	}
	state.Bodies = newBodies
}

func drawBottomBar(state *SimState, cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	barH := float32(46)
	barY := float32(screenH) - barH
	botRec := rl.NewRectangle(0, barY, float32(screenW), barH)
	rl.DrawRectangleRec(botRec, rl.NewColor(18, 22, 32, 235))
	rl.DrawRectangleLinesEx(botRec, 1, rl.NewColor(45, 55, 80, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, botRec)

	x := float32(14)
	btnY := barY + 9
	btnH := float32(28)

	// Spawner Toggle
	spawnBtnW := float32(120)
	spawnRec := rl.NewRectangle(x, btnY, spawnBtnW, btnH)
	spawnText := "SPAWN: OFF"
	if state.IsSpawning {
		spawnText = "SPAWN: ACTIVE"
	}
	if drawButton(spawnRec, spawnText, state.IsSpawning, rl.CheckCollisionPointRec(mousePos, spawnRec), rl.NewColor(30, 80, 60, 255)) {
		state.IsSpawning = !state.IsSpawning
	}
	x += spawnBtnW + 8

	// Spawn body types (when spawning is active)
	if state.IsSpawning {
		types := []struct {
			name string
			val  SpawnType
		}{
			{"Earth", SpawnEarth},
			{"Moon", SpawnMoon},
			{"Giant", SpawnGasGiant},
			{"Star", SpawnStar},
			{"BlackHole", SpawnBlackHole},
		}

		for _, t := range types {
			tW := float32(rl.MeasureText(t.name, 11) + 14)
			tRec := rl.NewRectangle(x, btnY, tW, btnH)
			isActive := (state.SpawnPreset == t.val)
			if drawButton(tRec, t.name, isActive, rl.CheckCollisionPointRec(mousePos, tRec), rl.NewColor(35, 55, 75, 255)) {
				state.SpawnPreset = t.val
			}
			x += tW + 4
		}

		rl.DrawText("(Click & Drag on plane to place + launch)", int32(x+10), int32(btnY+8), 11, rl.Gold)
	} else {
		// View toggles when not in spawn mode
		toggles := []struct {
			name  string
			value *bool
		}{
			{"Trails", &cfg.ShowTrails},
			{"Grid", &cfg.ShowGrid},
			{"Vectors", &cfg.ShowVectors},
			{"Forces", &cfg.ShowForces},
			{"Labels", &cfg.ShowLabels},
			{"Stars", &cfg.ShowStarfield},
		}

		for _, tg := range toggles {
			tW := float32(rl.MeasureText(tg.name, 11) + 16)
			tRec := rl.NewRectangle(x, btnY, tW, btnH)
			if drawButton(tRec, tg.name, *tg.value, rl.CheckCollisionPointRec(mousePos, tRec), rl.NewColor(35, 45, 65, 255)) {
				*tg.value = !*tg.value
			}
			x += tW + 6
		}

		// Collision mode toggle
		colModes := []string{"Merge", "Bounce", "Ghost"}
		colText := fmt.Sprintf("Collision: %s", colModes[cfg.Collision])
		colW := float32(rl.MeasureText(colText, 11) + 16)
		colRec := rl.NewRectangle(x, btnY, colW, btnH)
		if drawButton(colRec, colText, false, rl.CheckCollisionPointRec(mousePos, colRec), rl.NewColor(55, 40, 70, 255)) {
			cfg.Collision = (cfg.Collision + 1) % 3
		}
	}

	return mouseInside
}

func drawStatsHUD(state *SimState, cfg *Config, screenW int32) {
	fps := rl.GetFPS()
	bodyCount := len(state.Bodies)

	info := fmt.Sprintf("FPS: %d  |  Bodies: %d  |  G: %.1f", fps, bodyCount, cfg.G)
	infoW := rl.MeasureText(info, 12)
	hudX := screenW - infoW - 70
	rl.DrawText(info, hudX, 17, 12, rl.NewColor(180, 200, 230, 255))
}

func drawHelpModal(cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	modalW := float32(520)
	modalH := float32(420)
	modalX := (float32(screenW) - modalW) / 2
	modalY := (float32(screenH) - modalH) / 2

	// Backdrop
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 150))

	modalRec := rl.NewRectangle(modalX, modalY, modalW, modalH)
	rl.DrawRectangleRec(modalRec, rl.NewColor(22, 28, 42, 250))
	rl.DrawRectangleLinesEx(modalRec, 2, rl.Gold)

	mouseInside := rl.CheckCollisionPointRec(mousePos, modalRec)

	rl.DrawText("SIMULATION CONTROLS & MANUAL", int32(modalX+20), int32(modalY+16), 16, rl.Gold)

	closeRec := rl.NewRectangle(modalX+modalW-36, modalY+12, 24, 24)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		cfg.ShowHelp = false
		return true
	}

	lines := []string{
		"CAMERA CONTROLS:",
		"  * Right Mouse Drag: Orbit / Rotate 3D view",
		"  * Middle Mouse Drag or Shift + Right Drag: Pan camera",
		"  * Mouse Wheel: Zoom in / out",
		"  * W / A / S / D / Q / E: Keyboard camera pan",
		"  * F: Focus camera on selected body",
		"  * C: Toggle auto-tracking selected body",
		"  * R: Reset camera view to default",
		"",
		"SELECTION & FORCES:",
		"  * Left Click: Select body or deselect",
		"  * Inspector Panel: Make Stationary, Circularize orbit, Boost Prograde/Retrograde",
		"  * Delete key / Backspace: Remove selected body",
		"",
		"SPAWNING BODIES:",
		"  * Toggle SPAWN button in the bottom bar",
		"  * Choose type (Earth, Moon, Gas Giant, Star, Black Hole)",
		"  * Click & Drag on the orbital grid to place and impart initial launch velocity!",
		"",
		"HOTKEYS: Space (Pause), F1 (Help), F12 (Screenshot), 1-5 (Presets)",
	}

	y := int32(modalY + 54)
	for _, l := range lines {
		if len(l) > 0 && l[0] != ' ' {
			rl.DrawText(l, int32(modalX+20), y, 12, rl.Yellow)
		} else {
			rl.DrawText(l, int32(modalX+20), y, 12, rl.LightGray)
		}
		y += 18
	}

	return mouseInside
}
