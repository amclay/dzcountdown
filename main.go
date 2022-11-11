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
	"fyne.io/fyne/v2/widget"
	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	voices "github.com/hegedustibor/htgo-tts/voices"
)

var (
	timerSeconds = 1800
	timers       *sync.Map

	fyneApp = app.New()
	window  = fyneApp.NewWindow("SUS/DVS/SAS Countdown Timers")
	grid    = container.New(layout.NewGridLayout(2))

	red    = color.RGBA{255, 0, 0, 255}
	green  = color.RGBA{0, 100, 0, 255}
	blue   = color.RGBA{19, 13, 84, 255}
	black  = color.RGBA{0, 0, 0, 255}
	purple = color.RGBA{128, 0, 128, 255}
	orange = color.RGBA{128, 90, 20, 255}
	white  = color.RGBA{255, 255, 255, 255}
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

func (b *buttonAndText) updateText(id string, text string) {
	b.text.Objects[1].(*canvas.Text).Text = text
}

func (b *buttonAndText) updateColor(id string, color color.Color) {
	b.button.Objects[0].(*canvas.Rectangle).FillColor = color
}

func init() {

	fyneApp.Settings().SetTheme(myTheme{})
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
		showLandmarkTimers("east")
	}))
	grid.Add(widget.NewButton("West", func() {
		showLandmarkTimers("west")
	}))
	grid.Add(widget.NewButton("South", func() {
		showLandmarkTimers("south")
	}))

	window.SetContent(grid)

	// set window size
	window.Resize(fyne.NewSize(500, 600))
}

func showLandmarkTimers(region string) {
	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(2))

	for _, item := range getGridItemsForLayout(region) {
		grid.Add(item.button)
		grid.Add(item.text)
	}

	window.SetContent(grid)
}

func getGridItemsForLayout(region string) []*buttonAndText {
	// default to the first item being the "back" button
	dzText := canvas.NewText(region, white)
	dzText.TextSize = 30

	rows := []*buttonAndText{{
		button: container.NewMax(canvas.NewRectangle(purple), widget.NewButton("Zone Picker", showZonePicker)),
		text:   container.NewMax(canvas.NewRectangle(color.Black), dzText),
	}}

	// add landmarks to the items we will render
	for _, landmarkName := range dzLandmarksMap[region] {
		landmarkName := landmarkName

		tappedFunc := func() {
			updateButtonColor(landmarkName, orange)
			stopTimer(landmarkName)
			time.Sleep(1 * time.Second)
			startTimer(landmarkName, timerCallback, timerSeconds)
		}
		unknownText := canvas.NewText("unknown", white)
		unknownText.TextSize = 30

		buttonContainer := container.NewMax(canvas.NewRectangle(blue), widget.NewButton(landmarkName, tappedFunc))

		text := container.NewMax(canvas.NewRectangle(black), unknownText)

		rows = append(rows, &buttonAndText{buttonContainer, text})
	}
	return rows
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
