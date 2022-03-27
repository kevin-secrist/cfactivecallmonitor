package chesterfield

// GET https://api.chesterfield.gov/api/Police/V1.0/Traffic
// GET https://api.chesterfield.gov/api/Police/V1.0/Traffic/Henrico
// GET https://api.chesterfield.gov/api/Police/V1.0/Traffic/Richmond
type TrafficIncident []struct {
	Location  string `json:"location"`
	Direction string `json:"direction"`
	Status    string `json:"status"`
	Incident  string `json:"incident"`
	Type      string `json:"type"`
	Lon       string `json:"Lon"`
	Lat       string `json:"Lat"`
}
