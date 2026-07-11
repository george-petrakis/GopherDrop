package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/kek-Sec/gopherdrop/internal/config"
	"github.com/kek-Sec/gopherdrop/internal/models"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Send{})
	db.AutoMigrate(&models.StatCounters{})
	db.AutoMigrate(&models.StatDay{})
	db.Create(&models.StatCounters{ID: 1, Since: time.Now()})
	resetStatsCacheForTests()
	return db
}

func setupTestRouter(cfg config.Config, db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.POST("/send", CreateSend(cfg, db))
	r.POST("/send/text", CreateTextSend(cfg, db))
	r.POST("/send/file", CreateFileSend(cfg, db))
	r.GET("/send/:id", GetSend(cfg, db))
	r.GET("/send/:id/check", CheckPasswordProtection(db))
	r.GET("/stats", GetStats(db))
	return r
}

func createMultipartRequest(fieldName, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form field
	_ = writer.WriteField("type", "text")
	_ = writer.WriteField(fieldName, content)

	// Close writer to finalize boundary
	writer.Close()
	return body, writer.FormDataContentType()
}

func createMultipartFileRequest(fieldName, filename, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form field
	_ = writer.WriteField("type", "file")

	// Add file field
	part, _ := writer.CreateFormFile(fieldName, filename)
	io.WriteString(part, content)

	writer.Close()
	return body, writer.FormDataContentType()
}

func TestCheckPasswordProtection(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024, // 1MB
	}
	r := setupTestRouter(cfg, db)

	// Create a send with a password
	sendWithPassword := models.Send{
		Hash:      "protectedhash",
		Type:      "text",
		Data:      "encryptedDataHere",
		Password:  "password123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	db.Create(&sendWithPassword)

	// Create a send without a password
	sendWithoutPassword := models.Send{
		Hash:      "unprotectedhash",
		Type:      "text",
		Data:      "encryptedDataHere",
		Password:  "",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	db.Create(&sendWithoutPassword)

	// Test for send with password
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/send/protectedhash/check", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	expectedBody := `{"requiresPassword":true}`
	if w.Body.String() != expectedBody {
		t.Fatalf("expected body %s got %s", expectedBody, w.Body.String())
	}

	// Test for send without password
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/send/unprotectedhash/check", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	expectedBody = `{"requiresPassword":false}`
	if w.Body.String() != expectedBody {
		t.Fatalf("expected body %s got %s", expectedBody, w.Body.String())
	}
}

func TestCreateSendText(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024, // 1MB
	}
	r := setupTestRouter(cfg, db)

	body, contentType := createMultipartRequest("data", "This is a test message.")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestCreateTextSendWithRawBody(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024, // 1MB
	}
	r := setupTestRouter(cfg, db)

	body := bytes.NewBufferString("this came from curl raw body")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send/text", body)
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestCreateFileSendWithRawBody(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024, // 1MB
	}
	r := setupTestRouter(cfg, db)

	body := bytes.NewBufferString("file-bytes-from-curl")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send/file?filename=raw.txt", body)
	req.Header.Set("Content-Type", "application/octet-stream")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestCreateSendFileTooLarge(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 10, // Only allow 10 bytes
	}
	r := setupTestRouter(cfg, db)

	// Create a file with more than 10 bytes
	body, contentType := createMultipartFileRequest("file", "test.txt", "This file is too large.")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 got %d", w.Code)
	}
}

func TestGetNonExistentSend(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/send/unknownhash", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}

func TestExpiredSend(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	// Create expired send
	send := models.Send{
		Hash:      "expiredhash",
		Type:      "text",
		Data:      "encryptedDataHere",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	db.Create(&send)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/send/expiredhash", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", w.Code)
	}
}

func getStatCounters(db *gorm.DB) models.StatCounters {
	var c models.StatCounters
	db.Where("id = ?", 1).First(&c)
	return c
}

func TestCreateTextSendIncrementsStats(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	before := getStatCounters(db)

	text := "This is a test message."
	body, contentType := createMultipartRequest("data", text)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	after := getStatCounters(db)

	if after.TotalSends != before.TotalSends+1 {
		t.Fatalf("expected total_sends to increment by 1, got before=%d after=%d", before.TotalSends, after.TotalSends)
	}
	if after.TextSends != before.TextSends+1 {
		t.Fatalf("expected text_sends to increment by 1, got before=%d after=%d", before.TextSends, after.TextSends)
	}
	if after.FileSends != before.FileSends {
		t.Fatalf("expected file_sends to stay unchanged, got before=%d after=%d", before.FileSends, after.FileSends)
	}
	if after.TotalBytes != before.TotalBytes+int64(len(text)) {
		t.Fatalf("expected total_bytes to increment by %d, got before=%d after=%d", len(text), before.TotalBytes, after.TotalBytes)
	}
}

func TestCreateFileSendIncrementsStats(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	before := getStatCounters(db)

	content := "some file contents"
	body, contentType := createMultipartFileRequest("file", "test.txt", content)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	after := getStatCounters(db)

	if after.TotalSends != before.TotalSends+1 {
		t.Fatalf("expected total_sends to increment by 1, got before=%d after=%d", before.TotalSends, after.TotalSends)
	}
	if after.FileSends != before.FileSends+1 {
		t.Fatalf("expected file_sends to increment by 1, got before=%d after=%d", before.FileSends, after.FileSends)
	}
	if after.TextSends != before.TextSends {
		t.Fatalf("expected text_sends to stay unchanged, got before=%d after=%d", before.TextSends, after.TextSends)
	}
	if after.TotalBytes != before.TotalBytes+int64(len(content)) {
		t.Fatalf("expected total_bytes to increment by %d, got before=%d after=%d", len(content), before.TotalBytes, after.TotalBytes)
	}
}

func TestDeletingSendDoesNotDecrementStats(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	body, contentType := createMultipartRequest("data", "some data to delete")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	hash := resp["hash"]

	afterCreate := getStatCounters(db)

	// Force expiry and fetch, which triggers deleteSendAndFile.
	db.Model(&models.Send{}).Where("hash = ?", hash).UpdateColumn("expires_at", time.Now().Add(-1*time.Hour))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/send/"+hash, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for expired send, got %d", w.Code)
	}

	// Confirm the send was actually deleted from the database.
	var count int
	db.Model(&models.Send{}).Where("hash = ?", hash).Count(&count)
	if count != 0 {
		t.Fatalf("expected send to be deleted, but it still exists")
	}

	afterDelete := getStatCounters(db)

	if afterDelete.TotalSends != afterCreate.TotalSends {
		t.Fatalf("expected total_sends to remain unchanged after delete, before=%d after=%d", afterCreate.TotalSends, afterDelete.TotalSends)
	}
	if afterDelete.TextSends != afterCreate.TextSends {
		t.Fatalf("expected text_sends to remain unchanged after delete, before=%d after=%d", afterCreate.TextSends, afterDelete.TextSends)
	}
	if afterDelete.TotalBytes != afterCreate.TotalBytes {
		t.Fatalf("expected total_bytes to remain unchanged after delete, before=%d after=%d", afterCreate.TotalBytes, afterDelete.TotalBytes)
	}
}

func TestGetStatsShape(t *testing.T) {
	db := setupTestDB()
	cfg := config.Config{
		SecretKey:   "supersecretkeysupersecretkey32",
		MaxFileSize: 1024 * 1024,
	}
	r := setupTestRouter(cfg, db)

	// One active send, one already-expired send (should not count as active).
	db.Create(&models.Send{
		Hash:      "activehash",
		Type:      "text",
		Data:      "data",
		ExpiresAt: time.Now().Add(time.Hour),
	})
	db.Create(&models.Send{
		Hash:      "expiredstatshash",
		Type:      "text",
		Data:      "data",
		ExpiresAt: time.Now().Add(-time.Hour),
	})

	body, contentType := createMultipartRequest("data", "hello world")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/send", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 creating send, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}

	var stats struct {
		Lifetime struct {
			Total int64 `json:"total"`
			Texts int64 `json:"texts"`
			Files int64 `json:"files"`
			Bytes int64 `json:"bytes"`
		} `json:"lifetime"`
		Active struct {
			Total int64 `json:"total"`
		} `json:"active"`
		Daily []struct {
			Date  string `json:"date"`
			Texts int64  `json:"texts"`
			Files int64  `json:"files"`
			Bytes int64  `json:"bytes"`
		} `json:"daily"`
		Since string `json:"since"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to parse /stats response: %v", err)
	}

	if stats.Lifetime.Total != 1 {
		t.Fatalf("expected lifetime.total=1, got %d", stats.Lifetime.Total)
	}
	if stats.Lifetime.Texts != 1 {
		t.Fatalf("expected lifetime.texts=1, got %d", stats.Lifetime.Texts)
	}
	if stats.Lifetime.Bytes != int64(len("hello world")) {
		t.Fatalf("expected lifetime.bytes=%d, got %d", len("hello world"), stats.Lifetime.Bytes)
	}
	// Two pre-seeded sends (1 active + 1 expired), only the active one, plus
	// the newly created send, should count towards active.total.
	if stats.Active.Total != 2 {
		t.Fatalf("expected active.total=2, got %d", stats.Active.Total)
	}
	if len(stats.Daily) != 30 {
		t.Fatalf("expected exactly 30 daily entries, got %d", len(stats.Daily))
	}
	todayKey := time.Now().UTC().Format("2006-01-02")
	if stats.Daily[len(stats.Daily)-1].Date != todayKey {
		t.Fatalf("expected last daily entry to be today (%s), got %s", todayKey, stats.Daily[len(stats.Daily)-1].Date)
	}
	if stats.Daily[len(stats.Daily)-1].Texts != 1 {
		t.Fatalf("expected today's daily texts=1, got %d", stats.Daily[len(stats.Daily)-1].Texts)
	}
	if stats.Since == "" {
		t.Fatalf("expected since to be a non-empty RFC3339 timestamp")
	}
	if _, err := time.Parse(time.RFC3339, stats.Since); err != nil {
		t.Fatalf("expected since to be RFC3339, got %q: %v", stats.Since, err)
	}
}
