//go:build js && wasm

package main

import "math"

// Vector3 represents a 3D vector in Cartesian coordinates
type Vector3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Color represents an RGBA color
type Color struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

func NewVector3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

func Vector3Add(v1, v2 Vector3) Vector3 {
	return Vector3{X: v1.X + v2.X, Y: v1.Y + v2.Y, Z: v1.Z + v2.Z}
}

func Vector3Subtract(v1, v2 Vector3) Vector3 {
	return Vector3{X: v1.X - v2.X, Y: v1.Y - v2.Y, Z: v1.Z - v2.Z}
}

func Vector3Scale(v Vector3, scale float32) Vector3 {
	return Vector3{X: v.X * scale, Y: v.Y * scale, Z: v.Z * scale}
}

func Vector3Length(v Vector3) float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func Vector3LengthSqr(v Vector3) float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func Vector3Distance(v1, v2 Vector3) float32 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	dz := v1.Z - v2.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func Vector3Normalize(v Vector3) Vector3 {
	l := Vector3Length(v)
	if l < 1e-9 {
		return Vector3{}
	}
	inv := 1.0 / l
	return Vector3{X: v.X * inv, Y: v.Y * inv, Z: v.Z * inv}
}

func Vector3DotProduct(v1, v2 Vector3) float32 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

func Vector3CrossProduct(v1, v2 Vector3) Vector3 {
	return Vector3{
		X: v1.Y*v2.Z - v1.Z*v2.Y,
		Y: v1.Z*v2.X - v1.X*v2.Z,
		Z: v1.X*v2.Y - v1.Y*v2.X,
	}
}
