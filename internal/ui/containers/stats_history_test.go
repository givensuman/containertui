package containers

import (
	"math"
	"testing"
	"time"

	"github.com/givensuman/containertui/internal/client"
)

func TestStatsHistoryPushComputesRatesFromCumulativeNetworkCounters(t *testing.T) {
	history := newStatsHistory(10)
	base := time.Unix(100, 0)

	history.push(client.ContainerStats{NetRx: 100, NetTx: 200}, base)
	point := history.push(client.ContainerStats{NetRx: 160, NetTx: 320}, base.Add(2*time.Second))

	if point.NetRxRate != 30 {
		t.Fatalf("expected RX rate 30 bytes/s, got %v", point.NetRxRate)
	}

	if point.NetTxRate != 60 {
		t.Fatalf("expected TX rate 60 bytes/s, got %v", point.NetTxRate)
	}
}

func TestStatsHistoryPushComputesMemoryPercent(t *testing.T) {
	history := newStatsHistory(10)
	point := history.push(client.ContainerStats{MemUsage: 256, MemLimit: 1024}, time.Unix(200, 0))

	if math.Abs(point.MemPercent-25) > 0.0001 {
		t.Fatalf("expected memory percent 25, got %v", point.MemPercent)
	}
}

func TestStatsHistoryWindowIsBounded(t *testing.T) {
	history := newStatsHistory(2)
	base := time.Unix(300, 0)

	history.push(client.ContainerStats{CPUPercent: 1}, base)
	history.push(client.ContainerStats{CPUPercent: 2}, base.Add(time.Second))
	history.push(client.ContainerStats{CPUPercent: 3}, base.Add(2*time.Second))

	if len(history.points) != 2 {
		t.Fatalf("expected history length 2, got %d", len(history.points))
	}

	if history.points[0].CPUPercent != 2 {
		t.Fatalf("expected oldest retained point CPU=2, got %v", history.points[0].CPUPercent)
	}

	if history.points[1].CPUPercent != 3 {
		t.Fatalf("expected newest retained point CPU=3, got %v", history.points[1].CPUPercent)
	}
}

func TestStatsHistoryPushCounterResetProducesZeroRates(t *testing.T) {
	history := newStatsHistory(10)
	base := time.Unix(400, 0)

	history.push(client.ContainerStats{NetRx: 500, NetTx: 900}, base)
	point := history.push(client.ContainerStats{NetRx: 120, NetTx: 80}, base.Add(2*time.Second))

	if point.NetRxRate != 0 {
		t.Fatalf("expected RX rate 0 bytes/s after counter reset, got %v", point.NetRxRate)
	}

	if point.NetTxRate != 0 {
		t.Fatalf("expected TX rate 0 bytes/s after counter reset, got %v", point.NetTxRate)
	}
}

func TestStatsHistoryPushNonPositiveTimeDeltaProducesZeroRates(t *testing.T) {
	t.Run("zero delta", func(t *testing.T) {
		history := newStatsHistory(10)
		base := time.Unix(500, 0)

		history.push(client.ContainerStats{NetRx: 100, NetTx: 200}, base)
		point := history.push(client.ContainerStats{NetRx: 160, NetTx: 320}, base)

		if point.NetRxRate != 0 {
			t.Fatalf("expected RX rate 0 bytes/s for zero time delta, got %v", point.NetRxRate)
		}

		if point.NetTxRate != 0 {
			t.Fatalf("expected TX rate 0 bytes/s for zero time delta, got %v", point.NetTxRate)
		}
	})

	t.Run("negative delta", func(t *testing.T) {
		history := newStatsHistory(10)
		base := time.Unix(600, 0)

		history.push(client.ContainerStats{NetRx: 100, NetTx: 200}, base)
		point := history.push(client.ContainerStats{NetRx: 160, NetTx: 320}, base.Add(-1*time.Second))

		if point.NetRxRate != 0 {
			t.Fatalf("expected RX rate 0 bytes/s for negative time delta, got %v", point.NetRxRate)
		}

		if point.NetTxRate != 0 {
			t.Fatalf("expected TX rate 0 bytes/s for negative time delta, got %v", point.NetTxRate)
		}
	})
}
