// Copyright (C) 2018, Michael P. Gerlek (Flaxen Consulting)
//
// Portions of this code were derived from the PROJ.4 software
// In keeping with the terms of the PROJ.4 project, this software
// is provided under the MIT-style license in `LICENSE.md` and may
// additionally be subject to the copyrights of the PROJ.4 authors.

package proj

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/oahumap/proj/core"
	"github.com/oahumap/proj/support"

	// need to pull in the operations table entries
	_ "github.com/oahumap/proj/operations"
)

// EPSGCode is the enum type for coordinate systems
type EPSGCode int

// Supported EPSG codes
const (
	EPSG3395                    EPSGCode = 3395
	WorldMercator                        = EPSG3395
	EPSG3857                    EPSGCode = 3857
	WebMercator                          = EPSG3857
	EPSG4087                    EPSGCode = 4087
	WorldEquidistantCylindrical          = EPSG4087
	EPSG4326                    EPSGCode = 4326
	WGS84                                = EPSG4326
)

// Convert performs a conversion from a 4326 coordinate system (lon/lat
// degrees, 2D) to the given projected system (x/y meters, 2D).
//
// The input is assumed to be an array of lon/lat points, e.g. [lon0, lat0,
// lon1, lat1, lon2, lat2, ...]. The length of the array must, therefore, be
// even.
//
// The returned output is a similar array of x/y points, e.g. [x0, y0, x1,
// y1, x2, y2, ...].
// If the proj4 string represents WGS84 or a geographic coordinate system,
// returns the input coordinates unchanged.
func Convert(proj4 string, input []float64) ([]float64, error) {
	if isGeographicSystem(proj4) {
		result := make([]float64, len(input))
		copy(result, input)
		return result, nil
	}

	conv, err := newConversion(proj4)
	if err != nil {
		return nil, err
	}

	return conv.convert(input)
}

// Inverse converts from a projected X/Y of a coordinate system to
// 4326 (lat/lon, 2D).
//
// The input is assumed to be an array of x/y points, e.g. [x0, y0,
// x1, y1, x2, y2, ...]. The length of the array must, therefore, be
// even.
//
// The returned output is a similar array of lon/lat points, e.g. [lon0, lat0, lon1,
// lat1, lon2, lat2, ...].
func Inverse(proj4 string, input []float64) ([]float64, error) {
	conv, err := newConversion(proj4)
	if err != nil {
		return nil, err
	}

	return conv.inverse(input)
}

type Projection struct {
	Code    string
	Name    string
	Proj4   string
	OGCWKT  string
	ESRIWKT string
}

// GetInfoFromEPSG retrieves the info for a given EPSG code from epsg.io.
// It validates also if the proj4 string is supported by the library.
func GetInfoFromEPSG(epsg string) (*Projection, error) {
	proj4Str, err := getFromEPSGAPI(epsg, "proj4")
	if err != nil {
		return nil, err
	}
	ps, err := support.NewProjString(proj4Str)
	if err != nil {
		return nil, err
	}

	ogcWKT, err := getFromEPSGAPI(epsg, "prettywkt")
	if err != nil {
		return nil, err
	}
	esriWKT, err := getFromEPSGAPI(epsg, "esriwkt")
	if err != nil {
		return nil, err
	}
	jsonStr, err := getFromEPSGAPI(epsg, "json")
	if err != nil {
		return nil, err
	}

	// Parse the JSON string into a map
	var jsonData map[string]any
	err = json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return nil, err
	}

	_, _, err = core.NewSystem(ps)
	if err != nil {
		return nil, err
	}

	// Extract name from JSON data
	name := ""
	if nameValue, ok := jsonData["name"]; ok {
		if nameStr, ok := nameValue.(string); ok {
			name = nameStr
		}
	}

	return &Projection{
		Code:    epsg,
		Name:    name,
		Proj4:   proj4Str,
		OGCWKT:  ogcWKT,
		ESRIWKT: esriWKT,
	}, nil
}

func getFromEPSGAPI(epsg, what string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://epsg.io/%s.%s", epsg, what))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("epsg %s not found", epsg)
	}
	str, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

// isGeographicSystem checks if a proj4 string represents a geographic coordinate system
func isGeographicSystem(proj4 string) bool {
	ps, err := support.NewProjString(proj4)
	if err != nil {
		return false
	}
	proj, _ := ps.GetAsString("proj")
	return proj == "longlat" || proj == "latlong" || proj == "latlon" || proj == "lonlat"
}

//---------------------------------------------------------------------------

// conversion holds the objects needed to perform a conversion
type conversion struct {
	projString *support.ProjString
	system     *core.System
	operation  core.IOperation
	converter  core.IConvertLPToXY
}

// newConversion creates a conversion object for the destination systems.
func newConversion(proj4 string) (*conversion, error) {
	ps, err := support.NewProjString(proj4)
	if err != nil {
		return nil, err
	}

	sys, opx, err := core.NewSystem(ps)
	if err != nil {
		return nil, err
	}

	if !opx.GetDescription().IsConvertLPToXY() {
		return nil, fmt.Errorf("projection type is not supported")
	}

	conv := &conversion{
		projString: ps,
		system:     sys,
		operation:  opx,
		converter:  opx.(core.IConvertLPToXY),
	}

	return conv, nil
}

// convert performs the projection on the given input points
func (conv *conversion) convert(input []float64) ([]float64, error) {
	if conv == nil || conv.converter == nil {
		return nil, fmt.Errorf("conversion not initialized")
	}

	if len(input)%2 != 0 {
		return nil, fmt.Errorf("input array of lon/lat values must be an even number")
	}

	output := make([]float64, len(input))

	lp := &core.CoordLP{}

	for i := 0; i < len(input); i += 2 {
		lp.Lam = support.DDToR(input[i])
		lp.Phi = support.DDToR(input[i+1])

		xy, err := conv.converter.Forward(lp)
		if err != nil {
			return nil, err
		}

		output[i] = xy.X
		output[i+1] = xy.Y
	}

	return output, nil
}

func (conv *conversion) inverse(input []float64) ([]float64, error) {
	if conv == nil || conv.converter == nil {
		return nil, fmt.Errorf("conversion not initialized")
	}

	if len(input)%2 != 0 {
		return nil, fmt.Errorf("input array of x/y values must be an even number")
	}

	output := make([]float64, len(input))

	xy := &core.CoordXY{}

	for i := 0; i < len(input); i += 2 {
		xy.X = input[i]
		xy.Y = input[i+1]

		lp, err := conv.converter.Inverse(xy)

		if err != nil {
			return nil, err
		}

		l, p := lp.Lam, lp.Phi

		output[i] = support.RToDD(l)
		output[i+1] = support.RToDD(p)
	}

	return output, nil
}
