package main

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type landmarkRow struct {
	landmarkName    string
	buttonContainer *fyne.Container
	textContainer   *fyne.Container
}

func newLandmarkRow(landmarkName string) *landmarkRow {
	unknownText := canvas.NewText("unknown", white)
	unknownText.TextSize = 30
	unknownText.Alignment = fyne.TextAlignLeading
	tappedFunc := func() {
		moveToTop(landmarkName)
		updateButtonColor(landmarkName, orange)
		stopTimer(landmarkName)
		time.Sleep(1 * time.Second)
		startTimer(landmarkName, timerCallback, timerSeconds)
	}

	buttonContainer := container.NewMax(canvas.NewRectangle(blue), widget.NewButton(landmarkName, tappedFunc))
	textContainer := container.NewMax(canvas.NewRectangle(black), unknownText)

	return &landmarkRow{
		landmarkName:    landmarkName,
		buttonContainer: buttonContainer,
		textContainer:   textContainer,
	}
}

func (row *landmarkRow) toContainer() *fyne.Container {
	underlyingContainer := container.New(layout.NewHBoxLayout())
	underlyingContainer.Add(row.buttonContainer)
	underlyingContainer.Add(row.textContainer)
	return underlyingContainer
}

func (row *landmarkRow) updateText(s string) {
	row.textContainer.Objects[1].(*canvas.Text).Text = fmt.Sprintf("%-7s", s)
}

func (row *landmarkRow) updateButtonColor(c color.Color) {
	row.buttonContainer.Objects[0].(*canvas.Rectangle).FillColor = c
}
