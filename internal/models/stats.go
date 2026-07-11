package models

import "time"

// StatCounters holds monotonic lifetime counters for the stats feature.
// There is always exactly one row, with ID=1. These counters are only ever
// incremented at secret-creation time and must NEVER be decremented on
// delete/expiry.
type StatCounters struct {
	ID         uint `gorm:"primary_key"`
	TotalSends int64
	TextSends  int64
	FileSends  int64
	TotalBytes int64
	Since      time.Time
}

// StatDay holds per-UTC-day aggregate counters used to build the daily
// stats series. Rows are upserted at secret-creation time and never
// decremented.
type StatDay struct {
	Day   time.Time `gorm:"primary_key;type:date"`
	Texts int64
	Files int64
	Bytes int64
}
