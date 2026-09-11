package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// HandleRocheDisruption checks for moons or planetoids crossing the Roche limit of massive bodies
// and disintegrates them into an accretion ring of debris particles
func HandleRocheDisruption(state *SimState, cfg *Config) {
	if !cfg.EnableRocheLimit || len(state.Bodies) < 2 || len(state.Bodies) > 300 {
		return
	}

	var disruptedIndices []int
	var newFragments []*Body

	for i := 0; i < len(state.Bodies); i++ {
		b1 := state.Bodies[i]
		if b1.IsStationary || b1.IsStar || b1.Disrupted || b1.Mass <= 0.005 {
			continue
		}

		for j := 0; j < len(state.Bodies); j++ {
			if i == j {
				continue
			}
			primary := state.Bodies[j]

			// Primary must be substantially more massive (at least 15x)
			if primary.Mass < b1.Mass*15.0 {
				continue
			}

			dx := primary.Position.X - b1.Position.X
			dy := primary.Position.Y - b1.Position.Y
			dz := primary.Position.Z - b1.Position.Z
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

			// Must be outside direct physical contact
			minDist := primary.Radius + b1.Radius
			if dist <= minDist {
				continue
			}

			// Classical Roche limit: d_roche = 2.44 * r * (M / m)^(1/3)
			massRatio := primary.Mass / b1.Mass
			rocheDist := float32(2.44 * float64(b1.Radius) * math.Pow(massRatio, 1.0/3.0))

			// Cap Roche limit to reasonable visual bounds
			if rocheDist < primary.Radius*1.5 {
				rocheDist = primary.Radius * 1.5
			}
			if rocheDist > primary.Radius*6.0 {
				rocheDist = primary.Radius * 6.0
			}

			if dist < rocheDist {
				// Tidal disruption triggered!
				b1.Disrupted = true
				disruptedIndices = append(disruptedIndices, i)

				numFragments := 14
				fragMass := b1.Mass / float64(numFragments)
				fragRadius := float32(math.Max(0.18, float64(b1.Radius)/math.Pow(float64(numFragments), 1.0/3.0)))

				// Unit vector towards primary
				nx := dx / dist
				ny := dy / dist
				nz := dz / dist

				// Tangent direction perpendicular to radial on orbit plane
				tx := -nz
				tz := nx
				tLen := float32(math.Sqrt(float64(tx*tx + tz*tz)))
				if tLen > 0.001 {
					tx /= tLen
					tz /= tLen
				}

				for k := 0; k < numFragments; k++ {
					// Disperse along tidal axis and tangential shear
					radialOffset := (float32(k)/float32(numFragments) - 0.5) * (b1.Radius * 2.8)
					tangentOffset := (rand.Float32() - 0.5) * (b1.Radius * 2.0)
					vertOffset := (rand.Float32() - 0.5) * (b1.Radius * 0.8)

					fragPos := rl.NewVector3(
						b1.Position.X+nx*radialOffset+tx*tangentOffset,
						b1.Position.Y+ny*radialOffset+vertOffset,
						b1.Position.Z+nz*radialOffset+tz*tangentOffset,
					)

					// Orbital shear velocity: inner fragments orbit faster, outer fragments orbit slower
					shearDeltaV := -radialOffset * 0.45
					fragVel := rl.NewVector3(
						b1.Velocity.X+tx*shearDeltaV+(rand.Float32()-0.5)*0.25,
						b1.Velocity.Y+(rand.Float32()-0.5)*0.2,
						b1.Velocity.Z+tz*shearDeltaV+(rand.Float32()-0.5)*0.25,
					)

					// Color variation
					cDelta := int32(rand.Intn(40) - 20)
					fragCol := rl.NewColor(
						uint8(math.Max(0, math.Min(255, float64(int32(b1.Color.R)+cDelta)))),
						uint8(math.Max(0, math.Min(255, float64(int32(b1.Color.G)+cDelta)))),
						uint8(math.Max(0, math.Min(255, float64(int32(b1.Color.B)+cDelta)))),
						220,
					)

					frag := &Body{
						ID:            state.NextID,
						Name:          fmt.Sprintf("%s Ring Debris #%d", b1.Name, k+1),
						Position:      fragPos,
						Velocity:      fragVel,
						Mass:          fragMass,
						Radius:        fragRadius,
						Color:         fragCol,
						IsStationary:  false,
						IsStar:        false,
						Trail:         make([]rl.Vector3, 0, 80),
						TextureType:   TextureMoon,
						RotationAngle: rand.Float32() * 360.0,
						RotationSpeed: (rand.Float32() - 0.5) * 40.0,
						Disrupted:     true, // Do not disrupt fragments further
					}
					state.NextID++
					newFragments = append(newFragments, frag)
				}

				// HUD Notification
				cfg.NotificationText = fmt.Sprintf("ROCHE TIDAL BREAKUP: %s breached Roche limit of %s!", b1.Name, primary.Name)
				cfg.NotificationTimer = 4.5

				// If selected was disrupted, switch selection to primary
				if state.SelectedBodyID == b1.ID {
					state.SelectedBodyID = primary.ID
				}

				break // Disrupted by first dominant primary
			}
		}
	}

	// Remove disrupted bodies and add new ring fragments
	if len(disruptedIndices) > 0 {
		removeMap := make(map[int]bool)
		for _, idx := range disruptedIndices {
			removeMap[idx] = true
		}

		filtered := make([]*Body, 0, len(state.Bodies)-len(disruptedIndices)+len(newFragments))
		for idx, b := range state.Bodies {
			if !removeMap[idx] {
				filtered = append(filtered, b)
			}
		}
		filtered = append(filtered, newFragments...)
		state.Bodies = filtered
	}
}
