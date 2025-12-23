package main

import (
	"os"
	"pro10_tv_finder/helper"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var cache *helper.Cache
var debounceTimer *time.Timer
var debounceMs = 150
var maxResults = 123

func initFiles(root string, numWorkers int) {
	cache = &helper.Cache{}

	dirJobs := make(chan helper.File, 100)
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Go(func() {
			for job := range dirJobs {
				helper.ProcessFiles(job.Path, cache, dirJobs)
			}
		})
	}

	dirJobs <- helper.File{Path: root}

	go func() {
		wg.Wait()
		close(dirJobs)
	}()

	wg.Wait()
}

func previewFile(path string, previewArea *tview.TextView) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}

	data := string(contents)
	previewArea.SetText(helper.HighlightContent(data, path))
}

func feedToResList(resultsList *tview.List, results []string, previewArea *tview.TextView) {
	if len(results) == 0 {
		return
	}

	_, _, width, _ := resultsList.GetInnerRect()

	for _, p := range results {
		path := p
		resultsList.AddItem("-> "+helper.ShortenPath(path, width-6), "", 0, func() {
			previewFile(path, previewArea)
		})
	}
}

func initPreview(previewArea *tview.TextView) {
	previewArea.SetBorder(true).SetBorderColor(tcell.ColorBlack)
	previewArea.SetDynamicColors(true)
	previewArea.SetWrap(false)
	previewArea.SetBorderPadding(1, 1, 1, 1)

	previewArea.SetFocusFunc(func() {
		previewArea.SetBorderColor(tcell.ColorGhostWhite)
	})

	previewArea.SetBlurFunc(func() {
		previewArea.SetBorderColor(tcell.ColorBlack)
	})

	previewArea.SetScrollable(true)
	previewArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
}

func initResList(resultsList *tview.List) {
	resultsList.
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorBlack).
		SetBorderPadding(1, 1, 1, 1)

	resultsList.
		SetHighlightFullLine(false).
		ShowSecondaryText(false)

	resultsList.SetFocusFunc(func() {
		resultsList.SetBorderColor(tcell.ColorGhostWhite)
	})

	resultsList.SetBlurFunc(func() {
		resultsList.SetBorderColor(tcell.ColorBlack)
	})
}

func initInput(inputField *tview.InputField, resultsList *tview.List, previewArea *tview.TextView, app *tview.Application) {
	inputField.
		SetLabel(" > ").
		SetLabelStyle(tcell.StyleDefault.Bold(true)).
		SetLabelColor(tcell.ColorLightGreen).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorGhostWhite)

	inputField.SetFocusFunc(func() {
		inputField.SetFieldTextColor(tcell.ColorWhite)
		inputField.SetLabelColor(tcell.ColorLightGreen)
		inputField.SetBorderColor(tcell.ColorGhostWhite)
	})

	inputField.SetBlurFunc(func() {
		inputField.SetLabelColor(tcell.ColorDarkGreen)
		inputField.SetFieldTextColor(tcell.ColorDarkGray)
		inputField.SetBorderColor(tcell.ColorBlack)
	})

	inputField.SetChangedFunc(func(text string) {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}

		debounceTimer = time.AfterFunc(time.Duration(debounceMs)*time.Millisecond, func() {
			if text == "" {
				app.QueueUpdateDraw(func() {
					resultsList.Clear()
				})
				return
			}

			results := cache.Search(text, maxResults)

			app.QueueUpdateDraw(func() {
				resultsList.Clear()
				feedToResList(resultsList, results, previewArea)
			})
		})
	})
}

func main() {
	var (
		app      = tview.NewApplication()
		mainView = tview.NewGrid()

		inputField  = tview.NewInputField()
		resultsList = tview.NewList()
		previewArea = tview.NewTextView()
	)

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	go initFiles(root, 10)

	initInput(inputField, resultsList, previewArea, app)
	initResList(resultsList)
	initPreview(previewArea)

	mainView.
		SetBorders(false).
		SetColumns(0, 0).
		SetRows(-1, -5).
		SetBackgroundColor(tcell.ColorDefault)

	mainView.
		AddItem(inputField, 0, 0, 1, 1, 1, 1, true).
		AddItem(resultsList, 1, 0, 1, 1, 1, 1, false).
		AddItem(previewArea, 0, 1, 2, 1, 1, 5, false)

	app.SetRoot(mainView, true)
	app.SetFocus(inputField)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyDown:
			if !previewArea.HasFocus() && resultsList.GetItemCount() > 0 {
				app.SetFocus(resultsList)
			}

		case tcell.KeyUp:
			if !previewArea.HasFocus() && resultsList.GetCurrentItem() == 0 {
				app.SetFocus(inputField)
				app.QueueEvent(tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone))
			}

		case tcell.KeyRight:
			if !previewArea.HasFocus() {
				app.SetFocus(previewArea)
				return nil
			}

		case tcell.KeyLeft:
			_, c := previewArea.GetScrollOffset()

			if !resultsList.HasFocus() && c == 0 {
				app.SetFocus(resultsList)
			}
		}

		return event
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
