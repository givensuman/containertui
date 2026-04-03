package containers

import (
	"time"

	"github.com/givensuman/containertui/internal/client"
)

type statsPoint struct {
	Timestamp  time.Time
	CPUPercent float64
	MemUsage   float64
	MemLimit   float64
	MemPercent float64
	NetRx      float64
	NetTx      float64
	NetRxRate  float64
	NetTxRate  float64
}

type statsHistory struct {
	window int
	points []statsPoint
}

func newStatsHistory(window int) *statsHistory {
	if window <= 0 {
		window = 1
	}

	return &statsHistory{
		window: window,
		points: make([]statsPoint, 0, window),
	}
}

func (h *statsHistory) push(stats client.ContainerStats, at time.Time) statsPoint {
	point := statsPoint{
		Timestamp:  at,
		CPUPercent: stats.CPUPercent,
		MemUsage:   stats.MemUsage,
		MemLimit:   stats.MemLimit,
		NetRx:      stats.NetRx,
		NetTx:      stats.NetTx,
	}

	if stats.MemLimit > 0 {
		point.MemPercent = (stats.MemUsage / stats.MemLimit) * 100
	}

	if len(h.points) > 0 {
		prev := h.points[len(h.points)-1]
		deltaSeconds := at.Sub(prev.Timestamp).Seconds()
		if deltaSeconds > 0 {
			rxDelta := stats.NetRx - prev.NetRx
			if rxDelta < 0 {
				rxDelta = 0
			}

			txDelta := stats.NetTx - prev.NetTx
			if txDelta < 0 {
				txDelta = 0
			}

			point.NetRxRate = rxDelta / deltaSeconds
			point.NetTxRate = txDelta / deltaSeconds
		}
	}

	h.points = append(h.points, point)
	if len(h.points) > h.window {
		h.points = h.points[len(h.points)-h.window:]
	}

	return point
}
