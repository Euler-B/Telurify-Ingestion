package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Euler-B/telurify-ingestion/internal/store"
	"github.com/Euler-B/telurify-ingestion/internal/usgs"
	"github.com/lmittmann/tint"
)

const (
	minMagnitude = -1.0
	maxMagnitude = 10.0
)

func main() {
	ctx := context.Background()
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:   slog.LevelInfo,
		NoColor: !isTerminal(os.Stderr),
	}))
	slog.SetDefault(logger)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := store.Connect(ctx, databaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close(ctx)

	feed, err := usgs.Fetch()
	if err != nil {
		logger.Error("USGS feed request failed", "error", err)
		os.Exit(1)
	}
	logger.Info("USGS feed fetched", "features", len(feed.Features))

	var newRecords, duplicateRecords, invalidRecords, databaseErrors int
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
			databaseErrors++
			logger.Error("database lookup failed", "feature", processed+1, "external_id", feature.ID, "title", props.Title, "error", err)
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
			databaseErrors++
			logger.Error("database insert failed", "feature", processed+1, "external_id", feature.ID, "title", props.Title, "error", err)
			progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
			continue
		}

		newRecords++
		progress.update(processed+1, newRecords, duplicateRecords, invalidRecords)
	}
	progress.finish()

	logger.Info("ingestion completed",
		"new_records", newRecords,
		"duplicate_records", duplicateRecords,
		"invalid_records", invalidRecords,
		"database_errors", databaseErrors,
	)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
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
