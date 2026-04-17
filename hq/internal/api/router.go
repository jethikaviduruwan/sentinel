package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jethikaviduruwan/sentinel/hq/internal/db"
)

func NewRouter(database *db.DB) *gin.Engine {
	r := gin.Default()
	h := NewHandler(database)

	r.GET("/servers", h.GetServers)
	r.GET("/servers/:id/stats", h.GetServerStats)
	r.GET("/servers/:id/services", h.GetServerServices)

	return r
}