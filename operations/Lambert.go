// Copyright (C) 2018, Michael P. Gerlek (Flaxen Consulting)
//
// Portions of this code were derived from the PROJ.4 software
// In keeping with the terms of the PROJ.4 project, this software
// is provided under the MIT-style license in `LICENSE.md` and may
// additionally be subject to the copyrights of the PROJ.4 authors.

package operations

import (
	"math"

	"github.com/oahumap/proj/core"
	"github.com/oahumap/proj/merror"
	"github.com/oahumap/proj/support"
)

func init() {
	core.RegisterConvertLPToXY("lcc",
		"Lambert Conic Conformal (LCC)",
		"\n\tMisc Sph, no inv.\n\tno_cut lat_b=",
		NewLCC,
	)
}

const LCCIterationEpsilon = 1e-18

// LCC implements core.IOperation and core.ConvertLPToXY
type LCC struct {
	core.Operation

	n       float64 // scale factor of the cone
	F       float64 // cone constant
	rho0    float64 // radius at the origin parallel
	lambda0 float64 // longitude of origin
	phi0    float64 // latitude of origin
	phi1    float64 // first standard parallel
	phi2    float64 // second standard parallel
	x0      float64 // offset X
	y0      float64 // offset Y
}

// NewLCC returns a new LCC
func NewLCC(system *core.System, desc *core.OperationDescription) (core.IConvertLPToXY, error) {
	op := &LCC{}
	op.System = system

	err := op.lccSetup(system)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// Forward Operation
func (op *LCC) Forward(lp *core.CoordLP) (*core.CoordXY, error) {
	xy := &core.CoordXY{X: 0.0, Y: 0.0}
	var rho float64

	t := support.Tsfn(lp.Phi, math.Sin(lp.Phi), op.System.Ellipsoid.E)
	rho = op.F * math.Pow(t, op.n)

	xy.X = rho * math.Sin(op.n*(lp.Lam))
	xy.Y = op.rho0 - (rho * math.Cos(op.n*(lp.Lam)))

	return xy, nil
}

// Inverse Operation
func (op *LCC) Inverse(xy *core.CoordXY) (*core.CoordLP, error) {
	deltaE := xy.X
	deltaN := op.rho0 - xy.Y

	rPrime := math.Sqrt(deltaE*deltaE + deltaN*deltaN)
	if op.n < 0 {
		rPrime = -rPrime
	}

	tPrime := math.Pow(rPrime/op.F, 1.0/op.n)
	thetaPrime := math.Atan2(deltaE, deltaN)

	lon := (thetaPrime / op.n) - op.lambda0

	lat := math.Pi/2.0 - 2*math.Atan(tPrime)
	for range 10 { // 10 iterations limit for safety
		latNew := math.Pi/2.0 - 2*math.Atan(tPrime*math.Pow((1.0-op.System.Ellipsoid.E*math.Sin(lat))/(1.0+op.System.Ellipsoid.E*math.Sin(lat)), op.System.Ellipsoid.E/2.0))

		if math.Abs(latNew-lat) < LCCIterationEpsilon {
			lat = latNew
			break
		}
		lat = latNew
	}

	return &core.CoordLP{Phi: lat, Lam: lon}, nil
}

func (op *LCC) lccSetup(system *core.System) error {
	phi0, ok0 := system.ProjString.GetAsFloat("lat_0")
	if !ok0 {
		return merror.New(merror.InvalidArg)
	}
	phi1, ok1 := system.ProjString.GetAsFloat("lat_1")
	if !ok1 {
		return merror.New(merror.InvalidArg)
	}
	phi2, ok2 := system.ProjString.GetAsFloat("lat_2")
	if !ok2 {
		phi2 = phi1
	}
	lambda0, ok3 := system.ProjString.GetAsFloat("lon_0")
	if !ok3 {
		return merror.New(merror.InvalidArg)
	}
	x0, ok4 := system.ProjString.GetAsFloat("x_0")
	if !ok4 {
		x0 = 0.0
	}
	y0, ok5 := system.ProjString.GetAsFloat("y_0")
	if !ok5 {
		y0 = 0.0
	}

	op.phi0 = support.DDToR(phi0)
	op.phi1 = support.DDToR(phi1)
	op.phi2 = support.DDToR(phi2)
	op.lambda0 = support.DDToR(lambda0)
	op.x0 = x0
	op.y0 = y0

	PE := system.Ellipsoid

	m1 := support.Msfn(math.Sin(op.phi1), math.Cos(op.phi1), PE.Es)
	t1 := support.Tsfn(op.phi1, math.Sin(op.phi1), PE.E)
	m2 := support.Msfn(math.Sin(op.phi2), math.Cos(op.phi2), PE.Es)
	t2 := support.Tsfn(op.phi2, math.Sin(op.phi2), PE.E)
	op.n = math.Log(m1/m2) / math.Log(t1/t2)

	op.F = m1 / (op.n * math.Pow(t1, op.n))

	t0 := support.Tsfn(op.phi0, math.Sin(op.phi0), PE.E)
	op.rho0 = op.F * math.Pow(t0, op.n)

	return nil
}
