//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
)

// LagrangePoint holds calculated position, velocity, and properties of an equilibrium point
type LagrangePoint struct {
	Name      string  `json:"name"`
	Index     int     `json:"index"` // 1..5
	Position  Vector3 `json:"pos"`
	Velocity  Vector3 `json:"vel"`
	Potential float64 `json:"potential"`
	Stable    bool    `json:"stable"`
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

	rVec := Vector3Subtract(secondary.Position, primary.Position)
	R := float64(Vector3Length(rVec))
	if R < 1e-4 {
		return points
	}

	uHat := Vector3Scale(rVec, float32(1.0/R))

	// Determine orbital plane normal vector
	vRel := Vector3Subtract(secondary.Velocity, primary.Velocity)
	hVec := Vector3CrossProduct(rVec, vRel)
	hLen := float64(Vector3Length(hVec))
	var nHat Vector3
	if hLen > 1e-4 {
		nHat = Vector3Scale(hVec, float32(1.0/hLen))
	} else {
		// Default to Y-up normal for planar X-Z simulation
		nHat = Vector3{0, 1, 0}
	}

	// In-plane perpendicular vector (ahead in orbital direction)
	wHat := Vector3Normalize(Vector3CrossProduct(nHat, uHat))
	// Ensure wHat aligns with secondary velocity direction
	if Vector3DotProduct(wHat, vRel) < 0 {
		wHat = Vector3Scale(wHat, -1.0)
	}

	// Barycenter position and velocity
	baryPos := Vector3{
		X: float32((m1*float64(primary.Position.X) + m2*float64(secondary.Position.X)) / totalM),
		Y: float32((m1*float64(primary.Position.Y) + m2*float64(secondary.Position.Y)) / totalM),
		Z: float32((m1*float64(primary.Position.Z) + m2*float64(secondary.Position.Z)) / totalM),
	}
	baryVel := Vector3{
		X: float32((m1*float64(primary.Velocity.X) + m2*float64(secondary.Velocity.X)) / totalM),
		Y: float32((m1*float64(primary.Velocity.Y) + m2*float64(secondary.Velocity.Y)) / totalM),
		Z: float32((m1*float64(primary.Velocity.Z) + m2*float64(secondary.Velocity.Z)) / totalM),
	}

	// Mean motion / angular velocity
	omega := math.Sqrt((g * totalM) / (R * R * R))
	omegaVec := Vector3Scale(nHat, float32(omega))

	// Helper to compute inertial velocity from rotating frame offset
	computeVel := func(pos Vector3) Vector3 {
		relPos := Vector3Subtract(pos, baryPos)
		vRot := Vector3CrossProduct(omegaVec, relPos)
		return Vector3Add(baryVel, vRot)
	}

	// Effective potential in rotating frame along primary-secondary axis x (origin at barycenter)
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

			f := -(1.0-mu)*R*R*R/(d1*absD1) - mu*R*R*R/(d2*absD2) + x
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

	// 1. L1: Collinear between Primary and Secondary
	hillRadius := R * math.Pow(mu/3.0, 1.0/3.0)
	initL1 := x2 - hillRadius
	xL1 := evalCollinearRoot(initL1, x1+1e-4, x2-1e-4)
	posL1 := Vector3Add(baryPos, Vector3Scale(uHat, float32(xL1)))
	points[0] = LagrangePoint{
		Name:      "L1 (Inter-Orbital / SOHO)",
		Index:     1,
		Position:  posL1,
		Velocity:  computeVel(posL1),
		Potential: ComputeEffectivePotential(posL1, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 2. L2: Collinear beyond Secondary
	initL2 := x2 + hillRadius
	xL2 := evalCollinearRoot(initL2, x2+1e-4, x2+R*3.0)
	posL2 := Vector3Add(baryPos, Vector3Scale(uHat, float32(xL2)))
	points[1] = LagrangePoint{
		Name:      "L2 (Outer / JWST Halo)",
		Index:     2,
		Position:  posL2,
		Velocity:  computeVel(posL2),
		Potential: ComputeEffectivePotential(posL2, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 3. L3: Collinear behind Primary
	initL3 := x1 - R*(1.0-5.0*mu/12.0)
	xL3 := evalCollinearRoot(initL3, x1-R*3.0, x1-1e-4)
	posL3 := Vector3Add(baryPos, Vector3Scale(uHat, float32(xL3)))
	points[2] = LagrangePoint{
		Name:      "L3 (Counter-Orbital)",
		Index:     3,
		Position:  posL3,
		Velocity:  computeVel(posL3),
		Potential: ComputeEffectivePotential(posL3, primary, secondary, baryPos, g, omega),
		Stable:    false,
	}

	// 4. L4: Equilateral triangle 60 degrees ahead in orbit
	posL4 := Vector3Add(baryPos, Vector3Add(
		Vector3Scale(uHat, float32((0.5-mu)*R)),
		Vector3Scale(wHat, float32(math.Sqrt(3.0)/2.0*R)),
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
	posL5 := Vector3Add(baryPos, Vector3Add(
		Vector3Scale(uHat, float32((0.5-mu)*R)),
		Vector3Scale(wHat, float32(-math.Sqrt(3.0)/2.0*R)),
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

// ComputeEffectivePotential evaluates Jacobi effective potential in rotating frame
func ComputeEffectivePotential(pos Vector3, b1, b2 *Body, baryPos Vector3, g, omega float64) float64 {
	d1 := float64(Vector3Distance(pos, b1.Position))
	d2 := float64(Vector3Distance(pos, b2.Position))
	rBary := float64(Vector3Distance(pos, baryPos))

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

// GetLagrangePair determines the two primary bodies for Lagrange calculations
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
			primary := heaviest
			for _, b := range state.Bodies {
				if b.ID != selected.ID && b.Mass > selected.Mass*5.0 {
					dist := Vector3Distance(b.Position, selected.Position)
					primDist := Vector3Distance(primary.Position, selected.Position)
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

// jsGetLagrangePoints exports Lagrange points to JS
func jsGetLagrangePoints(this js.Value, args []js.Value) any {
	primary, secondary := GetLagrangePair(simState)
	if primary == nil || secondary == nil {
		return "{}"
	}

	pts := ComputeLagrangePoints(primary, secondary, simConfig.G)
	data, err := json.Marshal(map[string]any{
		"primary": map[string]any{
			"id":   primary.ID,
			"name": primary.Name,
			"pos":  primary.Position,
			"mass": primary.Mass,
		},
		"secondary": map[string]any{
			"id":   secondary.ID,
			"name": secondary.Name,
			"pos":  secondary.Position,
			"mass": secondary.Mass,
		},
		"points": pts,
	})
	if err != nil {
		return "{}"
	}
	return string(data)
}

// jsSpawnLagrangeProbe spawns a satellite directly into equilibrium orbit at L1..L5
func jsSpawnLagrangeProbe(this js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeNumber {
		return -1
	}
	lIndex := args[0].Int() // 1..5
	if lIndex < 1 || lIndex > 5 {
		return -1
	}

	primary, secondary := GetLagrangePair(simState)
	if primary == nil || secondary == nil {
		return -1
	}

	pts := ComputeLagrangePoints(primary, secondary, simConfig.G)
	pt := pts[lIndex-1]

	color := Color{100, 220, 255, 255}
	switch lIndex {
	case 1:
		color = Color{255, 220, 50, 255}
	case 2:
		color = Color{80, 220, 255, 255}
	case 3:
		color = Color{255, 120, 80, 255}
	case 4:
		color = Color{120, 255, 180, 255}
	case 5:
		color = Color{255, 180, 120, 255}
	}

	name := fmt.Sprintf("%s Probe", pt.Name)
	b := addBody(simState, name, pt.Position, pt.Velocity, 0.001, 0.45, color, false, false, 1)
	return b.ID
}
