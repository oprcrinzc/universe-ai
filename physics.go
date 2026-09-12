package main

import (
	"math"
	"runtime"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CalculateAccelerations computes gravitational accelerations and net forces on each body
func CalculateAccelerations(bodies []*Body, g float64, softening float64) []rl.Vector3 {
	n := len(bodies)
	accs := make([]rl.Vector3, n)
	eps2 := float32(softening * softening)
	gFloat := float32(g)

	// Reset net forces
	for i := 0; i < n; i++ {
		bodies[i].NetForce = rl.NewVector3(0, 0, 0)
	}

	if n < 64 {
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

				invDist3 := 1.0 / (distSq * dist)
				f1 := gFloat * float32(b2.Mass) * invDist3
				f2 := gFloat * float32(b1.Mass) * invDist3

				if !b1.IsStationary {
					accs[i].X += dx * f1
					accs[i].Y += dy * f1
					accs[i].Z += dz * f1

					netF := gFloat * float32(b1.Mass*b2.Mass) * invDist3
					bodies[i].NetForce.X += dx * netF
					bodies[i].NetForce.Y += dy * netF
					bodies[i].NetForce.Z += dz * netF
				}

				if !b2.IsStationary {
					accs[j].X -= dx * f2
					accs[j].Y -= dy * f2
					accs[j].Z -= dz * f2

					netF := gFloat * float32(b1.Mass*b2.Mass) * invDist3
					bodies[j].NetForce.X -= dx * netF
					bodies[j].NetForce.Y -= dy * netF
					bodies[j].NetForce.Z -= dz * netF
				}
			}
		}
		return accs
	}

	// Multi-threaded parallel direct N-body calculation for N >= 64
	numWorkers := runtime.GOMAXPROCS(0)
	if numWorkers < 1 {
		numWorkers = 1
	}
	chunkSize := (n + numWorkers - 1) / numWorkers
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}
		if start >= end {
			break
		}

		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				b1 := bodies[i]
				if b1.IsStationary {
					continue
				}
				var ax, ay, az float32
				var fx, fy, fz float32

				for j := 0; j < n; j++ {
					if i == j {
						continue
					}
					b2 := bodies[j]
					dx := b2.Position.X - b1.Position.X
					dy := b2.Position.Y - b1.Position.Y
					dz := b2.Position.Z - b1.Position.Z

					distSq := dx*dx + dy*dy + dz*dz + eps2
					dist := float32(math.Sqrt(float64(distSq)))
					if dist < 0.0001 {
						continue
					}

					invDist3 := 1.0 / (distSq * dist)
					f := gFloat * float32(b2.Mass) * invDist3
					ax += dx * f
					ay += dy * f
					az += dz * f

					netF := f * float32(b1.Mass)
					fx += dx * netF
					fy += dy * netF
					fz += dz * netF
				}
				accs[i] = rl.NewVector3(ax, ay, az)
				bodies[i].NetForce = rl.NewVector3(fx, fy, fz)
			}
		}(start, end)
	}

	wg.Wait()
	return accs
}

// computeTotalAccelerations evaluates gravitational forces via Barnes-Hut or Direct N-body,
// and optionally overlays post-Newtonian 1PN general relativity precession
func computeTotalAccelerations(bodies []*Body, cfg *Config) []rl.Vector3 {
	var accs []rl.Vector3
	n := len(bodies)

	// 1. GPU Compute Acceleration (Exact O(N^2) Direct on OpenGL Compute Shader)
	if cfg.UseGPUCompute && IsGPUComputeAvailable() {
		gpuAccs, ok := CalculateAccelerationsGPU(bodies, cfg.G, cfg.Softening)
		if ok && len(gpuAccs) == n {
			accs = gpuAccs
		}
	}

	// 2. CPU Fallback: Multi-threaded Barnes-Hut O(N log N) or Direct CPU O(N^2)
	if accs == nil {
		useBH := cfg.UseBarnesHut || n >= 350
		if useBH {
			accs = CalculateAccelerationsBarnesHut(bodies, cfg.G, cfg.Softening, cfg.BarnesHutTheta)
		} else {
			accs = CalculateAccelerations(bodies, cfg.G, cfg.Softening)
		}
	}

	if cfg.EnableRelativity {
		pnAccs := CalculateRelativisticCorrections(bodies, cfg.G, cfg.SpeedOfLight, cfg.Softening)
		for i := range accs {
			accs[i].X += pnAccs[i].X
			accs[i].Y += pnAccs[i].Y
			accs[i].Z += pnAccs[i].Z
		}
	}

	return accs
}

// stepYoshida4 advances the simulation by dt using the 4th-order symplectic Yoshida integrator
// Provides global energy conservation error O(dt^4), ideal for long-term planetary epochs and Lagrange libration
func stepYoshida4(state *SimState, cfg *Config, dt float32) {
	// Yoshida 4th-order coefficients
	const cbrt2 = 1.2599210498948731647672106
	const w1 = float32(1.0 / (2.0 - cbrt2))
	const w0 = float32(1.0 - 2.0*w1)

	// Step sizes for drift and kick
	c := [4]float32{w1 * 0.5 * dt, (w0 + w1) * 0.5 * dt, (w0 + w1) * 0.5 * dt, w1 * 0.5 * dt}
	d := [3]float32{w1 * dt, w0 * dt, w1 * dt}

	// Sub-step 1: Drift by c[0]
	for _, b := range state.Bodies {
		if !b.IsStationary {
			b.Position.X += b.Velocity.X * c[0]
			b.Position.Y += b.Velocity.Y * c[0]
			b.Position.Z += b.Velocity.Z * c[0]
		}
	}

	// Stage 1: Kick by d[0], Drift by c[1]
	a1 := computeTotalAccelerations(state.Bodies, cfg)
	for i, b := range state.Bodies {
		if !b.IsStationary {
			b.Velocity.X += a1[i].X * d[0]
			b.Velocity.Y += a1[i].Y * d[0]
			b.Velocity.Z += a1[i].Z * d[0]

			b.Position.X += b.Velocity.X * c[1]
			b.Position.Y += b.Velocity.Y * c[1]
			b.Position.Z += b.Velocity.Z * c[1]
		}
	}

	// Stage 2: Kick by d[1], Drift by c[2]
	a2 := computeTotalAccelerations(state.Bodies, cfg)
	for i, b := range state.Bodies {
		if !b.IsStationary {
			b.Velocity.X += a2[i].X * d[1]
			b.Velocity.Y += a2[i].Y * d[1]
			b.Velocity.Z += a2[i].Z * d[1]

			b.Position.X += b.Velocity.X * c[2]
			b.Position.Y += b.Velocity.Y * c[2]
			b.Position.Z += b.Velocity.Z * c[2]
		}
	}

	// Stage 3: Kick by d[2], Drift by c[3]
	a3 := computeTotalAccelerations(state.Bodies, cfg)
	for i, b := range state.Bodies {
		if !b.IsStationary {
			b.Velocity.X += a3[i].X * d[2]
			b.Velocity.Y += a3[i].Y * d[2]
			b.Velocity.Z += a3[i].Z * d[2]

			b.Position.X += b.Velocity.X * c[3]
			b.Position.Y += b.Velocity.Y * c[3]
			b.Position.Z += b.Velocity.Z * c[3]
		}
	}

	// Final acceleration evaluation for forces and telemetry
	aFinal := computeTotalAccelerations(state.Bodies, cfg)
	for i, b := range state.Bodies {
		b.Acceleration = aFinal[i]
		// Advance celestial axial rotation
		b.RotationAngle += b.RotationSpeed * dt * 25.0
		if b.RotationAngle >= 360.0 {
			b.RotationAngle -= 360.0
		} else if b.RotationAngle < 0 {
			b.RotationAngle += 360.0
		}
	}
}

// UpdatePhysics advances the simulation using the symplectic Velocity Verlet integrator
func UpdatePhysics(state *SimState, cfg *Config, dt float32) {
	startPhysics := time.Now()

	// Update notification timer even when paused
	if cfg.NotificationTimer > 0 {
		cfg.NotificationTimer -= dt
		if cfg.NotificationTimer <= 0 {
			cfg.NotificationText = ""
		}
	}

	if cfg.Paused || len(state.Bodies) == 0 {
		cfg.PhysicsTimeMs = 0
		return
	}

	n := len(state.Bodies)

	// Adaptive substeps to guarantee locked 144 FPS across all scene scales
	substeps := cfg.SubSteps
	if substeps < 1 {
		substeps = 1
	}
	if n > 2000 {
		substeps = 1
	} else if n > 600 && substeps > 2 {
		substeps = 2
	}

	stepDt := (dt * float32(cfg.TimeScale)) / float32(substeps)

	// Verify initial acceleration is populated
	initNeeded := false
	for _, b := range state.Bodies {
		if !b.IsStationary && b.Acceleration.X == 0 && b.Acceleration.Y == 0 && b.Acceleration.Z == 0 {
			initNeeded = true
			break
		}
	}
	if initNeeded {
		initAcc := computeTotalAccelerations(state.Bodies, cfg)
		for i, b := range state.Bodies {
			b.Acceleration = initAcc[i]
		}
	}

	for step := 0; step < substeps; step++ {
		curN := len(state.Bodies)
		if curN == 0 {
			break
		}

		if cfg.Integrator == IntegratorYoshida4 && curN < 1500 {
			stepYoshida4(state, cfg, stepDt)
		} else {
			halfDt := 0.5 * stepDt

			// 1. Kick velocity by halfDt and Drift position by stepDt
			for _, b := range state.Bodies {
				if !b.IsStationary {
					b.Velocity.X += b.Acceleration.X * halfDt
					b.Velocity.Y += b.Acceleration.Y * halfDt
					b.Velocity.Z += b.Acceleration.Z * halfDt

					b.Position.X += b.Velocity.X * stepDt
					b.Position.Y += b.Velocity.Y * stepDt
					b.Position.Z += b.Velocity.Z * stepDt
				}

				// Advance celestial axial rotation
				b.RotationAngle += b.RotationSpeed * stepDt * 25.0
				if b.RotationAngle >= 360.0 {
					b.RotationAngle -= 360.0
				} else if b.RotationAngle < 0 {
					b.RotationAngle += 360.0
				}
			}

			// 2. Compute new accelerations at updated positions (ONLY 1 EVALUATION PER SUBSTEP!)
			accNew := computeTotalAccelerations(state.Bodies, cfg)

			// 3. Second half kick velocity using new accelerations
			for i, b := range state.Bodies {
				if !b.IsStationary {
					b.Velocity.X += accNew[i].X * halfDt
					b.Velocity.Y += accNew[i].Y * halfDt
					b.Velocity.Z += accNew[i].Z * halfDt
				}
				b.Acceleration = accNew[i]
			}
		}

		// Handle Roche limit tidal disruption
		if cfg.EnableRocheLimit {
			HandleRocheDisruption(state, cfg)
		}

		// Handle collisions
		if cfg.Collision != CollisionNone {
			handleCollisions(state, cfg, cfg.Collision)
		}
	}

	// Update trails (optimized: skip small debris in large swarms)
	const maxTrailLen = 250
	recordInterval := float32(0.04)
	if n > 1000 {
		recordInterval = 0.08
	}

	for _, b := range state.Bodies {
		// In large swarms (N > 120), only store trails for selected body or major stars/planets
		if n > 120 && b.ID != state.SelectedBodyID && b.Mass < 5.0 && !b.IsStar {
			continue
		}

		b.TrailTimer += dt
		if b.TrailTimer >= recordInterval {
			b.TrailTimer = 0
			if len(b.Trail) == 0 {
				b.Trail = append(b.Trail, b.Position)
			} else {
				last := b.Trail[len(b.Trail)-1]
				dx := b.Position.X - last.X
				dy := b.Position.Y - last.Y
				dz := b.Position.Z - last.Z
				if dx*dx+dy*dy+dz*dz > 0.04 {
					b.Trail = append(b.Trail, b.Position)
					if len(b.Trail) > maxTrailLen {
						b.Trail = b.Trail[1:]
					}
				}
			}
		}
	}

	elapsed := float32(time.Since(startPhysics).Seconds() * 1000.0)
	cfg.PhysicsTimeMs = elapsed

	// Telemetry: calculate interactions per second
	fps := rl.GetFPS()
	if fps < 1 {
		fps = 144
	}
	if cfg.UseBarnesHut || n >= 350 {
		cfg.InteractionsPerSec = int64(n) * int64(math.Log2(float64(n)+1)*15) * int64(substeps) * int64(fps)
	} else {
		cfg.InteractionsPerSec = int64(n) * int64(n) * int64(substeps) * int64(fps)
	}
}

// handleCollisions resolves sphere collisions based on mode (Merge or Bounce)
func handleCollisions(state *SimState, cfg *Config, mode CollisionMode) {
	if mode == CollisionNone {
		return
	}

	n := len(state.Bodies)
	merged := make([]bool, n)
	isLargeSwarm := n > 250

	for i := 0; i < n; i++ {
		if merged[i] {
			continue
		}
		b1 := state.Bodies[i]
		isB1Major := b1.Mass >= 20.0 || b1.Radius >= 1.5 || b1.TextureType == TextureBlackHole || b1.IsPulsar

		for j := i + 1; j < n; j++ {
			if merged[j] {
				continue
			}
			b2 := state.Bodies[j]

			if isLargeSwarm {
				isB2Major := b2.Mass >= 20.0 || b2.Radius >= 1.5 || b2.TextureType == TextureBlackHole || b2.IsPulsar
				if !isB1Major && !isB2Major {
					continue // In massive swarms, star-star and dust-dust direct collision is negligible
				}
			}

			dx := b2.Position.X - b1.Position.X
			dy := b2.Position.Y - b1.Position.Y
			dz := b2.Position.Z - b1.Position.Z
			minDist := b1.Radius + b2.Radius
			distSq := dx*dx + dy*dy + dz*dz

			if distSq < minDist*minDist && distSq > 0.000001 {
				dist := float32(math.Sqrt(float64(distSq)))
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
						invTot := 1.0 / totalMass
						if !primary.IsStationary && !secondary.IsStationary {
							primary.Velocity.X = float32((float64(primary.Velocity.X)*primary.Mass + float64(secondary.Velocity.X)*secondary.Mass) * invTot)
							primary.Velocity.Y = float32((float64(primary.Velocity.Y)*primary.Mass + float64(secondary.Velocity.Y)*secondary.Mass) * invTot)
							primary.Velocity.Z = float32((float64(primary.Velocity.Z)*primary.Mass + float64(secondary.Velocity.Z)*secondary.Mass) * invTot)
						} else if secondary.IsStationary {
							primary.IsStationary = true
							primary.Velocity = rl.NewVector3(0, 0, 0)
						}

						vol1 := math.Pow(float64(primary.Radius), 3)
						vol2 := math.Pow(float64(secondary.Radius), 3)
						primary.Radius = float32(math.Pow(vol1+vol2, 1.0/3.0))
						primary.Mass = totalMass
					}

					if secondary.IsStar {
						primary.IsStar = true
					}

					if state.SelectedBodyID == secondary.ID {
						state.SelectedBodyID = primary.ID
					}

					// Trigger high-energy gravitational wave burst upon significant celestial collision
					if primary.Mass >= 0.2 || secondary.Mass >= 0.2 || primary.IsStar || secondary.IsStar || primary.TextureType == TextureBlackHole || secondary.TextureType == TextureBlackHole {
						EmitGravitationalWaveBurst(state, cfg, primary, secondary)
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

					rvx := b2.Velocity.X - b1.Velocity.X
					rvy := b2.Velocity.Y - b1.Velocity.Y
					rvz := b2.Velocity.Z - b1.Velocity.Z

					velAlongNormal := rvx*nx + rvy*ny + rvz*nz
					if velAlongNormal < 0 {
						restitution := float32(0.8)
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
							impulse := impulseMag / totalInvMass
							if !b1.IsStationary {
								b1.Velocity.X -= nx * impulse * invM1
								b1.Velocity.Y -= ny * impulse * invM1
								b1.Velocity.Z -= nz * impulse * invM1
							}
							if !b2.IsStationary {
								b2.Velocity.X += nx * impulse * invM2
								b2.Velocity.Y += ny * impulse * invM2
								b2.Velocity.Z += nz * impulse * invM2
							}
						}
					}
				}
			}
		}
	}

	if len(merged) > 0 {
		newBodies := make([]*Body, 0, len(state.Bodies)-len(merged))
		for i, b := range state.Bodies {
			if !merged[i] {
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
