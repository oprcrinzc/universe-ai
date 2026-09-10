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
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(6, 8, 14, 255))
		Render3DScene(state, cfg, camera)
		DrawUI(state, cfg, camera, w, h)
		rl.EndDrawing()
	}

	// 1. Solar System Showcase
	fmt.Println("[DemoCapture] 1/6 Capturing Solar System...")
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
	fmt.Println("[DemoCapture] 2/6 Capturing Binary Stars...")
	LoadPreset(state, cfg, PresetBinaryStars)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 165.0
	camera.Azimuth = 0.8
	camera.Elevation = 0.55
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = true
	cfg.ShowVectors = true

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
	fmt.Println("[DemoCapture] 3/6 Capturing 3-Body Choreography...")
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
	fmt.Println("[DemoCapture] 4/6 Capturing Galaxy Accretion Disk...")
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
	// Also copy as hero image
	rl.TakeScreenshot("assets/hero.png")

	// 5. Body Inspector & Orbital Maneuvers Showcase
	fmt.Println("[DemoCapture] 5/6 Capturing Body Inspector & Thrust...")
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

	// 6. Interactive Slingshot Spawner Showcase
	fmt.Println("[DemoCapture] 6/6 Capturing Slingshot Spawner...")
	state.SelectedBodyID = -1
	state.IsSpawning = true
	state.SpawnPreset = SpawnMoon
	state.IsDraggingSpawn = true
	state.SpawnWorldPos = rl.NewVector3(45, 0, 18)
	state.DragVelocity = rl.NewVector3(-14, 0, 20)
	renderFrame()
	rl.TakeScreenshot("assets/screenshots/spawner_slingshot.png")

	// 7. Video Frames Recording (Accretion disk swirl + Solar system)
	fmt.Println("[DemoCapture] Recording 180 video frames for animated GIF & MP4...")
	state.IsSpawning = false
	state.IsDraggingSpawn = false
	LoadPreset(state, cfg, PresetGalaxyDisk)
	camera.Target = rl.NewVector3(0, 0, 0)
	camera.Distance = 175.0
	camera.Azimuth = 0.6
	camera.Elevation = 0.62
	camera.UpdatePosition()
	cfg.ShowTrails = true
	cfg.ShowLabels = false
	cfg.ShowVectors = false

	// Warmup 100 frames
	for f := 0; f < 100; f++ {
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

	// 8. FFmpeg encoding
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

	fmt.Println("[DemoCapture] Encoding optimized GIF showcase...")
	gifFilter := "fps=20,scale=720:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer"
	gifCmd := exec.Command("ffmpeg", "-y", "-framerate", "30",
		"-i", filepath.Join(tempDir, "frame_%04d.png"),
		"-vf", gifFilter,
		"assets/showcase.gif")
	if out, err := gifCmd.CombinedOutput(); err != nil {
		fmt.Printf("[DemoCapture] GIF encoding warning: %v, output: %s\n", err, string(out))
	} else {
		fmt.Println("[DemoCapture] assets/showcase.gif generated successfully!")
	}

	_ = os.RemoveAll(tempDir)
	fmt.Println("[DemoCapture] Showcase generation complete!")
}
