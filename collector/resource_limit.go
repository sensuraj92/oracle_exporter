package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixResLimit = "resources"
	// Scrape querys
	resLimitQuery = `
      SELECT resource_name, current_utilization, max_utilization, to_number(initial_allocation) limit
      FROM   v$resource_limit
      WHERE  resource_name IN ('processes','sessions','max_shared_servers', 
                               'max_parallel_servers', 'enqueue_locks')`
)

var ()

// ScrapeResourceLimit gathers metrics about resource limits in v$resource_limit.
func ScrapeResourceLimit(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		resName                  string
		curUtil, maxUtil, iLimit float64
	)

	// Tablespace Usage
	resLimitRows, err := db.Query(resLimitQuery)
	if err != nil {
		return err
	}
	defer resLimitRows.Close()

	for resLimitRows.Next() {
		if err := resLimitRows.Scan(
			&resName, &curUtil, &maxUtil, &iLimit,
		); err != nil {
			return err
		}

		var (
			currentCount = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, prefixResLimit, resName+"_current_count"),
				"Current resource consumption.",
				[]string{}, nil,
			)
			maxCount = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, prefixResLimit, resName+"_max_count"),
				"High Water Mark of resource since instance startup.",
				[]string{}, nil,
			)
			initLimit = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, prefixResLimit, resName+"_limit"),
				"Configured limit in pfile or spfile.",
				[]string{}, nil,
			)
		)

		ch <- prometheus.MustNewConstMetric(
			currentCount,
			prometheus.GaugeValue,
			curUtil,
		)
		ch <- prometheus.MustNewConstMetric(
			maxCount,
			prometheus.GaugeValue,
			maxUtil,
		)
		ch <- prometheus.MustNewConstMetric(
			initLimit,
			prometheus.GaugeValue,
			iLimit,
		)
	}

	return nil
}
