package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Render3DScene renders all 3D objects within the camera perspective
func Render3DScene(state *SimState, cfg *Config, camera *OrbitCamera) {
	rl.BeginMode3D(camera.Camera)

	// 1. Draw Starfield
	if cfg.ShowStarfield {
		for i, starPos := range state.Stars {
			// Offset by camera target so stars stay at infinity
			pos := rl.Vector3Add(starPos, camera.Target)
			rl.DrawCube(pos, state.StarSizes[i], state.StarSizes[i], state.StarSizes[i], state.StarColors[i])
		}
	}

	// 2. Draw Reference Grid on XZ plane
	if cfg.ShowGrid {
		rl.DrawGrid(60, 4.0)
	}

	// 3. Draw Orbit Trails
	if cfg.ShowTrails {
		for _, b := range state.Bodies {
			trailLen := len(b.Trail)
			if trailLen < 2 {
				continue
			}

			col := b.Color
			for i := 0; i < trailLen-1; i++ {
				// Fade alpha towards the tail
				alpha := float32(i) / float32(trailLen)
				trailCol := rl.Fade(col, alpha*0.85)
				rl.DrawLine3D(b.Trail[i], b.Trail[i+1], trailCol)
			}
			// Connect last trail point to current position
			if trailLen > 0 {
				rl.DrawLine3D(b.Trail[trailLen-1], b.Position, col)
			}
		}
	}

	// 4. Draw Celestial Bodies
	for _, b := range state.Bodies {
		isSelected := (b.ID == state.SelectedBodyID)

		// Celestial body sphere
		rl.DrawSphere(b.Position, b.Radius, b.Color)

		// Glowing star aura
		if b.IsStar {
			rl.DrawSphere(b.Position, b.Radius*1.35, rl.Fade(b.Color, 0.28))
			rl.DrawSphere(b.Position, b.Radius*1.85, rl.Fade(b.Color, 0.12))
		}

		// Stationary indicator (lock ring)
		if b.IsStationary {
			rl.DrawSphereWires(b.Position, b.Radius*1.15, 6, 8, rl.NewColor(255, 60, 60, 200))
		}

		// Highlight selected body
		if isSelected {
			rl.DrawSphereWires(b.Position, b.Radius*1.35, 14, 14, rl.Yellow)

			// Drop altitude projection line to XZ plane (Y=0)
			dropPoint := rl.NewVector3(b.Position.X, 0, b.Position.Z)
			rl.DrawLine3D(b.Position, dropPoint, rl.Fade(rl.Yellow, 0.5))
			rl.DrawCircle3D(dropPoint, b.Radius*0.8, rl.NewVector3(1, 0, 0), 90, rl.Fade(rl.Yellow, 0.3))
		}

		// Velocity vector (Green)
		if cfg.ShowVectors && !b.IsStationary {
			vLen := rl.Vector3Length(b.Velocity)
			if vLen > 0.01 {
				// Scale for visual clarity
				vEnd := rl.Vector3Add(b.Position, rl.Vector3Scale(b.Velocity, 1.2))
				rl.DrawLine3D(b.Position, vEnd, rl.Lime)
			}
		}

		// Gravitational Force vector (Orange/Red)
		if cfg.ShowForces && !b.IsStationary {
			fLen := rl.Vector3Length(b.NetForce)
			if fLen > 0.001 {
				scale := float32(2.5) / float32(math.Sqrt(float64(fLen))+1.0)
				fEnd := rl.Vector3Add(b.Position, rl.Vector3Scale(b.NetForce, scale))
				rl.DrawLine3D(b.Position, fEnd, rl.Orange)
			}
		}
	}

	// 5. Draw interactive slingshot spawn vector
	if state.IsDraggingSpawn {
		rl.DrawSphere(state.SpawnWorldPos, 1.2, rl.Fade(rl.SkyBlue, 0.8))
		endPos := rl.Vector3Add(state.SpawnWorldPos, state.DragVelocity)
		rl.DrawLine3D(state.SpawnWorldPos, endPos, rl.Yellow)
		rl.DrawSphere(endPos, 0.4, rl.Gold)
	}

	// 6. Draw interactive force vector on selected body
	if state.IsDraggingForce && state.SelectedBodyID != -1 {
		var sel *Body
		for _, b := range state.Bodies {
			if b.ID == state.SelectedBodyID {
				sel = b
				break
			}
		}
		if sel != nil {
			endPos := rl.Vector3Add(sel.Position, state.ForceVector)
			rl.DrawLine3D(sel.Position, endPos, rl.Red)
			rl.DrawSphere(endPos, 0.5, rl.Orange)
		}
	}

	rl.EndMode3D()

	// 7. Draw floating 2D labels in screen space
	if cfg.ShowLabels {
		drawScreenLabels(state, camera)
	}
}

// drawScreenLabels draws nameplates and telemetry above celestial bodies
func drawScreenLabels(state *SimState, camera *OrbitCamera) {
	for _, b := range state.Bodies {
		screenPos := rl.GetWorldToScreen(b.Position, camera.Camera)

		// Don't draw if behind camera
		camToBody := rl.Vector3Subtract(b.Position, camera.Camera.Position)
		camForward := rl.Vector3Subtract(camera.Camera.Target, camera.Camera.Position)
		if rl.Vector3DotProduct(camToBody, camForward) <= 0 {
			continue
		}

		isSelected := (b.ID == state.SelectedBodyID)
		textX := int32(screenPos.X) + 12
		textY := int32(screenPos.Y) - 10

		label := b.Name
		if b.IsStationary {
			label += " [FIXED]"
		}

		textColor := rl.White
		if isSelected {
			textColor = rl.Yellow
			rl.DrawRectangle(textX-4, textY-3, rl.MeasureText(label, 12)+8, 18, rl.NewColor(0, 0, 0, 180))
			rl.DrawRectangleLines(textX-4, textY-3, rl.MeasureText(label, 12)+8, 18, rl.Yellow)
		} else {
			rl.DrawRectangle(textX-2, textY-2, rl.MeasureText(label, 11)+4, 15, rl.NewColor(0, 0, 0, 130))
		}

		rl.DrawText(label, textX, textY, 11, textColor)

		// Additional telemetry for selected
		if isSelected {
			speed := rl.Vector3Length(b.Velocity)
			info := fmt.Sprintf("M:%.1f | V:%.2f", b.Mass, speed)
			rl.DrawText(info, textX, textY+15, 10, rl.SkyBlue)
		}
	}
}
