package main

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	voices "github.com/hegedustibor/htgo-tts/voices"
)

var (
	timers *sync.Map

	fyneApp = app.New()
	window  = fyneApp.NewWindow("DZ Countdown Timers")
	grid    = container.New(layout.NewGridLayout(2))

	red    = color.RGBA{255, 0, 0, 255}
	green  = color.RGBA{0, 100, 0, 255}
	blue   = color.RGBA{19, 13, 84, 255}
	black  = color.RGBA{0, 0, 0, 255}
	purple = color.RGBA{128, 0, 128, 255}
	orange = color.RGBA{128, 90, 20, 255}
	white  = color.RGBA{255, 255, 255, 255}

	selectedDzRegion string
)

var dzLandmarksMap = map[string][]string{
	"east":  {"Labor Department", "Prison Bureau", "Tax Court", "Clock Tower", "Expansion", "Dead Park", "Lab", "Kitchen", "Wreck", "Palace", "DC-62 Storage", "Morgue", "The Grave", "Pool", "Collapse", "Stonehenge", "Reclaimed Forum"},
	"south": {"Military Camp", "Backfire", "Shantytown", "Bureau", "Shanghai Hotel", "The Swamp", "Garage", "Train Vault", "The Oven", "Chem Storage", "Stockpile", "USDA Theater", "USDA Cafeteria"},
	"west":  {"Ruined Harbor", "Graveyard", "Papermill", "Flooded Mall", "Deserted Suites", "Back Door", "Mansions", "Hotel 62", "DC-62 Plant"},
}

type buttonAndText struct {
	button *fyne.Container
	text   *fyne.Container
}

func init() {
	fyneApp.Settings().SetTheme(theme.DarkTheme())
}

func main() {
	timers = new(sync.Map)

	showZonePicker()

	window.ShowAndRun()
}

func showZonePicker() {
	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(1))

	grid.Add(widget.NewButton("East", func() {
		selectedDzRegion = "east"
		showLandmarkTimers()
	}))
	grid.Add(widget.NewButton("West", func() {
		selectedDzRegion = "west"
		showLandmarkTimers()
	}))
	grid.Add(widget.NewButton("South", func() {
		selectedDzRegion = "south"
		showLandmarkTimers()
	}))

	window.SetContent(grid)
}

func showLandmarkTimers() {
	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(2))

	// default to the first item being the "back" button
	dzText := canvas.NewText(selectedDzRegion, white)
	dzText.TextSize = 50

	rows := []buttonAndText{{
		button: container.NewMax(canvas.NewRectangle(purple), widget.NewButton("Zone Picker", showZonePicker)),
		text:   container.NewMax(canvas.NewRectangle(color.Black), dzText),
	}}

	// add landmarks to the items we will render
	for _, landmarkName := range dzLandmarksMap[selectedDzRegion] {
		landmarkName := landmarkName

		tappedFunc := func() {
			updateButtonColor(landmarkName, orange)
			stopTimer(landmarkName)
			time.Sleep(1 * time.Second)
			startTimer(landmarkName, timerCallback, 10)
		}
		unknownText := canvas.NewText("unknown", white)
		unknownText.TextSize = 50

		buttonCanvas := widget.NewButton(landmarkName, tappedFunc)

		buttonContainer := container.NewMax(canvas.NewRectangle(blue), buttonCanvas)
		text := container.NewMax(canvas.NewRectangle(black), unknownText)

		rows = append(rows, buttonAndText{buttonContainer, text})
	}

	// render our rows
	for _, row := range rows {
		grid.Add(row.button)
		grid.Add(row.text)
	}

	window.SetContent(grid)

}

/*
row 1
	container
		rectangle
		button
	container
		rectangle
		text
row 2
	container
		rectangle
		button
	container
		rectangle
		text
*/

func timerCallback(id string) {
	updateButtonColor(id, green)
	updateText(id, "00:00")
	tts(id)
}

// update button color for a given ID
func updateButtonColor(id string, color color.Color) {
	for i, item := range grid.Objects {
		i, item := i, item
		// skip header
		if i == 0 {
			continue
		}
		if button, isButton := item.(*fyne.Container).Objects[1].(*widget.Button); isButton {
			if button.Text == id {
				item.(*fyne.Container).Objects[0].(*canvas.Rectangle).FillColor = color
			}
		}
	}
	grid.Refresh()
}

// update a row text to have a new value
func updateText(id string, s string) {
	for i, row := range grid.Objects {
		i, row := i, row
		// skip header
		if i == 0 {
			continue
		}
		if button, isButton := row.(*fyne.Container).Objects[1].(*widget.Button); isButton {
			if button.Text == id {
				grid.Objects[i+1].(*fyne.Container).Objects[1].(*canvas.Text).Text = s
			}
		}
	}
	grid.Refresh()
}

func tts(id string) {
	speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.Native{}}
	speech.Speak(id + " is ready")
}
