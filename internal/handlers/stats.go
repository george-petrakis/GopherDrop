package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/kek-Sec/gopherdrop/internal/models"
)

// statsCacheTTL controls how long a built /stats payload is reused before
// being recomputed. This keeps the endpoint effectively free to call and
// removes the need for a rate limiter on it.
const statsCacheTTL = 30 * time.Second

var (
	statsCacheMu      sync.RWMutex
	statsCacheAt      time.Time
	statsCacheDB      *gorm.DB
	statsCachePayload gin.H
)

// dailyStat represents a single zero-filled day in the daily stats series.
type dailyStat struct {
	Date  string `json:"date"`
	Texts int64  `json:"texts"`
	Files int64  `json:"files"`
	Bytes int64  `json:"bytes"`
}

// recordSendStats best-effort increments the lifetime and daily stat
// counters for a newly created send. It is intentionally isolated from the
// request lifecycle: it never returns an error and must never influence the
// HTTP response for send creation. Counters are monotonic and must never be
// touched anywhere on the delete/expiry paths.
//
// It runs inline (synchronously) in the create handler rather than in a
// goroutine: the create endpoint is rate-limited to ~1 req/s (see
// routes.limiter), so single-row write contention on the two hot rows below is
// a non-issue in practice, and staying synchronous keeps the in-memory sqlite
// test DB deterministic. Both statements are upserts so they self-heal if the
// startup seed row is ever missing (fresh/reset external DB).
func recordSendStats(db *gorm.DB, stype string, payloadBytes int64) {
	if db == nil {
		return
	}

	var textInc, fileInc int64
	switch stype {
	case "text":
		textInc = 1
	case "file":
		fileInc = 1
	}

	if err := db.Exec(
		`INSERT INTO stat_counters (id, total_sends, text_sends, file_sends, total_bytes, since)
		 VALUES (1, 1, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET total_sends = stat_counters.total_sends + 1, text_sends = stat_counters.text_sends + ?, file_sends = stat_counters.file_sends + ?, total_bytes = stat_counters.total_bytes + ?`,
		textInc, fileInc, payloadBytes, time.Now().UTC(),
		textInc, fileInc, payloadBytes,
	).Error; err != nil {
		log.Println("stats: failed to update stat_counters:", err)
	}

	day := time.Now().UTC().Format("2006-01-02")
	if err := db.Exec(
		`INSERT INTO stat_days (day, texts, files, bytes) VALUES (?, ?, ?, ?)
		 ON CONFLICT(day) DO UPDATE SET texts = stat_days.texts + ?, files = stat_days.files + ?, bytes = stat_days.bytes + ?`,
		day, textInc, fileInc, payloadBytes,
		textInc, fileInc, payloadBytes,
	).Error; err != nil {
		log.Println("stats: failed to update stat_days:", err)
	}
}

// GetStats returns a public, aggregate-only, cached snapshot of lifetime and
// recent activity statistics.
func GetStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := buildStatsPayload(db)
		if err != nil {
			log.Println("stats: failed to build stats payload:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
			return
		}
		c.JSON(http.StatusOK, payload)
	}
}

func buildStatsPayload(db *gorm.DB) (gin.H, error) {
	if cached, ok := getCachedStats(db); ok {
		return cached, nil
	}

	statsCacheMu.Lock()
	defer statsCacheMu.Unlock()

	// Re-check under the write lock in case another goroutine already
	// rebuilt the cache while we were waiting.
	if statsCacheDB == db && time.Since(statsCacheAt) < statsCacheTTL && statsCachePayload != nil {
		return statsCachePayload, nil
	}

	var counters models.StatCounters
	if db.Where("id = ?", 1).First(&counters).RecordNotFound() {
		counters = models.StatCounters{ID: 1, Since: time.Now()}
	}

	var activeTotal int64
	if err := db.Model(&models.Send{}).Where("expires_at > ?", time.Now()).Count(&activeTotal).Error; err != nil {
		return nil, err
	}

	daily, err := buildDailySeries(db)
	if err != nil {
		return nil, err
	}

	payload := gin.H{
		"lifetime": gin.H{
			"total": counters.TotalSends,
			"texts": counters.TextSends,
			"files": counters.FileSends,
			"bytes": counters.TotalBytes,
		},
		"active": gin.H{
			"total": activeTotal,
		},
		"daily": daily,
		"since": counters.Since.UTC().Format(time.RFC3339),
	}

	statsCacheDB = db
	statsCachePayload = payload
	statsCacheAt = time.Now()

	return payload, nil
}

// resetStatsCacheForTests clears the in-memory stats cache. Production code
// never needs to call this; it exists purely so tests that spin up a fresh
// in-memory DB per test case aren't at the mercy of Go potentially reusing a
// prior *gorm.DB's memory address for the cache-identity check below.
func resetStatsCacheForTests() {
	statsCacheMu.Lock()
	defer statsCacheMu.Unlock()
	statsCacheDB = nil
	statsCacheAt = time.Time{}
	statsCachePayload = nil
}

func getCachedStats(db *gorm.DB) (gin.H, bool) {
	statsCacheMu.RLock()
	defer statsCacheMu.RUnlock()

	if statsCacheDB == db && time.Since(statsCacheAt) < statsCacheTTL && statsCachePayload != nil {
		return statsCachePayload, true
	}
	return nil, false
}

// buildDailySeries returns exactly 30 zero-filled daily entries, oldest
// first, covering the last 30 UTC days ending today.
func buildDailySeries(db *gorm.DB) ([]dailyStat, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	windowStart := today.AddDate(0, 0, -29)

	// Bound the scan to the visible window. Days are stored as ISO "YYYY-MM-DD"
	// so a lexicographic >= comparison is also chronological.
	var rows []models.StatDay
	if err := db.Where("day >= ?", windowStart.Format("2006-01-02")).Find(&rows).Error; err != nil {
		return nil, err
	}

	byDay := make(map[string]models.StatDay, len(rows))
	for _, r := range rows {
		byDay[r.Day.UTC().Format("2006-01-02")] = r
	}

	daily := make([]dailyStat, 0, 30)
	for i := 0; i < 30; i++ {
		key := windowStart.AddDate(0, 0, i).Format("2006-01-02")
		if row, ok := byDay[key]; ok {
			daily = append(daily, dailyStat{Date: key, Texts: row.Texts, Files: row.Files, Bytes: row.Bytes})
		} else {
			daily = append(daily, dailyStat{Date: key})
		}
	}

	return daily, nil
}
