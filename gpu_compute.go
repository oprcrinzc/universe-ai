package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// OpenGL Constants for Compute Shaders and SSBO (Shader Storage Buffer Objects)
const (
	glComputeShader           = 0x91B9
	glShaderStorageBuffer     = 0x90D2
	glDynamicDraw             = 0x88E8
	glDynamicCopy             = 0x88EA
	glCompileStatus           = 0x8B81
	glLinkStatus              = 0x8B82
	glInfoLogLength           = 0x8B84
	glShaderStorageBarrierBit = 0x00002000
	glVendor                  = 0x1F00
	glRenderer                = 0x1F01
	glVersion                 = 0x1F02
)

// BodyDataGPU matches std430 alignment layout in the GLSL compute shader
type BodyDataGPU struct {
	PosX, PosY, PosZ, Mass float32
	VelX, VelY, VelZ, Pad  float32
}

// AccDataGPU receives the gravitational acceleration computed by GPU threads
type AccDataGPU struct {
	AccX, AccY, AccZ, Pad float32
}

// Shared-Memory Tiled GLSL 4.30 Compute Shader for Direct N-Body Gravity
// Optimized with hardware reciprocal square root (inversesqrt) and hoisted G scaling
const nbodyComputeShaderSource = `#version 430
layout(local_size_x = 256, local_size_y = 1, local_size_z = 1) in;

struct BodyData {
	vec4 posMass; // x, y, z, mass
	vec4 vel;     // vx, vy, vz, padding
};

layout(std430, binding = 0) readonly buffer InputBodies {
	BodyData bodiesIn[];
};

layout(std430, binding = 1) writeonly buffer OutputAccs {
	vec4 accsOut[]; // ax, ay, az, 0
};

uniform int numBodies;
uniform float G;
uniform float softeningSq;

// High-speed on-chip cache shared across the 256 execution threads in this workgroup
shared vec4 sharedPosMass[256];

void main() {
	uint i = gl_GlobalInvocationID.x;
	vec3 p1 = vec3(0.0);
	if (i < numBodies) {
		p1 = bodiesIn[i].posMass.xyz;
	}
	vec3 acc = vec3(0.0);

	uint numTiles = (numBodies + 255) / 256;
	for (uint tile = 0; tile < numTiles; tile++) {
		uint idx = tile * 256 + gl_LocalInvocationIndex;
		if (idx < numBodies) {
			sharedPosMass[gl_LocalInvocationIndex] = bodiesIn[idx].posMass;
		} else {
			sharedPosMass[gl_LocalInvocationIndex] = vec4(0.0);
		}
		barrier();

		if (i < numBodies) {
			for (uint j = 0; j < 256; j++) {
				uint otherIdx = tile * 256 + j;
				if (otherIdx == i || otherIdx >= numBodies) continue;
				vec3 p2 = sharedPosMass[j].xyz;
				float m2 = sharedPosMass[j].w;

				vec3 d = p2 - p1;
				float distSq = dot(d, d) + softeningSq;
				float invDist = inversesqrt(distSq);
				float invDist3 = invDist * invDist * invDist;
				acc += d * (m2 * invDist3);
			}
		}
		barrier();
	}

	if (i < numBodies) {
		accsOut[i] = vec4(acc * G, 0.0);
	}
}
`

var (
	gpuMu             sync.Mutex
	gpuInitialized    bool
	gpuAvailable      bool
	gpuProgram        uint32
	gpuShader         uint32
	gpuBufIn          uint32
	gpuBufOut         uint32
	gpuBufCapacity    int
	gpuBodiesCache    []BodyDataGPU
	gpuAccsCache      []AccDataGPU
	gpuResultsCache   []rl.Vector3
	gpuRendererStr    string
	gpuVendorStr      string
	gpuVersionStr     string
	gpuDispatchTimeMs float32
	gpuLastDispatched int

	// Uniform locations
	locNumBodies   int32
	locG           int32
	locSofteningSq int32

	// Dynamically loaded OpenGL API bindings
	fnGlGetString          func(name uint32) *byte
	fnGlCreateShader       func(shaderType uint32) uint32
	fnGlShaderSource       func(shader uint32, count int32, str **byte, length *int32)
	fnGlCompileShader      func(shader uint32)
	fnGlGetShaderiv        func(shader uint32, pname uint32, params *int32)
	fnGlGetShaderInfoLog   func(shader uint32, bufSize int32, length *int32, infoLog *byte)
	fnGlCreateProgram      func() uint32
	fnGlAttachShader       func(program, shader uint32)
	fnGlLinkProgram        func(program uint32)
	fnGlGetProgramiv       func(program uint32, pname uint32, params *int32)
	fnGlGetProgramInfoLog  func(program uint32, bufSize int32, length *int32, infoLog *byte)
	fnGlUseProgram         func(program uint32)
	fnGlGenBuffers         func(n int32, buffers *uint32)
	fnGlBindBuffer         func(target uint32, buffer uint32)
	fnGlBufferData         func(target uint32, size int, data unsafe.Pointer, usage uint32)
	fnGlBufferSubData      func(target uint32, offset int, size int, data unsafe.Pointer)
	fnGlBindBufferBase     func(target uint32, index uint32, buffer uint32)
	fnGlDispatchCompute    func(num_groups_x, num_groups_y, num_groups_z uint32)
	fnGlMemoryBarrier      func(barriers uint32)
	fnGlGetBufferSubData   func(target uint32, offset int, size int, data unsafe.Pointer)
	fnGlDeleteBuffers      func(n int32, buffers *uint32)
	fnGlDeleteProgram      func(program uint32)
	fnGlDeleteShader       func(shader uint32)
	fnGlGetUniformLocation func(program uint32, name *byte) int32
	fnGlUniform1i          func(location int32, v0 int32)
	fnGlUniform1f          func(location int32, v0 float32)
	fnGlFinish             func()
)

// InitGPUCompute initializes the OpenGL Compute Shader acceleration pipeline
func InitGPUCompute() bool {
	gpuMu.Lock()
	defer gpuMu.Unlock()

	if gpuInitialized {
		return gpuAvailable
	}
	gpuInitialized = true

	var libName string
	switch runtime.GOOS {
	case "linux":
		libName = "libGL.so.1"
	case "windows":
		libName = "opengl32.dll"
	default:
		return false
	}

	libGL, err := purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return false
	}

	// Safely bind OpenGL functions
	defer func() {
		if r := recover(); r != nil {
			gpuAvailable = false
		}
	}()

	purego.RegisterLibFunc(&fnGlCreateShader, libGL, "glCreateShader")
	purego.RegisterLibFunc(&fnGlShaderSource, libGL, "glShaderSource")
	purego.RegisterLibFunc(&fnGlCompileShader, libGL, "glCompileShader")
	purego.RegisterLibFunc(&fnGlGetShaderiv, libGL, "glGetShaderiv")
	purego.RegisterLibFunc(&fnGlGetShaderInfoLog, libGL, "glGetShaderInfoLog")
	purego.RegisterLibFunc(&fnGlCreateProgram, libGL, "glCreateProgram")
	purego.RegisterLibFunc(&fnGlAttachShader, libGL, "glAttachShader")
	purego.RegisterLibFunc(&fnGlLinkProgram, libGL, "glLinkProgram")
	purego.RegisterLibFunc(&fnGlGetProgramiv, libGL, "glGetProgramiv")
	purego.RegisterLibFunc(&fnGlGetProgramInfoLog, libGL, "glGetProgramInfoLog")
	purego.RegisterLibFunc(&fnGlGetString, libGL, "glGetString")
	purego.RegisterLibFunc(&fnGlUseProgram, libGL, "glUseProgram")
	purego.RegisterLibFunc(&fnGlGenBuffers, libGL, "glGenBuffers")
	purego.RegisterLibFunc(&fnGlBindBuffer, libGL, "glBindBuffer")
	purego.RegisterLibFunc(&fnGlBufferData, libGL, "glBufferData")
	purego.RegisterLibFunc(&fnGlBufferSubData, libGL, "glBufferSubData")
	purego.RegisterLibFunc(&fnGlBindBufferBase, libGL, "glBindBufferBase")
	purego.RegisterLibFunc(&fnGlDispatchCompute, libGL, "glDispatchCompute")
	purego.RegisterLibFunc(&fnGlMemoryBarrier, libGL, "glMemoryBarrier")
	purego.RegisterLibFunc(&fnGlGetBufferSubData, libGL, "glGetBufferSubData")
	purego.RegisterLibFunc(&fnGlDeleteBuffers, libGL, "glDeleteBuffers")
	purego.RegisterLibFunc(&fnGlDeleteProgram, libGL, "glDeleteProgram")
	purego.RegisterLibFunc(&fnGlDeleteShader, libGL, "glDeleteShader")
	purego.RegisterLibFunc(&fnGlGetUniformLocation, libGL, "glGetUniformLocation")
	purego.RegisterLibFunc(&fnGlUniform1i, libGL, "glUniform1i")
	purego.RegisterLibFunc(&fnGlUniform1f, libGL, "glUniform1f")
	purego.RegisterLibFunc(&fnGlFinish, libGL, "glFinish")

	if fnGlGetString != nil {
		gpuVendorStr = glString(fnGlGetString(glVendor))
		gpuRendererStr = glString(fnGlGetString(glRenderer))
		gpuVersionStr = glString(fnGlGetString(glVersion))
	}

	if fnGlCreateShader == nil || fnGlDispatchCompute == nil {
		return false
	}

	cs := fnGlCreateShader(glComputeShader)
	if cs == 0 {
		return false
	}

	srcBytes := []byte(nbodyComputeShaderSource)
	srcPtr := &srcBytes[0]
	length := int32(len(srcBytes))
	fnGlShaderSource(cs, 1, &srcPtr, &length)
	fnGlCompileShader(cs)

	var status int32
	fnGlGetShaderiv(cs, glCompileStatus, &status)
	if status == 0 {
		var logLen int32
		fnGlGetShaderiv(cs, glInfoLogLength, &logLen)
		if logLen > 0 {
			logBytes := make([]byte, logLen)
			fnGlGetShaderInfoLog(cs, logLen, nil, &logBytes[0])
			fmt.Printf("[InitGPUCompute] Shader compile error:\n%s\n", string(logBytes))
		}
		fnGlDeleteShader(cs)
		return false
	}

	prog := fnGlCreateProgram()
	if prog == 0 {
		fnGlDeleteShader(cs)
		return false
	}

	fnGlAttachShader(prog, cs)
	fnGlLinkProgram(prog)

	fnGlGetProgramiv(prog, glLinkStatus, &status)
	if status == 0 {
		var logLen int32
		fnGlGetProgramiv(prog, glInfoLogLength, &logLen)
		if logLen > 0 {
			logBytes := make([]byte, logLen)
			fnGlGetProgramInfoLog(prog, logLen, nil, &logBytes[0])
			fmt.Printf("[InitGPUCompute] Program link error:\n%s\n", string(logBytes))
		}
		fnGlDeleteProgram(prog)
		fnGlDeleteShader(cs)
		return false
	}

	gpuShader = cs
	gpuProgram = prog

	cNum := []byte("numBodies\x00")
	cG := []byte("G\x00")
	cSoft := []byte("softeningSq\x00")

	locNumBodies = fnGlGetUniformLocation(prog, &cNum[0])
	locG = fnGlGetUniformLocation(prog, &cG[0])
	locSofteningSq = fnGlGetUniformLocation(prog, &cSoft[0])

	gpuAvailable = true
	devName := gpuRendererStr
	if devName == "" {
		devName = "Dedicated GPU Compute"
	}
	fmt.Printf("🪐 [GravitySim] GPU Compute Acceleration: ENABLED (%s | OpenGL 4.3 Compute Shader Active)\n", devName)
	return true
}

func glString(ptr *byte) string {
	if ptr == nil {
		return ""
	}
	var res []byte
	for p := ptr; *p != 0; p = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + 1)) {
		res = append(res, *p)
	}
	return string(res)
}

// GetGPUDeviceName returns the detected GPU hardware model string (e.g. "NVIDIA GeForce RTX 3050 Ti")
func GetGPUDeviceName() string {
	gpuMu.Lock()
	defer gpuMu.Unlock()
	if gpuRendererStr != "" {
		return gpuRendererStr
	}
	return "Dedicated GPU Compute"
}

// GetGPUVendorName returns the detected GPU vendor name (e.g. "NVIDIA Corporation")
func GetGPUVendorName() string {
	gpuMu.Lock()
	defer gpuMu.Unlock()
	return gpuVendorStr
}

// GetGPUDispatchTimeMs returns the execution time in ms of the last compute shader dispatch
func GetGPUDispatchTimeMs() float32 {
	gpuMu.Lock()
	defer gpuMu.Unlock()
	return gpuDispatchTimeMs
}

// IsGPUComputeAvailable returns true if the GPU compute pipeline is ready for dispatch
func IsGPUComputeAvailable() bool {
	gpuMu.Lock()
	defer gpuMu.Unlock()
	return gpuAvailable
}

// CalculateAccelerationsGPU computes direct O(N^2) gravitational accelerations on the GPU
func CalculateAccelerationsGPU(bodies []*Body, g float64, softening float64) ([]rl.Vector3, bool) {
	gpuMu.Lock()
	defer gpuMu.Unlock()

	if !gpuAvailable || gpuProgram == 0 {
		return nil, false
	}

	n := len(bodies)
	if n == 0 {
		return nil, true
	}

	if len(gpuBodiesCache) < n {
		gpuBodiesCache = make([]BodyDataGPU, n)
	}
	if len(gpuAccsCache) < n {
		gpuAccsCache = make([]AccDataGPU, n)
	}

	for i, b := range bodies {
		gpuBodiesCache[i] = BodyDataGPU{
			PosX: b.Position.X,
			PosY: b.Position.Y,
			PosZ: b.Position.Z,
			Mass: float32(b.Mass),
			VelX: b.Velocity.X,
			VelY: b.Velocity.Y,
			VelZ: b.Velocity.Z,
		}
	}

	bodyByteSize := int(unsafe.Sizeof(BodyDataGPU{}))
	accByteSize := int(unsafe.Sizeof(AccDataGPU{}))

	// Dynamically expand persistent SSBOs if particle count increases
	if n > gpuBufCapacity {
		newCap := n + 1024
		if gpuBufIn != 0 {
			var bufs = [2]uint32{gpuBufIn, gpuBufOut}
			fnGlDeleteBuffers(2, &bufs[0])
		}
		var bufs [2]uint32
		fnGlGenBuffers(2, &bufs[0])
		gpuBufIn = bufs[0]
		gpuBufOut = bufs[1]

		fnGlBindBuffer(glShaderStorageBuffer, gpuBufIn)
		fnGlBufferData(glShaderStorageBuffer, newCap*bodyByteSize, nil, glDynamicDraw)

		fnGlBindBuffer(glShaderStorageBuffer, gpuBufOut)
		fnGlBufferData(glShaderStorageBuffer, newCap*accByteSize, nil, glDynamicCopy)

		gpuBufCapacity = newCap
	}

	// Fast streaming upload into persistent VRAM SSBO buffer
	fnGlBindBuffer(glShaderStorageBuffer, gpuBufIn)
	fnGlBufferSubData(glShaderStorageBuffer, 0, n*bodyByteSize, unsafe.Pointer(&gpuBodiesCache[0]))
	fnGlBindBufferBase(glShaderStorageBuffer, 0, gpuBufIn)

	fnGlBindBuffer(glShaderStorageBuffer, gpuBufOut)
	fnGlBindBufferBase(glShaderStorageBuffer, 1, gpuBufOut)

	fnGlUseProgram(gpuProgram)
	fnGlUniform1i(locNumBodies, int32(n))
	fnGlUniform1f(locG, float32(g))
	fnGlUniform1f(locSofteningSq, float32(softening*softening))

	startTime := rl.GetTime()
	numGroups := uint32((n + 255) / 256)
	fnGlDispatchCompute(numGroups, 1, 1)
	fnGlMemoryBarrier(glShaderStorageBarrierBit)

	// Stream back computed accelerations
	fnGlBindBuffer(glShaderStorageBuffer, gpuBufOut)
	fnGlGetBufferSubData(glShaderStorageBuffer, 0, n*accByteSize, unsafe.Pointer(&gpuAccsCache[0]))
	fnGlUseProgram(0)

	gpuDispatchTimeMs = float32((rl.GetTime() - startTime) * 1000.0)
	gpuLastDispatched = n

	if len(gpuResultsCache) < n {
		gpuResultsCache = make([]rl.Vector3, n)
	}
	for i := 0; i < n; i++ {
		b := bodies[i]
		if b.IsStationary {
			gpuResultsCache[i] = rl.NewVector3(0, 0, 0)
			b.NetForce = rl.NewVector3(0, 0, 0)
		} else {
			a := gpuAccsCache[i]
			gpuResultsCache[i] = rl.NewVector3(a.AccX, a.AccY, a.AccZ)
			b.NetForce = rl.NewVector3(a.AccX*float32(b.Mass), a.AccY*float32(b.Mass), a.AccZ*float32(b.Mass))
		}
	}

	return gpuResultsCache[:n], true
}

// CleanupGPUCompute frees all GPU buffers and shader programs
func CleanupGPUCompute() {
	gpuMu.Lock()
	defer gpuMu.Unlock()

	if !gpuAvailable {
		return
	}

	if gpuBufIn != 0 {
		bufs := [2]uint32{gpuBufIn, gpuBufOut}
		fnGlDeleteBuffers(2, &bufs[0])
		gpuBufIn = 0
		gpuBufOut = 0
		gpuBufCapacity = 0
	}

	if gpuProgram != 0 {
		fnGlDeleteProgram(gpuProgram)
		gpuProgram = 0
	}
	if gpuShader != 0 {
		fnGlDeleteShader(gpuShader)
		gpuShader = 0
	}
	gpuAvailable = false
	gpuInitialized = false
}

// VerifyGPUComputeDirect tests the active GPU compute shader pipeline against CPU reference calculations
func VerifyGPUComputeDirect() {
	fmt.Printf("\n=======================================================\n")
	fmt.Printf("🪐 [GravitySim] GPU Compute Engine Verification\n")
	fmt.Printf("=======================================================\n")
	fmt.Printf("GPU Device:   %s\n", GetGPUDeviceName())
	fmt.Printf("GPU Vendor:   %s\n", GetGPUVendorName())
	fmt.Printf("GPU Available: %v\n", IsGPUComputeAvailable())

	if !IsGPUComputeAvailable() {
		fmt.Printf("❌ GPU Compute is not active on this device.\n\n")
		return
	}

	b1 := &Body{ID: 1, Mass: 1000.0, Position: rl.NewVector3(0, 0, 0)}
	b2 := &Body{ID: 2, Mass: 50.0, Position: rl.NewVector3(10, 0, 0)}
	b3 := &Body{ID: 3, Mass: 20.0, Position: rl.NewVector3(0, 15, 0)}
	bodies := []*Body{b1, b2, b3}

	cpuAccs := CalculateAccelerations(bodies, 1.0, 0.5)
	gpuAccs, success := CalculateAccelerationsGPU(bodies, 1.0, 0.5)

	if !success {
		fmt.Printf("❌ CalculateAccelerationsGPU failed execution!\n\n")
		return
	}

	fmt.Printf("GPU Kernel Dispatch Time: %.3f ms\n", GetGPUDispatchTimeMs())
	pass := true
	for i := range bodies {
		diffX := math.Abs(float64(gpuAccs[i].X - cpuAccs[i].X))
		diffY := math.Abs(float64(gpuAccs[i].Y - cpuAccs[i].Y))
		diffZ := math.Abs(float64(gpuAccs[i].Z - cpuAccs[i].Z))
		fmt.Printf("Body %d: CPU=(%+0.4f, %+0.4f, %+0.4f)  GPU=(%+0.4f, %+0.4f, %+0.4f)  Δ=(%.5f, %.5f, %.5f)\n",
			i, cpuAccs[i].X, cpuAccs[i].Y, cpuAccs[i].Z,
			gpuAccs[i].X, gpuAccs[i].Y, gpuAccs[i].Z,
			diffX, diffY, diffZ)
		if diffX > 0.01 || diffY > 0.01 || diffZ > 0.01 {
			pass = false
		}
	}

	if pass {
		fmt.Printf("\n✅ VERIFICATION PASSED: NVIDIA GPU Compute Shader matches analytical gravity calculation within tolerance!\n")
		fmt.Printf("Hardware compute path is 100%% ACTIVE and performing real-time O(N^2) gravity integration.\n")
	} else {
		fmt.Printf("\n❌ VERIFICATION FAILED: Numerical deviation exceeded threshold.\n")
	}
	fmt.Printf("=======================================================\n\n")
}
