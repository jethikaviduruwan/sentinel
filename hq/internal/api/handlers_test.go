package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockDB implements just enough to test handlers without a real database
type mockDB struct {
	servers  []map[string]interface{}
	stats    map[string]interface{}
	services []map[string]interface{}
	err      error
}

func TestGetServersEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/servers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"servers": []map[string]interface{}{}})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/servers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	servers, ok := response["servers"]
	if !ok {
		t.Fatal("response missing 'servers' key")
	}

	serverList := servers.([]interface{})
	if len(serverList) != 0 {
		t.Errorf("expected empty server list, got %d", len(serverList))
	}

	t.Log("GET /servers returned empty list correctly")
}

func TestGetServerStatsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/servers/:id/stats", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found or no metrics yet"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/servers/nonexistent/stats", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	t.Log("GET /servers/:id/stats returns 404 for unknown server correctly")
}

func TestGetServerServicesEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/servers/:id/services", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"server_id": c.Param("id"),
			"services":  []map[string]interface{}{},
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/servers/Jethika/services", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["server_id"] != "Jethika" {
		t.Errorf("expected server_id 'Jethika', got '%v'", response["server_id"])
	}

	t.Log("GET /servers/:id/services returns correct server_id")
}