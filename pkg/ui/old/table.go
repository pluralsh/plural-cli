package uiOld

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

var pages = tview.NewPages()

// Table demonstrates the Table.
func Table(c *cli.Context, nextSlide func()) (title string, content tview.Primitive) {
	table := tview.NewTable().
		SetFixed(1, 1)

	client := api.NewClient()

	//Load applications into table
	repos, _ := client.ListRepositories("")
	// if err != nil {
	// 	return err
	// }

	for row, line := range repos {
		textRow := []string{line.Name, line.Publisher.Name, line.Description}
		for column, cell := range textRow {
			color := tcell.ColorWhite
			if column == 2 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			if row == 0 {
				align = tview.AlignCenter
			} else if column == 1 || column >= 4 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(column != 2)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			table.SetCell(row, column, tableCell)
		}
	}
	table.SetBorder(true).SetTitle("Applications")

	table.SetSelectable(true, false).
		SetSeparator(' ')

	//Search for an application and update the table
	searchApp := func(query string, table *tview.Table) {

		table.Clear()
		repos, _ := client.ListRepositories(query)

		// if err != nil {
		// 	return err
		// }

		for row, line := range repos {
			textRow := []string{line.Name, line.Publisher.Name, line.Description}
			for column, cell := range textRow {
				color := tcell.ColorWhite
				if column == 2 {
					color = tcell.ColorDarkCyan
				}
				align := tview.AlignLeft
				if row == 0 {
					align = tview.AlignCenter
				} else if column == 1 || column >= 4 {
					align = tview.AlignRight
				}
				tableCell := tview.NewTableCell(cell).
					SetTextColor(color).
					SetAlign(align).
					SetSelectable(column != 2)
				if column >= 1 && column <= 3 {
					tableCell.SetExpansion(1)
				}
				table.SetCell(row, column, tableCell)
			}
		}
		table.SetBorder(true).SetTitle("Applications")

		table.SetSelectable(true, false).
			SetSeparator(' ')
	}

	list := tview.NewList()

	inputField := tview.NewInputField().
		SetLabel("Search for an application: ").
		SetFieldWidth(10).
		// SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				app.SetFocus(table)
			case tcell.KeyTab:
				app.SetFocus(table)
			case tcell.KeyEscape:
				app.SetFocus(list)
			case tcell.KeyBacktab:
				app.SetFocus(list)
			}
		}).SetChangedFunc(func(text string) {
		searchApp(text, table)
	})

	code := tview.NewTextView().
		SetWrap(false).
		SetDynamicColors(true)
	code.SetBorderPadding(1, 1, 2, 0)

	man, _ := manifest.FetchProject()

	// recipies, _ := client.ListRecipes("", strings.ToUpper(man.Provider))

	bundleTable := tview.NewTable()
	bundleTable.SetBorder(true).SetTitle("Bundles")
	bundleTable.SetSelectable(true, false).
		SetSeparator(' ')

	// selectRow := func() {
	// 	table.SetBorders(false).
	// 		SetSelectable(true, false).
	// 		SetSeparator(' ')
	// 	code.Clear()
	// 	fmt.Fprint(code, tableSelectRow)
	// }
	// fmt.Fprint(code, tableSelectRow)

	f := tview.NewForm().
		AddInputField("First name:", "", 20, nil, nil).
		AddInputField("Last name:", "", 20, nil, nil).
		AddDropDown("Role:", []string{"Engineer", "Manager", "Administration"}, 0, nil).
		AddCheckbox("On vacation:", false, nil).
		AddPasswordField("Password:", "", 10, '*', nil).
		AddButton("Save", func() {
			pages.HidePage("modal")
			app.SetFocus(table)
		}).
		AddButton("Cancel", func() { pages.HidePage("modal") })

	f.SetBorder(true).SetTitle("Employee Information")

	// modal := TestModal(f, 60, 60)

	navigate := func() {
		app.SetFocus(inputField)
		table.SetDoneFunc(func(key tcell.Key) {
			app.SetFocus(list)
		}).SetSelectedFunc(func(row int, column int) {
			GetRecipes(pages, row, client, repos, man, table, bundleTable)
			// pages.ShowPage("modal")
		})
	}

	list.ShowSecondaryText(false).
		// AddItem("Selectable rows", "", 'r', selectRow).
		AddItem("Navigate", "", 'n', navigate).
		AddItem("Exit", "", 'x', app.Stop)
	list.SetBorderPadding(1, 1, 2, 2)

	// box := tview.NewBox().
	// 	SetBorder(true).
	// 	SetTitle("Centered Box")

	// mainView := tview.NewFlex().
	// 	AddItem(tview.NewFlex().
	// 		SetDirection(tview.FlexRow).
	// 		AddItem(list, 10, 1, true).
	// 		AddItem(table, 0, 1, false), 0, 1, true).
	// 	AddItem(bundleTable, codeWidth, 1, false)

	pages.AddPage("app-table", table, true, true)
	// AddPage("modal", modal, true, true)

	// pages.HidePage("modal")

	// return "Table", pages
	return "Table", tview.NewFlex().
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(list, 10, 1, true).
			AddItem(inputField, 10, 1, true).
			AddItem(pages, 0, 1, false), 0, 1, true)
	// AddItem(pages, codeWidth, 1, false)
	// AddItem(installForm, codeWidth, 1, false)
}

func TestModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, false).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}

func GetRecipes(pages *tview.Pages, row int, client *api.Client, repos []*api.Repository, man *manifest.ProjectManifest, table, bundleTable *tview.Table) {
	app.SetFocus(bundleTable)

	// pages := tview.NewPages()

	// mainView := tview.NewFlex().
	// 	AddItem(tview.NewFlex().
	// 		SetDirection(tview.FlexRow).
	// 		AddItem(bundleTable, 10, 1, true), 0, 1, true)
	// 	// 	AddItem(table, 0, 1, false), 0, 1, true).
	// 	// AddItem(bundleTable, codeWidth, 1, false)

	// modal := TestModal(installForm, 60, 60)

	// pages.AddPage("main", mainView, true, true).
	// 	AddPage("modal", modal, true, true)

	recipies, _ := client.ListRecipes(repos[row].Name, strings.ToUpper(man.Provider))
	bundleTable.Clear()
	for row, line := range recipies {
		textRow := []string{line.Name, line.Provider, line.Description}
		for column, cell := range textRow {
			color := tcell.ColorWhite
			if column == 2 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			if row == 0 {
				align = tview.AlignCenter
			} else if column == 1 || column >= 4 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(column != 2)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			bundleTable.SetCell(row, column, tableCell)
		}
	}

	bundleTable.SetDoneFunc(func(key tcell.Key) {
		pages.RemovePage("bundle-table")
		app.SetFocus(table)
	}).SetSelectedFunc(func(bundleRow int, bundleColumn int) {

		// installForm := tview.NewForm()
		installForm := Install(nil, repos[row].Name, recipies[bundleRow].Name)

		// installForm := tview.NewForm().
		// 	SetText("Do you want to quit the application?").
		// 	AddButtons([]string{"Quit", "Cancel"}).
		// 	SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		// 		if buttonLabel == "Quit" {
		// 			app.Stop()
		// 		}
		// 	})

		modal := TestModal(installForm, 60, 60)
		pages.AddPage("modal", modal, true, true)

		pages.ShowPage("modal")
		// installForm.SetBorder(true).SetTitle("Employee Information")
		// // installForm = Install(app, c, repos[row].Name, recipies[bundleRow].Name, installForm)
		// app.SetFocus(installForm)
	})

	// app.SetFocus(bundleTable)
	pages.AddAndSwitchToPage("bundle-table", bundleTable, true)

}
