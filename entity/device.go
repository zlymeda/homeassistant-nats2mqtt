package entity

type Device struct {
	Id              string `json:"ids"`
	DisplayName     string `json:"name"`
	Name            string `json:"-"`
	Manufacturer    string `json:"mf"`
	Model           string `json:"mdl"`
	ModelId         string `json:"mdl_id"`
	SerialNumber    string `json:"sn"`
	SoftwareVersion string `json:"sw"`
	HardwareVersion string `json:"hw"`
}
