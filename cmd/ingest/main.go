package main

import (
	"context"
	"fmt"
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
	progress := newProgress(len(feed.Features))

	for processed, feature := range feed.Features {
		props := feature.Properties
		coords := feature.Geometry.Coordinates

		if props.Mag == nil || *props.Mag < minMagnitude || *props.Mag > maxMagnitude {
			invalidRecords++
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
			continue
		}

		exists, err := db.ExistsByTitle(ctx, props.Title)
		if err != nil {
			duplicateRecords++
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
			continue
		}

		if exists {
			duplicateRecords++
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
			continue
		}

		if len(coords) < 2 {
			invalidRecords++
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
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
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
			continue
		}

		newRecords++
		progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
	}
	progress.finish()

	log.Printf("Nuevos registros guardados: %d", newRecords)
	log.Printf("Registros duplicados omitidos: %d", duplicateRecords)
	log.Printf("Registros inválidos omitidos: %d", invalidRecords)
	log.Printf("Datos de los sismos obtenido, validados, y persistidos")
}

type progressBar struct {
	total       int
	interactive bool
}

func newProgress(total int) progressBar {
	info, err := os.Stdout.Stat()
	return progressBar{
		total:       total,
		interactive: err == nil && info.Mode()&os.ModeCharDevice != 0,
	}
}

func (p progressBar) update(processed, newRecords, duplicateRecords, invalidRecords int) {
	if !p.interactive || p.total == 0 {
		return
	}

	const width = 30
	completed := width * processed / p.total
	remaining := width - completed
	percent := 100 * processed / p.total
	fmt.Printf("\r[%s%s] %3d%% %d/%d | nuevos: %d | duplicados: %d | inválidos: %d",
		strings.Repeat("#", completed),
		strings.Repeat("-", remaining),
		percent,
		processed,
		p.total,
		newRecords,
		duplicateRecords,
		invalidRecords,
	)
}

func (p progressBar) finish() {
	if p.interactive {
		fmt.Println("\r\033[2K")
	}
}
