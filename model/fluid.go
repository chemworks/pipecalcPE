package model

import (
	"errors"
	"fmt"
	"math"
)

const (
	R         float64 = 0.082 // atm/l/molK
	Ps        float64 = 1.0
	Ts        float64 = 273.15
	PconvFact float64 = 1.033
)

// Returns Press Abs in kgf/cm2A from kgf/cm2G
// Pr: Pressure gauge in kgf/cm2g
// Returns Pressure in kgf/cm2A
func PressAbs(Pr float64) float64 { return Pr + Ps }

// Returns Temp Abs in K from C
// Tr: Temperature in °C
// Returns Temperature in K
func TempAbs(Tr float64) float64 { return Tr + Ts }

// Calcs settle pressure for a system with diferrents pressures
// press slice of float64
// fractions each vol frac slice float64
// press: Pressures in kgf/cm2G
// fractions: vol fraction of each pressure stage
// Returns pressure in kgf/cm2G
func SettlePress(press, fractions []float64) (float64, error) {
	// Check if Press = Frac
	if len(press) != len(fractions) {
		return 0, errors.New("Length of press and fraction must match")
	}
	sum := 0.0
	if len(press) == len(fractions) {
		for i := 0; i < (len(press)); i++ {
			sum = sum + press[i]*fractions[i]
		}
	}
	return sum, nil
}

// Returns vel of sound in m/s
// k: Cp/Cv
// MW: Mol Weight
// TK: Temperature in Kelvins
// Returns Vel of sound in m/s
func VelSound(k, MW, TK float64) float64 {
	const R = 8314.0
	return math.Pow(k*R*TK/MW, 0.5)
}

// Returns the mass flow througt an orifice for gas flow
// TODO check units
// P: Pressure
// k:  Cp/Cv
// Mw: Mol Weight
// T: Temperature in K
// Returns TODO PERRY
func FlowOrifice(P, k, Mw, T float64) float64 {
	// All unit must be absolute
	const R = 8314.0
	a := (k + 1) / (k - 1)
	b := 2.0 / (k + 1)
	c1 := k * Mw / (R * T)
	c2 := math.Pow(b, a) * c1
	return (P) * 98066.5 * math.Pow(c2, 0.5)
}

// Returns  the density of the gas in kg/m3
// P: Pressure in ATM
// T: Temperature in K
// MW: Mol weight
// Returns Density in kg/Am3
func DensIdeal(P, T, MW float64) float64 {
	return (P * MW) / (R * T)
}

// Returns  the density of the gas in kg/m3
// P: Pressure in kgf/cm2G
// T: Temperature in K
// MW: Mol weight
// z: Comp Factor
// Returns Density in kg/Am3
func DensReal(Pr, Tr, MW, z float64) float64 {
	Pa := Pr/PconvFact + Ps
	Ta := Tr + 273.15
	return DensIdeal(Pa, Ta, MW) / z
}

// Returns  the density of the gas in kg/m3
// P: Pressure in kgf/cm2G
// T: Temperature in K
// MW: Mol weight
// z: Comp Factor
// Returns Density in kg/Am3
func Dens(Pr, Tr, MW, z float64) float64 {
	if z == 0 {
		return DensIdeal(Pr, Tr, MW)
	}
	if z > 1 {
		return DensIdeal(Pr, Tr, MW)
	}
	return DensReal(Pr, Tr, MW, z)
}

// Returns Actual gas flow in Am3/timeunit
// Pa: Pressure in kgf/cm2A
// Ta: Temperature in K
// Flows: Std Flow rate in Sm3/time
// Returns Am3/s
func ConvFlowStdToActInAbs(Pa, Ta, Flows float64) float64 {
	return (Ps / Pa) * (Ta / Ts) * (Flows)
}

// Returns actual gas flow in Am3
// Pr: Pressure in kgf/cm2G
// Tr: Temperature in °C
// Flows: Std Flow rate in Sm3/time
// Returns Am3/s
func ConvFlowStdToActInKgfcm2(Pr, Tr, Flows float64) float64 {
	Pa := Pr/PconvFact + Ps
	Ta := Tr + 273.15
	return ConvFlowStdToActInAbs(Pa, Ta, Flows)
}

// Returns actual gas flow in Am3
// Pr: Pressure in kgf/cm2G
// Tr: Temperature in °C
// Flows: Std Flow rate in E3M3/D
// Returns Am3/s
func ConvFlowE3M3dToActInKgfcm2(Pr, Tr, Flows float64) float64 {
	Flows = Flows * 1000 / (24 * 60 * 60)
	fmt.Println("Flow", Flows)
	return ConvFlowStdToActInKgfcm2(Pr, Tr, Flows)
}

// Returns the actual vel
// Flow: Flow in Am3/timeunit
// Area: Flow Area in m2
func Vel(Flow, Area float64) float64 {
	return Flow / Area
}

// Returns Momentum
// dens: Density kg/m3
// Vel: Velocity m/s
// Returns Momentum in kg/ms2 see GPSA, Chapter 7
func JMomentum(dens, vel float64) float64 {
	return dens * math.Pow(vel, 2)
}

// returns vel Max calculated by momentum
// dens: Density kg/m3
// Returns the velocity to reach a J of 3750
func VelMaxJmomentum(dens float64) float64 {
	return math.Pow(3750/dens, 0.5)
}

// Returns area from diameter
// id: Diam in m2
// Returns Area in m2
func PipeArea(id float64) float64 {
	// Help this func calculates the internal area from id in meters

	return math.Pi * (math.Pow(id, 2)) * 0.25

}

// Returns area from diameter
// id: Diam in inches
// Returns Area in m2
func PipeAreaFromIn(id float64) float64 {
	// This function calculates the inside area in square meters from inchs
	diam := id * 25.4 / 1000
	return PipeArea(diam)
}

// TODO:
func MinDiam14E(vel float64) float64 {
	//
	return 0
}

// Returns mean density from slices
// fl: flow in the same unit
// ds: Dens in kg/m3
// Returns Mean Density kg/m3
func DensMed(fl, ds []float64) (float64, error) {
	//ftotal := fd[0][0][0] + fd[0][0][1] + fd[0][0][2]
	//dens := fd[0][0][0]*fd[0][1][0] + fd[0][0][1]*fd[0][1][1] + fd[0][0][2]*fd[0][1][2]
	if len(fl) != len(ds) {
		return 0, errors.New("Length must match")
	}
	flTotal := 0.0
	for _, j := range fl {
		flTotal = flTotal + j
	}
	return (fl[0]*ds[0] + fl[1]*ds[1] + fl[2]*ds[2]) / flTotal, nil

}

// Returns pipe area with flow and velocity
// flow: total flow in Am3/s
// vel: velocity in m/s
// Returns area in m2
func PipeReqArea(flow, vel float64) float64 {
	return flow / vel
}

// Returns the diameter from area
// area: pipe area in m2
// Returns diam in inches
func DiamFromA(area float64) float64 {
	return (math.Pow(((4 * area) / math.Pi), 0.5)) * 1000 / 25.4
}

// Returns the max vel according to API14E
// densmedia: density of the fluid in kg/m3
// Returns velocity m/s
func MaxVelAPI14E(densmedia float64) float64 {
	vel14E := (100 / math.Pow(densmedia*0.062427961, 0.5)) * 0.3048
	return vel14E
}
