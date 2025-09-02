// Copyright (C) 2018, Michael P. Gerlek (Flaxen Consulting)
//
// Portions of this code were derived from the PROJ.4 software
// In keeping with the terms of the PROJ.4 project, this software
// is provided under the MIT-style license in `LICENSE.md` and may
// additionally be subject to the copyrights of the PROJ.4 authors.

package operations_test

import (
	"fmt"
	"testing"

	"github.com/oahumap/proj/core"
	"github.com/oahumap/proj/support"
	"github.com/stretchr/testify/assert"
)

type data struct {
	proj  string
	delta float64
	fwd   [][]float64
	inv   [][]float64
}

var testdata = []data{
	{
		// builtins.gie:21
		proj:  "+proj=aea   +ellps=GRS80  +lat_1=0 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 222571.608757106, 110653.326743030},
			{2, -1, 222706.306508391, -110484.267144400},
			{-2, 1, -222571.608757106, 110653.326743030},
			{-2, -1, -222706.306508391, -110484.267144400},
		},
		inv: [][]float64{
			{200, 100, 0.001796631, 0.000904369},
			{200, -100, 0.001796630, -0.000904370},
			{-200, 100, -0.001796631, 0.000904369},
			{-200, -100, -0.001796630, -0.000904370},
		},
	}, {
		// builtins.gie:2317
		proj:  "+proj=leac +ellps=GRS80 +lat_1=0 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2.0, 1.0, 220685.140542979, 112983.500889396},
		},
		inv: [][]float64{
			{200, 100, 0.001796645, 0.000904352},
		},
	}, {
		// builtins.gie:1247
		proj:  "+proj=etmerc +ellps=GRS80 +ellps=GRS80 +lat_1=0.5 +lat_2=2 +n=0.5 +zone=30",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 222650.796797586, 110642.229411933},
		},
		inv: [][]float64{
			{200, 100, 0.001796631, 0.000904369},
		},
	}, {
		// builtins.gie:4684
		proj:  "+proj=utm +ellps=GRS80  +lat_1=0.5 +lat_2=2 +n=0.5 +zone=30",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 1057002.405491298, 110955.141175949},
		},
		inv: [][]float64{
			{200, 100, -7.486952083, 0.000901940},
		},
	}, {
		// builtins.gie:2626
		proj:  "+proj=merc   +ellps=GRS80  +lat_1=0.5 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 222638.981586547, 110579.965218250},
		},
		inv: [][]float64{
			{200, 100, 0.001796631, 0.000904369},
		},
	}, {
		// ellipsoid.gie:141
		proj:  "proj=utm zone=32   ellps=GRS80 b=6000000",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{12, 55, 699293.0880, 5674591.5295},
		},
	}, {
		// builtins.gie:309
		proj:  "+proj=airy   +a=6400000    +lat_1=0 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 189109.886908621, 94583.752387504},
		},
	}, {
		// builtins.gie:428
		proj:  "+proj=august   +a=6400000    +lat_1=0 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 223404.978180972, 111722.340289763},
		},
	}, {
		// builtins.gie:1104
		proj:  "proj=eqc   +a=6400000    +lat_1=0.5 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 223402.144255274, 111701.072127637},
			{2, -1, 223402.144255274, -111701.072127637},
			{-2, 1, -223402.144255274, 111701.072127637},
			{-2, -1, -223402.144255274, -111701.072127637},
		},
		inv: [][]float64{
			{200, 100, 0.001790493, 0.000895247},
			{200, -100, 0.001790493, -0.000895247},
			{-200, 100, -0.001790493, 0.000895247},
			{-200, -100, -0.001790493, -0.000895247},
		},
	}, {
		// builtins.gie:2251
		proj:  "+proj=lcc   +ellps=GRS80  +lat_1=0.5 +lat_2=2",
		delta: 0.1 * 0.001,
		fwd: [][]float64{
			{2, 1, 222588.439735968, 110660.533870800},
			{2, -1, 222756.879700279, -110532.797660827},
			{-2, 1, -222588.439735968, 110660.533870800},
			{-2, -1, -222756.879700279, -110532.797660827},
		},
		inv: [][]float64{
			{200, 100, 0.001796359, 0.000904232},
			{200, -100, 0.001796358, -0.000904233},
			{-200, 100, -0.001796359, 0.000904232},
			{-200, -100, -0.001796358, -0.000904233},
		},
	},
}

func TestConvert(t *testing.T) {
	assert := assert.New(t)

	for _, td := range testdata {

		ps, err := support.NewProjString(td.proj)
		assert.NoError(err)

		sys, opx, err := core.NewSystem(ps)
		assert.NoError(err)
		assert.NotNil(sys)
		assert.NotNil(opx)
		assert.EqualValues(sys, opx.GetSystem())

		op := opx.(core.IConvertLPToXY)

		for i, tc := range td.fwd {
			tag := fmt.Sprintf("%s (fwd/%d)", td.proj, i)
			input := &core.CoordLP{Lam: support.DDToR(tc[0]), Phi: support.DDToR(tc[1])}
			output, err := op.Forward(input)
			assert.NoError(err)

			x, y := output.X, output.Y
			assert.InDelta(tc[2], x, td.delta, tag)
			assert.InDelta(tc[3], y, td.delta, tag)
		}

		for i, tc := range td.inv {
			tag := fmt.Sprintf("%s (inv/%d)", td.proj, i)

			input := &core.CoordXY{X: tc[0], Y: tc[1]}
			output, err := op.Inverse(input)
			assert.NoError(err)

			l, p := output.Lam, output.Phi
			assert.InDelta(tc[2], support.RToDD(l), td.delta, tag)
			assert.InDelta(tc[3], support.RToDD(p), td.delta, tag)
		}
	}
}

func BenchmarkConvertEtMerc(b *testing.B) {

	ps, _ := support.NewProjString("+proj=utm +zone=32 +ellps=GRS80")
	_, opx, _ := core.NewSystem(ps)
	op := opx.(core.IConvertLPToXY)
	input := &core.CoordLP{Lam: support.DDToR(12.0), Phi: support.DDToR(55.0)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = op.Forward(input)
	}
}

func BenchmarkConvertAea(b *testing.B) {

	ps, _ := support.NewProjString("+proj=aea   +ellps=GRS80  +lat_1=0 +lat_2=2")
	_, opx, _ := core.NewSystem(ps)
	op := opx.(core.IConvertLPToXY)
	input := &core.CoordLP{Lam: support.DDToR(12.0), Phi: support.DDToR(55.0)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = op.Forward(input)
	}
}

func BenchmarkConvertAiry(b *testing.B) {

	ps, _ := support.NewProjString("+proj=airy   +a=6400000    +lat_1=0 +lat_2=2")
	_, opx, _ := core.NewSystem(ps)
	op := opx.(core.IConvertLPToXY)
	input := &core.CoordLP{Lam: support.DDToR(12.0), Phi: support.DDToR(55.0)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = op.Forward(input)
	}
}

func BenchmarkConvertEqc(b *testing.B) {
	ps, _ := support.NewProjString("+proj=eqc +a=6400000 +lat_1=0.5 +lat_2=2")
	_, opx, _ := core.NewSystem(ps)
	op := opx.(core.IConvertLPToXY)
	input := &core.CoordLP{Lam: support.DDToR(12.0), Phi: support.DDToR(55.0)}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = op.Forward(input)
	}
}
