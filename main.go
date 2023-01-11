package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"strings"
	"sync"

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

	fyneApp         = app.New()
	window          = fyneApp.NewWindow("SUS/DVS/SAS Countdown Timers")
	parentContainer = container.New(layout.NewHBoxLayout())
	grid            = container.New(layout.NewGridLayout(2))

	red          = color.RGBA{255, 0, 0, 255}
	green        = color.RGBA{0, 100, 0, 255}
	blue         = color.RGBA{19, 13, 84, 255}
	black        = color.RGBA{0, 0, 0, 255}
	purple       = color.RGBA{128, 0, 128, 255}
	orange       = color.RGBA{128, 90, 20, 255}
	white        = color.RGBA{255, 255, 255, 255}
	globalRegion = ""

	globalLandmarkRows = []*landmarkRow{}

	//go:embed south_annotated.png
	southImage []byte

	//go:embed west_annotated.png
	westImage []byte

	//go:embed east_annotated.png
	eastImage []byte
)

var dzLandmarksMap = map[string][]string{
	"east": {
		"Prison Bureau",
		"Clock Tower",
		"Expansion",
		"Lab",
		"Palace",
		"DC-62 Storage",
		"Morgue",
		"Pool",
		"Stonehenge",
		"CERA Camp",
		"Catacombs",
	},
	"south": {
		"Backfire",
		"Shantytown",
		"Shanghai Hotel",
		"Garage",
		"The Oven",
		"Chem Storage",
		"USDA Theater",
		"USDA Cafeteria",
	},
	"west": {
		"Graveyard",
		"Papermill",
		"Flooded Mall",
		"Deserted Suites",
		"Back Door",
		"Twin Courts",
		"Mansions",
		"DC-62 Plant",
	},
}

func init() {
	for _, zone := range dzLandmarksMap {
		for i, landmark := range zone {
			zone[i] = fmt.Sprintf("%-20s", landmark)
		}
	}
	fyneApp.Settings().SetTheme(myTheme{})
}

func main() {
	timers = new(sync.Map)

	showZonePicker()

	window.ShowAndRun()
}

func showZonePicker() {
	parentContainer.RemoveAll()
	parentContainer = container.New(layout.NewHBoxLayout())

	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(1))

	grid.Add(widget.NewButton("East", func() {
		globalRegion = "east"
		showLandmarkTimers(globalRegion)
	}))

	grid.Add(widget.NewButton("West", func() {
		globalRegion = "west"
		showLandmarkTimers(globalRegion)
	}))
	grid.Add(widget.NewButton("South", func() {
		globalRegion = "south"
		showLandmarkTimers(globalRegion)
	}))

	parentContainer.Add(grid)

	window.Resize(fyne.NewSize(100, 300))

	window.SetContent(parentContainer)
}

func showLandmarkTimers(region string) {
	parentContainer.RemoveAll()
	parentContainer = container.New(layout.NewHBoxLayout())

	grid.RemoveAll()

	globalLandmarkRows = getLandmarkRows(region)

	for _, item := range globalLandmarkRows {
		grid.Add(item.toContainer())
	}

	var imageFile image.Image
	switch region {
	case "east":
		imageFile, _, _ = image.Decode(bytes.NewReader(eastImage))
	case "west":
		imageFile, _, _ = image.Decode(bytes.NewReader(westImage))
	case "south":
		imageFile, _, _ = image.Decode(bytes.NewReader(southImage))
	default:
		log.Fatalf("unknown region: %s", region)
	}

	image := canvas.NewImageFromImage(imageFile)
	image.SetMinSize(fyne.NewSize(800, 800))
	parentContainer.Add(image)
	parentContainer.Add(grid)

	window.SetContent(parentContainer)
}

func getLandmarkRows(region string) []*landmarkRow {
	// default to the first item being the "back" button
	dzText := canvas.NewText(region, white)
	dzText.TextSize = 30

	rows := []*landmarkRow{{
		landmarkName:    "",
		buttonContainer: container.NewMax(canvas.NewRectangle(purple), widget.NewButton(fmt.Sprintf("%-20s", "Zone Picker"), showZonePicker)),
		textContainer:   container.NewMax(canvas.NewRectangle(color.Black), dzText),
	}}

	// add landmarks to the items we will render
	for _, landmarkName := range dzLandmarksMap[region] {
		landmarkName := landmarkName
		rows = append(rows, newLandmarkRow(landmarkName))
	}

	return rows
}

func timerCallback(id string) {
	updateButtonColor(id, green)
	updateText(id, fmt.Sprintf("%-7s", "00:00"))
	tts(id)
}

// update button color for a given ID
func updateButtonColor(id string, color color.Color) {
	for _, row := range globalLandmarkRows {
		if row.landmarkName == id {
			row.updateButtonColor(color)
		}
	}
	grid.Refresh()
}

// update a row text to have a new value
func updateText(id string, s string) {
	for _, row := range globalLandmarkRows {
		if row.landmarkName == id {
			row.updateText(s)
		}
	}
	grid.Refresh()
}

// move the timer to the top of the list
func moveToTop(id string) {
	for i, gridItem := range grid.Objects {
		if i == 0 {
			continue
		}
		// hocus pocus shit to get the button
		landmarkName := gridItem.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*widget.Button).Text
		if landmarkName == id {
			grid.Objects = moveItem(grid.Objects, i, 1)
		}
	}
	grid.Refresh()
}

func moveItem(array []fyne.CanvasObject, srcIndex int, dstIndex int) []fyne.CanvasObject {
	value := array[srcIndex]
	return insertItem(removeItem(array, srcIndex), value, dstIndex)
}

func insertItem(array []fyne.CanvasObject, value fyne.CanvasObject, index int) []fyne.CanvasObject {
	return append(array[:index], append([]fyne.CanvasObject{value}, array[index:]...)...)
}

func removeItem(array []fyne.CanvasObject, index int) []fyne.CanvasObject {
	return append(array[:index], array[index+1:]...)
}

func tts(id string) {
	idParts := strings.Split(id, ":")
	if len(idParts) > 1 {
		id = idParts[1]
	}
	speech := htgotts.Speech{Folder: "audio", Language: voices.EnglishUK, Handler: &handlers.Native{}}
	speech.Speak(id + " is ready")
}
