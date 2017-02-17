package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixResLimit = "resource_limit"
	// Scrape querys
	resLimitQuery = `
      SELECT resource_name, current_utilization, max_utilization, to_number(initial_allocation) init_limit
      FROM   v$resource_limit
      WHERE  resource_name IN ('processes','sessions','max_shared_servers', 
                               'max_parallel_servers', 'enqueue_locks')`
)

//
//
// TODO: Change metrics to be aggregatable...
//
//
//
//
//

var (
	currentCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixResLimit, "current_utilization"),
		"Current utilization of a resource limit.",
		[]string{"resource_name"}, nil,
	)
	maxCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixResLimit, "max_utilization"),
		"Max utilization of a resource limit since startup.",
		[]string{"resource_name"}, nil,
	)
	initLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixResLimit, "init_limit"),
		"Configured resource limit in pfile or spfile.",
		[]string{"resource_name"}, nil,
	)
)

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

		ch <- prometheus.MustNewConstMetric(
			currentCount,
			prometheus.GaugeValue,
			curUtil, resName,
		)
		ch <- prometheus.MustNewConstMetric(
			maxCount,
			prometheus.GaugeValue,
			maxUtil, resName,
		)
		ch <- prometheus.MustNewConstMetric(
			initLimit,
			prometheus.GaugeValue,
			iLimit, resName,
		)
	}

	return nil
}
