package sync

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"garmin-to-ido/internal/garmin"
	"garmin-to-ido/internal/ido"
)

// Syncer handles synchronization between Garmin and iDO
type Syncer struct {
	garminClient garmin.GarminClient
	idoClient    *ido.Client
}

// NewSyncer creates a new syncer
func NewSyncer(garminClient garmin.GarminClient, idoClient *ido.Client) *Syncer {
	return &Syncer{
		garminClient: garminClient,
		idoClient:    idoClient,
	}
}

// SyncBikeActivities synchronizes bike activities for a specific date
func (s *Syncer) SyncBikeActivities(date time.Time) error {
	// Get bike activities from Garmin
	activities, err := s.garminClient.GetBikeActivities(date)
	if err != nil {
		return fmt.Errorf("failed to get bike activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Printf("  No bike activities found for %s\n", date.Format("2006-01-02"))
		return nil
	}

	fmt.Printf("  Found %d bike activity(ies)\n", len(activities))

	// Upload each activity to iDO
	for i, activity := range activities {
		fmt.Printf("  [%d/%d] %s (%.2f km, %.0f min)\n",
			i+1, len(activities),
			activity.ActivityName,
			activity.Distance/1000,
			activity.Duration/60,
		)

		// Download activity data (this is a ZIP file from Garmin)
		zipData, err := s.garminClient.DownloadActivity(activity.ActivityID)
		if err != nil {
			fmt.Printf("    ✗ Failed to download: %v\n", err)
			continue
		}

		// Extract the FIT file from the ZIP
		zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
		if err != nil {
			fmt.Printf("    ✗ Failed to read ZIP: %v\n", err)
			continue
		}

		var fitData []byte
		var fitFilename string
		for _, file := range zipReader.File {
			if filepath.Ext(file.Name) == ".fit" {
				fitFilename = file.Name
				rc, err := file.Open()
				if err != nil {
					fmt.Printf("    ✗ Failed to open FIT file in ZIP: %v\n", err)
					break
				}
				fitData, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					fmt.Printf("    ✗ Failed to read FIT file: %v\n", err)
					break
				}
				break
			}
		}

		if len(fitData) == 0 {
			fmt.Printf("    ✗ No FIT file found in ZIP\n")
			continue
		}

		fmt.Printf("    → Extracted FIT file: %s (%d bytes)\n", fitFilename, len(fitData))

		// Save both ZIP and FIT files to disk
		fitDir := "downloaded_fits"
		if err := os.MkdirAll(fitDir, 0755); err != nil {
			fmt.Printf("    ✗ Failed to create directory: %v\n", err)
			continue
		}

		// Create filename with timestamp and activity ID
		timestamp := activity.StartTime.Format("20060102_150405")

		// Save the original ZIP file
		zipFilename := filepath.Join(fitDir, fmt.Sprintf("%s_%d_%s.zip", timestamp, activity.ActivityID, activity.ActivityType))
		if err := os.WriteFile(zipFilename, zipData, 0644); err != nil {
			fmt.Printf("    ✗ Failed to save ZIP file: %v\n", err)
		} else {
			fmt.Printf("    → Saved ZIP file: %s\n", zipFilename)
		}

		// Save the extracted FIT file
		fitFilePath := filepath.Join(fitDir, fmt.Sprintf("%s_%d_%s.fit", timestamp, activity.ActivityID, activity.ActivityType))
		if err := os.WriteFile(fitFilePath, fitData, 0644); err != nil {
			fmt.Printf("    ✗ Failed to save FIT file: %v\n", err)
		} else {
			fmt.Printf("    → Saved FIT file: %s\n", fitFilePath)
		}

		// Upload the extracted FIT data to iDO (not the ZIP)
		if err := s.idoClient.UploadActivity(fitData, activity.ActivityName, activity.ActivityType, activity.StartTime); err != nil {
			fmt.Printf("    ✗ Failed to upload: %v\n", err)
			continue
		}

		fmt.Printf("    ✓ Synced successfully\n")
	}

	return nil
}
