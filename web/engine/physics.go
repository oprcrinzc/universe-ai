package main

import (
	"math"
	"math/rand"
)

func UpdatePhysics(state *SimState, cfg *Config, dt float32) {
	if cfg.Paused || dt <= 0 || len(state.Bodies) == 0 {
		return
	}

	simDt := dt * float32(cfg.TimeScale)
	substeps := cfg.SubSteps
	if substeps < 1 {
		substeps = 1
	}

	n := len(state.Bodies)
	if n > 600 && substeps > 2 {
		substeps = 2
	} else if n > 2500 {
		substeps = 1
	}

	subDt := simDt / float32(substeps)

	for step := 0; step < substeps; step++ {
		// 1. Position update: r(t+dt) = r(t) + v(t)*dt + 0.5*a(t)*dt^2
		for _, b := range state.Bodies {
			if b.IsStationary {
				continue
			}
			b.Position.X += b.Velocity.X*subDt + 0.5*b.Acceleration.X*subDt*subDt
			b.Position.Y += b.Velocity.Y*subDt + 0.5*b.Acceleration.Y*subDt*subDt
			b.Position.Z += b.Velocity.Z*subDt + 0.5*b.Acceleration.Z*subDt*subDt
		}

		// 2. Compute accelerations at t+dt
		var newAccs []Vector3
		var interactions int64
		if cfg.EnableBarnesHut {
			newAccs, interactions = computeBarnesHutAccelerations(state, cfg)
		} else {
			newAccs, interactions = computeDirectAccelerations(state, cfg)
		}
		state.InteractionsPerSec = interactions

		// 3. Optional 1PN General Relativity corrections
		if cfg.EnableRelativity {
			applyRelativity1PN(state, cfg, newAccs)
		}

		// 4. Velocity update: v(t+dt) = v(t) + 0.5*(a(t) + a(t+dt))*dt
		for i, b := range state.Bodies {
			if b.IsStationary {
				continue
			}
			newAcc := newAccs[i]
			b.Velocity.X += 0.5 * (b.Acceleration.X + newAcc.X) * subDt
			b.Velocity.Y += 0.5 * (b.Acceleration.Y + newAcc.Y) * subDt
			b.Velocity.Z += 0.5 * (b.Acceleration.Z + newAcc.Z) * subDt
			b.Acceleration = newAcc
		}

		// 5. Roche tidal disruption
		if cfg.EnableRocheLimit {
			handleRocheBreakup(state, cfg)
		}

		// 6. Collisions
		handleCollisions(state, cfg.Collision)
	}

	// Trail recording (sample periodically)
	updateTrails(state)

	state.Time += float64(simDt)
}

func computeBarnesHutAccelerations(state *SimState, cfg *Config) ([]Vector3, int64) {
	n := len(state.Bodies)
	accs := make([]Vector3, n)
	tree := NewOctree(state.Bodies)
	var totalInteractions int64

	theta := cfg.BarnesHutTheta
	if theta <= 0.1 {
		theta = 0.7
	}

	for i, b := range state.Bodies {
		if b.IsStationary {
			continue
		}
		a, inter := tree.ComputeAcceleration(b, cfg.G, cfg.Softening, theta)
		accs[i] = a
		totalInteractions += inter
	}
	return accs, totalInteractions
}

func computeDirectAccelerations(state *SimState, cfg *Config) ([]Vector3, int64) {
	n := len(state.Bodies)
	accs := make([]Vector3, n)
	softSq := float32(cfg.Softening * cfg.Softening)
	g := float32(cfg.G)
	var totalInteractions int64

	for i := 0; i < n; i++ {
		b1 := state.Bodies[i]
		if b1.IsStationary {
			continue
		}
		var ax, ay, az float32
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			b2 := state.Bodies[j]
			dx := b2.Position.X - b1.Position.X
			dy := b2.Position.Y - b1.Position.Y
			dz := b2.Position.Z - b1.Position.Z
			distSq := dx*dx + dy*dy + dz*dz + softSq
			invDist := 1.0 / float32(math.Sqrt(float64(distSq)))
			invDist3 := invDist / distSq
			f := g * float32(b2.Mass) * invDist3
			ax += dx * f
			ay += dy * f
			az += dz * f
			totalInteractions++
		}
		accs[i] = Vector3{X: ax, Y: ay, Z: az}
	}
	return accs, totalInteractions
}

func applyRelativity1PN(state *SimState, cfg *Config, accs []Vector3) {
	c := cfg.SpeedOfLight
	if c <= 10.0 {
		c = 180.0
	}
	c2 := float32(c * c)
	g := float32(cfg.G)

	for i, b := range state.Bodies {
		if b.IsStationary {
			continue
		}
		for j, other := range state.Bodies {
			if i == j || other.Mass < 50.0 {
				continue
			}
			dx := b.Position.X - other.Position.X
			dy := b.Position.Y - other.Position.Y
			dz := b.Position.Z - other.Position.Z
			r := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz + 1.0)))
			if r < 0.1 {
				continue
			}

			vSq := b.Velocity.X*b.Velocity.X + b.Velocity.Y*b.Velocity.Y + b.Velocity.Z*b.Velocity.Z
			gm := g * float32(other.Mass)
			rDotV := (dx*b.Velocity.X + dy*b.Velocity.Y + dz*b.Velocity.Z) / r

			term1 := (4.0*gm/r - vSq) / (r * r)
			term2 := 4.0 * rDotV / (r * r)
			prefactor := gm / (c2 * r)

			accs[i].X += prefactor * (term1*dx + term2*b.Velocity.X)
			accs[i].Y += prefactor * (term1*dy + term2*b.Velocity.Y)
			accs[i].Z += prefactor * (term1*dz + term2*b.Velocity.Z)
		}
	}
}

func handleRocheBreakup(state *SimState, cfg *Config) {
	n := len(state.Bodies)
	if n > 250 {
		return
	}

	for i := 0; i < len(state.Bodies); i++ {
		primary := state.Bodies[i]
		if primary.Mass < 50.0 {
			continue
		}

		for j := 0; j < len(state.Bodies); j++ {
			if i == j {
				continue
			}
			victim := state.Bodies[j]
			if victim.Mass > primary.Mass*0.1 || victim.Mass < 0.01 || victim.IsStar || victim.IsBlackHole {
				continue
			}

			dx := victim.Position.X - primary.Position.X
			dy := victim.Position.Y - primary.Position.Y
			dz := victim.Position.Z - primary.Position.Z
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

			rocheDist := victim.Radius * 2.44 * float32(math.Pow(primary.Mass/victim.Mass, 1.0/3.0))
			if rocheDist < primary.Radius*1.5 {
				rocheDist = primary.Radius * 1.5
			}
			if rocheDist > primary.Radius*6.0 {
				rocheDist = primary.Radius * 6.0
			}
			if dist < rocheDist && dist > primary.Radius*1.15 {
				// Fragment victim into 8 debris particles
				numFrags := 8
				fragMass := victim.Mass / float64(numFrags)
				fragRad := victim.Radius * 0.4

				for k := 0; k < numFrags; k++ {
					angle := float64(k) * (2.0 * math.Pi / float64(numFrags))
					spread := victim.Radius * 0.8
					fx := victim.Position.X + float32(math.Cos(angle))*spread
					fy := victim.Position.Y + (rand.Float32()-0.5)*spread*0.3
					fz := victim.Position.Z + float32(math.Sin(angle))*spread

					tangentX := -float32(math.Sin(angle)) * 0.5
					tangentZ := float32(math.Cos(angle)) * 0.5

					frag := &Body{
						ID:       state.NextID,
						Name:     victim.Name + " Debris",
						Position: Vector3{X: fx, Y: fy, Z: fz},
						Velocity: Vector3{
							X: victim.Velocity.X + tangentX,
							Y: victim.Velocity.Y,
							Z: victim.Velocity.Z + tangentZ,
						},
						Mass:        fragMass,
						Radius:      fragRad,
						Color:       victim.Color,
						TextureType: 5, // Moon/Asteroid
					}
					state.NextID++
					state.Bodies = append(state.Bodies, frag)
				}

				// Remove victim
				state.Bodies = append(state.Bodies[:j], state.Bodies[j+1:]...)
				state.TidalBreakupCount++
				return
			}
		}
	}
}

func handleCollisions(state *SimState, mode CollisionMode) {
	if mode == CollisionGhost {
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
		isB1Major := b1.Mass >= 20.0 || b1.Radius >= 1.5 || b1.IsBlackHole || b1.IsPulsar

		for j := i + 1; j < n; j++ {
			if merged[j] {
				continue
			}
			b2 := state.Bodies[j]

			if isLargeSwarm {
				isB2Major := b2.Mass >= 20.0 || b2.Radius >= 1.5 || b2.IsBlackHole || b2.IsPulsar
				if !isB1Major && !isB2Major {
					continue
				}
			}

			dx := b2.Position.X - b1.Position.X
			dy := b2.Position.Y - b1.Position.Y
			dz := b2.Position.Z - b1.Position.Z
			minDist := b1.Radius + b2.Radius
			distSq := dx*dx + dy*dy + dz*dz

			if distSq < minDist*minDist && distSq > 1e-6 {
				dist := float32(math.Sqrt(float64(distSq)))
				if mode == CollisionMerge {
					// Inelastic merge: conserve momentum & volume
					totalMass := b1.Mass + b2.Mass
					if totalMass > 0 {
						b1.Velocity.X = float32((float64(b1.Velocity.X)*b1.Mass + float64(b2.Velocity.X)*b2.Mass) / totalMass)
						b1.Velocity.Y = float32((float64(b1.Velocity.Y)*b1.Mass + float64(b2.Velocity.Y)*b2.Mass) / totalMass)
						b1.Velocity.Z = float32((float64(b1.Velocity.Z)*b1.Mass + float64(b2.Velocity.Z)*b2.Mass) / totalMass)
					}
					r1_3 := float64(b1.Radius * b1.Radius * b1.Radius)
					r2_3 := float64(b2.Radius * b2.Radius * b2.Radius)
					b1.Radius = float32(math.Cbrt(r1_3 + r2_3))
					b1.Mass = totalMass
					merged[j] = true
					state.MergeCount++
				} else if mode == CollisionBounce {
					// Elastic bounce impulse
					nx := dx / dist
					ny := dy / dist
					nz := dz / dist
					relVx := b2.Velocity.X - b1.Velocity.X
					relVy := b2.Velocity.Y - b1.Velocity.Y
					relVz := b2.Velocity.Z - b1.Velocity.Z
					velAlongNormal := relVx*nx + relVy*ny + relVz*nz

					if velAlongNormal < 0 {
						restitution := float32(0.85)
						impulse := -(1.0 + restitution) * velAlongNormal
						invM1 := float32(1.0 / b1.Mass)
						invM2 := float32(1.0 / b2.Mass)
						jImp := impulse / (invM1 + invM2)

						b1.Velocity.X -= invM1 * jImp * nx
						b1.Velocity.Y -= invM1 * jImp * ny
						b1.Velocity.Z -= invM1 * jImp * nz

						b2.Velocity.X += invM2 * jImp * nx
						b2.Velocity.Y += invM2 * jImp * ny
						b2.Velocity.Z += invM2 * jImp * nz
					}
					state.CollisionCount++
				}
			}
		}
	}

	if mode == CollisionMerge {
		active := make([]*Body, 0, n)
		for i, b := range state.Bodies {
			if !merged[i] {
				active = append(active, b)
			}
		}
		state.Bodies = active
	}
}

func updateTrails(state *SimState) {
	if len(state.Bodies) > 120 {
		return
	}
	for _, b := range state.Bodies {
		if len(b.Trail) < MaxTrailPoints {
			b.Trail = append(b.Trail, b.Position)
		} else {
			if b.TrailHead >= len(b.Trail) {
				b.TrailHead = 0
			}
			b.Trail[b.TrailHead] = b.Position
			b.TrailHead = (b.TrailHead + 1) % len(b.Trail)
		}
	}
}
