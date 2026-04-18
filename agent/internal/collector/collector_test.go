package collector

import (
	"testing"
)

func TestCollectSystem(t *testing.T) {
	now := int64(1000000)
	metrics, err := collectSystem("test-server", now)

	if err != nil {
		t.Fatalf("collectSystem returned error: %v", err)
	}

	if metrics.ServerId != "test-server" {
		t.Errorf("expected server_id 'test-server', got '%s'", metrics.ServerId)
	}

	if metrics.Timestamp != now {
		t.Errorf("expected timestamp %d, got %d", now, metrics.Timestamp)
	}

	if metrics.CpuPercent < 0 || metrics.CpuPercent > 100 {
		t.Errorf("cpu_percent out of range: %.2f", metrics.CpuPercent)
	}

	if metrics.MemTotal == 0 {
		t.Error("mem_total should not be zero")
	}

	if metrics.MemUsed == 0 {
		t.Error("mem_used should not be zero")
	}

	if metrics.DiskTotal == 0 {
		t.Error("disk_total should not be zero")
	}

	t.Logf("CPU: %.2f%% | MEM: %dMB / %dMB | DISK: %dGB",
		metrics.CpuPercent,
		metrics.MemUsed/1024/1024,
		metrics.MemTotal/1024/1024,
		metrics.DiskTotal/1024/1024/1024,
	)
}

func TestCollectServices(t *testing.T) {
	services := []string{"bash", "nonexistent-service-xyz"}
	metrics := collectServices("test-server", services, 1000000)

	if len(metrics) != len(services) {
		t.Fatalf("expected %d service metrics, got %d", len(services), len(metrics))
	}

	// bash should be found
	bashFound := false
	for _, m := range metrics {
		if m.Name == "bash" && m.Running {
			bashFound = true
		}
	}
	if !bashFound {
		t.Error("expected bash to be running")
	}

	// nonexistent service should be down
	for _, m := range metrics {
		if m.Name == "nonexistent-service-xyz" && m.Running {
			t.Error("nonexistent service should not be running")
		}
	}

	t.Logf("services collected: %d", len(metrics))
}

func TestCollect(t *testing.T) {
	payload, err := Collect("test-server", []string{"bash"})

	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if payload.System == nil {
		t.Fatal("system metrics should not be nil")
	}

	if len(payload.Services) == 0 {
		t.Fatal("services should not be empty")
	}

	t.Logf("payload collected successfully for server: %s", payload.System.ServerId)
}