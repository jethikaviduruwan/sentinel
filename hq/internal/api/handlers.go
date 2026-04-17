package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jethikaviduruwan/sentinel/hq/internal/db"
)

type Handler struct {
	DB *db.DB
}

func NewHandler(database *db.DB) *Handler {
	return &Handler{DB: database}
}

// GET /servers
func (h *Handler) GetServers(c *gin.Context) {
	servers, err := h.DB.GetAllServers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if servers == nil {
		servers = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{"servers": servers})
}

// GET /servers/:id/stats
func (h *Handler) GetServerStats(c *gin.Context) {
	serverID := c.Param("id")

	stats, err := h.DB.GetLatestSystemMetrics(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found or no metrics yet"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GET /servers/:id/services
func (h *Handler) GetServerServices(c *gin.Context) {
	serverID := c.Param("id")

	services, err := h.DB.GetLatestServiceMetrics(c.Request.Context(), serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if services == nil {
		services = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{"server_id": serverID, "services": services})
}