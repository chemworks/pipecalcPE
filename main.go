package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	chengcroma "github.com/chemworks/croma"
	"github.com/chemworks/pipecalcPE/model"
)

const (
	fileNameLines = "Lines.json"
	fileNameCases = "Cases.json"
)

// TODO
// TODO
// TODO make a function to avoid if else on validation arround lines 142 >0
// this can be maked with a function func (a *widget.Entry) (val float64)

//TODO DOC
func compFileName(path, file string) string {
	path = path[7:]
	return fmt.Sprintf("%s/%s", path, file)
}

// function to create an empty widget with place holder and NON Wraping
// TODO mover esto no va aca en el medio
func widgetNewEntry(s string) *widget.Entry {
	w := widget.NewEntry()
	w.SetPlaceHolder(s)
	w.Wrapping = fyne.TextWrapOff
	return w
}

func setFloat64Entry(e *widget.Entry, v float64) {
	if v > 0 {
		//val := fmt.Sprintf("%.2f", v)
		val := fmt.Sprintf("%.0f", v)
		e.SetText(val)
	}
	if v == 0 {
		e.SetText("0") // TODO FIXME
	}
}

// Make the items to the Main App, this will be modified later
type chengPipesApp struct {
	// All lines And Cases
	Lines  *model.Lines
	Cases  *model.CaseList
	Cromas []*chengcroma.Croma
	// Current Line and current case
	Line  *model.Line
	Case  *model.Case
	Croma *chengcroma.Croma
	// Displayed items on ui for Lines
	ListLines *widget.List
	LineTag   *widget.Entry
	LineCases *widget.Entry
	// Displayed items on ui for Lines
	ListCases                            *widget.List
	CaseTag, GasFlow, Press, Temp, MW, Z *widget.Entry
	MultyFlow                            *widget.RadioGroup
	LlFlow, HlFlow, LlDens, HlDens       *widget.Entry
	// Cromas
	CC1    *widget.Entry
	CC2    *widget.Entry
	CC3    *widget.Entry
	CiC4   *widget.Entry
	CnC4   *widget.Entry
	CiC5   *widget.Entry
	CnC5   *widget.Entry
	CC6    *widget.Entry
	CC7    *widget.Entry
	CC8    *widget.Entry
	CN2    *widget.Entry
	CCO2   *widget.Entry
	CH2O   *widget.Entry
	CSH2   *widget.Entry
	CTotal *widget.Entry
	// Program data
	SavePath string
	// Return widget for results
	ResultsEntryCases *widget.Entry
	ResultsEntryLines *widget.Entry
}

// Sum all of the components for total
func (app *chengPipesApp) sumCroma() {
	c1 := StringToFloat64(app.CC1.Text)
	c2 := StringToFloat64(app.CC2.Text)
	c3 := StringToFloat64(app.CC3.Text)
	ci4 := StringToFloat64(app.CiC4.Text)
	cn4 := StringToFloat64(app.CnC4.Text)
	ci5 := StringToFloat64(app.CiC5.Text)
	cn5 := StringToFloat64(app.CnC5.Text)
	c6 := StringToFloat64(app.CC6.Text)
	c7 := StringToFloat64(app.CC7.Text)
	c8 := StringToFloat64(app.CC8.Text)
	cn2 := StringToFloat64(app.CN2.Text)
	cco2 := StringToFloat64(app.CCO2.Text)
	ch2o := StringToFloat64(app.CH2O.Text)
	csh2 := StringToFloat64(app.CSH2.Text)
	sum := c1 + c2 + c3 + ci4 + cn4 + ci5 + cn5 + c6 + c7 + c8 + cn2 + cco2 + ch2o + csh2
	sumtxt := fmt.Sprintf("%v", sum)
	app.CTotal.SetText(sumtxt)
}

func (app *chengPipesApp) getFilenames() (string, string) {
	if app.SavePath != "" {
		fileC := compFileName(app.SavePath, fileNameCases)
		fileL := compFileName(app.SavePath, fileNameLines)
		return fileL, fileC
	} else {
		return "Cannot Load File Path", "Cannot Load File Path"
	}
}

// TODO DOC
func (app *chengPipesApp) setFrmCasesToCases() {
	if app.CaseTag.Text != "" {
		actName := app.Case.CaseID
		linesLst := app.Cases.GetCasesList()
		idx := model.SliceItemIdx(linesLst, actName)
		app.Cases.Cas[idx].CaseID = app.CaseTag.Text
		app.Cases.Cas[idx].GasFlow = StringToFloat64(app.GasFlow.Text)
		app.Cases.Cas[idx].Pressure = StringToFloat64(app.Press.Text)
		app.Cases.Cas[idx].Temperature = StringToFloat64(app.Temp.Text)
		app.Cases.Cas[idx].MW = StringToFloat64(app.MW.Text)
		app.Cases.Cas[idx].Z = StringToFloat64(app.Z.Text)
		if app.MultyFlow.Selected == "True" {
			app.Cases.Cas[idx].Multiflow = true
		} else {
			app.Cases.Cas[idx].Multiflow = false
		}
		app.Cases.Cas[idx].LightLiquidFlow = StringToFloat64(app.LlFlow.Text)
		app.Cases.Cas[idx].LightLiquidDens = StringToFloat64(app.LlDens.Text)
		app.Cases.Cas[idx].HeavyLiquidFlow = StringToFloat64(app.HlFlow.Text)
		app.Cases.Cas[idx].HeavyLiquidDens = StringToFloat64(app.HlDens.Text)
		app.Cases.Cas[idx].Results = app.ResultsEntryCases.Text
	}
}

func (app *chengPipesApp) setFrmLinesToLines() {
	if app.LineTag.Text != "" {
		actName := app.Line.Tag
		fmt.Println("actName", actName)
		linesLst := app.Lines.GetLineList()
		idx := model.SliceItemIdx(linesLst, actName)
		fmt.Println("idx", idx)
		app.Lines.Line[idx].Tag = app.LineTag.Text
		app.Lines.Line[idx].CasesList = strings.Split(app.LineCases.Text, ",")
		app.Lines.Line[idx].Results = app.Line.Results
	}
}

// Method to ser the selected pipe to the forms
func (app *chengPipesApp) setPipe(singleline *model.Line) {
	app.Line = singleline
	app.LineTag.SetText(app.Line.Tag)
	casesString := strings.Join(app.Line.CasesList, ",")
	app.LineCases.SetText(casesString)
	res := strings.Join(app.Line.Results, "\n")
	app.ResultsEntryLines.SetText(res)
}

// Method to set the selected Case to the forms
func (app *chengPipesApp) setCase(singleCase *model.Case) {
	// first we add the single case to app.Case actual current case
	app.Case = singleCase
	app.CaseTag.SetText(app.Case.CaseID)
	setFloat64Entry(app.GasFlow, app.Case.GasFlow)
	setFloat64Entry(app.Temp, app.Case.Temperature)
	setFloat64Entry(app.Press, app.Case.Pressure)
	setFloat64Entry(app.MW, app.Case.MW)
	setFloat64Entry(app.Z, app.Case.Z)
	app.ResultsEntryCases.SetText(app.Case.Results)

	if app.Case.Multiflow {

		app.MultyFlow.SetSelected("True")
		setFloat64Entry(app.LlFlow, app.Case.LightLiquidFlow)
		setFloat64Entry(app.HlFlow, app.Case.HeavyLiquidFlow)
		setFloat64Entry(app.LlDens, app.Case.LightLiquidDens)
		setFloat64Entry(app.HlDens, app.Case.HeavyLiquidDens)
	} else {
		app.MultyFlow.SetSelected("False")
		setFloat64Entry(app.LlFlow, 0.0)
		setFloat64Entry(app.HlFlow, 0.0)
		setFloat64Entry(app.LlDens, 0.0)
		setFloat64Entry(app.HlDens, 0.0)
	}
}

// Function to make initialize the ui
// TODO FIXME!!!! I have to Pass win for the dialog (ShowFolder) function
// TODO
func (app *chengPipesApp) makeUI(win fyne.Window) fyne.CanvasObject {
	// i have to move this this is a main tool bar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// Just save to a file
			fileL, fileC := app.getFilenames()
			app.setFrmCasesToCases()
			app.setFrmLinesToLines()
			app.Lines.SaveJson(fileL)
			app.Cases.SaveJson(fileC)

			// TODO Manage errors
		}),
		widget.NewToolbarAction(theme.FileIcon(), func() {

			open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
				if list == nil {
					//log.Println("Cancelled")
					return
				}

				// children, err := list.List()
				// if err != nil {
				// 	dialog.ShowError(err, win)
				// 	return
				//}
				//out := fmt.Sprintf("Folder %s (%d children):\n%s", list.Name(), len(children), list.String())
				//out := fmt.Sprintf("Folder %s \n%s", list.Name(), list.String()) //
				app.SavePath = list.String()
				fileL, fileC := app.getFilenames()
				app.Lines = loadLinesFile(fileL)
				app.Cases = loadCaseFile(fileC)
				//dialog.ShowInformation("Folder Open", out, win)
			}, win)

			// _, err := os.Executable()
			// if err != nil {
			// 	panic(err)
			// }

			open.SetLocation(NewUriFromLocalPath())
			open.Show()
		}),
	)
	//TODO: DOC
	app.ListLines = widget.NewList(
		func() int {
			return len(app.Lines.Line)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template line")
		},
		func(i int, c fyne.CanvasObject) {
			// Change to Label as the return of the last function
			// if check item. Onchanged callback doesnt not work
			// check := c.(*widget.Check)
			check := c.(*widget.Label)
			check.Text = app.Lines.Line[i].Tag
			check.Refresh()
		})

	app.ListCases = widget.NewList(
		func() int {
			return len(app.Cases.Cas)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template line")
		},
		func(i int, c fyne.CanvasObject) {
			// Change to Label as the return of the last function
			// if check item. Onchanged callback doesnt not work
			// check := c.(*widget.Check)
			check := c.(*widget.Label)
			check.Text = app.Cases.Cas[i].CaseID
			check.Refresh()
		})

	//------------------------------------------------------------------------//
	//For Lines
	// app.ListLines... On Select
	app.ListLines.OnSelected = func(id int) {
		// display the selected line
		app.setPipe(app.Lines.Line[id])
	}
	//
	app.LineTag = widgetNewEntry("X\"-HGXXX-CAXX-B")
	// refresh on changed line works like a charm
	app.LineTag.OnChanged = func(s string) {
		if app.Line != nil {
			app.Line.Tag = s
			app.ListLines.Refresh() // real time update
			app.Lines.ParseWholeTags()
		}
	}
	// on change
	app.LineCases = widget.NewMultiLineEntry()
	app.LineCases.SetPlaceHolder("Use , as separator")
	app.LineCases.Wrapping = fyne.TextWrapOff
	// Real con change
	app.LineCases.OnChanged = func(s string) {
		s = strings.Replace(s, "\n", "", -1)
		s = strings.Replace(s, " ", "", -1)
		app.Lines.ParseWholeTags()
		lst := strings.Split(s, ",")
		app.Line.CasesList = lst
	}

	multiLineValidator := func(s string) error {
		entrys := strings.Split(s, ",")
		for i := 0; i < len(entrys); i++ {
			entrys[i] = strings.TrimSpace(entrys[i])
		}
		cases := app.Cases.GetCasesList()
		for i := 0; i < len(entrys); i++ {
			if !model.SliceHas(cases, entrys[i]) {
				return errors.New(" ")
			}
		}
		return nil
	}
	app.LineCases.Validator = multiLineValidator
	toolbarLines := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			// Make an empty Lines and add
			// fixme i cannot add more than one
			l := model.Line{Tag: "X-XX00X-CAXX-B"}
			app.Lines.AddLine(&l)
			err := app.Lines.Check()
			if err != nil {
				dialog.ShowError(err, win)
			}

			//nroOfItems := len(app.Lines.Line)
			//fmt.Println(nroOfItems)
			//app.setPipe(app.Lines.Line[nroOfItems-1])
		}),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			// TODO DEL select Line
			if len(app.Lines.Line) > 1 {
				tagLineToRemove := app.Line.Id
				app.Lines.RemoveLine(tagLineToRemove)
				nroOfItems := len(app.Lines.Line)
				app.setPipe(app.Lines.Line[nroOfItems-1])
				app.ListLines.Refresh()
			}
		}),
	)
	app.ResultsEntryLines = widget.NewMultiLineEntry()

	// app.ListCases On Select
	toolbarCases := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			c := model.Case{
				CaseID: "Case X",
			}
			app.Cases.AddCase(&c)
			//nroOfItems := len(app.Cases.Cas)
			//fmt.Println(nroOfItems)
			//app.setCase(app.Cases.Cas[nroOfItems-1])
		}),
		widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
			tagCaseToRemove := app.Case.CaseID
			if len(app.Cases.Cas) > 1 {
				app.Cases.RemoveCase(tagCaseToRemove)
				nroOfItems := len(app.Cases.Cas)
				app.setCase(app.Cases.Cas[nroOfItems-1])
				app.ListCases.Refresh()
			}
		}),
	)

	app.ListCases.OnSelected = func(id int) {
		// display the case selected
		//
		app.setCase(app.Cases.Cas[id])
	}

	//------------------------------------------------------------------------//
	// for Cases
	app.CaseTag = widgetNewEntry("Case N°")
	app.CaseTag.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.CaseID = s
			app.ListCases.Refresh()
		}
	}
	app.GasFlow = widgetNewEntry("500")
	app.GasFlow.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.GasFlow = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.Press = widgetNewEntry("1")
	app.Press.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.Pressure = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.Temp = widgetNewEntry("15")
	app.Temp.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.Temperature = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.MW = widgetNewEntry("18")
	app.MW.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.MW = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.Z = widgetNewEntry("0.9")
	app.Z.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.Z = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}

	app.MultyFlow = widget.NewRadioGroup([]string{"True", "False"}, func(string) {})
	app.LlFlow = widgetNewEntry("1")
	app.LlFlow.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.LightLiquidFlow = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.HlFlow = widgetNewEntry("1")
	app.HlFlow.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.HeavyLiquidFlow = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.LlDens = widgetNewEntry("650")
	app.LlDens.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.LightLiquidDens = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	app.HlDens = widgetNewEntry("1000")
	app.HlDens.OnChanged = func(s string) {
		if app.Case != nil {
			app.Case.HeavyLiquidDens = StringToFloat64(s)
			app.ListCases.Refresh()
		}
	}
	// setting from to croma and his functions on change
	app.CC1 = widgetNewEntry("100")
	app.CC1.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CC2 = widgetNewEntry("0")
	app.CC2.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CC3 = widgetNewEntry("0")
	app.CC3.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CiC4 = widgetNewEntry("0")
	app.CiC4.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CnC4 = widgetNewEntry("0")
	app.CnC4.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CiC5 = widgetNewEntry("0")
	app.CiC5.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CnC5 = widgetNewEntry("0")
	app.CnC5.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CC6 = widgetNewEntry("0")
	app.CC6.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CC7 = widgetNewEntry("0")
	app.CC7.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CC8 = widgetNewEntry("0")
	app.CC8.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CN2 = widgetNewEntry("0")
	app.CN2.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CCO2 = widgetNewEntry("0")
	app.CCO2.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CH2O = widgetNewEntry("0")
	app.CH2O.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CSH2 = widgetNewEntry("0")
	app.CSH2.OnChanged = func(s string) {
		if s != "" {
			app.sumCroma()
			app.CTotal.Refresh()
		}
		if s == "" {
			app.CTotal.SetText(s)
			app.CTotal.Refresh()
		}
	}
	app.CTotal = widgetNewEntry("0")

	app.ResultsEntryCases = widget.NewMultiLineEntry()
	app.ResultsEntryCases.Wrapping = fyne.TextWrapOff
	//		widget.NewFormItem("Cases", app.LineCases),
	//	)
	//	linesDetails.OnSubmit
	linesDetails := widget.NewForm(
		widget.NewFormItem("Tag", app.LineTag),
		widget.NewFormItem("Cases", app.LineCases),
		widget.NewFormItem("Results", app.ResultsEntryLines),
		widget.NewFormItem("", widget.NewButton("Calc", func() {
			//TODO TODO TODO
			//app.setFrmCasesToCases()
			app.setFrmLinesToLines()
			app.Lines.CalcAll(app.Cases)
			app.RefreshAll()
		})),
	)

	casesDetails := widget.NewForm(
		widget.NewFormItem("CaseID", app.CaseTag),
		widget.NewFormItem("Gas Flow [E3M3/D]", app.GasFlow),
		widget.NewFormItem("Press [kgf/cm2g]", app.Press),
		widget.NewFormItem("Temp [°C]", app.Temp),
		widget.NewFormItem("MW", app.MW),
		widget.NewFormItem("Z", app.Z),
		widget.NewFormItem("MultyFlow", app.MultyFlow),
		widget.NewFormItem("LlFlow [m3/D]", app.LlFlow),
		widget.NewFormItem("HlFlow [m3/D]", app.HlFlow),
		widget.NewFormItem("LlDens [kg/m3]", app.LlDens),
		widget.NewFormItem("HlDens [kg/m3]", app.HlDens),
		widget.NewFormItem("Results", app.ResultsEntryCases),
		widget.NewFormItem("", widget.NewButton("Calc", func() {
			// TODO El Nombre nada claro
			if app.CTotal.Text != "" {
				// TODO FIXME: must Be calculated
				// SET frm to Croma
				app.MW.SetText(app.CTotal.Text)
				app.Z.SetText(app.CTotal.Text)
			}
			app.setFrmCasesToCases()
			app.Cases.CalcAll()
			fmt.Println("Data refresh", app.ResultsEntryCases.Text)
			app.RefreshAll()
		})),
	)

	cromaDet := widget.NewForm(
		widget.NewFormItem("% or Mol Frac CH4", app.CC1),
		widget.NewFormItem("% or Mol Frac C2H6", app.CC2),
		widget.NewFormItem("% or Mol Frac C3H8", app.CC3),
		widget.NewFormItem("% or Mol Frac i-C4H10", app.CiC4),
		widget.NewFormItem("% or Mol Frac n-C4H10", app.CnC4),
		widget.NewFormItem("% or Mol Frac i-C5H12", app.CiC5),
		widget.NewFormItem("% or Mol Frac n-C5H12", app.CnC5),
		widget.NewFormItem("% or Mol Frac n-C6H14", app.CC6),
		widget.NewFormItem("% or Mol Frac n-C7H16", app.CC7),
		widget.NewFormItem("% or Mol Frac n-C8H18", app.CC8),
		widget.NewFormItem("% or Mol Frac N2", app.CN2),
		widget.NewFormItem("% or Mol Frac CO2", app.CCO2),
		widget.NewFormItem("% or Mol Frac H2O", app.CH2O),
		widget.NewFormItem("% or Mol Frac SH2", app.CSH2),
		widget.NewFormItem("Total % or Mol ", app.CTotal),
	)
	leftLines := container.NewBorder(toolbarLines, nil, app.ListLines, nil)
	middleLines := container.NewHBox(leftLines, linesDetails)

	leftCases := container.NewBorder(toolbarCases, nil, app.ListCases, nil)
	middleCases := container.NewHBox(leftCases, casesDetails, cromaDet)
	//	containerLines := container.NewBorder(toolbarLines, nil, hboxLines, nil) //app.ListLines, linesDetails)
	// containerLines := container.NewBorder(toolbarLines, nil, app.ListLines, linesDetails) //app.ListLines, linesDetails)
	//containerLines := container.New(
	//	layout.NewGridWrapLayout(fyne.NewSize(1064, 164)),
	//	container.NewBorder(toolbarLines, nil, app.ListLines, linesDetails),
	//)
	//
	//containerLines := container.New(layout.NewBorderLayout(toolbarLines, nil, middleLines, nil), toolbarLines, middleLines)
	containerLines := container.New(layout.NewBorderLayout(nil, nil, middleLines, nil), middleLines)

	// containerLines := container.NewBorder(toolbarLines, nil, app.ListLines, linesDetails)
	// containerLines := container.NewBorder(toolbarLines, nil, app.ListLines, nil)
	//containerCases := container.NewBorder(toolbarCases, nil, app.ListCases, casesDetails)
	//middleCase := container.NewHBox(app.ListCases, casesDetails)
	//containerCases := container.NewBorder(toolbarCases, nil, middleCase, nil)
	containerCases := container.New(layout.NewBorderLayout(nil, nil, middleCases, nil), middleCases)
	// containerCases := container.NewBorder(toolbarCases, nil, app.ListCases, nil)

	tabs := container.NewAppTabs(
		container.NewTabItem("Lines", containerLines),
		container.NewTabItem("Cases", containerCases),
	)
	return container.NewVBox(toolbar, tabs)
}

func (app *chengPipesApp) RefreshLines() {
	if app.LineTag.Text != "" {

		linesLst := app.Lines.GetLineList()
		idx := model.SliceItemIdx(linesLst, app.Line.Tag) // Get index of the actual case
		// Set the actual case
		fmt.Println("idx", idx)
		app.LineTag.SetText(app.Lines.Line[idx].Tag)
		app.LineCases.SetText(strings.Join(app.Lines.Line[idx].CasesList, ", "))
		app.ResultsEntryLines.SetText(strings.Join(app.Lines.Line[idx].Results, "\n"))
		app.ResultsEntryLines.Wrapping = fyne.TextWrapOff
	}
}

func (app *chengPipesApp) RefreshCases() {
	if app.CaseTag.Text != "" {
		casesLst := app.Cases.GetCasesList()
		idx := model.SliceItemIdx(casesLst, app.Case.CaseID)
		setFloat64Entry(app.GasFlow, app.Cases.Cas[idx].GasFlow)
		setFloat64Entry(app.Press, app.Cases.Cas[idx].Pressure)
		setFloat64Entry(app.Temp, app.Cases.Cas[idx].Temperature)
		setFloat64Entry(app.MW, app.Cases.Cas[idx].MW)
		setFloat64Entry(app.Z, app.Cases.Cas[idx].Z)
		setFloat64Entry(app.LlFlow, app.Cases.Cas[idx].LightLiquidFlow)
		setFloat64Entry(app.LlDens, app.Cases.Cas[idx].LightLiquidDens)
		setFloat64Entry(app.HlFlow, app.Cases.Cas[idx].HeavyLiquidFlow)
		setFloat64Entry(app.HlDens, app.Cases.Cas[idx].HeavyLiquidDens)
		app.ResultsEntryCases.SetText(app.Cases.Cas[idx].Results)
	}
}

func (app *chengPipesApp) RefreshAll() {
	// TODO FIXME Some errror when not selected any line
	fmt.Println("RefreshAll------FIXEM")
	app.RefreshLines()
	app.RefreshCases()
}

// This fuction creates an "instance" of model line for test
func dummyLines() *model.Lines {
	return &model.Lines{
		Line: []*model.Line{
			{Tag: "10-HG001-CA21-H"},
			{Tag: "8-HG002-CA21-H"},
			{Tag: "8-HG003-CA21-H"},
			{Tag: "8-HG004-CA21-H"},
		},
	}
}

// test to load the file
func loadLinesFile(filename string) *model.Lines {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		lines := emptyLines()
		return lines
	} else {
		// read our opened jsonFile as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)
		// TODO manage error
		// we initialize our lines array
		var lines *model.Lines
		json.Unmarshal(byteValue, &lines)
		return lines
	}
}

// This fuction creates an "instance" of model line for initialization
func emptyLines() *model.Lines {
	return &model.Lines{
		Line: []*model.Line{
			{},
		},
	}
}

func dummyCases() *model.CaseList {
	return &model.CaseList{
		Cas: []*model.Case{
			{CaseID: "1", GasFlow: 854, MW: 18, Pressure: 35, Temperature: 30, Multiflow: false},
			{CaseID: "2", GasFlow: 854, MW: 18, Pressure: 35, Temperature: 30, Multiflow: false},
			{CaseID: "3", GasFlow: 854, MW: 18, Pressure: 35, Temperature: 30, Multiflow: true, LightLiquidFlow: 30, HeavyLiquidFlow: 30, LightLiquidDens: 650, HeavyLiquidDens: 1000},
			{CaseID: "4", GasFlow: 854, MW: 18, Pressure: 35, Temperature: 30, Multiflow: true, LightLiquidFlow: 30, HeavyLiquidFlow: 30, LightLiquidDens: 650, HeavyLiquidDens: 1000},
			{CaseID: "5", GasFlow: 1000, MW: 18, Pressure: 35, Temperature: 30, Multiflow: true, LightLiquidFlow: 30, HeavyLiquidFlow: 30, LightLiquidDens: 650, HeavyLiquidDens: 1000},
		},
	}

}
func emptyCase() *model.CaseList {
	return &model.CaseList{
		Cas: []*model.Case{
			{},
		},
	}

}

func loadCaseFile(filename string) *model.CaseList {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return emptyCase()
	} else {
		// read our opened jsonFile as a byte array
		byteValue, _ := ioutil.ReadAll(jsonFile)
		//TODO Manage errors
		// we initialize our case array
		var cases *model.CaseList
		json.Unmarshal(byteValue, &cases)
		return cases
	}
}

func main() {
	//testLines()

	a := app.New()
	win := a.NewWindow("Cheng-pipes")
	var lines *model.Lines
	var cases *model.CaseList
	lines = emptyLines()
	cases = emptyCase()
	pipes := chengPipesApp{Lines: lines, Cases: cases}

	win.SetContent(pipes.makeUI(win))
	//win.Resize(fyne.NewSize(1000, 1000))
	win.ShowAndRun()
}
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func testLines() {
	lll := dummyLines()
	lll.ParseWholeTags()
	ccc := dummyCases()

	fmt.Println("lll", lll.Line[0].Tag)
	fmt.Println("ccc", ccc.Cas[0].CaseID)
	l1 := model.Line{Tag: "10-HG001-CA21-B"}
	l2 := model.Line{Tag: "10-HG002-CA21-B"}
	l3 := model.Line{Tag: "12-HG003-CA21-B"}
	l4 := model.Line{Tag: "14-HG004-CA21-B"}
	c1 := model.Case{CaseID: "Case 1"}
	c2 := model.Case{CaseID: "Case 2"}
	c3 := model.Case{CaseID: "Case 3"}
	c4 := model.Case{CaseID: "Case 4"}
	lines := model.Lines{}
	cases := model.CaseList{}
	cases.AddCase(&c3)
	cases.AddCase(&c4)
	cases.AddCase(&c2)
	cases.AddCase(&c1)
	lines.AddLine(&l1)
	lines.AddLine(&l2)
	lines.AddLine(&l2)
	lines.AddLine(&l3)
	lines.AddLine(&l4)
	lines.AddLine(&l4)
	lines.AddLine(&l4)
	// fmt.Println("l4", lines.l[4])
	// for i := range lines.l {
	// fmt.Println("l", i, lines.l[i])
	// }
	lines.AddCaseToLines("HG003", c1.CaseID)
	lines.AddCaseToLines("HG003", c3.CaseID)
	lines.AddCaseToLines("HG002", c2.CaseID)
	lines.AddCaseToLines("HG002", c1.CaseID)
	lines.AddCaseToLines("HG002", c4.CaseID)
	lines.AddCaseToLines("HG002", c2.CaseID)
	lines.AddCaseToLines("HG001", c2.CaseID)
	lines.AddCaseToLines("HG004", c2.CaseID)
	lines.AddCaseToLines("HG001", c2.CaseID)
	// lines.l[0].CasesList = []string{}
	// lines.l[0].c = CaseList{[]Case{&c1, &c2}}
	// lines.l[1].c = CaseList{[]Case{&c1, &c3}}
	// lines.l[2].c = CaseList{[]Case{&c4}}
	// lines.l[3].c = CaseList{[]Case{&c2}}
	// lines.l[0].c = append(lines.l[0].c, c1)
	// lines.parseWholeTags()
	lines.Check()
	lines.SaveJson("Lines.json")
	cases.SaveJson("Cases.json")
	fmt.Println("End of test")
	for i := 0; i < len(lines.Line); i++ {
		fmt.Println(lines.Line[i])
	}
}

func NewUriFromLocalPath() fyne.ListableURI {
	// well to start in the current directory its a little bit tricky
	// first we need to get the current path and prepend file:
	// if not the New uri will return nil and we will get an error
	path, _ := os.Getwd()
	path = fmt.Sprintf("file:%s", path)
	pathUri, _ := storage.ListerForURI(storage.NewURI(path))
	//exPath := fyne.ListableURI(uripath)
	return pathUri
}

func StringToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f

}
