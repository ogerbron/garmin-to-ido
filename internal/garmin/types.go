package garmin

import "time"

// Activity represents a Garmin activity
type Activity struct {
	ActivityID   int64     `json:"activityId"`
	ActivityName string    `json:"activityName"`
	ActivityType string    `json:"activityType"`
	StartTime    time.Time `json:"startTimeLocal"`
	Distance     float64   `json:"distance"`     // meters
	Duration     float64   `json:"duration"`     // seconds
	AvgSpeed     float64   `json:"averageSpeed"` // m/s
	Calories     float64   `json:"calories"`
}
