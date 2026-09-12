//go:build js && wasm

package main

import (
	"math"
)

type OrbitalElements struct {
	PrimaryID     int     `json:"primary_id"`
	PrimaryName   string  `json:"primary_name"`
	Distance      float32 `json:"distance"`
	SemiMajorAxis float32 `json:"semi_major_axis"`
	Eccentricity  float32 `json:"eccentricity"`
	Period        float32 `json:"period"`
	OrbitType     string  `json:"orbit_type"`
	CircularSpeed float32 `json:"circular_speed"`
}

func CalculateOrbitalElements(target *Body, bodies []*Body, g float64) OrbitalElements {
	elem := OrbitalElements{
		PrimaryID: -1,
		OrbitType: "None",
	}

	if target == nil || len(bodies) < 2 {
		return elem
	}

	// Find primary gravitational attractor
	var primary *Body
	var maxGForce float64 = -1.0

	for _, b := range bodies {
		if b.ID == target.ID {
			continue
		}
		dx := float64(b.Position.X - target.Position.X)
		dy := float64(b.Position.Y - target.Position.Y)
		dz := float64(b.Position.Z - target.Position.Z)
		distSq := dx*dx + dy*dy + dz*dz
		if distSq < 1e-4 {
			continue
		}
		// Gravitational influence: M / r^2
		influence := b.Mass / distSq
		if influence > maxGForce {
			maxGForce = influence
			primary = b
		}
	}

	if primary == nil || primary.Mass <= target.Mass*0.05 {
		return elem
	}

	elem.PrimaryID = primary.ID
	elem.PrimaryName = primary.Name

	rx := float64(target.Position.X - primary.Position.X)
	ry := float64(target.Position.Y - primary.Position.Y)
	rz := float64(target.Position.Z - primary.Position.Z)
	r := math.Sqrt(rx*rx + ry*ry + rz*rz)
	elem.Distance = float32(r)

	vx := float64(target.Velocity.X - primary.Velocity.X)
	vy := float64(target.Velocity.Y - primary.Velocity.Y)
	vz := float64(target.Velocity.Z - primary.Velocity.Z)
	vSq := vx*vx + vy*vy + vz*vz

	mu := g * (primary.Mass + target.Mass)
	if mu <= 0 {
		return elem
	}

	elem.CircularSpeed = float32(math.Sqrt(mu / r))

	// Specific energy
	energy := 0.5*vSq - mu/r
	if math.Abs(energy) > 1e-9 {
		a := -mu / (2.0 * energy)
		elem.SemiMajorAxis = float32(a)
		if a > 0 {
			elem.Period = float32(2.0 * math.Pi * math.Sqrt((a*a*a)/mu))
		}
	}

	// Angular momentum h = r x v
	hx := ry*vz - rz*vy
	hy := rz*vx - rx*vz
	hz := rx*vy - ry*vx

	// Eccentricity vector e = (v x h)/mu - r/r
	vXh_x := vy*hz - vz*hy
	vXh_y := vz*hx - vx*hz
	vXh_z := vx*hy - vy*hx

	ex := vXh_x/mu - rx/r
	ey := vXh_y/mu - ry/r
	ez := vXh_z/mu - rz/r
	ecc := math.Sqrt(ex*ex + ey*ey + ez*ez)
	elem.Eccentricity = float32(ecc)

	if ecc < 0.05 {
		elem.OrbitType = "Circular"
	} else if ecc < 0.95 {
		elem.OrbitType = "Elliptic"
	} else if ecc < 1.05 {
		elem.OrbitType = "Parabolic"
	} else {
		elem.OrbitType = "Hyperbolic"
	}

	return elem
}
