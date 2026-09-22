package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Holds json string arrays
type JSONStringSlice []string

func (j *JSONStringSlice) Scan(value any) error {
	if value == nil {
		*j = JSONStringSlice{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONStringSlice", value)
	}

	return json.Unmarshal(b, j)
}

func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

type TimeRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Schedule [7]*TimeRange

func (s *Schedule) Scan(value any) error {
	if value == nil {
		*s = Schedule{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into Schedule", value)
	}

	return json.Unmarshal(b, s)
}

func (s Schedule) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

type Student struct {
	ID           int             `json:"id"`
	UID          string          `json:"uid"` // card uid
	Name         string          `json:"name"`
	Status       string          `json:"status"`
	StartDate    string          `json:"start_date"` // time.DateOnly
	EndDate      string          `json:"end_date"`
	Schedule     Schedule        `json:"schedule"`
	DoneSeconds  int             `json:"done_seconds"`
	ExcludedDays JSONStringSlice `json:"excluded_days"`
	BreakTime    int             `json:"break_time"` // in seconds
	CreatedAt    string          `json:"created_at"`
}

type Device struct {
	ID        int
	Name      string
	SecretKey string
}

type Scan struct {
	ID        int    `json:"id"`
	UID       string `json:"uid"` // card uid
	Timestamp string `json:"timestamp"`
	StudentID int    `json:"student_id"`
}
