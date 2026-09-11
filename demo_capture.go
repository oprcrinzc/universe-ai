package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// RunDemoCapture renders high-resolution screenshots and video clips for documentation
func RunDemoCapture(state *SimState, cfg *Config, camera *OrbitCamera, w, h int32) {
	fmt.Println("[DemoCapture] Starting automatic showcase capture...")

	_ = os.MkdirAll("assets/screenshots", 0755)
	tempDir := "temp_frames"
	_ = os.RemoveAll(tempDir)
	_ = os.MkdirAll(tempDir, 0755)

	// Helper to draw a frame
	renderFrame := func() {
		currW := int32(rl.GetScreenWidth())
		currH := int32(rl.GetScreenHeight())
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(6, 8, 14, 255))
		Render3DScene(state, cfg, camera)
		Draw2DViewport(state, cfg, currW, currH)
		DrawUI(state, cfg, camera, currW, currH)
		rl.EndDrawing()
	}

	// 1. Solar System Showcase
	fmt.Println("[DemoCapture] 1/9 Capturing Solar System with Textures...")
	LoadPreset(state, cfg, PresetSolarSystem)
	camera.Target = rl.NewVector3(12, 0, 0)
	camera.Distance = 145.0
	camera.Azimuth = 0.95
	camera.Elevation = 0.52
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true
	cfg.ShowForces = false
	cfg.ShowTextures = true
	cfg.Show2DViewport = false
	state.IsSpawning = false

	for f := 0; f < 250; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	for _, b := range state.Bodies {
		if b.Name == "Earth" {
			state.SelectedBodyID = b.ID
			break
		}
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/solar_system.png")

	// 2. Binary Stars Showcase
	fmt.Println("[DemoCapture] 2/9 Capturing Binary Stars...")
	LoadPreset(state, cfg, PresetBinaryStars)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 165.0
	camera.Azimuth = 0.8
	camera.Elevation = 0.55
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true
	cfg.ShowTextures = true

	for f := 0; f < 280; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	for _, b := range state.Bodies {
		if b.Name == "Planet Tatooine I" {
			state.SelectedBodyID = b.ID
			break
		}
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/binary_stars.png")

	// 3. Three-Body Problem Showcase
	fmt.Println("[DemoCapture] 3/9 Capturing 3-Body Choreography...")
	LoadPreset(state, cfg, PresetThreeBody)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 88.0
	camera.Azimuth = 0.45
	camera.Elevation = 0.58
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true

	for f := 0; f < 320; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/three_body.png")

	// 4. Galaxy Accretion Disk Showcase
	fmt.Println("[DemoCapture] 4/9 Capturing Galaxy Accretion Disk...")
	LoadPreset(state, cfg, PresetGalaxyDisk)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 185.0
	camera.Azimuth = 0.55
	camera.Elevation = 0.68
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = false
	cfg.ShowVectors = false
	state.SelectedBodyID = -1

	for f := 0; f < 280; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/galaxy_disk.png")
	rl.TakeScreenshot("assets/hero.png")

	// 5. Barnes-Hut Milky Way Cluster (650+ Stars)
	fmt.Println("[DemoCapture] 5/9 Capturing Barnes-Hut Milky Way Cluster...")
	LoadPreset(state, cfg, PresetMilkyWayCluster)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 210.0
	camera.Azimuth = 0.72
	camera.Elevation = 0.65
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = false
	cfg.ShowVectors = false
	cfg.UseBarnesHut = true

	for f := 0; f < 180; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/barnes_hut_galaxy.png")

	// 6. Relativistic Rosette Precession (1PN General Relativity)
	fmt.Println("[DemoCapture] 6/9 Capturing Relativistic Rosette Precession...")
	LoadPreset(state, cfg, PresetRelativityRosette)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 65.0
	camera.Azimuth = 0.35
	camera.Elevation = 0.78
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true
	cfg.EnableRelativity = true

	for f := 0; f < 380; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/relativistic_rosette.png")
	cfg.EnableRelativity = false

	// 7. Roche Limit Tidal Disruption & Ring Formation
	fmt.Println("[DemoCapture] 7/9 Capturing Roche Limit Tidal Disruption...")
	LoadPreset(state, cfg, PresetRocheDisruption)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 75.0
	camera.Azimuth = 0.5
	camera.Elevation = 0.55
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = false
	cfg.EnableRocheLimit = true

	for f := 0; f < 220; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/roche_disruption.png")

	// 8. Body Inspector & Orbital Maneuvers Showcase
	fmt.Println("[DemoCapture] 8/9 Capturing Body Inspector & Thrust...")
	LoadPreset(state, cfg, PresetSolarSystem)
	camera.Target = rl.NewVector3(38, 0, 0)
	camera.Distance = 75.0
	camera.Azimuth = 0.65
	camera.Elevation = 0.42
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true
	cfg.ShowForces = true

	for f := 0; f < 160; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	for _, b := range state.Bodies {
		if b.Name == "Mars" {
			state.SelectedBodyID = b.ID
			break
		}
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/inspector_panel.png")

	// 9. Lagrange Trojans Showcase
	fmt.Println("[DemoCapture] Capturing Lagrange L4 & L5 Trojans...")
	LoadPreset(state, cfg, PresetLagrangeTrojans)
	camera.Target = rl.NewVector3(20, 0, 15)
	camera.Distance = 95.0
	camera.Azimuth = 0.8
	camera.Elevation = 0.72
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = false
	state.SelectedBodyID = -1

	for f := 0; f < 200; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/lagrange_trojans.png")

	// 10. Galaxy Collision Showcase
	fmt.Println("[DemoCapture] Capturing Galaxy Collision Tidal Tails...")
	LoadPreset(state, cfg, PresetGalaxyCollision)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 140.0
	camera.Azimuth = 0.65
	camera.Elevation = 0.60
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = false

	for f := 0; f < 240; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/galaxy_collision.png")

	// 11. Relativistic Pulsar Accretion Binary
	fmt.Println("[DemoCapture] Capturing Relativistic Pulsar Accretion Binary...")
	LoadPreset(state, cfg, PresetPulsarAccretion)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 48.0
	camera.Azimuth = 0.45
	camera.Elevation = 0.50
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true

	for f := 0; f < 180; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	for _, b := range state.Bodies {
		if b.IsPulsar {
			state.SelectedBodyID = b.ID
			break
		}
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/pulsar_accretion.png")

	// 12. Grand Solar System Showcase
	fmt.Println("[DemoCapture] Capturing Grand Solar System...")
	LoadPreset(state, cfg, PresetSolarSystemGrand)
	camera.Target = rl.NewVector3(45, 0, 0)
	camera.Distance = 170.0
	camera.Azimuth = 0.85
	camera.Elevation = 0.55
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true

	for f := 0; f < 220; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	for _, b := range state.Bodies {
		if b.Name == "Saturn" {
			state.SelectedBodyID = b.ID
			break
		}
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/solar_system_grand.png")

	// 13. Interactive Slingshot Spawner Showcase
	fmt.Println("[DemoCapture] Capturing Slingshot Spawner...")
	state.SelectedBodyID = -1
	state.IsSpawning = true
	state.SpawnPreset = SpawnNeutronStar
	state.IsDraggingSpawn = true
	state.SpawnWorldPos = rl.NewVector3(45, 0, 18)
	state.DragVelocity = rl.NewVector3(-14, 0, 20)
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/spawner_slingshot.png")

	// 14. Milky Way 10K Showcase
	fmt.Println("[DemoCapture] Capturing Milky Way Extreme 10K Particles...")
	state.IsSpawning = false
	state.IsDraggingSpawn = false
	LoadPreset(state, cfg, PresetMilkyWay10K)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 340.0
	camera.Azimuth = 0.82
	camera.Elevation = 0.58
	camera.UpdatePosition()
	cfg.ShowTrails = false
	cfg.ShowLabels = true
	cfg.ShowPotentialGrid = true
	for f := 0; f < 4; f++ {
		UpdatePhysics(state, cfg, 0.00694)
		renderFrame()
	}
	rl.TakeScreenshot("assets/screenshots/milky_way_10k.png")

	// 15. Asteroid Belt 2K Showcase
	fmt.Println("[DemoCapture] Capturing Asteroid Belt 2K Particles...")
	LoadPreset(state, cfg, PresetAsteroidBelt2K)
	camera.Target = rl.NewVector3(25, 0, 0)
	camera.Distance = 160.0
	camera.Azimuth = 0.90
	camera.Elevation = 0.52
	camera.UpdatePosition()
	cfg.ShowPotentialGrid = false
	for f := 0; f < 4; f++ {
		UpdatePhysics(state, cfg, 0.00694)
		renderFrame()
	}
	rl.TakeScreenshot("assets/screenshots/asteroid_belt_2k.png")

	// 16. Galaxy Collision 5K Showcase
	fmt.Println("[DemoCapture] Capturing Galaxy Collision 5K Particles...")
	LoadPreset(state, cfg, PresetGalaxyCollision5K)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 290.0
	camera.Azimuth = 0.75
	camera.Elevation = 0.55
	camera.UpdatePosition()
	for f := 0; f < 4; f++ {
		UpdatePhysics(state, cfg, 0.00694)
		renderFrame()
	}
	rl.TakeScreenshot("assets/screenshots/galaxy_collision_5k.png")

	// 17. Gargantua Black Hole Swarm 3K Showcase
	fmt.Println("[DemoCapture] Capturing Gargantua Black Hole Swarm 3K Particles...")
	LoadPreset(state, cfg, PresetBlackHoleSwarm3K)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 125.0
	camera.Azimuth = 0.68
	camera.Elevation = 0.48
	camera.UpdatePosition()
	cfg.ParticleGlowMode = true
	for f := 0; f < 4; f++ {
		UpdatePhysics(state, cfg, 0.00694)
		renderFrame()
	}
	rl.TakeScreenshot("assets/screenshots/blackhole_swarm_3k.png")
	cfg.ParticleGlowMode = false

	// 18. Real Scale Solar System Showcase (Astronomical AU scale, all planets, Kirkwood gaps, Halley, Voyager)
	fmt.Println("[DemoCapture] Capturing Real Scale Solar System with 2D Tactical Viewport...")
	LoadPreset(state, cfg, PresetRealSolarSystem)
	camera.Target = rl.NewVector3(120, 0, 0)
	camera.Distance = 550.0
	camera.Azimuth = 0.88
	camera.Elevation = 0.65
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = false
	cfg.ShowTextures = true
	cfg.Show2DViewport = true
	cfg.Viewport2DFullscreen = false
	cfg.Show2DHeatmap = true
	cfg.Show2DVectorField = true
	cfg.Viewport2DZoom = 0.22
	cfg.Viewport2DPan = rl.NewVector2(80, 0)
	cfg.Integrator = IntegratorYoshida4

	for f := 0; f < 100; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/real_scale_solar_system.png")

	// 19. Fullscreen 2D Viewport with Continuous Gravitational Heatmap & 2D Vector Field
	fmt.Println("[DemoCapture] Capturing Fullscreen 2D Viewport & Gravitational Heatmap...")
	cfg.Viewport2DFullscreen = true
	cfg.Show2DViewport = true
	cfg.Show2DHeatmap = true
	cfg.Show2DVectorField = true
	cfg.Viewport2DZoom = 0.40
	cfg.Viewport2DPan = rl.NewVector2(40, 0)
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/viewport_2d_heatmap.png")
	cfg.Viewport2DFullscreen = false

	// 20. Exact Analytical CR3BP Lagrange Points (L1-L5) & Trojan Libration Halos
	fmt.Println("[DemoCapture] Capturing Analytical Lagrange Points L1-L5 & Swarms...")
	LoadPreset(state, cfg, PresetLagrangeTrojans)
	camera.Target = rl.NewVector3(12, 0, 0)
	camera.Distance = 75.0
	camera.Azimuth = 0.75
	camera.Elevation = 0.70
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowLagrangePoints = true
	cfg.Show2DViewport = true
	cfg.Show2DHeatmap = true
	cfg.Show2DVectorField = true
	cfg.Viewport2DZoom = 1.3
	cfg.Viewport2DPan = rl.NewVector2(12, 0)
	cfg.Integrator = IntegratorYoshida4

	for f := 0; f < 120; f++ {
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
	}
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/lagrange_points_exact.png")

	// 21. Video Frames Recording (Real Scale Solar System + 2D Tactical Viewport with Heatmap)
	fmt.Println("[DemoCapture] Recording 180 video frames for animated GIF & MP4...")
	state.IsSpawning = false
	state.IsDraggingSpawn = false
	LoadPreset(state, cfg, PresetRealSolarSystem)
	camera.Target = rl.NewVector3(60, 0, 0)
	camera.Distance = 320.0
	camera.Azimuth = 0.75
	camera.Elevation = 0.58
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = false
	cfg.ShowTextures = true
	cfg.Show2DViewport = true
	cfg.Viewport2DFullscreen = false
	cfg.Show2DHeatmap = true
	cfg.Show2DVectorField = true
	cfg.Viewport2DZoom = 0.30
	cfg.Viewport2DPan = rl.NewVector2(45, 0)
	cfg.Integrator = IntegratorYoshida4

	// Warmup 40 frames
	for f := 0; f < 40; f++ {
		UpdatePhysics(state, cfg, 0.0166)
	}

	// Capture 180 frames with gentle camera rotation
	for f := 0; f < 180; f++ {
		camera.Azimuth += 0.005
		camera.UpdatePosition()
		UpdatePhysics(state, cfg, 0.0166)
		renderFrame()
		framePath := filepath.Join(tempDir, fmt.Sprintf("frame_%04d.png", f))
		rl.TakeScreenshot(framePath)
	}

	// 11. FFmpeg encoding
	fmt.Println("[DemoCapture] Encoding MP4 video showcase...")
	mp4Cmd := exec.Command("ffmpeg", "-y", "-framerate", "30",
		"-i", filepath.Join(tempDir, "frame_%04d.png"),
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-crf", "20",
		"assets/showcase.mp4")
	if out, err := mp4Cmd.CombinedOutput(); err != nil {
		fmt.Printf("[DemoCapture] MP4 encoding warning: %v, output: %s\n", err, string(out))
	} else {
		fmt.Println("[DemoCapture] assets/showcase.mp4 generated successfully!")
	}

	// Optimized web MP4 copy
	_ = exec.Command("cp", "assets/showcase.mp4", "assets/showcase_web.mp4").Run()

	fmt.Println("[DemoCapture] Encoding optimized GIF showcase...")
	gifFilter := "fps=20,scale=720:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer"
	gifCmd := exec.Command("ffmpeg", "-y", "-framerate", "30",
		"-i", filepath.Join(tempDir, "frame_%04d.png"),
		"-vf", gifFilter,
		"assets/showcase_preview.gif")
	if out, err := gifCmd.CombinedOutput(); err != nil {
		fmt.Printf("[DemoCapture] GIF encoding warning: %v, output: %s\n", err, string(out))
	} else {
		fmt.Println("[DemoCapture] assets/showcase_preview.gif generated successfully!")
	}

	_ = os.RemoveAll(tempDir)
	fmt.Println("[DemoCapture] Showcase generation complete!")
}
