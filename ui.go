package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// UIState handles UI interaction events
type UIState struct {
	HoveredButton bool
	MouseCaptured bool
}

// DrawUI renders all HUD elements, inspector panel, toolbars, modal dialogs, and help screens
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
	} else if state.SelectedLagrangeIndex > 0 {
		if drawLagrangeInspectorPanel(state, cfg, camera, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 2b. Camera Control Panel (Left side)
	if state.ShowCameraPanel {
		if drawCameraPanel(state, cfg, camera, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 3. Bottom Toolbar (Spawner & View toggles)
	if drawBottomBar(state, cfg, screenW, screenH, mousePos) {
		mouseCaptured = true
	}

	// 4. Scenario Selector Modal
	if state.ShowScenarioModal {
		if drawScenarioModal(state, cfg, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 5. Help Overlay
	if cfg.ShowHelp {
		if drawHelpModal(cfg, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	// 6. FPS & Physics Telemetry HUD (Top Left / Responsive in 2D Mode)
	hudW := drawStatsHUD(state, cfg, screenW, screenH)

	// 7. Temporary Notification Banner (Responsive placement to prevent overlapping)
	if cfg.NotificationText != "" && cfg.NotificationTimer > 0 {
		drawNotificationToast(state, cfg, screenW, screenH, hudW, state.SelectedBodyID != -1)
	}

	// 8. Gravitational Wave / LIGO Observatory HUD Card
	if cfg.ShowGravitationalWaves {
		if drawLIGOObservatoryHUD(state, cfg, screenW, screenH, mousePos) {
			mouseCaptured = true
		}
	}

	return mouseCaptured
}

// IsMouseOverUI returns true if the mouse cursor is over any active UI element or panel
func IsMouseOverUI(state *SimState, cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	// 1. Top bar
	if mousePos.Y <= 48 {
		return true
	}

	// 2. Presets dropdown
	if state.ShowMorePresetsDropdown && mousePos.Y <= 300 {
		return true
	}

	// 3. Modals
	if cfg.ShowHelp || state.ShowScenarioModal {
		return true
	}

	// 4. Bottom bar
	if mousePos.Y >= float32(screenH)-48 {
		return true
	}

	// 5. Spawner hint badge
	if state.IsSpawning && mousePos.X <= 450 && mousePos.Y >= float32(screenH)-85 {
		return true
	}

	// 6. Camera control panel
	if state.ShowCameraPanel {
		panelRec := rl.NewRectangle(14, 104, 300, 480)
		if rl.CheckCollisionPointRec(mousePos, panelRec) {
			return true
		}
	}

	// 7. Inspector panel
	if state.SelectedBodyID != -1 || state.SelectedLagrangeIndex > 0 {
		panelW := float32(310)
		panelX := float32(screenW) - panelW - 12
		panelRec := rl.NewRectangle(panelX, 52, panelW, 680)
		if rl.CheckCollisionPointRec(mousePos, panelRec) {
			return true
		}
	}

	// 8. 2D Tactical Viewport (Docked on right side or fullscreen)
	if cfg.Show2DViewport {
		vx, vy, vw, vh := get2DViewportRect(state, cfg, screenW, screenH)
		vpRec := rl.NewRectangle(float32(vx), float32(vy), float32(vw), float32(vh))
		if rl.CheckCollisionPointRec(mousePos, vpRec) {
			return true
		}
	}

	// 9. LIGO Observatory HUD
	if cfg.ShowGravitationalWaves {
		ligoRec := rl.NewRectangle(14, float32(screenH)-48-148, 274, 142)
		if rl.CheckCollisionPointRec(mousePos, ligoRec) {
			return true
		}
	}

	// 10. Stats HUD (Top Left or floating in 2D)
	if cfg.ShowStatsHUD {
		badgeX := float32(12)
		badgeY := float32(54)
		if cfg.Offload3D || cfg.Viewport2DFullscreen {
			vx, vy, _, _ := get2DViewportRect(state, cfg, screenW, screenH)
			badgeX = float32(vx + 12)
			badgeY = float32(vy) + 32 + 28 + 8
		}
		hudRec := rl.NewRectangle(badgeX, badgeY, 450, 40)
		if rl.CheckCollisionPointRec(mousePos, hudRec) {
			return true
		}
	}

	return false
}

// DrawTooltip draws a floating dark pill tooltip with high-contrast text and border
func DrawTooltip(text string, x, y int32) {
	tW := MeasureTextUI(text, FontSizeSmall)
	tH := float32(20)
	rec := rl.NewRectangle(float32(x), float32(y), tW+14, tH)
	rl.DrawRectangleRec(rec, rl.NewColor(12, 18, 30, 245))
	rl.DrawRectangleLinesEx(rec, 1.0, rl.NewColor(60, 140, 220, 220))
	DrawTextUI(text, x+7, y+4, FontSizeSmall, rl.RayWhite)
}

func drawButton(rec rl.Rectangle, text string, active, hovered bool, baseCol rl.Color) bool {
	clicked := false
	col := baseCol
	isDown := hovered && rl.IsMouseButtonDown(rl.MouseLeftButton)

	if isDown {
		col = rl.NewColor(55, 125, 220, 245)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			clicked = true
		}
	} else if hovered {
		col = rl.ColorAlpha(col, 0.92)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			clicked = true
		}
	} else {
		col = rl.ColorAlpha(col, 0.70)
	}

	if active && !isDown {
		col = rl.NewColor(50, 120, 220, 225)
	}

	rl.DrawRectangleRec(rec, col)
	borderCol := rl.NewColor(120, 160, 220, 200)
	if isDown {
		borderCol = rl.White
	} else if active {
		borderCol = rl.Gold
	}
	rl.DrawRectangleLinesEx(rec, 1.5, borderCol)

	fontSize := FontSizeButton
	if rec.Height < 24 {
		fontSize = FontSizeSmall
	} else if rec.Height < 28 {
		fontSize = FontSizeRegular
	}

	textLen := MeasureTextBoldUI(text, fontSize)
	textX := int32(rec.X + (rec.Width-textLen)/2)
	textY := int32(rec.Y + (rec.Height-fontSize)/2)
	if isDown {
		textY += 1
	}

	textColor := rl.White
	if active {
		textColor = rl.Yellow
	} else if isDown {
		textColor = rl.Gold
	}
	DrawTextBoldUI(text, textX, textY, fontSize, textColor)

	return clicked
}

func drawTopBar(state *SimState, cfg *Config, camera *OrbitCamera, screenW int32, mousePos rl.Vector2) bool {
	barHeight := float32(48)
	topBarRec := rl.NewRectangle(0, 0, float32(screenW), barHeight)
	rl.DrawRectangleRec(topBarRec, rl.NewColor(18, 22, 32, 240))
	rl.DrawRectangleLinesEx(topBarRec, 1, rl.NewColor(45, 55, 80, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, topBarRec)

	// Title
	DrawTextBoldUI("GRAVITY SIM 3D", 14, 15, FontSizeTitle, rl.Gold)

	x := float32(165)
	btnW := float32(64)
	btnH := float32(30)
	btnY := float32(9)

	// Play / Pause
	playText := "PAUSE"
	if cfg.Paused {
		playText = "PLAY"
	}
	playRec := rl.NewRectangle(x, btnY, btnW, btnH)
	if drawButton(playRec, playText, !cfg.Paused, rl.CheckCollisionPointRec(mousePos, playRec), rl.NewColor(30, 70, 40, 255)) {
		cfg.Paused = !cfg.Paused
		if cfg.Paused {
			cfg.NotificationText = "Simulation Paused"
		} else {
			cfg.NotificationText = "Simulation Resumed"
		}
		cfg.NotificationTimer = 1.2
	}
	x += btnW + 8

	// Speed controls with constant incremental stepping (0.1x, 0.5x, 2.0x)
	if cfg.TimeScaleStep <= 0 {
		cfg.TimeScaleStep = 0.5
	}
	step := cfg.TimeScaleStep

	speedText := fmt.Sprintf("%.1fx", cfg.TimeScale)
	slowRec := rl.NewRectangle(x, btnY, 26, btnH)
	slowHover := rl.CheckCollisionPointRec(mousePos, slowRec)
	if drawButton(slowRec, "<", false, slowHover, rl.NewColor(40, 45, 65, 255)) {
		newSpeed := math.Round((cfg.TimeScale-step)*10.0) / 10.0
		cfg.TimeScale = math.Max(0.1, newSpeed)
		cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (-%.1fx)", cfg.TimeScale, step)
		cfg.NotificationTimer = 1.2
	} else if slowHover && rl.IsMouseButtonPressed(rl.MouseRightButton) {
		cfg.TimeScale = math.Max(0.1, cfg.TimeScale*0.5)
		cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (Halved)", cfg.TimeScale)
		cfg.NotificationTimer = 1.2
	}
	x += 28

	spdRec := rl.NewRectangle(x, btnY, 52, btnH)
	spdHover := rl.CheckCollisionPointRec(mousePos, spdRec)
	spdBgCol := rl.NewColor(25, 30, 45, 255)
	if spdHover {
		spdBgCol = rl.NewColor(35, 45, 70, 255)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			cfg.TimeScale = 1.0
			cfg.NotificationText = "Speed reset to 1.0x (Normal)"
			cfg.NotificationTimer = 1.2
		}
	}
	rl.DrawRectangleRec(spdRec, spdBgCol)
	borderCol := rl.NewColor(60, 70, 95, 255)
	if spdHover {
		borderCol = rl.SkyBlue
	}
	rl.DrawRectangleLinesEx(spdRec, 1, borderCol)
	spdTextW := MeasureTextUI(speedText, FontSizeRegular)
	DrawTextUI(speedText, int32(x+(52-spdTextW)/2), int32(btnY+8), FontSizeRegular, rl.SkyBlue)
	x += 54

	fastRec := rl.NewRectangle(x, btnY, 26, btnH)
	fastHover := rl.CheckCollisionPointRec(mousePos, fastRec)
	if drawButton(fastRec, ">", false, fastHover, rl.NewColor(40, 45, 65, 255)) {
		newSpeed := math.Round((cfg.TimeScale+step)*10.0) / 10.0
		cfg.TimeScale = math.Min(10.0, newSpeed)
		cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (+%.1fx)", cfg.TimeScale, step)
		cfg.NotificationTimer = 1.2
	} else if fastHover && rl.IsMouseButtonPressed(rl.MouseRightButton) {
		cfg.TimeScale = math.Min(10.0, cfg.TimeScale*2.0)
		cfg.NotificationText = fmt.Sprintf("Speed: %.1fx (Doubled)", cfg.TimeScale)
		cfg.NotificationTimer = 1.2
	}
	x += 28

	// Step size selector button: cycles between 0.1x, 0.5x, 2.0x
	stepLabel := fmt.Sprintf("±%.1f", step)
	stepRec := rl.NewRectangle(x, btnY, 44, btnH)
	stepHover := rl.CheckCollisionPointRec(mousePos, stepRec)
	if drawButton(stepRec, stepLabel, false, stepHover, rl.NewColor(45, 55, 75, 255)) {
		if cfg.TimeScaleStep <= 0.15 {
			cfg.TimeScaleStep = 0.5
		} else if cfg.TimeScaleStep <= 0.6 {
			cfg.TimeScaleStep = 2.0
		} else {
			cfg.TimeScaleStep = 0.1
		}
		cfg.NotificationText = fmt.Sprintf("Speed Step: ±%.1fx", cfg.TimeScaleStep)
		cfg.NotificationTimer = 1.2
	}
	x += 48

	// Presets separator
	rl.DrawLineEx(rl.NewVector2(x, 10), rl.NewVector2(x, 38), 1, rl.NewColor(60, 70, 100, 255))
	x += 10

	// Action buttons aligned to right with distinct spacing
	helpW := float32(56)
	resetW := float32(58)
	archW := float32(72)
	saveW := float32(56)
	btnGap := float32(6)

	rx := float32(screenW) - 12 - helpW
	helpRec := rl.NewRectangle(rx, btnY, helpW, btnH)
	helpHover := rl.CheckCollisionPointRec(mousePos, helpRec)
	if drawButton(helpRec, "Help", cfg.ShowHelp, helpHover, rl.NewColor(50, 60, 85, 255)) {
		cfg.ShowHelp = !cfg.ShowHelp
	}
	if helpHover {
		DrawTooltip("Manual & Keyboard Shortcuts (H)", int32(rx-130), int32(btnY+btnH+6))
	}

	rx -= resetW + btnGap
	resetRec := rl.NewRectangle(rx, btnY, resetW, btnH)
	if drawButton(resetRec, "Reset", false, rl.CheckCollisionPointRec(mousePos, resetRec), rl.NewColor(75, 40, 35, 255)) {
		LoadPreset(state, cfg, state.CurrentPreset)
		cfg.NotificationText = fmt.Sprintf("Reset scenario '%s'", state.CurrentPreset)
		cfg.NotificationTimer = 2.0
	}

	rx -= archW + btnGap
	loadRec := rl.NewRectangle(rx, btnY, archW, btnH)
	if drawButton(loadRec, "Archive", state.ShowScenarioModal, rl.CheckCollisionPointRec(mousePos, loadRec), rl.NewColor(45, 60, 95, 255)) {
		state.AvailableScenarios = ScanScenariosDirectory()
		state.ShowScenarioModal = !state.ShowScenarioModal
	}

	rx -= saveW + btnGap
	saveRec := rl.NewRectangle(rx, btnY, saveW, btnH)
	if drawButton(saveRec, "Save", false, rl.CheckCollisionPointRec(mousePos, saveRec), rl.NewColor(30, 80, 80, 255)) {
		savedPath, err := SaveScenarioToFile(state, cfg, "")
		if err == nil {
			cfg.NotificationText = fmt.Sprintf("Scenario saved to %s", filepath.Base(savedPath))
			cfg.NotificationTimer = 3.5
		}
	}

	camW := float32(72)
	rx -= camW + btnGap
	camRec := rl.NewRectangle(rx, btnY, camW, btnH)
	if drawButton(camRec, "Camera", state.ShowCameraPanel, rl.CheckCollisionPointRec(mousePos, camRec), rl.NewColor(35, 60, 95, 255)) {
		state.ShowCameraPanel = !state.ShowCameraPanel
		if state.ShowCameraPanel {
			cfg.NotificationText = "Camera panel: OPEN (F4)"
		} else {
			cfg.NotificationText = "Camera panel: CLOSED"
		}
		cfg.NotificationTimer = 1.2
	}

	offW := float32(92)
	rx -= offW + btnGap
	offRec := rl.NewRectangle(rx, btnY, offW, btnH)
	offText := "2D Offload"
	offCol := rl.NewColor(30, 75, 85, 255)
	if cfg.Offload3D {
		offText = "3D OFF (2D)"
		offCol = rl.NewColor(0, 140, 140, 255)
	}
	if drawButton(offRec, offText, cfg.Offload3D, rl.CheckCollisionPointRec(mousePos, offRec), offCol) {
		cfg.Offload3D = !cfg.Offload3D
		if cfg.Offload3D {
			cfg.Show2DViewport = true
			cfg.NotificationText = "3D Render Offloaded: 2D Main Viewport Active"
		} else {
			cfg.NotificationText = "3D Render Restored: Standard Mode"
		}
		cfg.NotificationTimer = 2.0
	}

	rightEdgeLimit := rx - 12

	presets := []struct {
		name string
		val  PresetType
	}{
		{"Solar", PresetSolarSystem},
		{"Real Scale", PresetRealSolarSystem},
		{"MW 10K", PresetMilkyWay10K},
		{"Belt 2K", PresetAsteroidBelt2K},
		{"Collide 5K", PresetGalaxyCollision5K},
		{"BH Swarm 3K", PresetBlackHoleSwarm3K},
		{"Collapse", PresetGravitationalCollapse},
		{"Grand", PresetSolarSystemGrand},
		{"Binary", PresetBinaryStars},
		{"3-Body", PresetThreeBody},
		{"Trojans", PresetLagrangeTrojans},
		{"Pulsar", PresetPulsarAccretion},
		{"Cluster", PresetGlobularCluster},
		{"Trappist", PresetTrappistResonance},
		{"Rosette", PresetRelativityRosette},
		{"Roche", PresetRocheDisruption},
		{"Empty", PresetEmpty},
	}

	moreBtnW := float32(70)
	var overflowPresets []struct {
		name string
		val  PresetType
	}

	for i, p := range presets {
		pW := MeasureTextBoldUI(p.name, FontSizeButton) + 14
		reserve := float32(0)
		if i < len(presets)-1 {
			reserve = moreBtnW + 8
		}
		if x+pW > rightEdgeLimit-reserve {
			overflowPresets = presets[i:]
			break
		}
		pRec := rl.NewRectangle(x, btnY, pW, btnH)
		isActive := (state.CurrentPreset == p.val)
		if drawButton(pRec, p.name, isActive, rl.CheckCollisionPointRec(mousePos, pRec), rl.NewColor(35, 45, 70, 255)) {
			LoadPreset(state, cfg, p.val)
			state.ShowMorePresetsDropdown = false
			cfg.NotificationText = fmt.Sprintf("Loaded scenario: %s", p.name)
			cfg.NotificationTimer = 2.0
		}
		x += pW + 4
	}

	if len(overflowPresets) > 0 {
		moreRec := rl.NewRectangle(x, btnY, moreBtnW, btnH)
		if drawButton(moreRec, "More v", state.ShowMorePresetsDropdown, rl.CheckCollisionPointRec(mousePos, moreRec), rl.NewColor(45, 55, 80, 255)) {
			state.ShowMorePresetsDropdown = !state.ShowMorePresetsDropdown
		}

		if state.ShowMorePresetsDropdown {
			dropW := float32(140)
			dropH := float32(len(overflowPresets)*28 + 8)
			dropX := x
			dropY := barHeight + 2
			dropRec := rl.NewRectangle(dropX, dropY, dropW, dropH)
			rl.DrawRectangleRec(dropRec, rl.NewColor(20, 25, 38, 250))
			rl.DrawRectangleLinesEx(dropRec, 1.5, rl.NewColor(60, 80, 120, 255))

			if rl.CheckCollisionPointRec(mousePos, dropRec) {
				mouseInside = true
			}

			dy := dropY + 4
			for _, op := range overflowPresets {
				itemRec := rl.NewRectangle(dropX+4, dy, dropW-8, 24)
				isItemHover := rl.CheckCollisionPointRec(mousePos, itemRec)
				isItemActive := (state.CurrentPreset == op.val)
				if drawButton(itemRec, op.name, isItemActive, isItemHover, rl.NewColor(30, 38, 55, 255)) {
					LoadPreset(state, cfg, op.val)
					state.ShowMorePresetsDropdown = false
					cfg.NotificationText = fmt.Sprintf("Loaded scenario: %s", op.name)
					cfg.NotificationTimer = 2.0
				}
				dy += 28
			}
		}
	} else {
		state.ShowMorePresetsDropdown = false
	}

	return mouseInside
}

func drawCameraPanel(state *SimState, cfg *Config, camera *OrbitCamera, screenW, screenH int32, mousePos rl.Vector2) bool {
	panelW := float32(300)
	panelH := float32(470)
	panelX := float32(14)
	panelY := float32(104)

	if panelY+panelH > float32(screenH)-54 {
		panelH = float32(screenH) - 54 - panelY
	}

	panelRec := rl.NewRectangle(panelX, panelY, panelW, panelH)
	rl.DrawRectangleRec(panelRec, rl.NewColor(20, 26, 40, 248))
	rl.DrawRectangleLinesEx(panelRec, 1.5, rl.NewColor(70, 110, 175, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, panelRec)

	// Panel Title
	DrawTextBoldUI("CAMERA CONTROLS", int32(panelX+14), int32(panelY+12), FontSizeHeader, rl.Gold)

	// Close button [X]
	closeRec := rl.NewRectangle(panelX+panelW-28, panelY+8, 20, 20)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		state.ShowCameraPanel = false
		return true
	}

	y := panelY + 34

	// Telemetry / Status Indicator
	modeText := "Free Orbit"
	modeCol := rl.Lime
	var selBody *Body
	if state.SelectedBodyID != -1 {
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				selBody = b
				break
			}
		}
	}

	if state.FollowSelected && selBody != nil {
		modeText = "Locked: " + selBody.Name
		modeCol = rl.Yellow
	} else if selBody != nil {
		modeText = "Selected: " + selBody.Name
		modeCol = rl.Gold
	} else if state.FollowBarycenter {
		modeText = "Tracking Barycenter"
		modeCol = rl.SkyBlue
	} else if cfg.CinematicCamera {
		modeText = "Cinematic Tour"
		modeCol = rl.Magenta
	}

	viewName := "Custom / Free"
	switch camera.CurrentView {
	case "top":
		viewName = "Top-Down (90°)"
	case "side":
		viewName = "Edge-On (Side)"
	case "front":
		viewName = "Frontal"
	case "iso":
		viewName = "Isometric (3/4)"
	}

	DrawTextUI("Target: "+modeText, int32(panelX+14), int32(y), FontSizeRegular, modeCol)
	y += 18
	camInfo := fmt.Sprintf("View: %s | Dist: %.0f | FOV: %.0f°", viewName, camera.Distance, camera.Camera.Fovy)
	DrawTextUI(camInfo, int32(panelX+14), int32(y), FontSizeTelemetry, rl.LightGray)
	y += 22

	// Section 1: Targeting & Lock
	DrawTextBoldUI("TARGETING & LOCK:", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
	y += 18

	halfBtnW := (panelW - 34) / 2

	// Row 1: Reset View & Lock Nearest
	resetRec := rl.NewRectangle(panelX+14, y, halfBtnW, 24)
	if drawButton(resetRec, "Reset View (R)", false, rl.CheckCollisionPointRec(mousePos, resetRec), rl.NewColor(75, 40, 35, 255)) {
		camera.Reset(state, cfg)
	}

	lockNearRec := rl.NewRectangle(panelX+20+halfBtnW, y, halfBtnW, 24)
	if drawButton(lockNearRec, "Lock Near (N)", false, rl.CheckCollisionPointRec(mousePos, lockNearRec), rl.NewColor(35, 80, 115, 255)) {
		camera.LockToNearest(state, cfg)
		cfg.CinematicCamera = false
	}
	y += 28

	// Row 2: Frame All & Lock Heaviest
	frameRec := rl.NewRectangle(panelX+14, y, halfBtnW, 24)
	if drawButton(frameRec, "Frame All", false, rl.CheckCollisionPointRec(mousePos, frameRec), rl.NewColor(40, 65, 95, 255)) {
		camera.FrameAll(state, cfg)
		cfg.CinematicCamera = false
	}

	lockHeavyRec := rl.NewRectangle(panelX+20+halfBtnW, y, halfBtnW, 24)
	if drawButton(lockHeavyRec, "Lock Heaviest", false, rl.CheckCollisionPointRec(mousePos, lockHeavyRec), rl.NewColor(75, 45, 95, 255)) {
		camera.LockToHeaviest(state, cfg)
		cfg.CinematicCamera = false
	}
	y += 32

	// Section 2: Tracking Modes
	DrawTextBoldUI("TRACKING MODES:", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
	y += 18

	fullBtnW := panelW - 28

	trackSelText := "Track Body: Click to Lock Nearest"
	if state.FollowSelected && selBody != nil {
		trackSelText = fmt.Sprintf("Track '%s': ON (C)", selBody.Name)
	} else if selBody != nil {
		trackSelText = fmt.Sprintf("Track '%s': OFF (C)", selBody.Name)
	}

	trackSelRec := rl.NewRectangle(panelX+14, y, fullBtnW, 24)
	if drawButton(trackSelRec, trackSelText, state.FollowSelected, rl.CheckCollisionPointRec(mousePos, trackSelRec), rl.NewColor(45, 60, 90, 255)) {
		if state.SelectedBodyID != -1 && selBody != nil {
			state.FollowSelected = !state.FollowSelected
			if state.FollowSelected {
				state.FollowBarycenter = false
				cfg.CinematicCamera = false
				cfg.NotificationText = fmt.Sprintf("Camera tracking %s: ON", selBody.Name)
			} else {
				cfg.NotificationText = fmt.Sprintf("Camera tracking %s: OFF", selBody.Name)
			}
			cfg.NotificationTimer = 2.0
		} else {
			nearest := camera.LockToNearest(state, cfg)
			if nearest != nil {
				cfg.CinematicCamera = false
			}
		}
	}
	y += 28

	trackBaryText := "Track Barycenter: OFF"
	if state.FollowBarycenter {
		trackBaryText = "Track Barycenter: ON"
	}
	trackBaryRec := rl.NewRectangle(panelX+14, y, fullBtnW, 24)
	if drawButton(trackBaryRec, trackBaryText, state.FollowBarycenter, rl.CheckCollisionPointRec(mousePos, trackBaryRec), rl.NewColor(40, 75, 85, 255)) {
		if state.FollowBarycenter {
			state.FollowBarycenter = false
			cfg.NotificationText = "Barycenter tracking: OFF"
			cfg.NotificationTimer = 2.0
		} else {
			camera.LockToBarycenter(state, cfg)
			cfg.CinematicCamera = false
		}
	}
	y += 28

	cineText := "Cinematic Tour: OFF (V)"
	if cfg.CinematicCamera {
		cineText = "Cinematic Tour: ACTIVE"
	}
	cineRec := rl.NewRectangle(panelX+14, y, fullBtnW, 24)
	if drawButton(cineRec, cineText, cfg.CinematicCamera, rl.CheckCollisionPointRec(mousePos, cineRec), rl.NewColor(75, 50, 100, 255)) {
		cfg.CinematicCamera = !cfg.CinematicCamera
		if cfg.CinematicCamera {
			state.FollowBarycenter = false
			cfg.NotificationText = "Cinematic Auto-Tour: ON"
		} else {
			cfg.NotificationText = "Cinematic Auto-Tour: OFF"
		}
		cfg.NotificationTimer = 2.0
	}
	y += 32

	// Section 3: Preset Vantage Angles
	DrawTextBoldUI("VIEW PRESETS:", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
	y += 18

	col4W := (panelW - 28 - 18) / 4

	isTop := (camera.CurrentView == "top")
	topRec := rl.NewRectangle(panelX+14, y, col4W, 22)
	if drawButton(topRec, "Top", isTop, rl.CheckCollisionPointRec(mousePos, topRec), rl.NewColor(35, 50, 75, 255)) {
		camera.SetViewPreset("top")
		cfg.CinematicCamera = false
		cfg.NotificationText = "View: Top-Down (Orbital Plane)"
		cfg.NotificationTimer = 2.0
	}

	isSide := (camera.CurrentView == "side")
	sideRec := rl.NewRectangle(panelX+14+col4W+6, y, col4W, 22)
	if drawButton(sideRec, "Side", isSide, rl.CheckCollisionPointRec(mousePos, sideRec), rl.NewColor(35, 50, 75, 255)) {
		camera.SetViewPreset("side")
		cfg.CinematicCamera = false
		cfg.NotificationText = "View: Edge-On (Side Plane)"
		cfg.NotificationTimer = 2.0
	}

	isFront := (camera.CurrentView == "front")
	frontRec := rl.NewRectangle(panelX+14+(col4W+6)*2, y, col4W, 22)
	if drawButton(frontRec, "Front", isFront, rl.CheckCollisionPointRec(mousePos, frontRec), rl.NewColor(35, 50, 75, 255)) {
		camera.SetViewPreset("front")
		cfg.CinematicCamera = false
		cfg.NotificationText = "View: Frontal Perspective"
		cfg.NotificationTimer = 2.0
	}

	isIso := (camera.CurrentView == "iso")
	isoRec := rl.NewRectangle(panelX+14+(col4W+6)*3, y, col4W, 22)
	if drawButton(isoRec, "Iso", isIso, rl.CheckCollisionPointRec(mousePos, isoRec), rl.NewColor(35, 50, 75, 255)) {
		camera.SetViewPreset("iso")
		cfg.CinematicCamera = false
		cfg.NotificationText = "View: Isometric (3/4 Perspective)"
		cfg.NotificationTimer = 2.0
	}
	y += 28

	// Section 4: Zoom & Field of View
	DrawTextBoldUI("ZOOM & FIELD OF VIEW:", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
	y += 18

	isZ30 := math.Abs(float64(camera.Distance-30.0)) < 6.0
	zRec1 := rl.NewRectangle(panelX+14, y, col4W, 22)
	if drawButton(zRec1, "30", isZ30, rl.CheckCollisionPointRec(mousePos, zRec1), rl.NewColor(38, 48, 68, 255)) {
		camera.SetDistance(30.0)
		cfg.NotificationText = "Camera Distance: 30 units (Close)"
		cfg.NotificationTimer = 2.0
	}

	isZ120 := math.Abs(float64(camera.Distance-120.0)) < 15.0
	zRec2 := rl.NewRectangle(panelX+14+col4W+6, y, col4W, 22)
	if drawButton(zRec2, "120", isZ120, rl.CheckCollisionPointRec(mousePos, zRec2), rl.NewColor(38, 48, 68, 255)) {
		camera.SetDistance(120.0)
		cfg.NotificationText = "Camera Distance: 120 units (Standard)"
		cfg.NotificationTimer = 2.0
	}

	isZ350 := math.Abs(float64(camera.Distance-350.0)) < 35.0
	zRec3 := rl.NewRectangle(panelX+14+(col4W+6)*2, y, col4W, 22)
	if drawButton(zRec3, "350", isZ350, rl.CheckCollisionPointRec(mousePos, zRec3), rl.NewColor(38, 48, 68, 255)) {
		camera.SetDistance(350.0)
		cfg.NotificationText = "Camera Distance: 350 units (Deep Space)"
		cfg.NotificationTimer = 2.0
	}

	isZ1000 := math.Abs(float64(camera.Distance-1000.0)) < 85.0
	zRec4 := rl.NewRectangle(panelX+14+(col4W+6)*3, y, col4W, 22)
	if drawButton(zRec4, "1000", isZ1000, rl.CheckCollisionPointRec(mousePos, zRec4), rl.NewColor(38, 48, 68, 255)) {
		camera.SetDistance(1000.0)
		cfg.NotificationText = "Camera Distance: 1,000 units (Cosmic)"
		cfg.NotificationTimer = 2.0
	}
	y += 26

	isFov30 := math.Abs(float64(camera.Camera.Fovy-30.0)) < 0.5
	fovRec1 := rl.NewRectangle(panelX+14, y, col4W, 22)
	if drawButton(fovRec1, "30°", isFov30, rl.CheckCollisionPointRec(mousePos, fovRec1), rl.NewColor(38, 48, 68, 255)) {
		camera.SetFov(30.0)
		cfg.NotificationText = "Camera FOV: 30° (Telephoto)"
		cfg.NotificationTimer = 2.0
	}

	isFov45 := math.Abs(float64(camera.Camera.Fovy-45.0)) < 0.5
	fovRec2 := rl.NewRectangle(panelX+14+col4W+6, y, col4W, 22)
	if drawButton(fovRec2, "45°", isFov45, rl.CheckCollisionPointRec(mousePos, fovRec2), rl.NewColor(38, 48, 68, 255)) {
		camera.SetFov(45.0)
		cfg.NotificationText = "Camera FOV: 45° (Standard)"
		cfg.NotificationTimer = 2.0
	}

	isFov60 := math.Abs(float64(camera.Camera.Fovy-60.0)) < 0.5
	fovRec3 := rl.NewRectangle(panelX+14+(col4W+6)*2, y, col4W, 22)
	if drawButton(fovRec3, "60°", isFov60, rl.CheckCollisionPointRec(mousePos, fovRec3), rl.NewColor(38, 48, 68, 255)) {
		camera.SetFov(60.0)
		cfg.NotificationText = "Camera FOV: 60° (Wide Angle)"
		cfg.NotificationTimer = 2.0
	}

	isFov80 := math.Abs(float64(camera.Camera.Fovy-80.0)) < 0.5
	fovRec4 := rl.NewRectangle(panelX+14+(col4W+6)*3, y, col4W, 22)
	if drawButton(fovRec4, "80°", isFov80, rl.CheckCollisionPointRec(mousePos, fovRec4), rl.NewColor(38, 48, 68, 255)) {
		camera.SetFov(80.0)
		cfg.NotificationText = "Camera FOV: 80° (Ultra-Wide Cinematic)"
		cfg.NotificationTimer = 2.0
	}
	y += 28

	// Section 5: Hotkeys Footer
	rl.DrawLineEx(rl.NewVector2(panelX+14, y), rl.NewVector2(panelX+panelW-14, y), 1, rl.NewColor(55, 75, 110, 200))
	y += 6
	DrawTextUI("Hotkeys: R (Reset) | N (Lock Near) | F4 (Cam Panel)", int32(panelX+14), int32(y), FontSizeSmall, rl.NewColor(160, 195, 240, 255))
	y += 14
	DrawTextUI("C (Track) | F (Focus) | V (Cinematic) | Right Drag: Orbit", int32(panelX+14), int32(y), FontSizeSmall, rl.NewColor(160, 195, 240, 255))

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

	panelW := float32(310)
	panelH := float32(610)
	panelX := float32(screenW) - panelW - 12
	panelY := float32(52)

	panelRec := rl.NewRectangle(panelX, panelY, panelW, panelH)
	rl.DrawRectangleRec(panelRec, rl.NewColor(20, 25, 38, 245))
	rl.DrawRectangleLinesEx(panelRec, 1.5, rl.NewColor(65, 85, 125, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, panelRec)

	// Panel Title
	DrawTextBoldUI("BODY INSPECTOR & TELEMETRY", int32(panelX+14), int32(panelY+12), FontSizeHeader, rl.Gold)

	// Close button [X]
	closeRec := rl.NewRectangle(panelX+panelW-30, panelY+8, 22, 22)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		state.SelectedBodyID = -1
		return true
	}

	y := panelY + 36

	// Name & Badge
	DrawTextBoldUI(fmt.Sprintf("Name: %s", body.Name), int32(panelX+14), int32(y), FontSizeHeader, rl.White)
	y += 20

	// Status badge (Stationary vs Dynamic)
	statusText := "Status: ORBITING (Dynamic)"
	statusCol := rl.Lime
	if body.IsStationary {
		statusText = "Status: STATIONARY (Fixed Anchor)"
		statusCol = rl.Red
	}
	DrawTextUI(statusText, int32(panelX+14), int32(y), FontSizeRegular, statusCol)
	y += 22

	// Stationary Toggle Button
	statBtnText := "Make Stationary (Lock)"
	if body.IsStationary {
		statBtnText = "Unfreeze (Allow Orbit)"
	}
	statRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	if drawButton(statRec, statBtnText, body.IsStationary, rl.CheckCollisionPointRec(mousePos, statRec), rl.NewColor(70, 40, 90, 255)) {
		body.IsStationary = !body.IsStationary
		if body.IsStationary {
			body.Velocity = rl.NewVector3(0, 0, 0)
			cfg.NotificationText = fmt.Sprintf("Stationary anchor '%s': LOCKED", body.Name)
		} else {
			cfg.NotificationText = fmt.Sprintf("Body '%s': UNFROZEN (Dynamic)", body.Name)
		}
		cfg.NotificationTimer = 1.5
	}
	y += 30

	// Mass controls
	DrawTextUI(fmt.Sprintf("Mass: %.2f", body.Mass), int32(panelX+14), int32(y), FontSizeRegular, rl.White)
	y += 18

	btnW := (panelW - 28 - 18) / 4
	mRec1 := rl.NewRectangle(panelX+14, y, btnW, 22)
	if drawButton(mRec1, "-50%", false, rl.CheckCollisionPointRec(mousePos, mRec1), rl.NewColor(40, 50, 70, 255)) {
		body.Mass = math.Max(0.001, body.Mass*0.5)
		cfg.NotificationText = fmt.Sprintf("Mass '%s': %.2f (-50%%)", body.Name, body.Mass)
		cfg.NotificationTimer = 1.2
	}
	mRec2 := rl.NewRectangle(panelX+14+btnW+6, y, btnW, 22)
	if drawButton(mRec2, "-10%", false, rl.CheckCollisionPointRec(mousePos, mRec2), rl.NewColor(40, 50, 70, 255)) {
		body.Mass = math.Max(0.001, body.Mass*0.9)
		cfg.NotificationText = fmt.Sprintf("Mass '%s': %.2f (-10%%)", body.Name, body.Mass)
		cfg.NotificationTimer = 1.2
	}
	mRec3 := rl.NewRectangle(panelX+14+(btnW+6)*2, y, btnW, 22)
	if drawButton(mRec3, "+10%", false, rl.CheckCollisionPointRec(mousePos, mRec3), rl.NewColor(40, 50, 70, 255)) {
		body.Mass *= 1.1
		cfg.NotificationText = fmt.Sprintf("Mass '%s': %.2f (+10%%)", body.Name, body.Mass)
		cfg.NotificationTimer = 1.2
	}
	mRec4 := rl.NewRectangle(panelX+14+(btnW+6)*3, y, btnW, 22)
	if drawButton(mRec4, "x2", false, rl.CheckCollisionPointRec(mousePos, mRec4), rl.NewColor(40, 50, 70, 255)) {
		body.Mass *= 2.0
		cfg.NotificationText = fmt.Sprintf("Mass '%s': %.2f (x2)", body.Name, body.Mass)
		cfg.NotificationTimer = 1.2
	}
	y += 28

	// Radius controls
	DrawTextUI(fmt.Sprintf("Radius: %.2f", body.Radius), int32(panelX+14), int32(y), FontSizeRegular, rl.White)
	rSubRec := rl.NewRectangle(panelX+panelW-14-64, y-2, 30, 22)
	if drawButton(rSubRec, "-", false, rl.CheckCollisionPointRec(mousePos, rSubRec), rl.NewColor(40, 50, 70, 255)) {
		body.Radius = float32(math.Max(0.15, float64(body.Radius-0.2)))
		cfg.NotificationText = fmt.Sprintf("Radius '%s': %.2f", body.Name, body.Radius)
		cfg.NotificationTimer = 1.2
	}
	rAddRec := rl.NewRectangle(panelX+panelW-14-30, y-2, 30, 22)
	if drawButton(rAddRec, "+", false, rl.CheckCollisionPointRec(mousePos, rAddRec), rl.NewColor(40, 50, 70, 255)) {
		body.Radius += 0.2
		cfg.NotificationText = fmt.Sprintf("Radius '%s': %.2f", body.Name, body.Radius)
		cfg.NotificationTimer = 1.2
	}
	y += 28

	// Kinematics Telemetry
	speed := rl.Vector3Length(body.Velocity)
	DrawTextUI(fmt.Sprintf("Speed: %.2f units/s", speed), int32(panelX+14), int32(y), FontSizeRegular, rl.SkyBlue)
	y += 18
	DrawTextUI(fmt.Sprintf("Pos: (%.1f, %.1f, %.1f)", body.Position.X, body.Position.Y, body.Position.Z), int32(panelX+14), int32(y), FontSizeTelemetry, rl.LightGray)
	y += 17
	DrawTextUI(fmt.Sprintf("Vel: (%.1f, %.1f, %.1f)", body.Velocity.X, body.Velocity.Y, body.Velocity.Z), int32(panelX+14), int32(y), FontSizeTelemetry, rl.LightGray)
	y += 20

	// KEPLERIAN ORBITAL MECHANICS
	primary := FindPrimaryBody(state.Bodies, body)
	if primary != nil && !body.IsStationary {
		dx := body.Position.X - primary.Position.X
		dy := body.Position.Y - primary.Position.Y
		dz := body.Position.Z - primary.Position.Z
		rDist := math.Sqrt(float64(dx*dx + dy*dy + dz*dz))

		vx := body.Velocity.X - primary.Velocity.X
		vy := body.Velocity.Y - primary.Velocity.Y
		vz := body.Velocity.Z - primary.Velocity.Z
		vRelSq := float64(vx*vx + vy*vy + vz*vz)

		mu := cfg.G * (primary.Mass + body.Mass)
		if mu > 0.001 && rDist > 0.1 {
			// Specific orbital energy epsilon = v^2/2 - mu/r = -mu/(2a)
			eps := 0.5*vRelSq - mu/rDist
			a := -mu / (2.0 * eps)

			// Specific angular momentum vector h = r x v
			hx := dy*vz - dz*vy
			hy := dz*vx - dx*vz
			hz := dx*vy - dy*vx
			hSq := float64(hx*hx + hy*hy + hz*hz)

			// Eccentricity e = sqrt(1 + 2*eps*h^2/mu^2)
			eVal := 0.0
			eTerm := 1.0 + (2.0*eps*hSq)/(mu*mu)
			if eTerm > 0 {
				eVal = math.Sqrt(eTerm)
			}

			orbitType := "Elliptic"
			if eVal < 0.05 {
				orbitType = "Circular"
			} else if eVal >= 1.0 {
				orbitType = "Hyperbolic"
			}

			DrawTextBoldUI("KEPLERIAN ORBIT (around "+primary.Name+"):", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
			y += 18
			DrawTextUI(fmt.Sprintf("Dist: %.1f | Semi-Major a: %.1f", rDist, a), int32(panelX+14), int32(y), FontSizeTelemetry, rl.White)
			y += 17
			DrawTextUI(fmt.Sprintf("Eccentricity e: %.3f (%s)", eVal, orbitType), int32(panelX+14), int32(y), FontSizeTelemetry, rl.Yellow)
			y += 17

			if a > 0 && eVal < 1.0 {
				period := 2.0 * math.Pi * math.Sqrt(math.Pow(a, 3)/mu)
				DrawTextUI(fmt.Sprintf("Period: %.1f s", period), int32(panelX+14), int32(y), FontSizeTelemetry, rl.SkyBlue)
				y += 20
			} else {
				y += 6
			}
		}
	} else {
		y += 6
	}

	// ORBITAL MANEUVERS & FORCE SECTION
	DrawTextBoldUI("MANEUVERS & FORCES:", int32(panelX+14), int32(y), FontSizeRegular, rl.Gold)
	y += 18

	// Circularize Orbit Button
	circRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	if drawButton(circRec, "Circularize Orbit (Auto-V)", false, rl.CheckCollisionPointRec(mousePos, circRec), rl.NewColor(30, 80, 110, 255)) {
		if primary != nil {
			CircularizeOrbit(body, primary, cfg.G, true)
			cfg.NotificationText = fmt.Sprintf("Circularized orbit of '%s' around '%s'", body.Name, primary.Name)
		} else {
			cfg.NotificationText = fmt.Sprintf("No primary gravitational body found for '%s'", body.Name)
		}
		cfg.NotificationTimer = 2.0
	}
	y += 30

	// Prograde & Retrograde thrust buttons
	halfW := (panelW - 34) / 2
	proRec := rl.NewRectangle(panelX+14, y, halfW, 24)
	if drawButton(proRec, "Prograde (+10%)", false, rl.CheckCollisionPointRec(mousePos, proRec), rl.NewColor(30, 90, 50, 255)) {
		BoostPrograde(body, 0.10)
		cfg.NotificationText = fmt.Sprintf("Prograde burn (+10%%) on '%s'", body.Name)
		cfg.NotificationTimer = 1.5
	}
	retroRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 24)
	if drawButton(retroRec, "Retrograde (-10%)", false, rl.CheckCollisionPointRec(mousePos, retroRec), rl.NewColor(90, 45, 35, 255)) {
		BoostPrograde(body, -0.10)
		cfg.NotificationText = fmt.Sprintf("Retrograde burn (-10%%) on '%s'", body.Name)
		cfg.NotificationTimer = 1.5
	}
	y += 30

	// Vertical Nudge (+Y / -Y)
	nudgeUpRec := rl.NewRectangle(panelX+14, y, halfW, 22)
	if drawButton(nudgeUpRec, "Inclination +Y", false, rl.CheckCollisionPointRec(mousePos, nudgeUpRec), rl.NewColor(45, 55, 75, 255)) {
		ApplyImpulse(body, rl.NewVector3(0, 1.5, 0))
		cfg.NotificationText = fmt.Sprintf("Inclination +Y impulse on '%s'", body.Name)
		cfg.NotificationTimer = 1.5
	}
	nudgeDownRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 22)
	if drawButton(nudgeDownRec, "Inclination -Y", false, rl.CheckCollisionPointRec(mousePos, nudgeDownRec), rl.NewColor(45, 55, 75, 255)) {
		ApplyImpulse(body, rl.NewVector3(0, -1.5, 0))
		cfg.NotificationText = fmt.Sprintf("Inclination -Y impulse on '%s'", body.Name)
		cfg.NotificationTimer = 1.5
	}
	y += 28

	// Stop dead
	stopRec := rl.NewRectangle(panelX+14, y, panelW-28, 22)
	if drawButton(stopRec, "Zero Velocity (Stop Dead)", false, rl.CheckCollisionPointRec(mousePos, stopRec), rl.NewColor(75, 45, 60, 255)) {
		body.Velocity = rl.NewVector3(0, 0, 0)
		cfg.NotificationText = fmt.Sprintf("Zeroed velocity of '%s' (Stopped dead)", body.Name)
		cfg.NotificationTimer = 1.5
	}
	y += 28

	// Camera Focus & Track
	focRec := rl.NewRectangle(panelX+14, y, halfW, 24)
	if drawButton(focRec, "Focus View (F)", false, rl.CheckCollisionPointRec(mousePos, focRec), rl.NewColor(40, 65, 95, 255)) {
		camera.Target = body.Position
		cfg.NotificationText = fmt.Sprintf("Camera focused on '%s'", body.Name)
		cfg.NotificationTimer = 1.5
	}
	followRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 24)
	followText := "Track (C): OFF"
	if state.FollowSelected {
		followText = "Track (C): ON"
	}
	if drawButton(followRec, followText, state.FollowSelected, rl.CheckCollisionPointRec(mousePos, followRec), rl.NewColor(50, 70, 95, 255)) {
		state.FollowSelected = !state.FollowSelected
		if state.FollowSelected {
			state.FollowBarycenter = false
			cfg.CinematicCamera = false
			cfg.NotificationText = fmt.Sprintf("Tracking '%s': ON", body.Name)
		} else {
			cfg.NotificationText = fmt.Sprintf("Tracking '%s': OFF", body.Name)
		}
		cfg.NotificationTimer = 1.5
	}
	y += 30

	// Lagrange Satellites (if secondary in CR3BP)
	prim, sec := GetLagrangePair(state)
	if prim != nil && sec != nil && sec.ID == body.ID && cfg.ShowLagrangePoints {
		rl.DrawLineEx(rl.NewVector2(panelX+14, y), rl.NewVector2(panelX+panelW-14, y), 1, rl.NewColor(55, 75, 110, 200))
		y += 6
		DrawTextBoldUI("LAGRANGE EQUILIBRIUM (CR3BP)", int32(panelX+14), int32(y), FontSizeRegular, rl.NewColor(255, 220, 100, 255))
		y += 18

		pts := ComputeLagrangePoints(prim, sec, cfg.G)
		lBtnW := (panelW - 28 - 16) / 5
		for i := 0; i < 5; i++ {
			btnRec := rl.NewRectangle(panelX+14+float32(i)*(lBtnW+4), y, lBtnW, 20)
			lName := fmt.Sprintf("L%d", i+1)
			active := state.SelectedLagrangeIndex == i+1
			col := rl.NewColor(40, 60, 90, 255)
			if drawButton(btnRec, lName, active, rl.CheckCollisionPointRec(mousePos, btnRec), col) {
				state.SelectedLagrangeIndex = i + 1
				cfg.NotificationText = fmt.Sprintf("Inspecting %s", pts[i].Name)
				cfg.NotificationTimer = 2.0
			}
		}
		y += 24
		// Quick Spawn Probe buttons
		halfLBtnW := (panelW - 34) / 2
		spL1Rec := rl.NewRectangle(panelX+14, y, halfLBtnW, 20)
		if drawButton(spL1Rec, "+ Probe L1", false, rl.CheckCollisionPointRec(mousePos, spL1Rec), rl.NewColor(35, 75, 110, 255)) {
			SpawnProbeAtLagrange(state, cfg, pts[0], fmt.Sprintf("%s L1", sec.Name))
		}
		spL2Rec := rl.NewRectangle(panelX+20+halfLBtnW, y, halfLBtnW, 20)
		if drawButton(spL2Rec, "+ Probe L2", false, rl.CheckCollisionPointRec(mousePos, spL2Rec), rl.NewColor(35, 75, 110, 255)) {
			SpawnProbeAtLagrange(state, cfg, pts[1], fmt.Sprintf("%s L2", sec.Name))
		}
		y += 26
	}

	// Delete Body
	delRec := rl.NewRectangle(panelX+14, y, panelW-28, 24)
	if drawButton(delRec, "DELETE BODY", false, rl.CheckCollisionPointRec(mousePos, delRec), rl.NewColor(120, 30, 30, 255)) {
		name := body.Name
		removeBody(state, body.ID)
		state.SelectedBodyID = -1
		state.FollowSelected = false
		cfg.NotificationText = fmt.Sprintf("Deleted body '%s'", name)
		cfg.NotificationTimer = 2.0
	}

	return mouseInside
}

// drawLagrangeInspectorPanel renders full telemetry and interactive spawn actions for the selected Lagrange equilibrium point
func drawLagrangeInspectorPanel(state *SimState, cfg *Config, camera *OrbitCamera, screenW, screenH int32, mousePos rl.Vector2) bool {
	primary, secondary := GetLagrangePair(state)
	if primary == nil || secondary == nil || state.SelectedLagrangeIndex < 1 || state.SelectedLagrangeIndex > 5 {
		state.SelectedLagrangeIndex = 0
		return false
	}

	pts := ComputeLagrangePoints(primary, secondary, cfg.G)
	idx := state.SelectedLagrangeIndex - 1
	pt := pts[idx]

	panelW := float32(310)
	panelH := float32(420)
	panelX := float32(screenW) - panelW - 12
	panelY := float32(52)

	panelRec := rl.NewRectangle(panelX, panelY, panelW, panelH)
	rl.DrawRectangleRec(panelRec, rl.NewColor(16, 22, 36, 248))
	rl.DrawRectangleLinesEx(panelRec, 1.5, rl.NewColor(55, 85, 130, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, panelRec)

	// Panel Title
	DrawTextBoldUI("LAGRANGE EQUILIBRIUM POINT", int32(panelX+14), int32(panelY+12), FontSizeHeader, rl.Gold)

	// Close button [X]
	closeRec := rl.NewRectangle(panelX+panelW-30, panelY+8, 22, 22)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		state.SelectedLagrangeIndex = 0
		return true
	}

	y := panelY + 36

	// Point Badge Name
	colors := [5]rl.Color{
		rl.NewColor(255, 220, 50, 255),  // L1 Gold
		rl.NewColor(80, 220, 255, 255),  // L2 Sky Cyan
		rl.NewColor(255, 120, 80, 255),  // L3 Coral
		rl.NewColor(120, 255, 180, 255), // L4 Emerald
		rl.NewColor(255, 180, 120, 255), // L5 Amber
	}
	col := colors[idx]
	statBadge := "UNSTABLE SADDLE"
	statCol := rl.NewColor(255, 140, 100, 255)
	if pt.Stable {
		statBadge = "STABLE LIBRA"
		statCol = rl.NewColor(100, 240, 160, 255)
	}

	// Point Name Box
	nameRec := rl.NewRectangle(panelX+14, y, panelW-28, 26)
	rl.DrawRectangleRec(nameRec, rl.NewColor(22, 32, 52, 240))
	rl.DrawRectangleLinesEx(nameRec, 1.0, col)
	DrawTextBoldUI(pt.Name, int32(panelX+20), int32(y+5), FontSizeHeader, col)
	y += 32

	// System Context
	sysText := fmt.Sprintf("System: %s <-> %s", primary.Name, secondary.Name)
	DrawTextUI(sysText, int32(panelX+14), int32(y), FontSizeTelemetry, rl.NewColor(200, 225, 255, 240))
	y += 18

	// Stability status
	DrawTextUI(fmt.Sprintf("Stability: %s", statBadge), int32(panelX+14), int32(y), FontSizeTelemetry, statCol)
	y += 22

	// Position coordinates
	DrawTextBoldUI("CARTESIAN EQUILIBRIUM COORDINATES", int32(panelX+14), int32(y), FontSizeRegular, rl.NewColor(160, 205, 250, 255))
	y += 18
	posStr := fmt.Sprintf("X: %8.2f  Y: %8.2f  Z: %8.2f", pt.Position.X, pt.Position.Y, pt.Position.Z)
	DrawTextUI(posStr, int32(panelX+14), int32(y), FontSizeTelemetry, rl.White)
	y += 18

	// Orbital Velocity
	velMag := rl.Vector3Length(pt.Velocity)
	velStr := fmt.Sprintf("Vx: %7.2f  Vy: %7.2f  Vz: %7.2f  (|V|: %.2f)", pt.Velocity.X, pt.Velocity.Y, pt.Velocity.Z, velMag)
	DrawTextUI(velStr, int32(panelX+14), int32(y), FontSizeTelemetry, rl.NewColor(180, 230, 255, 240))
	y += 22

	// Effective Potential
	DrawTextUI(fmt.Sprintf("Effective Potential: %.2f J/kg", pt.Potential), int32(panelX+14), int32(y), FontSizeTelemetry, rl.NewColor(255, 220, 140, 240))
	y += 28

	// Action 1: Spawn Probe
	spRec := rl.NewRectangle(panelX+14, y, panelW-28, 26)
	if drawButton(spRec, fmt.Sprintf("Launch Satellite / Probe at %s", pt.Name[:2]), false, rl.CheckCollisionPointRec(mousePos, spRec), rl.NewColor(35, 95, 140, 255)) {
		SpawnProbeAtLagrange(state, cfg, pt, fmt.Sprintf("%s %s", secondary.Name, pt.Name[:2]))
	}
	y += 32

	// Action 2: If L4 or L5, offer Swarm deployment
	if idx == 3 || idx == 4 {
		swarmName := "Trojan"
		if idx == 4 {
			swarmName = "Greek"
		}
		swRec := rl.NewRectangle(panelX+14, y, panelW-28, 26)
		if drawButton(swRec, fmt.Sprintf("Deploy %s Libration Swarm (12x)", swarmName), false, rl.CheckCollisionPointRec(mousePos, swRec), rl.NewColor(45, 110, 80, 255)) {
			SpawnLagrangeSwarm(state, cfg, primary, secondary, pt, 12)
		}
		y += 32
	}

	// Action 3: Focus & Track
	halfW := (panelW - 34) / 2
	focRec := rl.NewRectangle(panelX+14, y, halfW, 24)
	if drawButton(focRec, "Focus View (F)", false, rl.CheckCollisionPointRec(mousePos, focRec), rl.NewColor(40, 65, 95, 255)) {
		camera.Target = pt.Position
		cfg.NotificationText = fmt.Sprintf("Camera focused on %s", pt.Name)
		cfg.NotificationTimer = 1.5
	}
	followRec := rl.NewRectangle(panelX+20+halfW, y, halfW, 24)
	followText := "Track (C): OFF"
	if state.FollowSelected {
		followText = "Track (C): ON"
	}
	if drawButton(followRec, followText, state.FollowSelected, rl.CheckCollisionPointRec(mousePos, followRec), rl.NewColor(50, 70, 95, 255)) {
		state.FollowSelected = !state.FollowSelected
		if state.FollowSelected {
			state.FollowBarycenter = false
			cfg.CinematicCamera = false
			cfg.NotificationText = fmt.Sprintf("Tracking %s: ON", pt.Name)
		} else {
			cfg.NotificationText = fmt.Sprintf("Tracking %s: OFF", pt.Name)
		}
		cfg.NotificationTimer = 1.5
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
	barH := float32(48)
	barY := float32(screenH) - barH
	botRec := rl.NewRectangle(0, barY, float32(screenW), barH)
	rl.DrawRectangleRec(botRec, rl.NewColor(18, 22, 32, 240))
	rl.DrawRectangleLinesEx(botRec, 1, rl.NewColor(45, 55, 80, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, botRec)

	x := float32(12)
	btnY := barY + 9
	btnH := float32(30)

	// Spawner Toggle
	spawnBtnW := float32(120)
	spawnRec := rl.NewRectangle(x, btnY, spawnBtnW, btnH)
	spawnText := "SPAWN: OFF"
	if state.IsSpawning {
		spawnText = "SPAWN: ACTIVE"
	}
	if drawButton(spawnRec, spawnText, state.IsSpawning, rl.CheckCollisionPointRec(mousePos, spawnRec), rl.NewColor(30, 80, 60, 255)) {
		state.IsSpawning = !state.IsSpawning
		if state.IsSpawning {
			cfg.NotificationText = "Spawner Mode: ENABLED (Drag to place/launch)"
		} else {
			cfg.NotificationText = "Spawner Mode: DISABLED"
		}
		cfg.NotificationTimer = 1.5
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
			{"Gas Giant", SpawnGasGiant},
			{"Ice Giant", SpawnIceGiant},
			{"Sun", SpawnStar},
			{"Red Giant", SpawnRedGiant},
			{"White Dwarf", SpawnWhiteDwarf},
			{"Pulsar", SpawnNeutronStar},
			{"Black Hole", SpawnBlackHole},
			{"Dwarf", SpawnDwarfPlanet},
			{"Comet", SpawnComet},
			{"Asteroids", SpawnAsteroidRing},
			{"Swarm 50x", SpawnStarCluster50},
			{"Galaxy 150x", SpawnMiniGalaxy150},
			{"Collapse 100x", SpawnCollapseCloud100},
		}

		for _, t := range types {
			tW := MeasureTextBoldUI(t.name, FontSizeButton) + 12
			if x+tW > float32(screenW)-12 {
				break
			}
			tRec := rl.NewRectangle(x, btnY, tW, btnH)
			isActive := (state.SpawnPreset == t.val)
			if drawButton(tRec, t.name, isActive, rl.CheckCollisionPointRec(mousePos, tRec), rl.NewColor(35, 55, 75, 255)) {
				state.SpawnPreset = t.val
				cfg.NotificationText = fmt.Sprintf("Selected template: %s", t.name)
				cfg.NotificationTimer = 1.2
			}
			x += tW + 4
		}

		// Dedicated floating instruction badge above bottom bar on the left so it never overflows or collides with buttons
		hintText := "(Click & Drag on plane to place + launch)"
		hintW := MeasureTextBoldUI(hintText, FontSizeRegular) + 24
		hintH := float32(26)
		hintRec := rl.NewRectangle(12, barY-32, hintW, hintH)
		rl.DrawRectangleRec(hintRec, rl.NewColor(18, 24, 38, 230))
		rl.DrawRectangleLinesEx(hintRec, 1.2, rl.Gold)
		DrawTextBoldUI(hintText, 24, int32(barY-27), FontSizeRegular, rl.Yellow)
	} else {
		gpuBtnName := "GPU N²"
		gpuBtnCol := rl.NewColor(20, 80, 75, 255)
		if IsGPUComputeAvailable() {
			gpuBtnName = "RTX 3050 Ti"
			gpuBtnCol = rl.NewColor(20, 110, 80, 255)
		}

		// View & Physics toggles
		toggles := []struct {
			name  string
			value *bool
			col   rl.Color
		}{
			{"Spacetime", &cfg.ShowPotentialGrid, rl.NewColor(35, 75, 115, 255)},
			{"GW Waves", &cfg.ShowGravitationalWaves, rl.NewColor(85, 45, 135, 255)},
			{"3D Vectors", &cfg.ShowVectorField, rl.NewColor(35, 95, 80, 255)},
			{"Dense 3D", &cfg.DenseVectorField, rl.NewColor(30, 110, 85, 255)},
			{"2D Map", &cfg.Show2DViewport, rl.NewColor(30, 85, 120, 255)},
			{"2D Offload", &cfg.Offload3D, rl.NewColor(20, 110, 110, 255)},
			{"2D Icons", &cfg.Show2DIcons, rl.NewColor(25, 95, 120, 255)},
			{"2D Circles", &cfg.Show2DCircles, rl.NewColor(20, 100, 120, 255)},
			{"Lagrange", &cfg.ShowLagrangePoints, rl.NewColor(110, 85, 30, 255)},
			{"Heatmap", &cfg.Show2DHeatmap, rl.NewColor(100, 60, 30, 255)},
			{"3D Heatmap", &cfg.Show3DHeatmapPlane, rl.NewColor(90, 50, 35, 255)},
			{"Trails", &cfg.ShowTrails, rl.NewColor(35, 45, 65, 255)},
			{"Grid", &cfg.ShowGrid, rl.NewColor(35, 45, 65, 255)},
			{"Barnes-Hut", &cfg.UseBarnesHut, rl.NewColor(30, 75, 70, 255)},
			{gpuBtnName, &cfg.UseGPUCompute, gpuBtnCol},
			{"1PN GR", &cfg.EnableRelativity, rl.NewColor(70, 45, 80, 255)},
			{"Roche", &cfg.EnableRocheLimit, rl.NewColor(80, 50, 40, 255)},
			{"Textures", &cfg.ShowTextures, rl.NewColor(40, 60, 80, 255)},
			{"Labels", &cfg.ShowLabels, rl.NewColor(35, 45, 65, 255)},
			{"Vectors", &cfg.ShowVectors, rl.NewColor(35, 45, 65, 255)},
			{"Cinematic", &cfg.CinematicCamera, rl.NewColor(80, 60, 110, 255)},
			{"Glow", &cfg.ParticleGlowMode, rl.NewColor(30, 85, 90, 255)},
			{"Forces", &cfg.ShowForces, rl.NewColor(35, 45, 65, 255)},
			{"Stars", &cfg.ShowStarfield, rl.NewColor(35, 45, 65, 255)},
		}

		for _, tg := range toggles {
			tW := MeasureTextBoldUI(tg.name, FontSizeButton) + 14
			if x+tW > float32(screenW)-110 {
				break
			}
			tRec := rl.NewRectangle(x, btnY, tW, btnH)
			if drawButton(tRec, tg.name, *tg.value, rl.CheckCollisionPointRec(mousePos, tRec), tg.col) {
				*tg.value = !*tg.value
				if tg.name == "Dense 3D" && *tg.value {
					cfg.ShowVectorField = true
				}
				if tg.name == "GW Waves" && *tg.value {
					cfg.ShowPotentialGrid = true
				}
				status := "OFF"
				if *tg.value {
					status = "ON"
				}
				if tg.name == gpuBtnName {
					if *tg.value {
						cfg.NotificationText = fmt.Sprintf("NVIDIA GPU Acceleration: ON (%s | O(N²) SSBO)", GetGPUDeviceName())
					} else {
						cfg.NotificationText = "GPU Acceleration: OFF (Switched to CPU Physics)"
					}
					cfg.NotificationTimer = 2.5
				} else if tg.name == "2D Offload" {
					if *tg.value {
						cfg.Show2DViewport = true
						cfg.NotificationText = "3D Render Offloaded: 2D Main Viewport Active"
					} else {
						cfg.NotificationText = "3D Render Restored: Standard Mode"
					}
					cfg.NotificationTimer = 2.0
				} else if tg.name == "2D Icons" {
					if *tg.value {
						cfg.NotificationText = "2D Object Icons: ON (Celestial Pictures / Glyphs Active)"
					} else {
						cfg.NotificationText = "2D Object Icons: OFF (Clear Circles Mode)"
					}
					cfg.NotificationTimer = 2.0
				} else if tg.name == "2D Circles" {
					if *tg.value {
						cfg.NotificationText = "2D Circles & Dots: ON"
					} else {
						cfg.NotificationText = "2D Circles & Dots: OFF"
					}
					cfg.NotificationTimer = 1.8
				} else {
					cfg.NotificationText = fmt.Sprintf("%s: %s", tg.name, status)
					cfg.NotificationTimer = 1.2
				}
			}
			x += tW + 4
		}

		// Integrator toggle button (Yoshida 4th vs Verlet 2nd)
		intText := "Int: Verlet"
		if cfg.Integrator == IntegratorYoshida4 {
			intText = "Int: Yoshida4"
		}
		intW := MeasureTextBoldUI(intText, FontSizeButton) + 14
		if x+intW <= float32(screenW)-160 {
			intRec := rl.NewRectangle(x, btnY, intW, btnH)
			if drawButton(intRec, intText, cfg.Integrator == IntegratorYoshida4, rl.CheckCollisionPointRec(mousePos, intRec), rl.NewColor(60, 75, 35, 255)) {
				if cfg.Integrator == IntegratorVerlet {
					cfg.Integrator = IntegratorYoshida4
					cfg.NotificationText = "Integrator: Yoshida 4th-Order Symplectic O(dt^4)"
				} else {
					cfg.Integrator = IntegratorVerlet
					cfg.NotificationText = "Integrator: Velocity Verlet 2nd-Order Symplectic"
				}
				cfg.NotificationTimer = 2.0
			}
			x += intW + 4
		}

		// Collision mode toggle
		colModes := []string{"Merge", "Bounce", "Ghost"}
		colText := fmt.Sprintf("Col: %s", colModes[cfg.Collision])
		colW := MeasureTextBoldUI(colText, FontSizeButton) + 14
		if x+colW <= float32(screenW)-12 {
			colRec := rl.NewRectangle(x, btnY, colW, btnH)
			if drawButton(colRec, colText, false, rl.CheckCollisionPointRec(mousePos, colRec), rl.NewColor(55, 40, 70, 255)) {
				cfg.Collision = (cfg.Collision + 1) % 3
				cfg.NotificationText = fmt.Sprintf("Collision Mode: %s", colModes[cfg.Collision])
				cfg.NotificationTimer = 1.5
			}
		}
	}

	return mouseInside
}

func drawStatsHUD(state *SimState, cfg *Config, screenW, screenH int32) float32 {
	if !cfg.ShowStatsHUD {
		return 0
	}

	fps := rl.GetFPS()
	bodyCount := len(state.Bodies)

	method := "Direct CPU O(N^2)"
	gpuActive := cfg.UseGPUCompute && IsGPUComputeAvailable()
	if gpuActive {
		gpuDev := GetGPUDeviceName()
		if strings.Contains(gpuDev, "RTX 3050 Ti") {
			gpuDev = "NVIDIA RTX 3050 Ti"
		} else if strings.Contains(gpuDev, "NVIDIA") {
			gpuDev = "NVIDIA GPU"
		}
		gpuTime := GetGPUDispatchTimeMs()
		method = fmt.Sprintf("GPU: %s (N² SSBO %.1fms)", gpuDev, gpuTime)
	} else if cfg.UseBarnesHut || bodyCount >= 350 {
		method = fmt.Sprintf("Barnes-Hut (th=%.2f)", cfg.BarnesHutTheta)
	}

	intStr := ""
	if cfg.InteractionsPerSec >= 1000000 {
		intStr = fmt.Sprintf("%.1fM int/s", float64(cfg.InteractionsPerSec)/1000000.0)
	} else if cfg.InteractionsPerSec >= 1000 {
		intStr = fmt.Sprintf("%.1fK int/s", float64(cfg.InteractionsPerSec)/1000.0)
	} else {
		intStr = fmt.Sprintf("%d int/s", cfg.InteractionsPerSec)
	}

	line1 := fmt.Sprintf("FPS: %d / 144  |  Bodies: %d  |  Physics: %.1fms  |  Render: %.1fms  |  %s", fps, bodyCount, cfg.PhysicsTimeMs, cfg.RenderTimeMs, method)
	if cfg.EnableRelativity {
		line1 += "  |  1PN GR"
	}
	if cfg.CinematicCamera {
		line1 += "  |  CINEMATIC"
	}

	intName := "Verlet 2nd"
	if cfg.Integrator == IntegratorYoshida4 {
		intName = "Yoshida 4th O(dt^4)"
	}
	line2 := fmt.Sprintf("Throughput: %s  |  G: %.1f  |  SubSteps: %d  |  %s", intStr, cfg.G, cfg.SubSteps, intName)
	if cfg.Show2DViewport {
		line2 += "  |  2D Map [M]"
	}
	if cfg.ShowPotentialGrid {
		res := cfg.SpacetimeResolution
		if res <= 0 {
			res = 120
		}
		scale := cfg.SpacetimeScale
		if scale <= 0 {
			scale = 1.0
		}
		span := StaticBaseSpacetimeSpan * scale
		line2 += fmt.Sprintf("  |  Spacetime (Span: %.0f, %dx%d)", span, res, res)
	}
	if cfg.ShowGravitationalWaves {
		line2 += "  |  GW/LIGO [J]"
	}
	if cfg.ShowVectorField {
		if cfg.DenseVectorField {
			line2 += "  |  Dense 3D Vectors"
		} else {
			line2 += "  |  VectorField (32x32)"
		}
	}
	if cfg.ParticleGlowMode {
		line2 += "  |  Velocity Glow"
	}

	badgeX := float32(12)
	badgeY := float32(54)
	badgeH := float32(40)

	// If 2D Viewport is running as Main Viewport (Offload 3D or Fullscreen),
	// position the Stats HUD strictly below the 2D Viewport's top header and toolbar
	// so it NEVER blocks the viewport title or any toolbar buttons!
	if cfg.Offload3D || cfg.Viewport2DFullscreen {
		vx, vy, _, _ := get2DViewportRect(state, cfg, screenW, screenH)
		headerH := float32(32)
		toolbarH := float32(28)
		badgeX = float32(vx + 12)
		badgeY = float32(vy) + headerH + toolbarH + 8 // 48 + 32 + 28 + 8 = 116 (completely clear of toolbar buttons!)
	} else if cfg.Show2DViewport {
		// Docked mode on the right side: ensure badge doesn't overlap the docked 2D viewport
		vx, _, _, _ := get2DViewportRect(state, cfg, screenW, screenH)
		rightBound := float32(vx) - 8
		w1 := MeasureTextUI(line1, FontSizeTelemetry)
		w2 := MeasureTextUI(line2, FontSizeTelemetry)
		maxW := w1
		if w2 > maxW {
			maxW = w2
		}
		if badgeX+maxW+36 > rightBound && rightBound > badgeX+200 {
			// Compact strings to fit cleanly before the docked 2D viewport
			line1 = fmt.Sprintf("FPS: %d  |  Bodies: %d  |  Phys: %.1fms  |  %s", fps, bodyCount, cfg.PhysicsTimeMs, method)
			line2 = fmt.Sprintf("%s  |  G: %.1f  |  SubSteps: %d", intStr, cfg.G, cfg.SubSteps)
		}
	}

	w1 := MeasureTextUI(line1, FontSizeTelemetry)
	w2 := MeasureTextUI(line2, FontSizeTelemetry)
	maxW := w1
	if w2 > maxW {
		maxW = w2
	}
	badgeW := maxW + 40

	// Clamp badge width to screen margin
	maxAllowedW := float32(screenW) - badgeX - 12
	if badgeW > maxAllowedW {
		badgeW = maxAllowedW
	}

	badgeRec := rl.NewRectangle(badgeX, badgeY, badgeW, badgeH)
	rl.DrawRectangleRec(badgeRec, rl.NewColor(16, 22, 34, 225))
	borderCol := rl.NewColor(55, 80, 125, 220)
	if gpuActive {
		borderCol = rl.NewColor(40, 200, 130, 240)
	} else if fps >= 130 {
		borderCol = rl.NewColor(60, 190, 120, 220)
	}
	rl.DrawRectangleLinesEx(badgeRec, 1.2, borderCol)

	fpsCol := rl.NewColor(200, 225, 255, 255)
	if fps >= 135 {
		fpsCol = rl.Lime
	} else if fps < 60 {
		fpsCol = rl.Orange
	}

	DrawTextUI(line1, int32(badgeX+10), int32(badgeY+5), FontSizeTelemetry, fpsCol)
	DrawTextUI(line2, int32(badgeX+10), int32(badgeY+21), FontSizeTelemetry, rl.NewColor(160, 205, 245, 220))

	// Close / Dismiss button [×] on the top-right of the stats badge
	closeRec := rl.NewRectangle(badgeX+badgeW-20, badgeY+4, 16, 16)
	mPos := rl.GetMousePosition()
	isCloseHover := rl.CheckCollisionPointRec(mPos, closeRec)
	closeCol := rl.NewColor(130, 160, 200, 200)
	if isCloseHover {
		closeCol = rl.Red
		rl.DrawRectangleRec(closeRec, rl.NewColor(50, 25, 30, 230))
		rl.DrawRectangleLinesEx(closeRec, 1.0, rl.Red)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			cfg.ShowStatsHUD = false
			cfg.NotificationText = "Stats HUD: Hidden (Press F11 or [HUD] to restore)"
			cfg.NotificationTimer = 2.0
		}
	}
	DrawTextUI("x", int32(closeRec.X+4), int32(closeRec.Y+1), FontSizeTelemetry, closeCol)

	return badgeW
}

func drawNotificationToast(state *SimState, cfg *Config, screenW, screenH int32, hudW float32, inspectorOpen bool) {
	text := cfg.NotificationText
	textW := MeasureTextBoldUI(text, FontSizeHeader)
	pillW := textW + 36
	pillH := float32(32)

	// If 2D Viewport is running as Main Viewport (Offload 3D or Fullscreen),
	// place toast centered near bottom of viewport above status bar so it NEVER blocks top toolbar!
	if cfg.Offload3D || cfg.Viewport2DFullscreen {
		pillX := (float32(screenW) - pillW) * 0.5
		pillY := float32(screenH) - 48 - 22 - 38 // 38px above the 22px status bar and 48px bottom bar
		rect := rl.NewRectangle(pillX, pillY, pillW, pillH)
		rl.DrawRectangleRec(rect, rl.NewColor(15, 20, 32, 245))
		rl.DrawRectangleLinesEx(rect, 1.5, rl.Gold)
		DrawTextBoldUI(text, int32(pillX+18), int32(pillY+7), FontSizeHeader, rl.Yellow)
		return
	}

	// Available space calculation to prevent any overlapping in standard/docked mode
	leftLimit := float32(14)
	if hudW > 0 {
		leftLimit = 14 + hudW + 14
	}

	rightLimit := float32(screenW) - 14
	if inspectorOpen {
		rightLimit = float32(screenW) - 310 - 24
	}
	// Avoid overlapping 2D Viewport when docked on the right
	if cfg.Show2DViewport && !cfg.Viewport2DFullscreen && !cfg.Offload3D {
		vx, _, _, _ := get2DViewportRect(state, cfg, screenW, screenH)
		if float32(vx)-14 < rightLimit {
			rightLimit = float32(vx) - 14
		}
	}

	availW := rightLimit - leftLimit
	pillY := float32(54)

	var pillX float32
	if pillW <= availW {
		pillX = leftLimit + (availW-pillW)*0.5
	} else {
		pillY = 100
		if inspectorOpen && pillW <= (float32(screenW)-310-28) {
			pillX = 14 + (float32(screenW)-310-28-pillW)*0.5
		} else {
			pillX = (float32(screenW) - pillW) * 0.5
		}
	}

	if pillX < 12 {
		pillX = 12
	}

	rect := rl.NewRectangle(pillX, pillY, pillW, pillH)
	rl.DrawRectangleRec(rect, rl.NewColor(15, 20, 32, 245))
	rl.DrawRectangleLinesEx(rect, 1.5, rl.Gold)

	DrawTextBoldUI(text, int32(pillX+18), int32(pillY+7), FontSizeHeader, rl.Yellow)
}

func drawScenarioModal(state *SimState, cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	modalW := float32(540)
	modalH := float32(460)
	modalX := (float32(screenW) - modalW) / 2
	modalY := (float32(screenH) - modalH) / 2

	// Modal overlay backdrop
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 160))

	modalRec := rl.NewRectangle(modalX, modalY, modalW, modalH)
	rl.DrawRectangleRec(modalRec, rl.NewColor(22, 28, 42, 252))
	rl.DrawRectangleLinesEx(modalRec, 1.5, rl.NewColor(60, 80, 130, 255))

	mouseInside := rl.CheckCollisionPointRec(mousePos, modalRec)

	// Modal Header
	DrawTextBoldUI("SCENARIO ARCHIVE", int32(modalX+20), int32(modalY+16), FontSizeHeader, rl.Gold)

	closeRec := rl.NewRectangle(modalX+modalW-32, modalY+12, 20, 20)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		state.ShowScenarioModal = false
		return true
	}

	// List scenarios
	y := modalY + 54
	if len(state.AvailableScenarios) == 0 {
		DrawTextUI("No scenario files found in ./scenarios/", int32(modalX+24), int32(y), FontSizeRegular, rl.LightGray)
	} else {
		for _, scenPath := range state.AvailableScenarios {
			baseName := filepath.Base(scenPath)
			itemRec := rl.NewRectangle(modalX+16, y, modalW-32, 34)
			isHovered := rl.CheckCollisionPointRec(mousePos, itemRec)

			itemBg := rl.NewColor(30, 38, 55, 255)
			if isHovered {
				itemBg = rl.NewColor(45, 60, 90, 255)
			}
			rl.DrawRectangleRec(itemRec, itemBg)
			rl.DrawRectangleLinesEx(itemRec, 1, rl.NewColor(55, 75, 110, 255))

			DrawTextBoldUI(baseName, int32(modalX+26), int32(y+8), FontSizeRegular, rl.White)

			loadBtnRec := rl.NewRectangle(modalX+modalW-100, y+5, 72, 24)
			if drawButton(loadBtnRec, "Load", false, rl.CheckCollisionPointRec(mousePos, loadBtnRec), rl.NewColor(35, 70, 110, 255)) {
				err := LoadScenarioFromFile(state, cfg, scenPath)
				if err == nil {
					state.ShowScenarioModal = false
					cfg.NotificationText = fmt.Sprintf("Loaded scenario: %s", baseName)
					cfg.NotificationTimer = 2.5
				}
			}

			y += 40
		}
	}

	return mouseInside
}

func drawHelpModal(cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	modalW := float32(720)
	modalH := float32(670)
	modalX := (float32(screenW) - modalW) / 2
	modalY := (float32(screenH) - modalH) / 2

	// Backdrop
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 160))

	modalRec := rl.NewRectangle(modalX, modalY, modalW, modalH)
	rl.DrawRectangleRec(modalRec, rl.NewColor(22, 28, 42, 252))
	rl.DrawRectangleLinesEx(modalRec, 2, rl.Gold)

	mouseInside := rl.CheckCollisionPointRec(mousePos, modalRec)

	DrawTextBoldUI("GRAVITY SIM 3D - 144 FPS UNIVERSE ENGINE MANUAL", int32(modalX+24), int32(modalY+16), FontSizeTitle, rl.Gold)

	closeRec := rl.NewRectangle(modalX+modalW-36, modalY+12, 24, 24)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(80, 30, 30, 255)) {
		cfg.ShowHelp = false
		return true
	}

	lines := []string{
		"CAMERA & PERSPECTIVE:",
		"  * Camera Button or F4: Toggle Camera Control Panel (Vantage points, FOV, modes)",
		"  * R: Reset camera view to origin | N: Lock & track nearest celestial body",
		"  * Right Mouse Drag: Orbit 3D camera | Shift+Right / Middle Drag: Pan target",
		"  * Mouse Wheel: Smooth logarithmic zoom in / out",
		"  * W / A / S / D / Q / E: 6-DOF free camera navigation",
		"  * V: Toggle Cinematic Camera (smooth auto-orbit tour)",
		"  * F: Focus on selected body | C: Auto-track selected body",
		"",
		"SIMULATION SPEED & HOTKEYS:",
		"  * , / . (or < / >): Step speed constantly by active step (±0.1x, ±0.5x, ±2.0x)",
		"  * \\ (Backslash): Reset speed to normal 1.0x | Shift+, / .: Fine 0.1x speed stepping",
		"  * Top Bar: Click < or > to step, right-click to halve/double, click speed box to reset to 1.0x",
		"  * ±Step Button: Cycle step increment (0.1x -> 0.5x -> 2.0x)",
		"",
		"GIGANTIC COSMIC SCENES (1,000 - 10,000 PARTICLES):",
		"  * F5: Version 1 - Milky Way Galaxy Extreme (10,000 particles & Sagittarius A*)",
		"  * F6: Version 2 - Solar System Asteroid Belt (2,000 asteroids & Kirkwood gaps)",
		"  * F7: Version 3 - Milky Way & Andromeda Collision (5,000 colliding stars)",
		"  * F8: Version 4 - Gargantua Kerr Black Hole Swarm (3,000 relativistic bodies)",
		"  * F9: Version 5 - Gravitational Collapse (1,600 cold gas particles collapsing into sphere)",
		"  * F10: Real Scale Solar System (Astronomical AU, authentic distances & Keplerian periods)",
		"",
		"ORBITAL MANEUVERS & SPAWNER:",
		"  * Left Click: Select body | Circularize (Auto-V) | Prograde / Retrograde (+/-10%)",
		"  * Spawner: Stars, Planets, Asteroid Ring, Swarm (50x), Mini-Galaxy (150x), Collapse Cloud (100x)",
		"",
		"VISUALIZATION & PHYSICS FEATURES:",
		"  * J: Toggle Gravitational Waves & LIGO Detector (Einstein quadrupole metric ripples)",
		"  * O: Toggle 3D Vector Field (Standard -> Ultra-Dense 3,448 volumetric vectors -> Off)",
		"  * M: Toggle 2D Tactical Viewport (Minimap / Fullscreen docked on right)",
		"  * K: Toggle 2D Map Celestial Labels (Clean celestial circles vs annotated tags)",
		"  * H: Toggle Gravitational Heatmap (2D Viewport & 3D Orbital Plane Grid)",
		"  * Shift+H / Alt+H or Alt+[ / Alt+]: Step 2D Heatmap Resolution (16 up to 200 cols)",
		"  * I: Cycle Symplectic Integrator (Yoshida 4th-Order O(dt^4) vs Velocity Verlet 2nd-Order)",
		"  * U: Toggle GPU Compute Shader Acceleration (NVIDIA/OpenGL 4.3 SSBO Direct N²)",
		"  * P: Toggle 3D Spacetime Curvature Potential Grid (Einstein gravity wells)",
		"  * [ / ]: Decrease / Increase Spacetime Grid Scale (0.5x - 3.5x)",
		"  * { / } or 2D Toolbar [RES-]/[RES+]: Step Spacetime / 2D Heatmap Resolution",
		"  * Shift+T or Toolbar [TRL]: Toggle Orbit Reference Trails",
		"  * Shift+C or Toolbar [CIR]: Toggle 2D Body Circles and Center Dots",
		"  * Y or Toolbar [ICON]: Toggle 2D Celestial Object Icons / Clear Circles",
		"  * V or Toolbar [VEC]: Toggle Object Velocity & Direction Vectors",
		"  * L: Toggle Particle Kinetic Velocity Color Glow",
		"  * B: Toggle Barnes-Hut O(N log N) multi-threaded octree",
		"  * G: Toggle 1PN General Relativity precession",
		"",
		"SHORTCUTS: Space (Pause) | J (GW) | O (Vec) | M (2D Map) | K (LBL) | Y (ICON) | { / } (Res) | 1-9, 0, F5-F10",
	}

	y := int32(modalY + 48)
	for _, l := range lines {
		if len(l) > 0 && l[0] != ' ' {
			DrawTextBoldUI(l, int32(modalX+24), y, FontSizeHeader, rl.Yellow)
		} else {
			DrawTextUI(l, int32(modalX+24), y, FontSizeTelemetry, rl.LightGray)
		}
		y += 20
	}

	return mouseInside
}

// drawLIGOObservatoryHUD renders the high-tech Laser Interferometer Gravitational-Wave Observatory
// HUD card with live quadrupole strain, frequency, gravitational power, and real-time waveform oscilloscope.
func drawLIGOObservatoryHUD(state *SimState, cfg *Config, screenW, screenH int32, mousePos rl.Vector2) bool {
	cardW := float32(274)
	cardH := float32(142)
	cardX := float32(14)
	cardY := float32(screenH) - 48 - cardH - 6

	cardRec := rl.NewRectangle(cardX, cardY, cardW, cardH)
	mouseInside := rl.CheckCollisionPointRec(mousePos, cardRec)

	// Panel background & neon relativistic border
	rl.DrawRectangleRec(cardRec, rl.NewColor(14, 18, 28, 245))
	burstActive := len(state.GWBursts) > 0
	borderColor := rl.NewColor(90, 60, 150, 240)
	if burstActive {
		pulse := float32(math.Sin(float64(rl.GetTime()*8.0)))*0.5 + 0.5
		borderColor = rl.NewColor(uint8(180+pulse*75), 40, uint8(220+pulse*35), 255)
	}
	rl.DrawRectangleLinesEx(cardRec, 1.4, borderColor)

	// Header
	DrawTextBoldUI("LIGO / VIRGO OBSERVATORY", int32(cardX+10), int32(cardY+8), FontSizeHeader, rl.NewColor(80, 220, 255, 255))

	// Status indicator
	statusText := "[ONLINE] 4.0km Fabry-Perot Arms"
	statusCol := rl.NewColor(80, 240, 140, 255)
	if burstActive {
		statusText = "[EVENT] BURST DETECTED!"
		statusCol = rl.NewColor(255, 70, 180, 255)
	}
	DrawTextUI(statusText, int32(cardX+10), int32(cardY+24), FontSizeTelemetry, statusCol)

	// Dismiss button [X]
	closeRec := rl.NewRectangle(cardX+cardW-24, cardY+6, 18, 18)
	if drawButton(closeRec, "X", false, rl.CheckCollisionPointRec(mousePos, closeRec), rl.NewColor(60, 30, 45, 255)) {
		cfg.ShowGravitationalWaves = false
		cfg.NotificationText = "Gravitational Waves: OFF"
		cfg.NotificationTimer = 1.5
		return true
	}

	// Telemetry row
	strainStr := fmt.Sprintf("Strain h: %+.2e", state.LastGWStrain)
	DrawTextBoldUI(strainStr, int32(cardX+10), int32(cardY+40), FontSizeSmall, rl.NewColor(255, 225, 90, 255))

	freqStr := fmt.Sprintf("f: %.2f Hz", state.LastGWFreq)
	DrawTextUI(freqStr, int32(cardX+168), int32(cardY+40), FontSizeSmall, rl.NewColor(160, 220, 255, 255))

	powerStr := fmt.Sprintf("P_gw: %.2e W", state.LastGWPower)
	DrawTextUI(powerStr, int32(cardX+10), int32(cardY+56), FontSizeTelemetry, rl.LightGray)

	// Live Oscilloscope Waveform Box
	scopeX := cardX + 8
	scopeY := cardY + 74
	scopeW := cardW - 16
	scopeH := float32(58)
	scopeRec := rl.NewRectangle(scopeX, scopeY, scopeW, scopeH)
	rl.DrawRectangleRec(scopeRec, rl.NewColor(8, 10, 16, 255))
	rl.DrawRectangleLinesEx(scopeRec, 1.0, rl.NewColor(40, 50, 75, 200))

	// Center baseline
	midY := scopeY + scopeH*0.5
	rl.DrawLine(int32(scopeX), int32(midY), int32(scopeX+scopeW), int32(midY), rl.NewColor(30, 45, 65, 180))

	// Oscilloscope plot from circular history buffer
	nHist := len(state.GWWaveformHistory)
	if nHist > 1 {
		head := state.GWWaveformHead
		stepX := scopeW / float32(nHist-1)

		waveCol := rl.NewColor(60, 230, 255, 255)
		if burstActive {
			waveCol = rl.NewColor(255, 80, 190, 255)
		}

		for s := 0; s < nHist-1; s++ {
			idx1 := (head + s) % nHist
			idx2 := (head + s + 1) % nHist
			v1 := state.GWWaveformHistory[idx1]
			v2 := state.GWWaveformHistory[idx2]

			// Scale & clamp to scope height
			y1 := midY - v1*scopeH*2.0
			y2 := midY - v2*scopeH*2.0
			if y1 < scopeY+2 {
				y1 = scopeY + 2
			} else if y1 > scopeY+scopeH-2 {
				y1 = scopeY + scopeH - 2
			}
			if y2 < scopeY+2 {
				y2 = scopeY + 2
			} else if y2 > scopeY+scopeH-2 {
				y2 = scopeY + scopeH - 2
			}

			x1 := scopeX + float32(s)*stepX
			x2 := scopeX + float32(s+1)*stepX
			rl.DrawLine(int32(x1), int32(y1), int32(x2), int32(y2), waveCol)
		}
	}

	return mouseInside
}
