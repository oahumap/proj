// Copyright (C) 2018, Michael P. Gerlek (Flaxen Consulting)
//
// Portions of this code were derived from the PROJ.4 software
// In keeping with the terms of the PROJ.4 project, this software
// is provided under the MIT-style license in `LICENSE.md` and may
// additionally be subject to the copyrights of the PROJ.4 authors.

package proj_test

import (
	"fmt"
	"testing"

	"github.com/oahumap/proj"
	"github.com/stretchr/testify/assert"
)

var inputA = []float64{
	-0.127758, 51.507351, // London
	2.352222, 48.856614, // Paris
	12.496366, 41.902783, // Rome
}
var inputB = []float64{
	-77.625583, 38.833846, // mpg
}

type testcase struct {
	proj4     string
	expectedA []float64
	expectedB []float64
}

var projStrings = map[string]string{
	"3395": "+proj=merc +lon_0=0 +k=1 +x_0=0 +y_0=0 +datum=WGS84",                            // TODO: support +units=m +no_defs
	"3857": "+proj=merc +a=6378137 +b=6378137 +lat_ts=0.0 +lon_0=0.0 +x_0=0.0 +y_0=0 +k=1.0", // TODO: support +units=m +nadgrids=@null +wktext +no_defs
	"4087": "+proj=eqc +lat_ts=0 +lat_0=0 +lon_0=0 +x_0=0 +y_0=0 +datum=WGS84",               // TODO: support +units=m +no_defs
}

var testcases = []testcase{
	{
		proj4: projStrings["3395"],
		expectedA: []float64{
			-14221.96, 6678068.96,
			261848.16, 6218371.80,
			1391089.10, 5117883.04,
		},
		expectedB: []float64{
			-8641240.37, 4671101.60,
		},
	},
	{
		proj4: projStrings["3857"],
		expectedA: []float64{
			-14221.96, 6711533.71,
			261848.16, 6250566.72,
			1391089.10, 5146427.91,
		},
		expectedB: []float64{
			-8641240.37, 4697899.31,
		},
	},
	{
		proj4: projStrings["4087"],
		expectedA: []float64{
			-14221.96, 5733772.09,
			261848.16, 5438693.39,
			1391089.10, 4664596.47,
		},
		expectedB: []float64{
			-8641240.37, 4322963.96,
		},
	},
}

func TestConvert(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range testcases {

		outputA, err := proj.Convert(tc.proj4, inputA)
		assert.NoError(err)

		outputB, err := proj.Convert(tc.proj4, inputB)
		assert.NoError(err)

		invA, err := proj.Inverse(tc.proj4, tc.expectedA)
		assert.NoError(err)

		invB, err := proj.Inverse(tc.proj4, tc.expectedB)
		assert.NoError(err)

		const tol = 1.0e-2

		for i := range tc.expectedA {
			tag := fmt.Sprintf("epsg:%s, input=A.%d", tc.proj4, i)
			assert.InDelta(tc.expectedA[i], outputA[i], tol, tag)
			assert.InDelta(tc.expectedA[i], outputA[i], tol, tag)
		}
		for i := range tc.expectedB {
			tag := fmt.Sprintf("epsg:%s, input=B.%d", tc.proj4, i)
			assert.InDelta(tc.expectedB[i], outputB[i], tol, tag)
			assert.InDelta(tc.expectedB[i], outputB[i], tol, tag)
		}

		for i := range tc.expectedA {
			tag := fmt.Sprintf("inverse: epsg:%s, input=A.%d", tc.proj4, i)
			assert.InDelta(invA[i], inputA[i], tol, tag)
		}

		for i := range tc.expectedB {
			tag := fmt.Sprintf("inverse: epsg:%s, input=B.%d", tc.proj4, i)
			assert.InDelta(invB[i], inputB[i], tol, tag)
		}
	}
}

func TestEnsureRaisedError(t *testing.T) {
	type testcase struct {
		op          string
		pt          []float64
		expectedErr string
		epsgCode    string
	}

	fn := func(tc testcase) func(t *testing.T) {
		return func(t *testing.T) {
			var err error

			if tc.op == "convert" {
				_, err = proj.Convert(tc.epsgCode, tc.pt)
			} else {
				_, err = proj.Inverse(tc.epsgCode, tc.pt)
			}

			if err == nil {
				t.Errorf("didn't get expected error: %v", tc.expectedErr)
				return
			}

			if err.Error() != tc.expectedErr {
				t.Errorf("error: %v not equal to expected error: %v", err.Error(), tc.expectedErr)
			}
		}
	}

	tests := map[string]testcase{
		"3857 out of bounds WGS84": {
			op:          "convert",
			epsgCode:    "3857",
			pt:          []float64{-180.0, 90.0},
			expectedErr: "tolerance condition error",
		},
		"4326 not supported as source srid": {
			op:          "convert",
			epsgCode:    "4326",
			pt:          []float64{0, 0},
			expectedErr: "epsg code is not a supported projection",
		},
		"convert bad point count": {
			op:          "convert",
			epsgCode:    "3395",
			pt:          []float64{-180.0, 90.0, 11.0},
			expectedErr: "input array of lon/lat values must be an even number",
		},
		"inverse bad point count": {
			op:          "inverse",
			epsgCode:    "3395",
			pt:          []float64{-180.0, 90.0, 11.0},
			expectedErr: "input array of x/y values must be an even number",
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func ExampleConvert() {

	var dd = []float64{
		-77.625583, 38.833846,
	}

	xy, err := proj.Convert("3395", dd)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%.2f, %.2f\n", xy[0], xy[1])
	// Output: -8641240.37, 4671101.60
}

// func TestIsSupported(t *testing.T) {
// 	proj.IsProjectionSupported(28992)  // Should be supported
// 	proj.IsProjectionSupported(999999) // Should not be supported
// }

// func TestConvertProj4String(t *testing.T) {
// 	output, err := proj.ConvertProj4String(2154, 2.352222, 48.856614)
// 	assert.NoError(t, err)
// 	assert.InDelta(t, 102970.56, output.X, 0.01)
// 	assert.InDelta(t, 6864593.23, output.Y, 0.01)
// }
