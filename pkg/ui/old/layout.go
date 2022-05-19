package uiOld

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

// Slide is a function which returns the slide's main primitive and its title.
// It receives a "nextSlide" function which can be called to advance the
// presentation to the next slide.
type Slide func(*cli.Context, func()) (title string, content tview.Primitive)

// The application.
var app = tview.NewApplication()

// func InteractiveLayout(c *cli.Context) error {
// 	modal := func(p tview.Primitive, width, height int) tview.Primitive {
// 		return tview.NewFlex().
// 			AddItem(nil, 0, 1, false).
// 			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
// 				AddItem(nil, 0, 1, false).
// 				AddItem(p, height, 1, false).
// 				AddItem(nil, 0, 1, false), width, 1, false).
// 			AddItem(nil, 0, 1, false)
// 	}

// 	background := tview.NewTextView().
// 		SetTextColor(tcell.ColorBlue).
// 		SetText(strings.Repeat("background ", 1000))

// 	box := tview.NewBox().
// 		SetBorder(true).
// 		SetTitle("Centered Box")

// 	pages := tview.NewPages().
// 		AddPage("background", background, true, true).
// 		AddPage("modal", modal(box, 40, 10), true, true)

// 	pages.HidePage("modal")

// 	if err := app.SetRoot(pages, true).Run(); err != nil {
// 		return err
// 	}
// 	return nil
// }

func InteractiveLayout(c *cli.Context) error {
	// The presentation slides.
	slides := []Slide{
		Init,
		Table,
		// Modal,
	}

	pages := tview.NewPages()

	// The bottom row has some info on where we are.
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			pages.SwitchToPage(added[0])
		})

	// Create the pages for all slides.
	previousSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide - 1 + len(slides)) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	nextSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide + 1) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	for index, slide := range slides {
		title, primitive := slide(c, nextSlide)
		pages.AddPage(strconv.Itoa(index), primitive, true, index == 0)
		fmt.Fprintf(info, `%d ["%d"][darkcyan]%s[white][""]  `, index+1, index, title)
	}
	info.Highlight("0")

	// box := tview.NewBox().
	// 	SetBorder(true).
	// 	SetTitle("Centered Box")

	// pages.AddPage("error", ErrorModal(c, fmt.Errorf("test error"), 40, 10), true, false)

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 1, 1, false)

	// Shortcuts to navigate the slides.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextSlide()
			return nil
		} else if event.Key() == tcell.KeyCtrlP {
			previousSlide()
			return nil
		}
		return event
	})

	// pages.ShowPage("error")
	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}
