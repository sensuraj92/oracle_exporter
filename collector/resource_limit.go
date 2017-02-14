package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixTbs = "resource_limit"
	// Scrape querys
	// Query tbsUsageQuery is based on Tim Hall's ts_fee_space.sql script.
	//   Detailed information about UNDO tbs was added by @odbaeu.
	resLimitQuery = `
      SELECT resource_name, current_utilization, max_utilization, initial_allocation init_limit
      FROM   v$resource_limit
      WHERE  resource_name IN ('processes','sessions','max_shared_servers', 
                               'max_parallel_servers', 'enqueue_locks')`
)

var (
	processesGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "processes"),
		"",
		[]string{"tablespace_name"}, nil,
	)
	sessionsGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "sessions"),
		"",
		[]string{"tablespace_name"}, nil,
	)
	maxSharedServerGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "max_shared_servers"),
		"",
		[]string{"tablespace_name"}, nil,
	)
	maxParallelServersGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "max_parallel_servers"),
		"",
		[]string{"tablespace_name"}, nil,
	)
	enqueueLocksGauge = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "enqueue_locks"),
		"",
		[]string{"tablespace_name"}, nil,
	)
)

// ScrapeTablespaceUsage gathers metrics about tablespace usage.
func ScrapeTablespaceUsage(db *sql.DB, ch chan<- prometheus.Metric) error {
	/*	var (
			tbsName                        string
			sizeBytes, freeBytes           float64
			maxSizeBytes, maxSizeFreeBytes float64
		)

		// Tablespace Usage
		tbsRows, err := db.Query(tbsUsageQuery)
		if err != nil {
			return err
		}
		defer tbsRows.Close()

		for tbsRows.Next() {
			if err := tbsRows.Scan(
				&tbsName, &sizeBytes, &freeBytes, &maxSizeBytes, &maxSizeFreeBytes,
			); err != nil {
				return err
			}

			ch <- prometheus.MustNewConstMetric(
				tablespaceSizeBytes,
				prometheus.GaugeValue,
				sizeBytes, tbsName,
			)
			ch <- prometheus.MustNewConstMetric(
				tablespaceFreeBytes,
				prometheus.GaugeValue,
				freeBytes, tbsName,
			)
			ch <- prometheus.MustNewConstMetric(
				tablespaceMaxSizeBytes,
				prometheus.GaugeValue,
				maxSizeBytes, tbsName,
			)
			ch <- prometheus.MustNewConstMetric(
				tablespaceMaxFreeBytes,
				prometheus.GaugeValue,
				maxSizeFreeBytes, tbsName,
			)
		}
	*/
	return nil
}
