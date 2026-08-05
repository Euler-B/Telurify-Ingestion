package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/Euler-B/telurify-ingestion/internal/store"
	"github.com/Euler-B/telurify-ingestion/internal/usgs"
)

const (
	minMagnitude = -1.0
	maxMagnitude = 10.0
)

func main() {
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := store.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal("database connection failed")
	}
	defer db.Close(ctx)

	feed, err := usgs.Fetch()
	if err != nil {
		log.Fatal("USGS feed request failed")
	}

	var newRecords, duplicateRecords, invalidRecords int

	for _, feature := range feed.Features {
		props := feature.Properties
		coords := feature.Geometry.Coordinates

		if props.Mag == nil || *props.Mag < minMagnitude || *props.Mag > maxMagnitude {
			invalidRecords++
			log.Printf("Registro inválido (magnitud): %s", safeLogValue(props.Title))
			continue
		}

		exists, err := db.ExistsByTitle(ctx, props.Title)
		if err != nil {
			log.Printf("Error verificando duplicado para %q", safeLogValue(props.Title))
			duplicateRecords++
			continue
		}

		if exists {
			duplicateRecords++
			log.Printf("Registro duplicado: %s", safeLogValue(props.Title))
			continue
		}

		if len(coords) < 2 {
			invalidRecords++
			log.Printf("Registro inválido (coordenadas): %s", safeLogValue(props.Title))
			continue
		}

		record := store.SismoRecord{
			Title:      props.Title,
			URL:        props.URL,
			Place:      props.Place,
			MagType:    props.MagType,
			Mag:        *props.Mag,
			Longitude:  coords[0],
			Latitude:   coords[1],
			Tsunami:    props.Tsunami == 1,
			ExternalID: feature.ID,
		}

		if err := db.Insert(ctx, record); err != nil {
			duplicateRecords++
			log.Printf("Error al guardar el registro %q", safeLogValue(props.Title))
			continue
		}

		newRecords++
	}

	log.Printf("Nuevos registros guardados: %d", newRecords)
	log.Printf("Registros duplicados omitidos: %d", duplicateRecords)
	log.Printf("Registros inválidos omitidos: %d", invalidRecords)
	log.Printf("Datos de los sismos obtenido, validados, y persistidos")
}

func safeLogValue(value string) string {
	value = strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(value)
	if len(value) > 200 {
		return value[:200] + "..."
	}
	return value
}
