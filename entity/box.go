package entity

type Box struct {
	Code     string  `json:"code"`
	PostDate string  `json:"postDate"`
	Events   []Event `json:"events"`
}

