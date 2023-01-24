package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

type landmarkRow struct {
	landmarkName    string
	buttonContainer *fyne.Container
	textContainer   *fyne.Container
}

func timerLoop() {
	// timer that ticks every 1 second
	for {
		time.Sleep(100 * time.Millisecond)
		for _, landmarkName := range dzLandmarksMap[globalRegion] {
			// get time until timer expires
			stopTime, ok := timers.Load(landmarkName)
			if !ok {
				continue
			}
			durationLeft := time.Until(stopTime.(time.Time))
			minutes := int(durationLeft.Minutes())
			seconds := int(durationLeft.Seconds()) - (minutes * 60)
			if minutes <= 0 && seconds <= 0 {
				log.Println("no time remaining for", landmarkName)
				timers.Delete(landmarkName)
				updateText(landmarkName, fmt.Sprintf("%-7s", "00:00"))
				go tts(landmarkName)
				updateButtonColor(landmarkName, green)
				continue
			}
			updateText(landmarkName, fmt.Sprintf("%02d:%02d", minutes, seconds))
		}
	}
}

func newLandmarkRow(landmarkName string) *landmarkRow {
	unknownText := canvas.NewText("unknown", white)
	unknownText.TextSize = 30
	unknownText.Alignment = fyne.TextAlignLeading
	tappedFunc := func() {
		moveToTop(landmarkName)
		updateButtonColor(landmarkName, orange)
		timers.Delete(landmarkName)
		time.Sleep(100 * time.Millisecond)
		startTimer(landmarkName, timerSeconds)
	}

	buttonObject := newExtendedButton(landmarkName, tappedFunc)

	buttonContainer := container.NewMax(canvas.NewRectangle(blue), buttonObject)
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
