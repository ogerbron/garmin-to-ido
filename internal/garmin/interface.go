package garmin

import "time"

// GarminClient is an interface for Garmin Connect clients
type GarminClient interface {
	Login() error
	GetActivities(date time.Time) ([]Activity, error)
	GetBikeActivities(date time.Time) ([]Activity, error)
	DownloadActivity(activityID int64) ([]byte, error)
	Logout() error
}
