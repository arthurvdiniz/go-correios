package gateway

import "github.com/arthurvdiniz/go-correios/entity"

type CorreiosGateway interface {
	GetTrackerCodeContent(box *entity.Box) error
}