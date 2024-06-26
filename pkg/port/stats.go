package port

func Statistics(p Port) map[string]any {
	// TODO: Implement this function
	return map[string]any{
		"rx_bytes":   p.Attrs().Statistics.RxBytes,
		"rx_packets": p.Attrs().Statistics.RxPackets,
		"tx_bytes":   p.Attrs().Statistics.TxBytes,
		"tx_packets": p.Attrs().Statistics.TxPackets,
		"rx_dropped": p.Attrs().Statistics.RxDropped,
		"tx_dropped": p.Attrs().Statistics.TxDropped,
		"rx_errors":  p.Attrs().Statistics.RxErrors,
		"tx_errors":  p.Attrs().Statistics.TxErrors,
		"collisions": p.Attrs().Statistics.Collisions,
	}
}
