package collector

import (
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	pb "github.com/jethikaviduruwan/sentinel/proto/gen"
)

// Collect gathers system and service metrics concurrently
func Collect(serverID string, services []string) (*pb.MetricPayload, error) {
	var (
		sys    *pb.SystemMetrics
		svcs   []*pb.ServiceMetric
		sysErr error
		mu     sync.Mutex
		wg     sync.WaitGroup
	)

	now := time.Now().UnixMilli()

	// Goroutine 1: collect system metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := collectSystem(serverID, now)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			sysErr = err
			return
		}
		sys = s
	}()

	// Goroutine 2: collect service metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := collectServices(serverID, services, now)
		mu.Lock()
		defer mu.Unlock()
		svcs = s
	}()

	wg.Wait()

	if sysErr != nil {
		return nil, sysErr
	}

	return &pb.MetricPayload{
		System:   sys,
		Services: svcs,
	}, nil
}

func collectSystem(serverID string, now int64) (*pb.SystemMetrics, error) {
	cpuPcts, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	d, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &pb.SystemMetrics{
		ServerId:   serverID,
		Timestamp:  now,
		CpuPercent: cpuPcts[0],
		MemTotal:   vm.Total,
		MemUsed:    vm.Used,
		MemFree:    vm.Free,
		DiskTotal:  d.Total,
		DiskUsed:   d.Used,
		DiskFree:   d.Free,
	}, nil
}

func collectServices(serverID string, services []string, now int64) []*pb.ServiceMetric {
	procs, err := process.Processes()
	if err != nil {
		log.Printf("[collector] failed to list processes: %v", err)
		return nil
	}

	// Build a map of process name -> process for fast lookup
	procMap := make(map[string]*process.Process)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		procMap[name] = p
	}

	var metrics []*pb.ServiceMetric
	for _, svcName := range services {
		metric := &pb.ServiceMetric{
			ServerId:  serverID,
			Timestamp: now,
			Name:      svcName,
			Running:   false,
		}

		if p, found := procMap[svcName]; found {
			cpuPct, _ := p.CPUPercent()
			memInfo, _ := p.MemoryInfo()
			metric.Running = true
			metric.CpuPercent = cpuPct
			if memInfo != nil {
				metric.MemRss = memInfo.RSS
			}
		}

		metrics = append(metrics, metric)
	}

	return metrics
}