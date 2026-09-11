package main

import (
	"math"
	"math/rand"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestBarnesHutOctreeConstruction(t *testing.T) {
	bodies := []*Body{
		{ID: 1, Position: rl.NewVector3(0, 0, 0), Mass: 100.0, Radius: 2.0},
		{ID: 2, Position: rl.NewVector3(10, 0, 0), Mass: 10.0, Radius: 1.0},
		{ID: 3, Position: rl.NewVector3(-10, 5, 2), Mass: 5.0, Radius: 1.0},
		{ID: 4, Position: rl.NewVector3(0, -15, 8), Mass: 20.0, Radius: 1.5},
	}

	root := BuildOctree(bodies)
	if root == nil {
		t.Fatal("Expected non-nil octree root")
	}

	expectedTotalMass := 100.0 + 10.0 + 5.0 + 20.0
	if math.Abs(root.TotalMass-expectedTotalMass) > 0.001 {
		t.Fatalf("Expected root total mass %f, got %f", expectedTotalMass, root.TotalMass)
	}

	if root.Count != 4 {
		t.Fatalf("Expected root count 4, got %d", root.Count)
	}
}

func TestBarnesHutAccelerationComparison(t *testing.T) {
	// Generate random cluster of 50 bodies
	r := rand.New(rand.NewSource(42))
	n := 50
	bodies := make([]*Body, n)

	for i := 0; i < n; i++ {
		pos := rl.NewVector3(
			float32((r.Float64()-0.5)*100.0),
			float32((r.Float64()-0.5)*100.0),
			float32((r.Float64()-0.5)*100.0),
		)
		vel := rl.NewVector3(
			float32((r.Float64()-0.5)*5.0),
			float32((r.Float64()-0.5)*5.0),
			float32((r.Float64()-0.5)*5.0),
		)
		mass := 1.0 + r.Float64()*50.0
		bodies[i] = &Body{
			ID:       int64(i + 1),
			Position: pos,
			Velocity: vel,
			Mass:     mass,
			Radius:   1.0,
		}
	}

	g := 1.0
	softening := 0.5
	theta := float32(0.6) // Conservative opening angle for high accuracy

	directAccs := CalculateAccelerations(bodies, g, softening)
	bhAccs := CalculateAccelerationsBarnesHut(bodies, g, softening, theta)

	// Verify that Barnes-Hut approximates Direct calculation within close margin
	var maxRelErr float64 = 0.0
	for i := 0; i < n; i++ {
		dLen := float64(rl.Vector3Length(directAccs[i]))
		diff := rl.Vector3Subtract(directAccs[i], bhAccs[i])
		diffLen := float64(rl.Vector3Length(diff))

		if dLen > 0.001 {
			relErr := diffLen / dLen
			if relErr > maxRelErr {
				maxRelErr = relErr
			}
		}
	}

	t.Logf("Barnes-Hut max relative acceleration error vs Direct: %.4f%%", maxRelErr*100)
	if maxRelErr > 0.15 { // Expect < 15% error for theta=0.6 on random cloud
		t.Fatalf("Barnes-Hut error too high: %.4f%%", maxRelErr*100)
	}
}
