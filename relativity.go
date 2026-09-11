package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CalculateRelativisticCorrections computes the 1st Post-Newtonian (1PN) General Relativity
// acceleration corrections that produce perihelion advance (precession rosette patterns)
func CalculateRelativisticCorrections(bodies []*Body, g float64, c float64, softening float64) []rl.Vector3 {
	n := len(bodies)
	pnAccs := make([]rl.Vector3, n)
	if c <= 1.0 {
		return pnAccs
	}

	c2 := float32(c * c)
	eps2 := float32(softening * softening)

	primaries := make([]*Body, 0, 8)
	for _, b := range bodies {
		if b.Mass >= 50.0 || b.IsPulsar || b.TextureType == TextureBlackHole {
			primaries = append(primaries, b)
		}
	}
	if len(primaries) == 0 {
		return pnAccs
	}

	for i := 0; i < n; i++ {
		b1 := bodies[i]
		if b1.IsStationary {
			continue
		}

		for _, b2 := range primaries {
			if b1.ID == b2.ID {
				continue
			}

			// Vector from b1 to b2
			rx := b2.Position.X - b1.Position.X
			ry := b2.Position.Y - b1.Position.Y
			rz := b2.Position.Z - b1.Position.Z

			distSq := rx*rx + ry*ry + rz*rz + eps2
			dist := float32(math.Sqrt(float64(distSq)))
			if dist < 0.1 {
				continue
			}

			// Relative velocity: v1 - v2
			vx := b1.Velocity.X - b2.Velocity.X
			vy := b1.Velocity.Y - b2.Velocity.Y
			vz := b1.Velocity.Z - b2.Velocity.Z
			vSq := vx*vx + vy*vy + vz*vz

			// Radial velocity projection: (r . v)
			rDotV := rx*vx + ry*vy + rz*vz

			// 1PN Precession term:
			// a_1PN = (G * M / (c^2 * r^3)) * [ (4*G*M/r - v^2) * r + 4*(r . v) * v ]
			invDist3 := 1.0 / (distSq * dist)
			mu := float32(g * b2.Mass)
			termScalar := (4.0*mu/dist - vSq)
			factor := (mu / c2) * invDist3

			ax := factor * (termScalar*rx + 4.0*rDotV*vx)
			ay := factor * (termScalar*ry + 4.0*rDotV*vy)
			az := factor * (termScalar*rz + 4.0*rDotV*vz)

			pnAccs[i].X += ax
			pnAccs[i].Y += ay
			pnAccs[i].Z += az

			// Accumulate into NetForce for visualization
			bodies[i].NetForce.X += ax * float32(b1.Mass)
			bodies[i].NetForce.Y += ay * float32(b1.Mass)
			bodies[i].NetForce.Z += az * float32(b1.Mass)
		}
	}

	return pnAccs
}
