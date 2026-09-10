package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CalculateAccelerations computes gravitational accelerations and net forces on each body
func CalculateAccelerations(bodies []*Body, g float64, softening float64) []rl.Vector3 {
	n := len(bodies)
	accs := make([]rl.Vector3, n)
	eps2 := float32(softening * softening)

	// Reset net forces
	for i := 0; i < n; i++ {
		bodies[i].NetForce = rl.NewVector3(0, 0, 0)
	}

	for i := 0; i < n; i++ {
		b1 := bodies[i]
		for j := i + 1; j < n; j++ {
			b2 := bodies[j]

			dx := b2.Position.X - b1.Position.X
			dy := b2.Position.Y - b1.Position.Y
			dz := b2.Position.Z - b1.Position.Z

			distSq := dx*dx + dy*dy + dz*dz + eps2
			dist := float32(math.Sqrt(float64(distSq)))
			if dist < 0.0001 {
				continue
			}

			// F = G * m1 * m2 / distSq
			// a1 = G * m2 / distSq, towards b2
			// a2 = G * m1 / distSq, towards b1
			invDist3 := 1.0 / (distSq * dist)

			f1 := float32(g*b2.Mass) * invDist3
			f2 := float32(g*b1.Mass) * invDist3

			if !b1.IsStationary {
				accs[i].X += dx * f1
				accs[i].Y += dy * f1
				accs[i].Z += dz * f1

				// Store net force vector for visual arrows
				netF := float32(g * b1.Mass * b2.Mass) * invDist3
				bodies[i].NetForce.X += dx * netF
				bodies[i].NetForce.Y += dy * netF
				bodies[i].NetForce.Z += dz * netF
			}

			if !b2.IsStationary {
				accs[j].X -= dx * f2
				accs[j].Y -= dy * f2
				accs[j].Z -= dz * f2

				netF := float32(g * b1.Mass * b2.Mass) * invDist3
				bodies[j].NetForce.X -= dx * netF
				bodies[j].NetForce.Y -= dy * netF
				bodies[j].NetForce.Z -= dz * netF
			}
		}
	}

	return accs
}

// UpdatePhysics advances the simulation using the symplectic Velocity Verlet integrator
func UpdatePhysics(state *SimState, cfg *Config, dt float32) {
	if cfg.Paused || len(state.Bodies) == 0 {
		return
	}

	substeps := cfg.SubSteps
	if substeps < 1 {
		substeps = 1
	}

	stepDt := (dt * float32(cfg.TimeScale)) / float32(substeps)

	for step := 0; step < substeps; step++ {
		n := len(state.Bodies)
		if n == 0 {
			break
		}

		// 1. Initial accelerations
		acc1 := CalculateAccelerations(state.Bodies, cfg.G, cfg.Softening)

		// 2. Update positions: x(t + dt) = x(t) + v(t)*dt + 0.5*a(t)*dt^2
		halfDtSq := 0.5 * stepDt * stepDt
		for i, b := range state.Bodies {
			if !b.IsStationary {
				b.Position.X += b.Velocity.X*stepDt + acc1[i].X*halfDtSq
				b.Position.Y += b.Velocity.Y*stepDt + acc1[i].Y*halfDtSq
				b.Position.Z += b.Velocity.Z*stepDt + acc1[i].Z*halfDtSq
			}
			b.Acceleration = acc1[i]
		}

		// 3. New accelerations at x(t + dt)
		acc2 := CalculateAccelerations(state.Bodies, cfg.G, cfg.Softening)

		// 4. Update velocities: v(t + dt) = v(t) + 0.5*(a(t) + a(t + dt))*dt
		halfDt := 0.5 * stepDt
		for i, b := range state.Bodies {
			if !b.IsStationary {
				b.Velocity.X += (acc1[i].X + acc2[i].X) * halfDt
				b.Velocity.Y += (acc1[i].Y + acc2[i].Y) * halfDt
				b.Velocity.Z += (acc1[i].Z + acc2[i].Z) * halfDt
			}
		}

		// Handle collisions
		if cfg.Collision != CollisionNone {
			handleCollisions(state, cfg.Collision)
		}
	}

	// Update trails
	const maxTrailLen = 300
	for _, b := range state.Bodies {
		b.TrailTimer += dt
		if b.TrailTimer >= 0.04 { // record trail point every 40ms
			b.TrailTimer = 0
			// Only record if moved noticeably or initial
			if len(b.Trail) == 0 {
				b.Trail = append(b.Trail, b.Position)
			} else {
				last := b.Trail[len(b.Trail)-1]
				dx := b.Position.X - last.X
				dy := b.Position.Y - last.Y
				dz := b.Position.Z - last.Z
				if dx*dx+dy*dy+dz*dz > 0.05 {
					b.Trail = append(b.Trail, b.Position)
					if len(b.Trail) > maxTrailLen {
						b.Trail = b.Trail[1:]
					}
				}
			}
		}
	}
}

// handleCollisions resolves sphere collisions based on mode (Merge or Bounce)
func handleCollisions(state *SimState, mode CollisionMode) {
	if mode == CollisionNone {
		return
	}

	merged := make(map[int]bool)

	for i := 0; i < len(state.Bodies); i++ {
		if merged[i] {
			continue
		}
		b1 := state.Bodies[i]

		for j := i + 1; j < len(state.Bodies); j++ {
			if merged[j] {
				continue
			}
			b2 := state.Bodies[j]

			dx := b2.Position.X - b1.Position.X
			dy := b2.Position.Y - b1.Position.Y
			dz := b2.Position.Z - b1.Position.Z
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

			minDist := b1.Radius + b2.Radius
			if dist < minDist && dist > 0.0001 {
				if mode == CollisionMerge {
					// Merge smaller body into larger body
					primary := b1
					secondary := b2
					removeIdx := j
					if b2.Mass > b1.Mass {
						primary = b2
						secondary = b1
						removeIdx = i
					}

					totalMass := primary.Mass + secondary.Mass
					if totalMass > 0 {
						// Conservation of momentum: v = (m1*v1 + m2*v2) / (m1+m2)
						if !primary.IsStationary && !secondary.IsStationary {
							primary.Velocity.X = float32((float64(primary.Velocity.X)*primary.Mass + float64(secondary.Velocity.X)*secondary.Mass) / totalMass)
							primary.Velocity.Y = float32((float64(primary.Velocity.Y)*primary.Mass + float64(secondary.Velocity.Y)*secondary.Mass) / totalMass)
							primary.Velocity.Z = float32((float64(primary.Velocity.Z)*primary.Mass + float64(secondary.Velocity.Z)*secondary.Mass) / totalMass)
						} else if secondary.IsStationary {
							primary.IsStationary = true
							primary.Velocity = rl.NewVector3(0, 0, 0)
						}

						// Radius scales with volume (m^1/3)
						vol1 := math.Pow(float64(primary.Radius), 3)
						vol2 := math.Pow(float64(secondary.Radius), 3)
						primary.Radius = float32(math.Pow(vol1+vol2, 1.0/3.0))

						primary.Mass = totalMass
					}

					if secondary.IsStar {
						primary.IsStar = true
					}

					// If selected body was merged, update selection
					if state.SelectedBodyID == secondary.ID {
						state.SelectedBodyID = primary.ID
					}

					merged[removeIdx] = true
					if removeIdx == i {
						break
					}
				} else if mode == CollisionBounce {
					// Elastic bounce along collision normal
					nx := dx / dist
					ny := dy / dist
					nz := dz / dist

					// Separate slightly to avoid sticking
					overlap := 0.5 * (minDist - dist)
					if !b1.IsStationary {
						b1.Position.X -= nx * overlap
						b1.Position.Y -= ny * overlap
						b1.Position.Z -= nz * overlap
					}
					if !b2.IsStationary {
						b2.Position.X += nx * overlap
						b2.Position.Y += ny * overlap
						b2.Position.Z += nz * overlap
					}

					// Relative velocity
					rvx := b2.Velocity.X - b1.Velocity.X
					rvy := b2.Velocity.Y - b1.Velocity.Y
					rvz := b2.Velocity.Z - b1.Velocity.Z

					velAlongNormal := rvx*nx + rvy*ny + rvz*nz
					if velAlongNormal < 0 { // moving towards each other
						restitution := float32(0.8) // coefficient of restitution
						impulseMag := -(1.0 + restitution) * velAlongNormal
						invM1 := float32(0)
						invM2 := float32(0)
						if !b1.IsStationary && b1.Mass > 0 {
							invM1 = float32(1.0 / b1.Mass)
						}
						if !b2.IsStationary && b2.Mass > 0 {
							invM2 = float32(1.0 / b2.Mass)
						}

						totalInvMass := invM1 + invM2
						if totalInvMass > 0 {
							impulseMag /= totalInvMass
							if !b1.IsStationary {
								b1.Velocity.X -= nx * impulseMag * invM1
								b1.Velocity.Y -= ny * impulseMag * invM1
								b1.Velocity.Z -= nz * impulseMag * invM1
							}
							if !b2.IsStationary {
								b2.Velocity.X += nx * impulseMag * invM2
								b2.Velocity.Y += ny * impulseMag * invM2
								b2.Velocity.Z += nz * impulseMag * invM2
							}
						}
					}
				}
			}
		}
	}

	// Filter out merged bodies
	if len(merged) > 0 {
		newBodies := make([]*Body, 0, len(state.Bodies)-len(merged))
		for idx, b := range state.Bodies {
			if !merged[idx] {
				newBodies = append(newBodies, b)
			}
		}
		state.Bodies = newBodies
	}
}

// FindPrimaryBody finds the most massive body (other than itself) or nearest heavy body
func FindPrimaryBody(bodies []*Body, target *Body) *Body {
	var best *Body
	var maxInfluence float64 = -1

	for _, b := range bodies {
		if b.ID == target.ID {
			continue
		}
		dx := target.Position.X - b.Position.X
		dy := target.Position.Y - b.Position.Y
		dz := target.Position.Z - b.Position.Z
		distSq := float64(dx*dx + dy*dy + dz*dz)
		if distSq < 0.001 {
			continue
		}
		// Influence is M / r^2 (gravitational pull)
		influence := b.Mass / distSq
		if influence > maxInfluence {
			maxInfluence = influence
			best = b
		}
	}
	return best
}

// CircularizeOrbit sets the target body's velocity to enter a stable circular orbit around the primary body
func CircularizeOrbit(target *Body, primary *Body, g float64, clockwise bool) {
	if target == nil || primary == nil || target.IsStationary {
		return
	}

	// Radial vector from primary to target on XZ plane
	dx := target.Position.X - primary.Position.X
	dz := target.Position.Z - primary.Position.Z
	dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

	if dist < 0.1 {
		return
	}

	// Circular orbit speed v = sqrt(G * M / r)
	speed := float32(math.Sqrt((g * primary.Mass) / float64(dist)))

	// Tangent vector on XZ plane
	// Perpendicular to (dx, dz): (-dz, dx) or (dz, -dx)
	var tx, tz float32
	if clockwise {
		tx = dz / dist
		tz = -dx / dist
	} else {
		tx = -dz / dist
		tz = dx / dist
	}

	// Target velocity = primary velocity + tangent * speed
	target.Velocity = rl.NewVector3(
		primary.Velocity.X+tx*speed,
		primary.Velocity.Y,
		primary.Velocity.Z+tz*speed,
	)
}

// ApplyImpulse adds a velocity impulse (delta-v) to a body
func ApplyImpulse(b *Body, deltaV rl.Vector3) {
	if b == nil || b.IsStationary {
		return
	}
	b.Velocity.X += deltaV.X
	b.Velocity.Y += deltaV.Y
	b.Velocity.Z += deltaV.Z
}

// BoostPrograde modifies speed along current direction of travel
func BoostPrograde(b *Body, percent float32) {
	if b == nil || b.IsStationary {
		return
	}
	b.Velocity.X *= (1.0 + percent)
	b.Velocity.Y *= (1.0 + percent)
	b.Velocity.Z *= (1.0 + percent)
}
