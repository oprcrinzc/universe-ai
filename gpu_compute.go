package main

import (
	"fmt"
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
				float dist = sqrt(distSq);
				float invDist3 = 1.0 / (distSq * dist);
				acc += d * (G * m2 * invDist3);
			}
		}
		barrier();
	}

	if (i < numBodies) {
		accsOut[i] = vec4(acc, 0.0);
	}
}
`

var (
	gpuMu          sync.Mutex
	gpuInitialized bool
	gpuAvailable   bool
	gpuProgram     uint32
	gpuShader      uint32
	gpuBufIn       uint32
	gpuBufOut      uint32
	gpuBufCapacity int
	gpuBodiesCache []BodyDataGPU
	gpuAccsCache   []AccDataGPU

	// Uniform locations
	locNumBodies   int32
	locG           int32
	locSofteningSq int32

	// Dynamically loaded OpenGL API bindings
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
	fmt.Println("🪐 [GravitySim] GPU Compute Acceleration: ENABLED (NVIDIA/OpenGL 4.3 Compute Shader Active)")
	return true
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

	numGroups := uint32((n + 255) / 256)
	fnGlDispatchCompute(numGroups, 1, 1)
	fnGlMemoryBarrier(glShaderStorageBarrierBit)

	// Stream back computed accelerations
	fnGlBindBuffer(glShaderStorageBuffer, gpuBufOut)
	fnGlGetBufferSubData(glShaderStorageBuffer, 0, n*accByteSize, unsafe.Pointer(&gpuAccsCache[0]))
	fnGlUseProgram(0)

	result := make([]rl.Vector3, n)
	for i := 0; i < n; i++ {
		b := bodies[i]
		if b.IsStationary {
			result[i] = rl.NewVector3(0, 0, 0)
			b.NetForce = rl.NewVector3(0, 0, 0)
		} else {
			a := gpuAccsCache[i]
			result[i] = rl.NewVector3(a.AccX, a.AccY, a.AccZ)
			b.NetForce = rl.NewVector3(a.AccX*float32(b.Mass), a.AccY*float32(b.Mass), a.AccZ*float32(b.Mass))
		}
	}

	return result, true
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
