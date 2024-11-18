package conf

type ExifToolInfos []ExifToolInfo

type ExifToolInfo struct {
	SourceFile          string `json:"SourceFile,omitempty"`
	ImageDataHash       string `json:"ImageDataHash,omitempty"`
	Title               string `json:"Title,omitempty"`
	ObjectName          string `json:"ObjectName,omitempty"`
	ContentLocationName string `json:"ContentLocationName,omitempty"`
	CaptionAbstract     string `json:"Caption-Abstract,omitempty"`
	ImageDescription    string `json:"ImageDescription,omitempty"`
	Description         string `json:"Description,omitempty"`
	City                string `json:"City,omitempty"`
	ProvinceState       string `json:"Province-State,omitempty"`
	State               string `json:"State,omitempty"`
	Country             string `json:"Country,omitempty"`
	CountryCode         string `json:"CountryCode,omitempty"`
}
