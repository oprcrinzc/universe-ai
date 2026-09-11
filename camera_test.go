package main

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestOrbitCameraReset(t *testing.T) {
	cam := NewOrbitCamera()
	state := &SimState{
		FollowSelected:   true,
		FollowBarycenter: true,
	}
	cfg := &Config{
		CinematicCamera: true,
	}

	// Change camera properties away from default
	cam.Target = rl.NewVector3(100, 50, -200)
	cam.Distance = 500.0
	cam.Azimuth = 2.5
	cam.Elevation = -0.5
	cam.Camera.Fovy = 75.0

	cam.Reset(state, cfg)

	if cam.Target.X != 0 || cam.Target.Y != 0 || cam.Target.Z != 0 {
		t.Fatalf("Expected target (0,0,0), got %+v", cam.Target)
	}
	if cam.Distance != 120.0 {
		t.Fatalf("Expected distance 120.0, got %f", cam.Distance)
	}
	if cam.Azimuth != 0.8 || cam.Elevation != 0.45 {
		t.Fatalf("Expected azimuth 0.8 and elevation 0.45, got azim=%f, elev=%f", cam.Azimuth, cam.Elevation)
	}
	if cam.Camera.Fovy != 45.0 {
		t.Fatalf("Expected FOV 45.0, got %f", cam.Camera.Fovy)
	}
	if state.FollowSelected {
		t.Fatalf("Expected FollowSelected false")
	}
	if state.FollowBarycenter {
		t.Fatalf("Expected FollowBarycenter false")
	}
	if cfg.CinematicCamera {
		t.Fatalf("Expected CinematicCamera false after reset")
	}
}

func TestOrbitCameraLockToNearest(t *testing.T) {
	cam := NewOrbitCamera()
	cam.Camera.Position = rl.NewVector3(10, 0, 10)

	body1 := &Body{ID: 1, Name: "Far Star", Position: rl.NewVector3(100, 0, 100), Radius: 5.0, Mass: 100}
	body2 := &Body{ID: 2, Name: "Near Planet", Position: rl.NewVector3(12, 0, 11), Radius: 1.0, Mass: 1}
	body3 := &Body{ID: 3, Name: "Mid Planet", Position: rl.NewVector3(25, 0, 20), Radius: 2.0, Mass: 5}

	state := &SimState{
		Bodies: []*Body{body1, body2, body3},
	}
	cfg := &Config{}

	nearest := cam.LockToNearest(state, cfg)
	if nearest == nil {
		t.Fatalf("Expected to find nearest body, got nil")
	}
	if nearest.ID != 2 {
		t.Fatalf("Expected nearest body ID 2 (%s), got ID %d (%s)", body2.Name, nearest.ID, nearest.Name)
	}
	if state.SelectedBodyID != 2 {
		t.Fatalf("Expected state.SelectedBodyID == 2, got %d", state.SelectedBodyID)
	}
	if !state.FollowSelected {
		t.Fatalf("Expected state.FollowSelected == true")
	}
	if state.FollowBarycenter {
		t.Fatalf("Expected state.FollowBarycenter == false")
	}
	if cam.Target != body2.Position {
		t.Fatalf("Expected cam.Target == body2.Position, got %+v", cam.Target)
	}

	// Calling LockToNearest while already locked to body2 should cycle to the next nearest body (body3)
	nextNearest := cam.LockToNearest(state, cfg)
	if nextNearest == nil {
		t.Fatalf("Expected cycling to find next nearest body, got nil")
	}
	if nextNearest.ID != 3 {
		t.Fatalf("Expected next nearest body ID 3 (%s), got ID %d (%s)", body3.Name, nextNearest.ID, nextNearest.Name)
	}
	if state.SelectedBodyID != 3 {
		t.Fatalf("Expected state.SelectedBodyID == 3, got %d", state.SelectedBodyID)
	}
}

func TestOrbitCameraLockToHeaviest(t *testing.T) {
	cam := NewOrbitCamera()
	body1 := &Body{ID: 1, Name: "Small Asteroid", Position: rl.NewVector3(5, 0, 5), Radius: 0.5, Mass: 0.1}
	body2 := &Body{ID: 2, Name: "Supermassive Black Hole", Position: rl.NewVector3(0, 0, 0), Radius: 8.0, Mass: 50000.0}
	body3 := &Body{ID: 3, Name: "Medium Star", Position: rl.NewVector3(50, 0, 0), Radius: 3.0, Mass: 500.0}

	state := &SimState{
		Bodies: []*Body{body1, body2, body3},
	}
	cfg := &Config{}

	heaviest := cam.LockToHeaviest(state, cfg)
	if heaviest == nil {
		t.Fatalf("Expected heaviest body, got nil")
	}
	if heaviest.ID != 2 {
		t.Fatalf("Expected body ID 2, got %d (%s)", heaviest.ID, heaviest.Name)
	}
	if state.SelectedBodyID != 2 || !state.FollowSelected {
		t.Fatalf("Expected SelectedBodyID=2 and FollowSelected=true, got id=%d, follow=%v", state.SelectedBodyID, state.FollowSelected)
	}
}

func TestOrbitCameraLockToBarycenter(t *testing.T) {
	cam := NewOrbitCamera()
	body1 := &Body{ID: 1, Position: rl.NewVector3(-10, 0, 0), Mass: 100.0}
	body2 := &Body{ID: 2, Position: rl.NewVector3(10, 0, 0), Mass: 100.0}

	state := &SimState{
		Bodies: []*Body{body1, body2},
	}
	cfg := &Config{}

	cam.LockToBarycenter(state, cfg)
	if !state.FollowBarycenter {
		t.Fatalf("Expected FollowBarycenter to be true")
	}
	if state.FollowSelected {
		t.Fatalf("Expected FollowSelected to be false")
	}
	// COM should be at (0, 0, 0)
	if math.Abs(float64(cam.Target.X)) > 0.001 || math.Abs(float64(cam.Target.Y)) > 0.001 || math.Abs(float64(cam.Target.Z)) > 0.001 {
		t.Fatalf("Expected Target around (0,0,0), got %+v", cam.Target)
	}
}

func TestOrbitCameraFrameAll(t *testing.T) {
	cam := NewOrbitCamera()
	body1 := &Body{ID: 1, Position: rl.NewVector3(-100, 0, 0), Radius: 5.0}
	body2 := &Body{ID: 2, Position: rl.NewVector3(100, 0, 0), Radius: 5.0}

	state := &SimState{
		Bodies: []*Body{body1, body2},
	}
	cfg := &Config{}

	cam.FrameAll(state, cfg)

	// Center should be roughly at (0, 0, 0)
	if math.Abs(float64(cam.Target.X)) > 0.01 {
		t.Fatalf("Expected center X near 0, got %f", cam.Target.X)
	}
	// Distance should be large enough to contain +/-100 radius
	if cam.Distance < 200.0 {
		t.Fatalf("Expected framed distance >= 200, got %f", cam.Distance)
	}
}

func TestOrbitCameraControls(t *testing.T) {
	cam := NewOrbitCamera()

	cam.SetViewAngle(0, 1.50)
	if cam.Azimuth != 0 || cam.Elevation != 1.50 {
		t.Fatalf("Expected azim=0, elev=1.50, got azim=%f, elev=%f", cam.Azimuth, cam.Elevation)
	}

	cam.SetDistance(350.0)
	if cam.Distance != 350.0 {
		t.Fatalf("Expected distance 350, got %f", cam.Distance)
	}

	cam.SetFov(60.0)
	if cam.Camera.Fovy != 60.0 {
		t.Fatalf("Expected Fovy 60.0, got %f", cam.Camera.Fovy)
	}

	// Clamp tests
	cam.SetFov(10.0)
	if cam.Camera.Fovy != 20.0 {
		t.Fatalf("Expected clamped FOV 20.0, got %f", cam.Camera.Fovy)
	}
	cam.SetFov(150.0)
	if cam.Camera.Fovy != 100.0 {
		t.Fatalf("Expected clamped FOV 100.0, got %f", cam.Camera.Fovy)
	}
}
