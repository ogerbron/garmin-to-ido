package sync

import (
	"fmt"
	"time"

	"garmin-to-ido/internal/garmin"
	"garmin-to-ido/internal/ido"
)

// Syncer handles synchronization between Garmin and iDO
type Syncer struct {
	garminClient *garmin.Client
	idoClient    *ido.Client
}

// NewSyncer creates a new syncer
func NewSyncer(garminClient *garmin.Client, idoClient *ido.Client) *Syncer {
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

		// Download activity data
		activityData, err := s.garminClient.DownloadActivity(activity.ActivityID)
		if err != nil {
			fmt.Printf("    ✗ Failed to download: %v\n", err)
			continue
		}

		// Upload to iDO
		if err := s.idoClient.UploadActivity(activityData, activity.ActivityName); err != nil {
			fmt.Printf("    ✗ Failed to upload: %v\n", err)
			continue
		}

		fmt.Printf("    ✓ Synced successfully\n")
	}

	return nil
}
