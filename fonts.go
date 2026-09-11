package main

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FontManager stores pre-loaded fonts for clean, modern rendering
type FontManager struct {
	Initialized bool
	Regular     rl.Font
	Bold        rl.Font
}

var GlobalFonts = &FontManager{
	Initialized: false,
}

const (
	FontSizeTitle     float32 = 16.0
	FontSizeHeader    float32 = 14.5
	FontSizeButton    float32 = 13.5
	FontSizeRegular   float32 = 13.0
	FontSizeTelemetry float32 = 12.5
	FontSizeSmall     float32 = 12.0
)

// InitFontManager loads modern TTF fonts from assets/fonts
// Must be called after rl.InitWindow
func InitFontManager() {
	if GlobalFonts.Initialized {
		return
	}

	regPath := "assets/fonts/regular.ttf"
	boldPath := "assets/fonts/bold.ttf"

	if _, err := os.Stat(regPath); err == nil {
		GlobalFonts.Regular = rl.LoadFontEx(regPath, 42, nil, 0)
		rl.SetTextureFilter(GlobalFonts.Regular.Texture, rl.FilterBilinear)
	} else {
		GlobalFonts.Regular = rl.GetFontDefault()
	}

	if _, err := os.Stat(boldPath); err == nil {
		GlobalFonts.Bold = rl.LoadFontEx(boldPath, 42, nil, 0)
		rl.SetTextureFilter(GlobalFonts.Bold.Texture, rl.FilterBilinear)
	} else {
		GlobalFonts.Bold = GlobalFonts.Regular
	}

	GlobalFonts.Initialized = true
}

// UnloadFontManager releases font resources from GPU memory
func UnloadFontManager() {
	if !GlobalFonts.Initialized {
		return
	}
	defFont := rl.GetFontDefault()
	if GlobalFonts.Regular.Texture.ID != defFont.Texture.ID {
		rl.UnloadFont(GlobalFonts.Regular)
	}
	if GlobalFonts.Bold.Texture.ID != defFont.Texture.ID && GlobalFonts.Bold.Texture.ID != GlobalFonts.Regular.Texture.ID {
		rl.UnloadFont(GlobalFonts.Bold)
	}
	GlobalFonts.Initialized = false
}

// DrawTextUI renders regular text using the modern font
func DrawTextUI(text string, x, y int32, size float32, col rl.Color) {
	if !GlobalFonts.Initialized {
		rl.DrawText(text, x, y, int32(size), col)
		return
	}
	pos := rl.NewVector2(float32(x), float32(y))
	rl.DrawTextEx(GlobalFonts.Regular, text, pos, size, 1.0, col)
}

// DrawTextBoldUI renders bold text using the modern font
func DrawTextBoldUI(text string, x, y int32, size float32, col rl.Color) {
	if !GlobalFonts.Initialized {
		rl.DrawText(text, x, y, int32(size), col)
		return
	}
	pos := rl.NewVector2(float32(x), float32(y))
	rl.DrawTextEx(GlobalFonts.Bold, text, pos, size, 1.0, col)
}

// MeasureTextUI measures width of regular text
func MeasureTextUI(text string, size float32) float32 {
	if !GlobalFonts.Initialized {
		return float32(rl.MeasureText(text, int32(size)))
	}
	v := rl.MeasureTextEx(GlobalFonts.Regular, text, size, 1.0)
	return v.X
}

// MeasureTextBoldUI measures width of bold text
func MeasureTextBoldUI(text string, size float32) float32 {
	if !GlobalFonts.Initialized {
		return float32(rl.MeasureText(text, int32(size)))
	}
	v := rl.MeasureTextEx(GlobalFonts.Bold, text, size, 1.0)
	return v.X
}
