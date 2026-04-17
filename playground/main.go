package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	fmt.Println(" Sentinel Playground ")
	fmt.Println()

	//CPU
	cpuPercents, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		fmt.Printf("CPU error: %v\n", err)
	} else {
		fmt.Printf("CPU Usage:    %.2f%%\n", cpuPercents[0])
	}

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("Memory error: %v\n", err)
	} else {
		fmt.Printf("Memory Total: %v MB\n", vmStat.Total/1024/1024)
		fmt.Printf("Memory Used:  %v MB\n", vmStat.Used/1024/1024)
		fmt.Printf("Memory Free:  %v MB\n", vmStat.Free/1024/1024)
	}

	// Disk
	diskStat, err := disk.Usage("/")
	if err != nil {
		fmt.Printf("Disk error: %v\n", err)
	} else {
		fmt.Printf("Disk Total:   %v GB\n", diskStat.Total/1024/1024/1024)
		fmt.Printf("Disk Used:    %v GB\n", diskStat.Used/1024/1024/1024)
		fmt.Printf("Disk Free:    %v GB\n", diskStat.Free/1024/1024/1024)
	}

	// Services
	fmt.Println()
	fmt.Println("=== Process Monitor ===")

	services := []string{"nginx", "postgres", "bash"}

	processes, err := process.Processes()
	if err != nil {
		fmt.Printf("Process error: %v\n", err)
		return
	}

	for _, svc := range services {
		found := false
		for _, p := range processes {
			name, err := p.Name()
			if err != nil {
				continue
			}
			if name == svc {
				cpuPct, _ := p.CPUPercent()
				memInfo, _ := p.MemoryInfo()
				fmt.Printf("%-12s  UP    CPU: %.2f%%  MEM: %v KB\n",
					svc, cpuPct, memInfo.RSS/1024)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%-12s  DOWN\n", svc)
		}
	}
}