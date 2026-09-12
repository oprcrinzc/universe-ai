package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// LagrangePoint holds calculated position, velocity, and properties of an equilibrium point
type LagrangePoint struct {
	Name      string
	Index     int // 1..5
	Position  rl.Vector3
	Velocity  rl.Vector3
	Potential float64
	Stable    bool
}

// ComputeLagrangePoints calculates exact coordinates and velocities for L1 through L5
// in the inertial frame for two primary bodies (e.g. Sun and Jupiter or Earth and Moon)
func ComputeLagrangePoints(b1, b2 *Body, g float64) [5]LagrangePoint {
	var points [5]LagrangePoint

	m1 := b1.Mass
	m2 := b2.Mass
	primary := b1
	secondary := b2
	if m2 > m1 {
		m1, m2 = b2.Mass, b1.Mass
		primary = b2
		secondary = b1
	}

	totalM := m1 + m2
	if totalM <= 0 {
		return points
	}
	mu := m2 / totalM // Mass ratio <= 0.5

	rVec := rl.Vector3Subtract(secondary.Position, primary.Position)
	R := float64(rl.Vector3Length(rVec))
	if R < 1e-4 {
		return points
	}

	uHat := rl.Vector3Scale(rVec, float32(1.0/R))

	// Determine orbital plane normal vector
	vRel := rl.Vector3Subtract(secondary.Velocity, primary.Velocity)
	hVec := rl.Vector3CrossProduct(rVec, vRel)
	hLen := float64(rl.Vector3Length(hVec))
	var nHat rl.Vector3
	if hLen > 1e-4 {
		nHat = rl.Vector3Scale(hVec, float32(1.0/hLen))
	} else {
		// Default to Y-up normal for planar X-Z simulation
		nHat = rl.NewVector3(0, 1, 0)
	}

	// In-plane perpendicular vector (ahead in orbital direction)
	wHat := rl.Vector3Normalize(rl.Vector3CrossProduct(nHat, uHat))
	// Ensure wHat aligns with secondary velocity direction
	if rl.Vector3DotProduct(wHat, vRel) < 0 {
		wHat = rl.Vector3Scale(wHat, -1.0)
	}

	// Barycenter position and velocity
	baryPos := rl.NewVector3(
		float32((m1*float64(primary.Position.X)+m2*float64(secondary.Position.X))/totalM),
		float32((m1*float64(primary.Position.Y)+m2*float64(secondary.Position.Y))/totalM),
		float32((m1*float64(primary.Position.Z)+m2*float64(secondary.Position.Z))/totalM),
	)
	baryVel := rl.NewVector3(
		float32((m1*float64(primary.Velocity.X)+m2*float64(secondary.Velocity.X))/totalM),
		float32((m1*float64(primary.Velocity.Y)+m2*float64(secondary.Velocity.Y))/totalM),
		float32((m1*float64(primary.Velocity.Z)+m2*float64(secondary.Velocity.Z))/totalM),
	)

	// Mean motion / angular velocity
	omega := math.Sqrt((g * totalM) / (R * R * R))
	omegaVec := rl.Vector3Scale(nHat, float32(omega))

	// Helper to compute inertial velocity from rotating frame offset
	computeVel := func(pos rl.Vector3) rl.Vector3 {
		relPos := rl.Vector3Subtract(pos, baryPos)
		vRot := rl.Vector3CrossProduct(omegaVec, relPos)
		return rl.Vector3Add(baryVel, vRot)
	}

	// Effective potential in rotating frame along primary-secondary axis x (origin at barycenter)
	// Primary is at -mu * R, Secondary is at (1 - mu) * R
	x1 := -mu * R
	x2 := (1.0 - mu) * R

	evalCollinearRoot := func(initX float64, minX, maxX float64) float64 {
		x := initX
		for iter := 0; iter < 16; iter++ {
			d1 := x - x1
			d2 := x - x2
			absD1 := math.Abs(d1)
			absD2 := math.Abs(d2)
			if absD1 < 1e-6 || absD2 < 1e-6 {
				break
			}

			// f(x) = - (1-mu)*R^3 * sgn(d1)/d1^2 - mu*R^3 * sgn(d2)/d2^2 + x
			f := -(1.0-mu)*R*R*R/(d1*absD1) - mu*R*R*R/(d2*absD2) + x
			// f'(x) = 2*(1-mu)*R^3 / |d1|^3 + 2*mu*R^3 / |d2|^3 + 1
			fPrime := 2.0*(1.0-mu)*R*R*R/(absD1*absD1*absD1) + 2.0*mu*R*R*R/(absD2*absD2*absD2) + 1.0

			step := f / fPrime
			x -= step
			if x <= minX {
				x = minX + 1e-4
			} else if x >= maxX {
				x = maxX - 1e-4
			}
			if math.Abs(step) < 1e-9*R {
				break
			}
		}
		return x
	}

	// 1. L1: Collinear between Primary and Secondary (x1 < x < x2)
	// First approximation: Hill sphere distance r_H = R * (mu/3)^(1/3)
	hillRadius := R * math.Pow(mu/3.0, 1.0/3.0)
	initL1 := x2 - hillRadius
	xL1 := evalCollinearRoot(initL1, x1+1e-4, x2-1e-4)
	posL1 := rl.Vector3Add(baryPos, rl.Vector3Scale(uHat, float32(xL1)))
	points[0] = LagrangePoint{
		Name:      "L1 (Inter-Orbital / SOHO)",
		Index:     1,
		Position:  posL1,
		Velocity:  computeVel(posL1),
		Potential: ComputeEffectivePotential(posL1, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 2. L2: Collinear beyond Secondary (x > x2)
	initL2 := x2 + hillRadius
	xL2 := evalCollinearRoot(initL2, x2+1e-4, x2+R*3.0)
	posL2 := rl.Vector3Add(baryPos, rl.Vector3Scale(uHat, float32(xL2)))
	points[1] = LagrangePoint{
		Name:      "L2 (Outer / JWST Halo)",
		Index:     2,
		Position:  posL2,
		Velocity:  computeVel(posL2),
		Potential: ComputeEffectivePotential(posL2, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 3. L3: Collinear behind Primary (x < x1)
	initL3 := x1 - R*(1.0-5.0*mu/12.0)
	xL3 := evalCollinearRoot(initL3, x1-R*3.0, x1-1e-4)
	posL3 := rl.Vector3Add(baryPos, rl.Vector3Scale(uHat, float32(xL3)))
	points[2] = LagrangePoint{
		Name:      "L3 (Counter-Orbital)",
		Index:     3,
		Position:  posL3,
		Velocity:  computeVel(posL3),
		Potential: ComputeEffectivePotential(posL3, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 4. L4: Equilateral triangle 60 degrees ahead in orbit
	posL4 := rl.Vector3Add(baryPos, rl.Vector3Add(
		rl.Vector3Scale(uHat, float32((0.5-mu)*R)),
		rl.Vector3Scale(wHat, float32(math.Sqrt(3.0)/2.0*R)),
	))
	points[3] = LagrangePoint{
		Name:      "L4 (Trojan Leading +60°)",
		Index:     4,
		Position:  posL4,
		Velocity:  computeVel(posL4),
		Potential: ComputeEffectivePotential(posL4, primary, secondary, baryPos, g, omega),
		Stable:    mu < 0.0385208965, // Routh stability criterion
	}

	// 5. L5: Equilateral triangle 60 degrees trailing in orbit
	posL5 := rl.Vector3Add(baryPos, rl.Vector3Add(
		rl.Vector3Scale(uHat, float32((0.5-mu)*R)),
		rl.Vector3Scale(wHat, float32(-math.Sqrt(3.0)/2.0*R)),
	))
	points[4] = LagrangePoint{
		Name:      "L5 (Greek Trailing -60°)",
		Index:     5,
		Position:  posL5,
		Velocity:  computeVel(posL5),
		Potential: ComputeEffectivePotential(posL5, primary, secondary, baryPos, g, omega),
		Stable:    mu < 0.0385208965,
	}

	return points
}

// ComputeEffectivePotential evaluates Jacobi effective potential in rotating frame:
// Phi_eff(r) = - G*M1/|r - r1| - G*M2/|r - r2| - 0.5 * omega^2 * |r - r_bary|^2
func ComputeEffectivePotential(pos rl.Vector3, b1, b2 *Body, baryPos rl.Vector3, g, omega float64) float64 {
	d1 := float64(rl.Vector3Distance(pos, b1.Position))
	d2 := float64(rl.Vector3Distance(pos, b2.Position))
	rBary := float64(rl.Vector3Distance(pos, baryPos))

	if d1 < 0.1 {
		d1 = 0.1
	}
	if d2 < 0.1 {
		d2 = 0.1
	}

	gravPot := -(g*b1.Mass)/d1 - (g*b2.Mass)/d2
	centrifugalPot := -0.5 * omega * omega * rBary * rBary
	return gravPot + centrifugalPot
}

// FormatLagrangeSummary returns a human-readable telemetry string of the 5 points
func FormatLagrangeSummary(points [5]LagrangePoint) string {
	return fmt.Sprintf("L1: (%.1f, %.1f) | L2: (%.1f, %.1f) | L3: (%.1f, %.1f) | L4: (%.1f, %.1f) | L5: (%.1f, %.1f)",
		points[0].Position.X, points[0].Position.Z,
		points[1].Position.X, points[1].Position.Z,
		points[2].Position.X, points[2].Position.Z,
		points[3].Position.X, points[3].Position.Z,
		points[4].Position.X, points[4].Position.Z,
	)
}

// GetLagrangePair determines the two primary bodies for Lagrange equilibrium calculations.
// If a body is currently selected and is not the sole heaviest star (e.g. Earth, Jupiter, Moon),
// it finds the dominant gravitational attractor (e.g. the Sun or parent planet) and pairs them.
// Otherwise, it selects the two most massive bodies in the simulation.
func GetLagrangePair(state *SimState) (*Body, *Body) {
	if state == nil || len(state.Bodies) < 2 {
		return nil, nil
	}

	// 1. If a secondary body is selected (and not the single most massive body)
	if state.SelectedBodyID != -1 {
		var selected *Body
		var heaviest *Body
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				selected = b
			}
			if heaviest == nil || b.Mass > heaviest.Mass {
				heaviest = b
			}
		}

		if selected != nil && heaviest != nil && selected.ID != heaviest.ID && selected.Mass > 0.00001 {
			// Find closest heavier primary or default to heaviest star
			var primary *Body = heaviest
			// If selected is a moon orbiting a planet, find its immediate parent planet
			for _, b := range state.Bodies {
				if b.ID != selected.ID && b.Mass > selected.Mass*5.0 {
					dist := rl.Vector3Distance(b.Position, selected.Position)
					primDist := rl.Vector3Distance(primary.Position, selected.Position)
					// If this body is significantly closer and sufficiently massive to dominate local gravity
					if dist < primDist*0.4 && b.Mass >= selected.Mass*10.0 {
						primary = b
					}
				}
			}
			return primary, selected
		}
	}

	// 2. Default: find two most massive bodies
	var primary, secondary *Body
	for _, b := range state.Bodies {
		if primary == nil || b.Mass > primary.Mass {
			secondary = primary
			primary = b
		} else if secondary == nil || b.Mass > secondary.Mass {
			secondary = b
		}
	}

	if primary == nil || secondary == nil || secondary.Mass < 0.0001 {
		return nil, nil
	}
	return primary, secondary
}

// SpawnProbeAtLagrange creates a scientific probe / satellite body precisely at the given Lagrange point
func SpawnProbeAtLagrange(state *SimState, cfg *Config, pt LagrangePoint, namePrefix string) *Body {
	if namePrefix == "" {
		namePrefix = pt.Name
	}
	name := fmt.Sprintf("%s Probe", namePrefix)
	color := rl.NewColor(100, 220, 255, 255)
	switch pt.Index {
	case 1:
		color = rl.NewColor(255, 220, 50, 255) // Gold
	case 2:
		color = rl.NewColor(80, 220, 255, 255) // Sky Cyan
	case 3:
		color = rl.NewColor(255, 120, 80, 255) // Coral
	case 4:
		color = rl.NewColor(120, 255, 180, 255) // Emerald
	case 5:
		color = rl.NewColor(255, 180, 120, 255) // Amber
	}

	body := addBodyTex(state, name, pt.Position, pt.Velocity, 0.001, 0.45, color, false, false, TextureMoon, 1.0)
	body.CelestialIcon = IconProbe
	if cfg != nil {
		cfg.NotificationText = fmt.Sprintf("[PROBE DEPLOYED] '%s' placed into equilibrium orbit!", name)
		cfg.NotificationTimer = 2.5
	}
	return body
}

// SpawnLagrangeSwarm spawns a cluster of asteroids in stable libration around L4 or L5
func SpawnLagrangeSwarm(state *SimState, cfg *Config, primary, secondary *Body, pt LagrangePoint, count int) {
	if count <= 0 {
		count = 12
	}
	m1 := primary.Mass
	m2 := secondary.Mass
	totalM := m1 + m2
	if totalM <= 0 {
		return
	}
	rVec := rl.Vector3Subtract(secondary.Position, primary.Position)
	R := float64(rl.Vector3Length(rVec))
	if R < 1e-4 {
		return
	}
	omega := math.Sqrt((cfg.G * totalM) / (R * R * R))

	baryPos := rl.NewVector3(
		float32((m1*float64(primary.Position.X)+m2*float64(secondary.Position.X))/totalM),
		float32((m1*float64(primary.Position.Y)+m2*float64(secondary.Position.Y))/totalM),
		float32((m1*float64(primary.Position.Z)+m2*float64(secondary.Position.Z))/totalM),
	)
	baryVel := rl.NewVector3(
		float32((m1*float64(primary.Velocity.X)+m2*float64(secondary.Velocity.X))/totalM),
		float32((m1*float64(primary.Velocity.Y)+m2*float64(secondary.Velocity.Y))/totalM),
		float32((m1*float64(primary.Velocity.Z)+m2*float64(secondary.Velocity.Z))/totalM),
	)

	vRel := rl.Vector3Subtract(secondary.Velocity, primary.Velocity)
	hVec := rl.Vector3CrossProduct(rVec, vRel)
	hLen := float64(rl.Vector3Length(hVec))
	var nHat rl.Vector3
	if hLen > 1e-4 {
		nHat = rl.Vector3Scale(hVec, float32(1.0/hLen))
	} else {
		nHat = rl.NewVector3(0, 1, 0)
	}
	omegaVec := rl.Vector3Scale(nHat, float32(omega))

	typeName := "Trojan"
	col := rl.NewColor(140, 200, 255, 255)
	if pt.Index == 5 {
		typeName = "Greek"
		col = rl.NewColor(255, 190, 130, 255)
	}

	for i := 0; i < count; i++ {
		// Small radial and angular perturbations around equilibrium point
		angOffset := (float64(i%4) - 1.5) * 0.035
		radOffset := (float64(i/4) - 1.5) * 0.012 * R

		// Perturb position relative to barycenter
		relP := rl.Vector3Subtract(pt.Position, baryPos)
		dist := float64(rl.Vector3Length(relP))
		if dist < 1e-4 {
			continue
		}
		dir := rl.Vector3Scale(relP, float32(1.0/dist))
		newDist := float32(dist + radOffset)

		cosA := float32(math.Cos(angOffset))
		sinA := float32(math.Sin(angOffset))
		perp := rl.Vector3CrossProduct(nHat, dir)

		newDir := rl.Vector3Add(rl.Vector3Scale(dir, cosA), rl.Vector3Scale(perp, sinA))
		pos := rl.Vector3Add(baryPos, rl.Vector3Scale(newDir, newDist))
		pos.Y += float32((i%3)-1) * float32(0.004*R)

		// Exact rotating frame velocity
		relPosNew := rl.Vector3Subtract(pos, baryPos)
		vRot := rl.Vector3CrossProduct(omegaVec, relPosNew)
		vel := rl.Vector3Add(baryVel, vRot)

		addBodyTex(state, fmt.Sprintf("%s #%d", typeName, i+1),
			pos, vel, 0.002, 0.35, col, false, false, TextureMoon, 1.0)
	}

	if cfg != nil {
		cfg.NotificationText = fmt.Sprintf("[SWARM DEPLOYED] %d %s asteroids placed into libration swarm!", count, typeName)
		cfg.NotificationTimer = 2.5
	}
}
