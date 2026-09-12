package main

import (
	"image"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextureManager handles procedural generation and caching of celestial surface textures and models
type TextureManager struct {
	Initialized bool
	Textures    map[BodyTextureType]rl.Texture2D
	Models      map[BodyTextureType]rl.Model
}

var GlobalTextures = &TextureManager{
	Textures: make(map[BodyTextureType]rl.Texture2D),
	Models:   make(map[BodyTextureType]rl.Model),
}

// InitTextureManager generates all procedural textures and caches 3D sphere models
func InitTextureManager() {
	if GlobalTextures.Initialized {
		return
	}

	texTypes := []BodyTextureType{
		TextureTerrestrial,
		TextureDesert,
		TextureGasGiant,
		TextureIceGiant,
		TextureMoon,
		TextureSun,
		TextureBlackHole,
		TextureRedGiant,
		TextureWhiteDwarf,
		TextureNeutronStar,
		TextureDwarfPlanet,
		TextureComet,
	}

	const w, h = 512, 256

	for _, tt := range texTypes {
		var img *image.RGBA
		switch tt {
		case TextureTerrestrial:
			img = generateTerrestrialMap(w, h)
		case TextureDesert:
			img = generateDesertMap(w, h)
		case TextureGasGiant:
			img = generateGasGiantMap(w, h)
		case TextureIceGiant:
			img = generateIceGiantMap(w, h)
		case TextureMoon:
			img = generateMoonMap(w, h)
		case TextureSun:
			img = generateSunMap(w, h)
		case TextureBlackHole:
			img = generateBlackHoleMap(w, h)
		case TextureRedGiant:
			img = generateRedGiantMap(w, h)
		case TextureWhiteDwarf:
			img = generateWhiteDwarfMap(w, h)
		case TextureNeutronStar:
			img = generateNeutronStarMap(w, h)
		case TextureDwarfPlanet:
			img = generateDwarfPlanetMap(w, h)
		case TextureComet:
			img = generateCometMap(w, h)
		}

		rlImg := rl.NewImageFromImage(img)
		tex := rl.LoadTextureFromImage(rlImg)
		rl.GenTextureMipmaps(&tex)
		rl.SetTextureFilter(tex, rl.FilterTrilinear)

		mesh := rl.GenMeshSphere(1.0, 36, 36)
		model := rl.LoadModelFromMesh(mesh)
		rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, tex)

		GlobalTextures.Textures[tt] = tex
		GlobalTextures.Models[tt] = model
	}

	GlobalTextures.Initialized = true
}

// UnloadTextureManager frees GPU textures and mesh resources
func UnloadTextureManager() {
	if !GlobalTextures.Initialized {
		return
	}
	for _, tex := range GlobalTextures.Textures {
		rl.UnloadTexture(tex)
	}
	for _, model := range GlobalTextures.Models {
		rl.UnloadModel(model)
	}
	GlobalTextures.Initialized = false
}

// Simple fractal noise helper (value noise with octaves)
func noise2D(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)

	// Smoothstep interpolation
	u := xf * xf * (3.0 - 2.0*xf)
	v := yf * yf * (3.0 - 2.0*yf)

	hash := func(i, j int) float64 {
		n := i + j*57
		n = (n << 13) ^ n
		nn := (n*(n*n*15731+789221) + 1376312589) & 0x7fffffff
		return 1.0 - float64(nn)/1073741824.0
	}

	g00 := hash(xi, yi)
	g10 := hash(xi+1, yi)
	g01 := hash(xi, yi+1)
	g11 := hash(xi+1, yi+1)

	x1 := g00*(1.0-u) + g10*u
	x2 := g01*(1.0-u) + g11*u
	return x1*(1.0-v) + x2*v
}

func fbm(x, y float64, octaves int) float64 {
	val := 0.0
	amp := 0.5
	freq := 1.0
	for i := 0; i < octaves; i++ {
		val += amp * noise2D(x*freq, y*freq)
		amp *= 0.5
		freq *= 2.0
	}
	return val
}

func generateTerrestrialMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		lat := (float64(y)/float64(h) - 0.5) * math.Pi
		cosLat := math.Cos(lat)

		for x := 0; x < w; x++ {
			lon := (float64(x)/float64(w) - 0.5) * 2.0 * math.Pi

			// 3D coordinates on unit sphere to eliminate cylindrical seam
			nx := cosLat * math.Cos(lon)
			ny := math.Sin(lat)
			nz := cosLat * math.Sin(lon)

			elevation := fbm(nx*2.2+10.0, ny*2.2+10.0, 5) + 0.5*fbm(nz*2.2+20.0, ny*2.2, 4)
			cloud := fbm(nx*4.5+30.0, ny*4.5+30.0, 4)

			var r, g, b uint8
			absLat := math.Abs(lat)

			// Polar Ice Caps
			if absLat > 1.25 {
				r, g, b = 240, 248, 255
			} else if elevation < 0.46 {
				// Deep ocean to shallow turquoise coast
				depth := elevation / 0.46
				r = uint8(10 + depth*25)
				g = uint8(45 + depth*75)
				b = uint8(120 + depth*90)
			} else if elevation < 0.50 {
				// Sandy beaches
				r, g, b = 215, 195, 140
			} else if elevation < 0.68 {
				// Grasslands and dense temperate forests
				r = uint8(35 + (elevation-0.50)*80)
				g = uint8(115 + (elevation-0.50)*70)
				b = uint8(45 + (elevation-0.50)*30)
			} else if elevation < 0.80 {
				// Mountains
				r = uint8(125 + (elevation-0.68)*140)
				g = uint8(110 + (elevation-0.68)*120)
				b = uint8(95 + (elevation-0.68)*110)
			} else {
				// Snow-capped peaks
				r, g, b = 235, 240, 250
			}

			// Swirling clouds overlay
			if cloud > 0.58 && absLat < 1.3 {
				cAlpha := (cloud - 0.58) / 0.35
				if cAlpha > 1.0 {
					cAlpha = 1.0
				}
				r = uint8(float64(r)*(1.0-cAlpha*0.8) + 250.0*cAlpha*0.8)
				g = uint8(float64(g)*(1.0-cAlpha*0.8) + 250.0*cAlpha*0.8)
				b = uint8(float64(b)*(1.0-cAlpha*0.8) + 255.0*cAlpha*0.8)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateDesertMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		lat := (float64(y)/float64(h) - 0.5) * math.Pi
		cosLat := math.Cos(lat)

		for x := 0; x < w; x++ {
			lon := (float64(x)/float64(w) - 0.5) * 2.0 * math.Pi
			nx := cosLat * math.Cos(lon)
			ny := math.Sin(lat)
			nz := cosLat * math.Sin(lon)

			n := fbm(nx*3.0+15.0, ny*3.0+15.0, 5) + 0.3*fbm(nz*6.0, ny*6.0, 3)

			var r, g, b uint8
			if math.Abs(lat) > 1.35 {
				// Dry ice polar cap
				r, g, b = 245, 245, 250
			} else {
				// Martian ochre, canyon rust, and dark volcanic basalt
				r = uint8(160 + n*85)
				g = uint8(60 + n*60)
				b = uint8(30 + n*40)
			}
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateGasGiantMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := float64(y) / float64(h)
		bandFreq := yNorm * 22.0

		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			// Horizontal shear turbulence
			turb := fbm(xNorm*8.0, yNorm*12.0, 4) * 0.15
			v := math.Sin(bandFreq*math.Pi + turb*6.0)

			// Great Red Spot (southern mid-latitude vortex around x=0.65, y=0.68)
			dx := (xNorm - 0.65) * 6.0
			dy := (yNorm - 0.68) * 14.0
			spotDist := dx*dx + dy*dy

			var r, g, b uint8
			if spotDist < 1.0 {
				spotAlpha := 1.0 - spotDist
				r = uint8(210 + spotAlpha*45)
				g = uint8(75 + spotAlpha*20)
				b = uint8(45 + spotAlpha*20)
			} else {
				// Cream, amber, rust, and tan cloud bands
				blend := (v + 1.0) * 0.5
				r = uint8(190 + blend*55)
				g = uint8(140 + blend*65)
				b = uint8(95 + blend*75)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateIceGiantMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			band := math.Sin(yNorm*16.0*math.Pi + fbm(xNorm*5.0, yNorm*6.0, 3)*3.0)
			blend := (band + 1.0) * 0.5

			// Vivid cyan, azure, and methane blue streaks
			var r, g, b uint8
			r = uint8(45 + blend*40)
			g = uint8(130 + blend*70)
			b = uint8(200 + blend*55)

			// Subtle high-altitude white methane clouds
			if fbm(xNorm*10.0, yNorm*15.0, 3) > 0.65 {
				r = uint8(math.Min(255, float64(r)+80))
				g = uint8(math.Min(255, float64(g)+80))
				b = uint8(math.Min(255, float64(b)+80))
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateMoonMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		lat := (float64(y)/float64(h) - 0.5) * math.Pi
		cosLat := math.Cos(lat)

		for x := 0; x < w; x++ {
			lon := (float64(x)/float64(w) - 0.5) * 2.0 * math.Pi
			nx := cosLat * math.Cos(lon)
			ny := math.Sin(lat)
			nz := cosLat * math.Sin(lon)

			// Cratered cellular basalt texture
			craterNoise := fbm(nx*5.0+5.0, ny*5.0+nz*3.0, 5)
			craters := math.Abs(math.Sin(craterNoise * 14.0))

			base := 95.0 + craterNoise*90.0 + craters*35.0
			if base > 255 {
				base = 255
			}
			val := uint8(base)

			img.Set(x, y, color.RGBA{R: val, G: val, B: uint8(float64(val) * 0.95), A: 255})
		}
	}
	return img
}

func generateSunMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			// Solar convective granulation cells
			cells := fbm(xNorm*18.0, yNorm*18.0, 4)
			flares := fbm(xNorm*6.0+10.0, yNorm*6.0+10.0, 3)

			r := uint8(math.Min(255, 230+cells*25))
			g := uint8(math.Min(255, 140+cells*80+flares*30))
			b := uint8(math.Min(255, 20+cells*50))

			// Sunspot cooler regions
			if cells < 0.28 {
				r = 170
				g = 60
				b = 10
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateBlackHoleMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := (float64(y)/float64(h) - 0.5) * 2.0
		for x := 0; x < w; x++ {
			xNorm := (float64(x)/float64(w) - 0.5) * 2.0

			dist := math.Sqrt(xNorm*xNorm + yNorm*yNorm)

			var r, g, b uint8
			if dist < 0.5 {
				// Pure event horizon
				r, g, b = 5, 2, 8
			} else if dist < 0.65 {
				// Relativistic photon sphere ring
				photon := (dist - 0.5) / 0.15
				r = uint8(240 + photon*15)
				g = uint8(190 + photon*50)
				b = uint8(255)
			} else {
				// Accretion disk gradient
				accretion := (1.0 - (dist-0.65)/0.35)
				if accretion < 0 {
					accretion = 0
				}
				r = uint8(180 * accretion)
				g = uint8(90 * accretion)
				b = uint8(220 * accretion)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateRedGiantMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			// Massive turbulent convection cells
			cells := fbm(xNorm*12.0, yNorm*12.0, 4)
			prominences := fbm(xNorm*4.0+5.0, yNorm*4.0+5.0, 3)

			r := uint8(math.Min(255, 190+cells*65))
			g := uint8(math.Min(255, 30+cells*90+prominences*40))
			b := uint8(math.Min(255, 10+cells*30))

			if cells < 0.22 {
				// Cool convection downdraft zones
				r = 140
				g = 15
				b = 5
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateWhiteDwarfMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		yNorm := float64(y) / float64(h)
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			stripes := math.Sin(yNorm*40.0 + fbm(xNorm*8.0, yNorm*8.0, 3)*4.0)
			glow := fbm(xNorm*14.0, yNorm*14.0, 3)

			r := uint8(math.Min(255, 220+glow*35))
			g := uint8(math.Min(255, 235+glow*20))
			b := uint8(255)

			if stripes > 0.7 {
				r = uint8(math.Max(180, float64(r)-25))
				g = uint8(math.Max(210, float64(g)-15))
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateNeutronStarMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		lat := (float64(y)/float64(h) - 0.5) * math.Pi
		absLat := math.Abs(lat)
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)

			magField := math.Sin(lat*8.0 + math.Sin(xNorm*16.0*math.Pi)*0.5)
			crust := fbm(xNorm*20.0, lat*6.0, 4)

			var r, g, b uint8
			if absLat > 1.35 {
				// Intense magnetic polar hotspot
				r = 255
				g = 255
				b = 255
			} else {
				// Dense relativistic crust with glowing magnetic flux ropes
				r = uint8(60 + crust*70)
				g = uint8(120 + crust*90 + (magField+1.0)*20)
				b = uint8(220 + crust*35)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateDwarfPlanetMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		lat := (float64(y)/float64(h) - 0.5) * math.Pi
		cosLat := math.Cos(lat)
		for x := 0; x < w; x++ {
			lon := (float64(x)/float64(w) - 0.5) * 2.0 * math.Pi
			nx := cosLat * math.Cos(lon)
			ny := math.Sin(lat)
			nz := cosLat * math.Sin(lon)

			elevation := fbm(nx*3.0+15.0, ny*3.0+15.0, 5)
			tholin := fbm(nz*2.0+5.0, ny*2.0+5.0, 3)

			var r, g, b uint8
			if elevation > 0.60 {
				// Nitrogen / methane bright ice sheets (Sputnik Planitia)
				r, g, b = 235, 230, 220
			} else if tholin > 0.52 {
				// Reddish-brown organic tholin macula
				r, g, b = 175, 75, 45
			} else {
				// Cratered dark bedrock
				r, g, b = 130, 115, 105
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func generateCometMap(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			xNorm := float64(x) / float64(w)
			yNorm := float64(y) / float64(h)

			rough := fbm(xNorm*25.0, yNorm*25.0, 4)
			vents := fbm(xNorm*8.0+10.0, yNorm*8.0+10.0, 3)

			var r, g, b uint8
			if vents > 0.65 {
				// Active outgassing ice vent
				r, g, b = 220, 245, 255
			} else {
				// Very dark carbonaceous chondrite regolith
				val := uint8(45 + rough*45)
				r, g, b = val, val+5, val+10
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// DrawTexturedCelestialSphere draws a sphere mapped with procedural celestial textures and axial rotation
func DrawTexturedCelestialSphere(b *Body, isSelected bool) {
	if !GlobalTextures.Initialized {
		rl.DrawSphere(b.Position, b.Radius, b.Color)
		return
	}

	model, exists := GlobalTextures.Models[b.TextureType]
	if !exists || b.TextureType == TextureNone {
		rl.DrawSphere(b.Position, b.Radius, b.Color)
		return
	}

	// Axial spin around Y axis
	rotAxis := rl.NewVector3(0, 1, 0)
	scale := rl.NewVector3(b.Radius, b.Radius, b.Radius)

	// Draw textured 3D model
	rl.DrawModelEx(model, b.Position, rotAxis, b.RotationAngle, scale, rl.White)

	// Atmospheric limb glow and coronal auras
	switch b.TextureType {
	case TextureTerrestrial:
		rl.DrawSphereWires(b.Position, b.Radius*1.04, 8, 8, rl.NewColor(90, 160, 255, 60))
	case TextureGasGiant:
		rl.DrawSphereWires(b.Position, b.Radius*1.03, 8, 8, rl.NewColor(240, 180, 110, 50))
	case TextureRedGiant:
		rl.DrawSphere(b.Position, b.Radius*1.25, rl.Fade(rl.Red, 0.22))
		rl.DrawSphere(b.Position, b.Radius*1.55, rl.Fade(rl.Orange, 0.12))
	case TextureWhiteDwarf:
		rl.DrawSphere(b.Position, b.Radius*1.40, rl.Fade(rl.White, 0.35))
		rl.DrawSphere(b.Position, b.Radius*1.85, rl.Fade(rl.SkyBlue, 0.18))
	case TextureNeutronStar:
		rl.DrawSphere(b.Position, b.Radius*1.45, rl.Fade(rl.Purple, 0.30))
		rl.DrawSphereWires(b.Position, b.Radius*1.70, 8, 8, rl.NewColor(140, 100, 255, 90))
	}

	// Draw planetary dust ring if defined
	if b.RingInnerRadius > 0 && b.RingOuterRadius > b.RingInnerRadius {
		ringSegments := 48
		ringColor := rl.NewColor(210, 195, 170, 120)
		for s := 0; s < ringSegments; s++ {
			a1 := float64(s) * (2.0 * math.Pi / float64(ringSegments))
			a2 := float64(s+1) * (2.0 * math.Pi / float64(ringSegments))

			p1Inner := rl.NewVector3(b.Position.X+b.RingInnerRadius*float32(math.Cos(a1)), b.Position.Y, b.Position.Z+b.RingInnerRadius*float32(math.Sin(a1)))
			p1Outer := rl.NewVector3(b.Position.X+b.RingOuterRadius*float32(math.Cos(a1)), b.Position.Y, b.Position.Z+b.RingOuterRadius*float32(math.Sin(a1)))
			p2Inner := rl.NewVector3(b.Position.X+b.RingInnerRadius*float32(math.Cos(a2)), b.Position.Y, b.Position.Z+b.RingInnerRadius*float32(math.Sin(a2)))
			p2Outer := rl.NewVector3(b.Position.X+b.RingOuterRadius*float32(math.Cos(a2)), b.Position.Y, b.Position.Z+b.RingOuterRadius*float32(math.Sin(a2)))

			rl.DrawLine3D(p1Inner, p1Outer, ringColor)
			rl.DrawLine3D(p1Inner, p2Inner, ringColor)
			rl.DrawLine3D(p1Outer, p2Outer, ringColor)
		}
	}
}
