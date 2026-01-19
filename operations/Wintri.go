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
	core.RegisterConvertLPToXY("wintri",
		"Winkel Tripel",
		"\n\tPCyl., Sph.\n\tlat_1= (default: 50.467°)",
		NewWintri,
	)
}

// Wintri implements core.IOperation and core.ConvertLPToXY
type Wintri struct {
	core.Operation
	lat1    float64 // standard parallel (radiants)
	cosLat1 float64 // cosin of standard parallel
}

// NewWintri returns a new Winkel Tripel projection
func NewWintri(system *core.System, desc *core.OperationDescription) (core.IConvertLPToXY, error) {
	op := &Wintri{}
	op.System = system

	err := op.wintriSetup(system)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// Forward Operation
func (op *Wintri) Forward(lp *core.CoordLP) (*core.CoordXY, error) {
	var xy core.CoordXY

	lat := lp.Phi
	lon := lp.Lam

	x1 := lon * op.cosLat1
	y1 := lat

	var x2, y2 float64

	cosLat := math.Cos(lat)
	cosHalfLon := math.Cos(lon * 0.5)
	alpha := math.Acos(cosLat * cosHalfLon)

	if alpha < eps10 {
		x2 = lon
		y2 = lat
	} else {
		sinAlpha := math.Sin(alpha)
		if sinAlpha < eps10 {
			x2 = 0.0
			y2 = 0.0
		} else {
			factor := alpha / sinAlpha
			x2 = 2.0 * cosLat * math.Sin(lon*0.5) * factor
			y2 = math.Sin(lat) * factor
		}
	}

	xy.X = 0.5 * (x1 + x2)
	xy.Y = 0.5 * (y1 + y2)

	return &xy, nil
}

// Inverse Operation
func (op *Wintri) Inverse(xy *core.CoordXY) (*core.CoordLP, error) {
	var lp core.CoordLP

	x := xy.X
	y := xy.Y

	const maxIter = 30
	const tolerance = 1e-14

	phi := y
	lam := x / op.cosLat1

	if phi > math.Pi*0.5 {
		phi = math.Pi * 0.5
	} else if phi < -math.Pi*0.5 {
		phi = -math.Pi * 0.5
	}

	if lam > math.Pi {
		lam = math.Pi
	} else if lam < -math.Pi {
		lam = -math.Pi
	}

	for range maxIter {
		testLP := core.CoordLP{Phi: phi, Lam: lam}
		testXY, err := op.Forward(&testLP)
		if err != nil {
			return nil, err
		}

		dx := testXY.X - x
		dy := testXY.Y - y
		if math.Abs(dx) < tolerance && math.Abs(dy) < tolerance {
			break
		}

		if math.Abs(dx) > 10 || math.Abs(dy) > 10 {
			phi = y * 0.9
			lam = x * 0.9 / op.cosLat1
			continue
		}

		delta := math.Max(1e-8, math.Min(1e-6, math.Max(math.Abs(phi), math.Abs(lam))*1e-8))

		testLP1 := core.CoordLP{Phi: phi + delta, Lam: lam}
		testXY1, err1 := op.Forward(&testLP1)
		if err1 != nil {
			delta *= 0.5
			continue
		}
		dxdPhi := (testXY1.X - testXY.X) / delta
		dydPhi := (testXY1.Y - testXY.Y) / delta

		testLP2 := core.CoordLP{Phi: phi, Lam: lam + delta}
		testXY2, err2 := op.Forward(&testLP2)
		if err2 != nil {
			delta *= 0.5
			continue
		}
		dxdLam := (testXY2.X - testXY.X) / delta
		dydLam := (testXY2.Y - testXY.Y) / delta

		det := dxdPhi*dydLam - dydPhi*dxdLam
		if math.Abs(det) < 1e-15 {
			return nil, merror.New(merror.ToleranceCondition, "Jacobian determinant too small in Winkel Tripel inverse")
		}

		dphi := (dydLam*dx - dxdLam*dy) / det
		dlam := (dxdPhi*dy - dydPhi*dx) / det

		damping := 1.0
		if math.Abs(dphi) > 0.1 || math.Abs(dlam) > 0.1 {
			damping = 0.5
		}

		phi -= damping * dphi
		lam -= damping * dlam

		if phi > math.Pi*0.5 {
			phi = math.Pi * 0.5
		} else if phi < -math.Pi*0.5 {
			phi = -math.Pi * 0.5
		}

		for lam > math.Pi {
			lam -= 2 * math.Pi
		}
		for lam < -math.Pi {
			lam += 2 * math.Pi
		}
	}

	lp.Phi = phi
	lp.Lam = lam

	return &lp, nil
}

func (op *Wintri) wintriSetup(system *core.System) error {
	system.Ellipsoid.Es = 0.0
	op.lat1 = math.Acos(2.0 / math.Pi)

	if val, ok := system.ProjString.GetAsFloat("lat_1"); ok {
		op.lat1 = support.DDToR(val)
	}

	op.cosLat1 = math.Cos(op.lat1)

	if math.Abs(op.lat1) > math.Pi*0.5 {
		op.lat1 = math.Copysign(math.Pi*0.5, op.lat1)
		op.cosLat1 = math.Cos(op.lat1)
	}

	return nil
}
