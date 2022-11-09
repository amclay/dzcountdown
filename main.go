package main

import (
	"fmt"
	"image/color"
	"log"
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
	ttlcache "github.com/jellydator/ttlcache/v3"
)

// globals cause yolo
var (
	fyneApp         = app.New()
	window          = fyneApp.NewWindow("DZ Countdown Timers")
	timers          = make(map[string]time.Time)
	timerResetFuncs = make(map[string]func())
	dzRegion        = "east"
	grid            = container.New(layout.NewGridLayout(2))
	red             = color.RGBA{255, 0, 0, 255}
	green           = color.RGBA{0, 100, 0, 255}
	blue            = color.RGBA{19, 13, 84, 255}
	black           = color.RGBA{0, 0, 0, 255}
	purple          = color.RGBA{128, 0, 128, 255}
	orange          = color.RGBA{128, 90, 20, 255}
	white           = color.RGBA{255, 255, 255, 255}

	cache = ttlcache.New[string, string]()
)

func main() {
	go cache.Start()

	showZonePicker()

	window.ShowAndRun()
}

func showZonePicker() {
	for k := range timerResetFuncs {
		delete(timerResetFuncs, k)
	}
	for k := range timers {
		delete(timers, k)
	}

	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(1))

	grid.Add(widget.NewButton("East", func() {
		dzRegion = "east"
		showLandmarkTimers()
	}))
	grid.Add(widget.NewButton("West", func() {
		dzRegion = "west"
		showLandmarkTimers()
	}))
	grid.Add(widget.NewButton("South", func() {
		dzRegion = "south"
		showLandmarkTimers()
	}))

	window.SetContent(grid)
}

type buttonAndText struct {
	button *fyne.Container
	text   *fyne.Container
}

func showLandmarkTimers() {
	grid.RemoveAll()
	grid = container.New(layout.NewGridLayout(2))

	ticker := time.NewTicker(100 * time.Millisecond)

	// default to the first item being the "back" button
	canvasItems := []buttonAndText{{
		text:   container.NewMax(canvas.NewRectangle(color.Black), widget.NewLabel(dzRegion)),
		button: container.NewMax(canvas.NewRectangle(purple), widget.NewButton("Zone Picker", showZonePicker)),
	}}

	landmarks := []string{}

	switch dzRegion {
	case "east":
		landmarks = []string{"Labor Department", "Prison Bureau", "Tax Court", "Clock Tower", "Expansion", "Dead Park", "Lab", "Kitchen", "Wreck", "Palace", "DC-62 Storage", "Morgue", "The Grave", "Pool", "Collapse", "Stonehenge", "Reclaimed Forum"}
	case "south":
		landmarks = []string{"Military Camp", "Backfire", "Bureau", "USDA Theater", "USDA Cafeteria", "Stockpile", "The Oven", "Train Vault", "Garage", "Shanghai Hotel", "The Swamp"}
	case "west":
		landmarks = []string{"Ruined Harbor", "Graveyard", "Papermill", "Flooded Mall", "Deserted Suites", "Back Door", "Mansions", "Deserted Suites", "Hotel 62", "DC-62 Plant"}
	}

	// setup the gui and timers for the landmarks
	for _, landmark := range landmarks {
		landmarkName := landmark
		timerResetFuncs[landmarkName] = func() {
			timers[landmarkName] = time.Now().Add(30 * time.Minute)
			for _, item := range canvasItems {
				if item.button.Objects[1].(*widget.Button).Text == landmarkName {
					item.button.Objects[0].(*canvas.Rectangle).FillColor = orange
				}
			}
		}
		button := widget.NewButton(landmarkName, timerResetFuncs[landmarkName])
		max := container.NewMax(canvas.NewRectangle(blue), button)
		canvasItems = append(canvasItems, buttonAndText{
			text:   container.NewMax(canvas.NewRectangle(color.Black), canvas.NewText("unknown", white)),
			button: max,
		})
	}

	for _, item := range canvasItems {
		grid.Add(item.button)
		grid.Add(item.text)
	}

	window.SetContent(grid)

	// poll our timers, yikes
	go func() {
		for range ticker.C {
			for i, item := range canvasItems {
				i, item := i, item
				// ignore "zone picker" button
				if i == 0 {
					continue
				}
				item.text.Objects[1].(*canvas.Text).Text = getTimeLabelForLocation(item.button.Objects[1].(*widget.Button).Text)
				item.text.Objects[1].(*canvas.Text).Color = white
				item.text.Objects[1].(*canvas.Text).TextStyle.Bold = true
				if item.text.Objects[1].(*canvas.Text).Text == "00:01" {
					go func() {
						locationName := item.button.Objects[1].(*widget.Button).Text
						// check if spoken before (since we're polling our timer objects yikes)
						cacheItem := cache.Get(locationName)
						if cacheItem == nil {
							log.Println(item.button.Objects[1].(*widget.Button).Text, "is about to expire")
							cache.Set(locationName, "", 10*time.Second)
							speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.Native{}}
							speech.Speak(item.button.Objects[1].(*widget.Button).Text + " is ready")
						}
					}()
					item.button.Objects[0].(*canvas.Rectangle).FillColor = green
				}
			}
			grid.Refresh()
		}
	}()

}

func getTimeLabelForLocation(location string) string {
	durationLeft := time.Until(timers[location])

	minutes := int(durationLeft.Minutes())
	seconds := int(durationLeft.Seconds()) - (minutes * 60)
	if minutes < -9999 {
		return "unknown"
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
