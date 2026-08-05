package usgs

// FeatureCollection is the top-level USGS GeoJSON response.
type FeatureCollection struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
	Geometry   Geometry   `json:"geometry"`
}

type Properties struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Place   string   `json:"place"`
	MagType string   `json:"magType"`
	Mag     *float64  `json:"mag"`
	Tsunami int      `json:"tsunami"`
}

// Coordinates is [longitude, latitude, depth].
type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
}
