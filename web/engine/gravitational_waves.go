//go:build js && wasm

package main

import (
	"fmt"
	"math"
)

// GravitationalWaveBurst represents a transient gravitational wave event (merger ringdown)
type GravitationalWaveBurst struct {
	Position   Vector3 `json:"pos"`
	StartTime  float64 `json:"start_time"`
	PeakStrain float32 `json:"peak_strain"`
	Frequency  float32 `json:"frequency"`
	DecayTime  float32 `json:"decay_time"`
	WaveSpeed  float32 `json:"wave_speed"`
	Source     string  `json:"source"`
}

// UpdateGravitationalWaves computes continuous quadrupole radiation from orbiting bodies
// and advances active merger wave bursts
func UpdateGravitationalWaves(state *SimState, cfg *Config, dt float64) {
	if len(state.Bodies) == 0 {
		return
	}

	curTime := state.Time
	c := cfg.SpeedOfLight
	if c <= 0 {
		c = 150.0
	}

	// 1. Identify primary binary or most significant accelerating mass pair
	var primary, secondary *Body
	if len(state.Bodies) >= 2 {
		for _, b := range state.Bodies {
			if primary == nil || b.Mass > primary.Mass {
				secondary = primary
				primary = b
			} else if secondary == nil || b.Mass > secondary.Mass {
				secondary = b
			}
		}
	}

	var curStrain float32 = 0.0
	var curFreq float32 = 0.0
	var curPower float64 = 0.0

	if primary != nil && secondary != nil && secondary.Mass > 0.0005 {
		dx := float64(secondary.Position.X - primary.Position.X)
		dy := float64(secondary.Position.Y - primary.Position.Y)
		dz := float64(secondary.Position.Z - primary.Position.Z)
		sepSq := dx*dx + dy*dy + dz*dz
		sep := math.Sqrt(sepSq)
		if sep < 0.2 {
			sep = 0.2
		}

		m1 := primary.Mass
		m2 := secondary.Mass
		mTot := m1 + m2
		mu := (m1 * m2) / mTot

		omegaOrb := math.Sqrt((cfg.G * mTot) / (sep * sep * sep))
		omegaGW := 2.0 * omegaOrb
		curFreq = float32(omegaGW / (2.0 * math.Pi))

		c5 := math.Pow(c, 5)
		g4 := math.Pow(cfg.G, 4)
		if c5 > 1e-6 {
			curPower = (32.0 / 5.0) * (g4 / c5) * (mu * mu * mTot * mTot * mTot) / math.Pow(sep, 5)
		}

		const rDet = 120.0
		c4 := math.Pow(c, 4)
		if c4 > 1e-6 {
			h0 := (4.0 * cfg.G * cfg.G * mu * mTot) / (c4 * sep * rDet)
			normH := float32(h0 * 1200.0)
			if normH > 0.25 {
				normH = 0.25
			}
			curStrain = normH * float32(math.Cos(omegaGW*curTime))
		}
	}

	// 2. Add contributions from active merger ringdown bursts
	activeBursts := make([]GravitationalWaveBurst, 0, len(state.GWBursts))
	for _, b := range state.GWBursts {
		age := float32(curTime - b.StartTime)
		if age < b.DecayTime {
			activeBursts = append(activeBursts, b)
			burstAmp := b.PeakStrain * float32(math.Exp(float64(-age*1.4))) * float32(math.Cos(float64(b.Frequency*age)))
			curStrain += burstAmp
			if math.Abs(float64(b.Frequency)) > float64(curFreq) {
				curFreq = b.Frequency / (2.0 * float32(math.Pi))
			}
		}
	}
	state.GWBursts = activeBursts

	state.LastGWStrain = curStrain
	state.LastGWFreq = curFreq
	state.LastGWPower = curPower

	if len(state.GWWaveformHistory) > 0 {
		state.GWWaveformHistory[state.GWWaveformHead] = curStrain
		state.GWWaveformHead = (state.GWWaveformHead + 1) % len(state.GWWaveformHistory)
	}
}

// EmitGravitationalWaveBurst creates a ringdown burst event when bodies merge
func EmitGravitationalWaveBurst(state *SimState, cfg *Config, primary, secondary *Body) {
	curTime := state.Time
	c := cfg.SpeedOfLight
	if c <= 0 {
		c = 150.0
	}

	mTot := primary.Mass + secondary.Mass
	mu := (primary.Mass * secondary.Mass) / mTot

	compactBoost := float32(1.0)
	if primary.IsBlackHole || secondary.IsBlackHole {
		compactBoost = 2.8
	} else if primary.IsPulsar || secondary.IsPulsar {
		compactBoost = 2.0
	}

	peakStrain := float32(math.Min(0.35, math.Max(0.015, math.Log10(mu+1.0)*0.08))) * compactBoost
	rCol := float64(primary.Radius + secondary.Radius)
	if rCol < 0.5 {
		rCol = 0.5
	}
	ringFreq := float32(math.Sqrt((cfg.G*mTot)/(rCol*rCol*rCol))) * 2.8

	burst := GravitationalWaveBurst{
		Position:   primary.Position,
		StartTime:  curTime,
		PeakStrain: peakStrain,
		Frequency:  ringFreq,
		DecayTime:  7.5,
		WaveSpeed:  float32(math.Max(50.0, c*0.65)),
		Source:     fmt.Sprintf("%s + %s Merger", primary.Name, secondary.Name),
	}
	state.GWBursts = append(state.GWBursts, burst)
}

// ComputeGravitationalWaveDisplacement calculates wave metric ripple offset and intensity at (gx, gz)
func ComputeGravitationalWaveDisplacement(state *SimState, cfg *Config, gx, gz float32, curTime float64) (float32, float32) {
	c := float32(cfg.SpeedOfLight)
	if c <= 0 {
		c = 150.0
	}

	var totalDisp float32 = 0.0
	var strainIntensity float32 = 0.0

	// 1. Continuous Rotating Einstein Quadrupole Spiral from primary binary
	if len(state.Bodies) >= 2 {
		var p1, p2 *Body
		for _, b := range state.Bodies {
			if p1 == nil || b.Mass > p1.Mass {
				p2 = p1
				p1 = b
			} else if p2 == nil || b.Mass > p2.Mass {
				p2 = b
			}
		}

		if p1 != nil && p2 != nil && p2.Mass > 0.0005 {
			baryX := float32((p1.Position.X*float32(p1.Mass) + p2.Position.X*float32(p2.Mass)) / float32(p1.Mass+p2.Mass))
			baryZ := float32((p1.Position.Z*float32(p1.Mass) + p2.Position.Z*float32(p2.Mass)) / float32(p1.Mass+p2.Mass))

			dx := gx - baryX
			dz := gz - baryZ
			r := float32(math.Sqrt(float64(dx*dx + dz*dz)))
			theta := float32(math.Atan2(float64(dz), float64(dx)))

			sep := float32(math.Hypot(float64(p2.Position.X-p1.Position.X), float64(p2.Position.Z-p1.Position.Z)))
			if sep < 0.2 {
				sep = 0.2
			}
			mTot := p1.Mass + p2.Mass
			mu := (p1.Mass * p2.Mass) / mTot

			omegaOrb := float32(math.Sqrt((cfg.G * mTot) / float64(sep*sep*sep)))
			omegaGW := 2.0 * omegaOrb

			tRet := float32(curTime) - (r / c)
			phase := 2.0*theta - omegaGW*tRet

			falloff := float32(sep*3.5) / (r + float32(sep*3.5))
			quadAmp := float32(math.Min(2.5, math.Log10(mu+1.0)*0.85)) * falloff
			ripple := quadAmp * float32(math.Cos(float64(phase)))

			totalDisp += ripple
			strainIntensity += float32(math.Abs(float64(ripple)))
		}
	}

	// 2. Propagating Wavefronts from active Merger Bursts
	for _, b := range state.GWBursts {
		age := float32(curTime - b.StartTime)
		if age <= 0 || age > b.DecayTime {
			continue
		}

		dx := gx - b.Position.X
		dz := gz - b.Position.Z
		dist := float32(math.Sqrt(float64(dx*dx + dz*dz)))

		waveFrontR := b.WaveSpeed * age
		dFront := dist - waveFrontR

		const packetWidth = 9.0
		spatialEnvelope := float32(math.Exp(float64(-(dFront * dFront) / (2.0 * packetWidth * packetWidth))))
		temporalDecay := float32(math.Exp(float64(-age * 0.8)))
		geomFalloff := float32(18.0) / float32(math.Sqrt(float64(dist+12.0)))

		wavenumber := b.Frequency / (b.WaveSpeed * 0.25)
		if wavenumber < 0.2 {
			wavenumber = 0.2
		} else if wavenumber > 2.0 {
			wavenumber = 2.0
		}

		oscillation := float32(math.Cos(float64(wavenumber * dFront)))
		burstDisp := b.PeakStrain * 14.0 * spatialEnvelope * temporalDecay * geomFalloff * oscillation

		totalDisp += burstDisp
		strainIntensity += float32(math.Abs(float64(burstDisp)))
	}

	return totalDisp, strainIntensity
}
