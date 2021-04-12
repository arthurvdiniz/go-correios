package gocorreios

import (
	"github.com/arthurvdiniz/go-correios/entity"
	"github.com/arthurvdiniz/go-correios/gateway"
)

const (
	ScraperTrackerURL = "https://www2.correios.com.br/sistemas/rastreamento/resultado.cfm"
)

type ParcelUseCase struct {
	method string
	gateway.CorreiosGateway
}

func Get(code string, method string) (*entity.Box, error) {

	uc := ParcelUseCase{}

	uc.method = method
	if method == "" {
		uc.method = "scraper"
	}

	switch uc.method {
	case "scraper":
		uc.CorreiosGateway = &gateway.CorreiosScraperGateway{TrackerURL: ScraperTrackerURL}
	}

	box := &entity.Box{
		Code: code,
		PostDate: "",
		Events: []entity.Event{},
	}

	err := uc.CorreiosGateway.GetTrackerCodeContent(box)
	if err != nil {
		return nil, err
	}

	return box, nil
}