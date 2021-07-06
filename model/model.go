// Package main provides the structs of our data
package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"
)

var lineDict = map[string]string{
	"HG": "Gaseous HC",
	"GH": "Gaseous HC",
	"MF": "Multy Phase",
	"HL": "Liquid HC",
	"LH": "Liquid HC",
	"VG": "Gaseous Vent",
	"R":  "Relief",
	"OD": "Open Drain",
	"DR": "Drain",
	"LO": "Lub Oil",
	"PF": "Process Fluid",
	"GL": "Lub Oil",
	"IA": "Instrument Air",
	"AI": "Instrument Air",
	"GI": "Instrument Gas",
	"IG": "Instrument Gas",
	"FG": "Fuel Gas",
}

// Case struct, defines the data form of a new calculation for each pipe
type Case struct {
	// TODO: Add Unit
	// CaseID just the name
	CaseID string `json:"CaseID"`
	// Gas flow rate in std conditions E3M3/D to get SM3/D mult by 1000
	GasFlow float64 `json:"GasFlow"`
	// Gas Molecular Weight in g/mol
	MW float64 `json:"MW"`
	// Comp factor
	Z float64 `json:"Z"`
	// Pressure in kgf/cm2(g) and Temp in C
	Pressure    float64 `json:"Pressure"`
	Temperature float64 `json:"Temperature"`
	// True or False in case of multi flow
	Multiflow bool `json:"Multiflow"`
	// Different liquid in feed line in M3/d
	LightLiquidFlow float64 `json:"LLFlow"`
	HeavyLiquidFlow float64 `json:"HLFlow"`
	// Different liquid densities kg/m3
	LightLiquidDens float64 `json:"LLDens"`
	HeavyLiquidDens float64 `json:"HLDens"`
	//------------------------------------------------------------------------//
	// To Be populated by calc
	// Max vel according to API 14E
	MaxVel14E float64 `json:"MaxVel14E"`
	// Max Vel Momentum
	MaxVelMomentum float64 `json:"MaxVelMomentum"`
	// Max Vel 18m/s
	MaxVel float64 `json:"MaxVel"`
	// Actual Flow
	ActFlow float64 `json:"ActGasFlow"`
	// Total Flow
	TotalFlow float64 `json:"TotalFlow"`
	// Required Area 14E
	ReqArea14E float64 `json:"ReqArea14E"`
	// Required Area Momentum
	ReqAreaMomentum float64 `json:"ReqAreaMomentum"`
	// Required Area 18ms
	ReqArea float64 `json:"ReqArea"`
	// Required Diam
	ReqDiam float64 `json:"ReqDiam"`
	// Required Diam
	ReqDiam14E float64 `json:"ReqDiam14E"`
	// Required Diam
	ReqDiamMomentum float64 `json:"ReqDiamMomentum"`
	//Density
	DensActGas float64 `json:"DensactGas"`
	//Density
	DensMix float64 `json:"DensMix"`
	// Saves the resutls to a case string
	Results string `json:"Results"`
}

// TODO Refactor
func (c *Case) Calc() error {
	c.ActFlow = ConvFlowE3M3dToActInKgfcm2(c.Pressure, c.Temperature, c.GasFlow)
	c.DensActGas = DensReal(c.Pressure, c.Temperature, c.MW, 1)
	if c.Multiflow {
		fl := []float64{c.ActFlow, c.LightLiquidFlow / 24 / 60 / 60, c.HeavyLiquidFlow / 24 / 60 / 60}
		ds := []float64{c.DensActGas, c.LightLiquidDens, c.HeavyLiquidDens}
		densMix, err := DensMed(fl, ds)
		if err != nil {
			return err
		}
		c.DensMix = densMix
		c.TotalFlow = c.ActFlow + c.LightLiquidFlow/24/60/60 + c.HeavyLiquidFlow/24/60/60
	} else {
		c.DensMix = c.DensActGas
		c.TotalFlow = c.ActFlow
	}
	fmt.Println(c.ActFlow)
	c.MaxVel = 18.0
	c.MaxVelMomentum = VelMaxJmomentum(c.DensMix)
	c.MaxVel14E = MaxVelAPI14E(c.DensMix)
	c.ReqArea14E = PipeReqArea(c.TotalFlow, c.MaxVel14E)
	c.ReqAreaMomentum = PipeReqArea(c.TotalFlow, c.MaxVelMomentum)
	c.ReqArea = PipeReqArea(c.TotalFlow, c.MaxVel)
	c.ReqDiam = DiamFromA(c.ReqArea)
	c.ReqDiam14E = DiamFromA(c.ReqArea14E)
	c.ReqDiamMomentum = DiamFromA(c.ReqAreaMomentum)

	// TODO Refactor a LOT
	fmt.Println("-----RESU")
	//temp1 := template.New("result")
	//temp1, _ = temp1.Parse("Actual Gas Flow {{.ActFlow}}")
	temp1 := template.Must(template.ParseFiles("model/tempCases.gohtml"))
	// var temp2 strin

	var tpl bytes.Buffer
	err2 := temp1.Execute(&tpl, c)
	if err2 != nil {
		fmt.Println(err2)
	}
	// TODO DOC
	c.Results = tpl.String()
	return nil
}

// CaseList struct, define the list of all cases
type CaseList struct {
	Cas []*Case
}

// Calcs for each in CaseList.Cas
func (cs *CaseList) CalcAll() {
	for i := 0; i < len(cs.Cas); i++ {
		cs.Cas[i].Calc()
	}
}

// Add a case to the Case list
func (cs *CaseList) AddCase(cas *Case) {
	if cs.Cas == nil {
		cs.Cas = append(cs.Cas, cas)
	}
	if cs.Cas != nil {
		// first populate a slice of line names
		sliceId := []string{}
		for id := 0; id < len(cs.Cas); id++ {
			sliceId = append(sliceId, cs.Cas[id].CaseID)
		}
		// if case is not in the case list add
		if !SliceHas(sliceId, cas.CaseID) {
			cs.Cas = append(cs.Cas, cas)
		}
	}

}

// Remove case from Caselist
func (cs *CaseList) RemoveCase(caseID string) {
	// Make a new empty list to overwrite
	newList := &CaseList{Cas: []*Case{}}
	if cs.Cas != nil {
		// first populate a slice of line names
		sliceId := []string{}
		for id := 0; id < len(cs.Cas); id++ {
			sliceId = append(sliceId, cs.Cas[id].CaseID)
		}
		// If the case is in remove if not do nothing
		if SliceHas(sliceId, caseID) {
			for id := 0; id < len(cs.Cas); id++ {
				if caseID != cs.Cas[id].CaseID {
					if newList != nil {
						newList.Cas = append(newList.Cas, cs.Cas[id])
					}
				}
			}
			// Then replace the list with the new list must be pointer to pointer
			*cs = *newList
		}
	}
}

// Save the struct to json file
func (cs *CaseList) SaveJson(filename string) {
	data, _ := json.MarshalIndent(cs, "", " ")
	ioutil.WriteFile(filename, data, 0644)

}

// Return the currect Case List
func (cs *CaseList) GetCasesList() []string {
	sl := []string{}
	for i := 0; i < len(cs.Cas); i++ {
		sl = append(sl, cs.Cas[i].CaseID)
	}
	return sl
}

// Line struct, its has a Tag and a case list.
type Line struct {
	// Line Tag
	Tag string `json:"Tags"`
	// Line Diam must be populated with parseTag
	Diam float64 `json:"Diam"`
	// Id must be populated with parseTag
	Id string `json:"Id"`
	// FluidType must be populated with parseTag
	FluidType string `json:"FluidType"`
	// Class of mat must be populated with parseTag
	ClassMat string `json:"ClassMat"`
	// Class Press must be populated with parseTag
	ClassPres int `json:"ClassPres"`
	// Corrosion Allowance must be populated with parseTag
	CorrotionAl int `json:"CorrotionAl"`
	// Insulation bare or traced or insulated
	Insulation string `json:"Insulation"`
	// Case list populated with must be populated with parseTag
	CasesList []string
	// Straight line pipe to calculate pressure lost
	StrightLine float64
	// Actual velocities for each case
	Velocities map[string]float64
	// Actual thov2 for each case
	RhoV2 map[string]float64
	// if verifies diameter
	CheckDiam map[string]bool
	// Result string to Print
	Results      []string `json:"Results"`
	ResultsCases []string `json:"ResultCases"`
}

// Method to fill cases list
func (l *Line) AddCase(simpleCaseId string) {
	if l.CasesList == nil {
		l.CasesList = append(l.CasesList, simpleCaseId)
		// i Have to initialize the map
		l.Velocities = make(map[string]float64)
		l.RhoV2 = make(map[string]float64)
		l.CheckDiam = make(map[string]bool)
	}
	if l.CasesList != nil {
		// First Populate a list of cases to compare
		sl := l.GetCases()
		// Then make the comparission
		if !SliceHas(sl, simpleCaseId) {
			l.CasesList = append(l.CasesList, simpleCaseId)
		}
	}
}

// TODO: improve DOC Method to populate id, FluidType ...
// TODO: Handle Errors
func (l *Line) ParseTag() {
	// fmt.Println("l", l.Tag)
	/// Split in - - - Parts, then remove "
	tag := strings.Split(l.Tag, "-")
	v, _ := strconv.ParseFloat(strings.Replace(tag[0], "\"", "", -1), 64)
	l.Diam = v
	if len(tag) == 4 {
		l.Id = tag[1]
		if len(l.Id) == 5 {
			// This if is to match HG0001
			l.FluidType = tag[1][:2]
			l.ClassMat = tag[2][:2]
		}
		// This if is fot CA21
		if len(tag[2]) == 4 {
			presval, _ := strconv.Atoi(tag[2][2:3])
			l.ClassPres = presval
			presca, _ := strconv.Atoi(tag[2][3:])
			l.CorrotionAl = presca
		}
	}
}

// Get slice of cases
func (l *Line) GetCases() []string {
	sl := []string{}
	// First Populate a list of cases to compare
	for i := 0; i < len(l.CasesList); i++ {
		sl = append(sl, l.CasesList[i])
	}
	return sl
}

// Lines struct
type Lines struct {
	Line []*Line `json:"Line"`
}

// Method to parse the whole tags
func (ls *Lines) ParseWholeTags() {
	// Run parse tag for each
	for i := 0; i < len(ls.Line); i++ {
		ls.Line[i].ParseTag()
	}

}

// Method to check repeated tags
func (ls *Lines) Check() error {
	// we make a slice of tags and pass to the dupes function
	// list.parseWholeTags()
	id := []string{}
	for i := 0; i < len(ls.Line); i++ {
		id = append(id, ls.Line[i].Id)
	}

	dupes := dup_count(id)
	fmt.Println("DUPES", dupes)

	for in, val := range dupes {
		if val > 1 {
			// TODO Esto puede ser un error a dialog
			err := fmt.Sprintf("Please FIX duplicated line tag %v, %v", in, val)
			return errors.New(err)
		}
	}
	return nil
}

// Method to add a case to list of strings of the line
func (ls *Lines) AddCaseToLines(lineId, caseTag string) {
	for id := 0; id < len(ls.Line); id++ {
		if lineId == ls.Line[id].Id {
			ls.Line[id].AddCase(caseTag)
		}
	}
}

// Method to simply append a line
func (ls *Lines) AddLine(simpleLine *Line) {
	// TODO DOC: make a documentatiion
	simpleLine.ParseTag()
	//fmt.Println("Inside Add Line To check", simpleLine.Tag)
	if ls.Line == nil {
		ls.Line = append(ls.Line, simpleLine)
	} else {

		sliceId := []string{}
		// first populate a slice of line names
		for id := 0; id < len(ls.Line); id++ {
			sliceId = append(sliceId, ls.Line[id].Id)
		}
		// Then compare
		if !SliceHas(sliceId, simpleLine.Id) {
			ls.Line = append(ls.Line, simpleLine)
		}
		//	fmt.Println(sliceId)
	}
}

// Method to remove a line tag
func (ls *Lines) RemoveLine(lineID string) {
	// TODO: DOC
	//var newList *Lines
	newList := &Lines{Line: []*Line{}}
	fmt.Println("Inside Remove")
	if ls.Line != nil {

		sliceId := []string{}
		// first populate a slice of line names
		for id := 0; id < len(ls.Line); id++ {
			sliceId = append(sliceId, ls.Line[id].Id)
		}
		// Then compare
		if SliceHas(sliceId, lineID) {
			for id := 0; id < len(ls.Line); id++ {

				fmt.Println(id, ls.Line[id].Id)
				if lineID != ls.Line[id].Id {
					if newList != nil {
						newList.Line = append(newList.Line, ls.Line[id])
						//	newList.Line = append(newList.Line, list.Line[id])
					}
				}
			}
			// Then replace the list with the new list must be pointer to pointer
			*ls = *newList
		}
	}
}

// Save the lines to json
func (ls *Lines) SaveJson(filename string) {
	data, _ := json.MarshalIndent(ls, "", " ")
	ioutil.WriteFile(filename, data, 0644)

}

func (ls *Lines) CalcAll(cs *CaseList) {
	for i := 0; i < len(ls.Line); i++ {
		ls.Line[i].Calc(cs)
	}
}

// Return the Line List
func (ls *Lines) GetLineList() []string {
	sl := []string{}
	for i := 0; i < len(ls.Line); i++ {
		sl = append(sl, ls.Line[i].Tag)
	}
	return sl
}

type ResultsLines struct {
	Tag  string
	Case string
	Vel  float64
	Rv   float64
}

// TODO DOC
func (l *Line) Calc(c *CaseList) {
	fmt.Println("Size of Caseslst", l.CasesList, len(l.CasesList))
	l.Results = []string{}
	l.ResultsCases = []string{}
	l.Velocities = make(map[string]float64)
	l.RhoV2 = make(map[string]float64)
	fmt.Println("Vel", l.Velocities)
	if (l.CasesList[0]) != "" {
		totalCaseLst := c.GetCasesList()
		fmt.Println("Cases", l.CasesList)
		Area := PipeAreaFromIn(l.Diam)
		for i := 0; i < len(l.CasesList); i++ {
			idx := SliceItemIdx(totalCaseLst, l.CasesList[i])
			ActualVel := c.Cas[idx].TotalFlow / Area
			l.Velocities[c.Cas[idx].CaseID] = ActualVel
			l.RhoV2[c.Cas[idx].CaseID] = c.Cas[idx].DensMix * ActualVel * ActualVel

			res := ResultsLines{Tag: l.Tag,
				Case: c.Cas[idx].CaseID,
				Vel:  ActualVel,
				Rv:   c.Cas[idx].DensMix * ActualVel * ActualVel,
			}

			fmt.Println("Inside For", i, l.CasesList[i], c.Cas[idx].CaseID, l.Velocities)
			temp1 := template.Must(template.ParseFiles("model/tempLines.gohtml"))
			// var temp2 strin

			var tpl bytes.Buffer
			err2 := temp1.Execute(&tpl, res)
			if err2 != nil {
				fmt.Println(err2)
			}
			// TODO DOC
			l.Results = append(l.Results, tpl.String())
			l.ResultsCases = append(l.ResultsCases, c.Cas[idx].Results)

		}
	}
}

// ------------------------------------------------------------------------------------

// Function to counr douplicated items
func dup_count(list []string) map[string]int {

	duplicate_frequency := make(map[string]int)
	for _, item := range list {
		// check if the item/element exist in the duplicate_frequency map
		_, exist := duplicate_frequency[item]
		if exist {
			duplicate_frequency[item] += 1 // increase counter by 1 if already in the map
		} else {
			duplicate_frequency[item] = 1 // else start counting from 1
		}
	}
	return duplicate_frequency
}

// https://stackoverflow.com/questions/15323767/does-go-have-if-x-in-construct-similar-to-python
/*
func (list StrSlice) Has(a string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func main() {
    var testList = StrSlice{"The", "big", "dog", "has", "fleas"}

    if testList.Has("dog") {
        fmt.Println("Yay!")
    }
}

*/
// Function to check if a slice has an item
func SliceHas(lst []string, match string) bool {
	// TODO make doc
	for _, val := range lst {
		if val == match {
			return true
		}
	}
	return false
}

func SliceItemIdx(lst []string, match string) int {
	for i := 0; i < len(lst); i++ {
		if lst[i] == match {
			return i
		}
	}
	return 0
}
