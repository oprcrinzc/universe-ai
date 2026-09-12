package main

import (
	"strings"
	"testing"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TestGPUMemoryLayout verifies std430 SSBO alignment constraints for GLSL compute shader
func TestGPUMemoryLayout(t *testing.T) {
	bodySize := unsafe.Sizeof(BodyDataGPU{})
	if bodySize != 32 {
		t.Fatalf("BodyDataGPU must be 32 bytes (2 x vec4) for std430 SSBO alignment, got %d", bodySize)
	}

	accSize := unsafe.Sizeof(AccDataGPU{})
	if accSize != 16 {
		t.Fatalf("AccDataGPU must be 16 bytes (1 x vec4) for std430 SSBO alignment, got %d", accSize)
	}

	// Verify field offsets
	var b BodyDataGPU
	posOffset := unsafe.Offsetof(b.PosX)
	velOffset := unsafe.Offsetof(b.VelX)
	if posOffset != 0 || velOffset != 16 {
		t.Fatalf("Invalid field offsets: PosX=%d, VelX=%d", posOffset, velOffset)
	}
}

// TestGPUComputeShaderSource verifies required GLSL compute shader directives
func TestGPUComputeShaderSource(t *testing.T) {
	src := nbodyComputeShaderSource

	if !strings.Contains(src, "#version 430") {
		t.Fatalf("Compute shader must specify #version 430")
	}
	if !strings.Contains(src, "local_size_x = 256") {
		t.Fatalf("Compute shader must use 256 thread workgroups")
	}
	if !strings.Contains(src, "shared vec4 sharedPosMass[256]") {
		t.Fatalf("Compute shader must declare shared memory on-chip cache")
	}
	if !strings.Contains(src, "inversesqrt") {
		t.Fatalf("Compute shader should use hardware inversesqrt instruction for performance")
	}
	if !strings.Contains(src, "uniform float softeningSq") {
		t.Fatalf("Compute shader must have softeningSq uniform")
	}
}

// TestGPUSafeGetters verifies telemetry getters execute safely without deadlocking
func TestGPUSafeGetters(t *testing.T) {
	name := GetGPUDeviceName()
	if name == "" {
		t.Fatalf("GetGPUDeviceName returned empty string")
	}

	_ = GetGPUVendorName()
	_ = GetGPUDispatchTimeMs()
	_ = IsGPUComputeAvailable()
}

// TestGPUAnalyticalEquivalence verifies the baseline N-body gravity math
func TestGPUAnalyticalEquivalence(t *testing.T) {
	b1 := &Body{ID: 1, Mass: 1000.0, Position: rl.NewVector3(0, 0, 0)}
	b2 := &Body{ID: 2, Mass: 50.0, Position: rl.NewVector3(10, 0, 0)}
	bodies := []*Body{b1, b2}

	accs := CalculateAccelerations(bodies, 1.0, 0.0)
	if len(accs) != 2 {
		t.Fatalf("Expected 2 accelerations, got %d", len(accs))
	}

	// Body 2 is at distance 10 from body 1 (mass 1000).
	// Expected acceleration towards body 1: a = G * M / r^2 = 1.0 * 1000 / 100 = 10.0 in -X direction
	if accs[1].X > -9.9 || accs[1].X < -10.1 {
		t.Fatalf("Expected acceleration ~ -10.0, got %f", accs[1].X)
	}
}
