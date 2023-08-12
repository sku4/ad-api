package model

import (
	"github.com/sku4/ad-parser/pkg/ad/model"
	"github.com/sku4/ad-parser/pkg/ad/street"
)

type StreetsExt struct {
	Streets []*model.StreetTnt
	Types   map[uint8]*street.Type
}

type Street struct {
	ID   uint64 `json:"i"`
	Name string `json:"n"`
}
