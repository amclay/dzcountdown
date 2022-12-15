package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type myTheme struct{}

var _ fyne.Theme = (*myTheme)(nil)

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DarkTheme().Color(name, variant)
}

func (m myTheme) Font(textStyle fyne.TextStyle) fyne.Resource {
	textStyle.Monospace = true
	return theme.DarkTheme().Font(textStyle)
}

func (m myTheme) Icon(themeIconName fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(themeIconName)
}
func (m myTheme) Size(themeSizeName fyne.ThemeSizeName) float32 {
	if themeSizeName == theme.SizeNameText {
		return 20
	}
	return theme.DarkTheme().Size(themeSizeName)
}
