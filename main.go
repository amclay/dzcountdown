package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"runtime"
	"strings"
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
	"github.com/kindlyfire/go-keylogger"
)

var (
	timerSeconds = 1800
	timers       *sync.Map

	fyneApp         = app.New()
	window          = fyneApp.NewWindow("Xangold's Countdown Timers")
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

	finishedPhrasesPrefixes = []string{"Get ready for", "Get your ass moving to", "Prepare for", "Hurry up to", "Run to", "Get going to", "Lets get going to ", "Next up is"}
	finishedPhrasesSuffixes = []string{"You idiot", "You donkey", "dumb agent", "slowpoke", "moron", "poo poo head", "dummy", "brainwashed puppet"}
)

var dzLandmarksMap = map[string][]string{
	"east": {
		"Lab",
		"Clock Tower",
		"Expansion",
		"Prison Bureau",
		"Pool",
		"Stonehenge",
		"Catacombs",
		"CERA Camp",
		"Morgue",
		"DC-62 Storage",
		"Palace",
	},
	"south": {
		"Backfire",
		"Shantytown",
		"Shanghai Hotel",
		"Garage",
		"The Oven",
		"USDA Cafeteria",
		"USDA Theater",
	},
	"west": {
		"DC-62 Plant",
		"Mansions",
		"Back Door",
		"Twin Courts",
		"Graveyard",
		"Papermill",
		"Flooded Mall",
		"Deserted Suites",
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

type extendedButton struct {
	widget.Button
}

func newExtendedButton(label string, tapped func()) *extendedButton {
	ret := &extendedButton{}
	ret.ExtendBaseWidget(ret)
	ret.Text = label
	ret.OnTapped = tapped
	return ret
}

func (eb *extendedButton) TappedSecondary(pe *fyne.PointEvent) {
	id := eb.Text
	stopTime, ok := timers.Load(id)
	if !ok {
		return
	}

	t := stopTime.(time.Time)
	t2 := t.Add(-30 * time.Second)

	timers.Store(id, t2)
}

func main() {

	timers = new(sync.Map)

	showZonePicker()

	go timerLoop()
	go listenForF5()

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

func getLastGridItemName() string {
	return grid.Objects[len(grid.Objects)-1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*extendedButton).Text
}

// move the timer to the top of the list
func moveToTop(id string) {
	for i, gridItem := range grid.Objects {
		if i == 0 {
			continue
		}
		// hocus pocus shit to get the landmark name from the canvas
		landmarkName := gridItem.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*extendedButton).Text
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

	prefix := finishedPhrasesPrefixes[rand.Intn(len(finishedPhrasesPrefixes))]
	suffix := finishedPhrasesSuffixes[rand.Intn(len(finishedPhrasesSuffixes))]

	speech.Speak(fmt.Sprintf("%s %s, %s", prefix, id, suffix))
}

func listenForF5() {
	// listen to F5 key, and move the last grid item to the top on press if using windows
	if runtime.GOOS == "windows" {
		keylogger := keylogger.NewKeylogger()
		go func() {
			for {
				time.Sleep(10 * time.Millisecond)
				if globalRegion == "" {
					continue
				}
				key := keylogger.GetKey()
				if key.Keycode == 116 && !key.Empty {
					id := getLastGridItemName()
					moveToTop(id)
					updateButtonColor(id, orange)

					timers.Delete(id)

					startTimer(id, timerSeconds)

					moveToTop(id)
				}
			}
		}()
	}
}
