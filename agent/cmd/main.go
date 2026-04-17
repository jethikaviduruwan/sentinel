package main

import (
	"context"
	"log"
	"time"

	"github.com/jethikaviduruwan/sentinel/agent/internal/collector"
	"github.com/jethikaviduruwan/sentinel/agent/internal/config"
	"github.com/jethikaviduruwan/sentinel/agent/internal/sender"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("[Agent] failed to load config: %v", err)
	}

	log.Printf("[Agent] starting, server_id=%s, hq=%s", cfg.ServerID, cfg.HQAddress)
	log.Printf("[Agent] monitoring services: %v", cfg.Services)

	// Connect to HQ
	s, err := sender.New(cfg.HQAddress)
	if err != nil {
		log.Fatalf("[Agent] could not connect to HQ: %v", err)
	}
	defer s.Close()

	log.Println("[Agent] connected to HQ, starting metric stream...")

	ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Collect concurrently
		payload, err := collector.Collect(cfg.ServerID, cfg.Services)
		if err != nil {
			log.Printf("[Agent] collect error: %v", err)
			continue
		}

		if err := s.Send(context.Background(), payload); err != nil {
			log.Printf("[Agent] send error: %v", err)
			continue
		}

		log.Printf("[Agent] metrics sent — CPU: %.2f%% MEM: %dMB",
			payload.System.CpuPercent,
			payload.System.MemUsed/1024/1024,
		)
	}
}