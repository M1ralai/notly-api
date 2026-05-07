package dto

import "time"

// SyncRequest is the payload from the frontend asking for updates
type SyncRequest struct {
	Since *string `json:"since"` // ISO8601 string, null if full sync
}

// DeltaSyncResponse is the unified payload returned to the frontend
type DeltaSyncResponse struct {
	Timestamp time.Time    `json:"timestamp"`
	Changes   ChangesDelta `json:"changes"`
}

type ChangesDelta struct {
	Tasks     ModuleDelta `json:"tasks"`
	Habits    ModuleDelta `json:"habits"`
	LifeAreas ModuleDelta `json:"life_areas"`
	Goals     ModuleDelta `json:"goals"`
	Courses   ModuleDelta `json:"courses"`
	Events    ModuleDelta `json:"events"`
	Semesters ModuleDelta `json:"semesters"`
	Notes     ModuleDelta `json:"notes"`
}

type ModuleDelta struct {
	Updated interface{} `json:"updated"`
	Deleted []int       `json:"deleted"`
}

// SyncSignalRequest is a generic wrapper for broadcasting real-time signals
type SyncSignalRequest struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}
