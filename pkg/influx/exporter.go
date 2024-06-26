package influx

import (
	"context"
	"fmt"
	"time"

	"github.com/ai4networks/net4me/pkg/topology"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/sirupsen/logrus"
)

func Graph(writer api.WriteAPIBlocking) error {
	nodePoints := make([]*write.Point, 0)
	for _, h := range topology.Hosts() {
		if h.State() == topology.HostStateRunning {
			stats, err := h.Stats()
			if err != nil {
				logrus.WithError(err).Errorln("failed to get host stats for export")
			}
			nodePoints = append(nodePoints, influxdb2.NewPoint("nodes",
				map[string]string{
					"node": h.ID(),
				},
				map[string]interface{}{
					"topology":      topology.ID(),
					"id":            h.ID(),
					"title":         fmt.Sprintf("Name: %s", h.Name()),
					"subtitle":      fmt.Sprintf("Device: %s", h.Device()),
					"mainstat":      stats["mainstat"],
					"secondarystat": stats["secondarystat"],
					"icon":          h.Node().Manager().Icon(),
					"color":         h.Node().Manager().Color(),
				},
				time.Now(),
			))
		}
	}
	if err := writer.WritePoint(context.Background(), nodePoints...); err != nil {
		return fmt.Errorf("failed to write host data to influx: %w", err)
	} else {
		logrus.
			WithField("timestamp", time.Now()).
			WithField("node_count", len(nodePoints)).
			Debugln("graph: exported node data to influx")
	}

	linkPoints := make([]*write.Point, 0)
	for _, l := range topology.Links() {
		linkPoints = append(linkPoints, influxdb2.NewPoint("edges",
			map[string]string{
				"edges": l.SelfPort().Attrs().Name,
			},
			map[string]interface{}{
				"topology":      topology.ID(),
				"id":            l.SelfPort().Attrs().Index,
				"source":        l.SelfHost().ID(),
				"target":        l.PeerHost().ID(),
				"mainstat":      l.SelfPort().Attrs().Statistics.TxBytes,
				"secondarystat": l.SelfPort().Attrs().Statistics.TxPackets,
			},
			time.Now(),
		))
	}
	if err := writer.WritePoint(context.Background(), linkPoints...); err != nil {
		return fmt.Errorf("failed to write link data to influx: %w", err)
	} else {
		logrus.
			WithField("timestamp", time.Now()).
			WithField("link_count", len(linkPoints)).
			Debugln("graph: exported link data to influx")
	}

	return nil
}

func RunExporter(URL, token, org, bucket string, interval int) (<-chan error, error) {
	errCh := make(chan error)
	client := influxdb2.NewClientWithOptions(URL, token, influxdb2.DefaultOptions().SetBatchSize(20).AddDefaultTag("net4me", "true").AddDefaultTag("topology", topology.ID()))
	alive, err := client.Ping(context.Background())
	if err != nil {
		return errCh, fmt.Errorf("failed to ping influx instance: %w", err)
	}
	if !alive {
		return errCh, fmt.Errorf("influx instance is not alive")
	}
	writeAPI := client.WriteAPIBlocking(org, bucket)
	go func() {
		defer close(errCh)
		for {
			if err := Graph(writeAPI); err != nil {
				errCh <- err
				return
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
	return errCh, nil
}
