// Package repository implements data abstraction.
package repository

import (
	"github.com/Pklerik/urlshortener/internal/config"
	"github.com/Pklerik/urlshortener/internal/model"
)

var mapShortener model.MapShortener = loadLocalData()

func loadLocalData() model.MapShortener {
	return make(model.MapShortener, config.MapSize)
}

// MapShorts provides pointer to data structure.
func MapShorts() *model.MapShortener {
	return &mapShortener
}
