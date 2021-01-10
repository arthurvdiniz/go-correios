package gocorreios

type Box struct {
	Code     string  `json:"code"`
	PostDate string  `json:"postDate"`
	Events   []Event `json:"events"`
}

type Event struct {
	Date     string `json:"date"`
	Location string `json:"location"`
	Info     string `json:"info"`
}
