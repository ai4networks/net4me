package dind

import (
	"context"
	"encoding/json"
)

type DockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64   `json:"total_usage"`
			PercpuUsage       []uint64 `json:"percpu_usage"`
			UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64   `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
		ThrottlingData struct {
			Periods          uint64 `json:"periods"`
			ThrottledPeriods uint64 `json:"throttled_periods"`
			ThrottledTime    uint64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"cpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Stats struct {
			ActiveAnon        uint64 `json:"active_anon"`
			ActiveFile        uint64 `json:"active_file"`
			Cache             uint64 `json:"cache"`
			Dirty             uint64 `json:"dirty"`
			Hierarchical      uint64 `json:"hierarchical_memory_limit"`
			HierarchicalUsage uint64 `json:"hierarchical_memsw_limit"`
			InactiveAnon      uint64 `json:"inactive_anon"`
			InactiveFile      uint64 `json:"inactive_file"`
			Mapped            uint64 `json:"mapped"`
			Pgfault           uint64 `json:"pgfault"`
			Pgmajfault        uint64 `json:"pgmajfault"`
			Pgpgin            uint64 `json:"pgpgin"`
			Pgpgout           uint64 `json:"pgpgout"`
			Rss               uint64 `json:"rss"`
			RssHuge           uint64 `json:"rss_huge"`
			TotalActive       uint64 `json:"total_active"`
			TotalInactive     uint64 `json:"total_inactive"`
			TotalPgfaults     uint64 `json:"total_pgfaults"`
			TotalPgpgin       uint64 `json:"total_pgpgin"`
			TotalPgpgout      uint64 `json:"total_pgpgout"`
			TotalRss          uint64 `json:"total_rss"`
			TotalRssHuge      uint64 `json:"total_rss_huge"`
			TotalUnevictable  uint64 `json:"total_unevictable"`
			TotalWriteback    uint64 `json:"total_writeback"`
			Unevictable       uint64 `json:"unevictable"`
			Writeback         uint64 `json:"writeback"`
		} `json:"stats"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64   `json:"total_usage"`
			PercpuUsage       []uint64 `json:"percpu_usage"`
			UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64   `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
		ThrottlingData struct {
			Periods          uint64 `json:"periods"`
			ThrottledPeriods uint64 `json:"throttled_periods"`
			ThrottledTime    uint64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"precpu_stats"`
}

func (n *Node) Stats() (map[string]any, error) {
	n.manager.lock.RLock()
	defer n.manager.lock.RUnlock()
	statsResp, err := n.manager.clientDocker.ContainerStats(context.Background(), n.id, false)
	if err != nil {
		return nil, err
	}
	defer statsResp.Body.Close()
	var stats DockerStats
	decoder := json.NewDecoder(statsResp.Body)
	if err := decoder.Decode(&stats); err != nil {
		return nil, err
	}
	statsMap := make(map[string]any)
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage) - float64(stats.PreCPUStats.SystemCPUUsage)
	usedMemory := float64(stats.MemoryStats.Usage) - float64(stats.MemoryStats.Stats.Cache)
	availableMemory := float64(stats.MemoryStats.Limit)
	statsMap["memory_percent"] = float64(usedMemory/availableMemory) * float64(100)
	statsMap["cpu_percent"] = float64(cpuDelta/systemDelta) * float64(100) * float64(stats.CPUStats.OnlineCPUs)

	statsMap["mainstat"] = statsMap["cpu_percent"]
	statsMap["secondarystat"] = statsMap["memory_percent"]
	return statsMap, nil
}
