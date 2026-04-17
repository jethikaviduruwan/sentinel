package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/jethikaviduruwan/sentinel/agent/internal/sender"
	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

func main() {
	hqAddr := "localhost:50051"
	serverID, _ := os.Hostname()
	services := []string{"nginx", "postgres", "bash"}

	log.Printf("[Agent] starting, server_id=%s", serverID)

	s, err := sender.New(hqAddr)
	if err != nil {
		log.Fatalf("[Agent] could not connect to HQ: %v", err)
	}
	defer s.Close()

	log.Println("[Agent] connected to HQ, starting metric stream...")

	// Send metrics every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		payload, err := collectMetrics(serverID, services)
		if err != nil {
			log.Printf("[Agent] collect error: %v", err)
			continue
		}

		if err := s.Send(context.Background(), payload); err != nil {
			log.Printf("[Agent] send error: %v", err)
			continue
		}

		log.Printf("[Agent] metrics sent successfully")
	}
}

func collectMetrics(serverID string, services []string) (*pb.MetricPayload, error) {
	now := time.Now().UnixMilli()

	// CPU
	cpuPcts, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	// Memory
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Disk
	d, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	sys := &pb.SystemMetrics{
		ServerId:   serverID,
		Timestamp:  now,
		CpuPercent: cpuPcts[0],
		MemTotal:   vm.Total,
		MemUsed:    vm.Used,
		MemFree:    vm.Free,
		DiskTotal:  d.Total,
		DiskUsed:   d.Used,
		DiskFree:   d.Free,
	}

	// Services
	var svcMetrics []*pb.ServiceMetric
	procs, _ := process.Processes()

	for _, svcName := range services {
		metric := &pb.ServiceMetric{
			ServerId:  serverID,
			Timestamp: now,
			Name:      svcName,
			Running:   false,
		}
		for _, p := range procs {
			name, err := p.Name()
			if err != nil {
				continue
			}
			if name == svcName {
				cpuPct, _ := p.CPUPercent()
				memInfo, _ := p.MemoryInfo()
				metric.Running = true
				metric.CpuPercent = cpuPct
				metric.MemRss = memInfo.RSS
				break
			}
		}
		svcMetrics = append(svcMetrics, metric)
	}

	return &pb.MetricPayload{
		System:   sys,
		Services: svcMetrics,
	}, nil
}