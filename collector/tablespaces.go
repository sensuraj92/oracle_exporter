package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixTbs = "tablespaces"
	// Scrape querys
	tbsUsageQuery = `
      SELECT a.tablespace_name,
             b.tbs_size size_bytes,
             CASE 
			   WHEN c.contents = 'UNDO'
			   THEN b.tbs_size-d.used_bytes
			   ELSE a.free
			 END free_bytes,
             b.max_size max_size_bytes,
             CASE
			   WHEN c.contents = 'UNDO' 
			   THEN b.tbs_size-d.used_bytes
			   ELSE a.free
			 END + (b.max_size - b.tbs_size) AS max_free_bytes
      FROM   (SELECT tablespace_name,
                     SUM(bytes) free
              FROM   dba_free_space
              GROUP BY tablespace_name) a,
             (SELECT tablespace_name,
                     SUM(bytes) tbs_size,
                     SUM(GREATEST(bytes,maxbytes)) max_size
              FROM   dba_data_files
              GROUP BY tablespace_name) b,
             (SELECT tablespace_name, contents
              FROM   dba_tablespaces) c,
             (SELECT tablespace_name, SUM(bytes) used_bytes
              FROM   dba_undo_extents
              WHERE  status in ('ACTIVE', 'UNEXPIRED')
              GROUP BY tablespace_name) d
      WHERE  b.tablespace_name = a.tablespace_name
      AND    b.tablespace_name = c.tablespace_name
      AND    b.tablespace_name = d.tablespace_name(+)`
)

var (
	tablespaceSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "size_bytes"),
		"Tablespace size in Bytes.",
		[]string{"tablespace_name"}, nil,
	)
	tablespaceFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "free_bytes"),
		"Tablespace free space in Bytes.",
		[]string{"tablespace_name"}, nil,
	)
	tablespaceMaxSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "max_size_bytes"),
		"Tablespace maximum size in Bytes. Maxsize is the limit of autoextend config. If autoextend is disabled then max_size_bytes = size_bytes.",
		[]string{"tablespace_name"}, nil,
	)
	tablespaceMaxFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixTbs, "max_free_bytes"),
		"Tablespace maximum free space in Bytes. Maxsize is the limit of autoexend config. If autoextend is disabled then max_size_bytes = size_bytes.",
		[]string{"tablespace_name"}, nil,
	)
)

// ScrapeTablespaceUsage gathers metrics about tablespace usage.
func ScrapeTablespaceUsage(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
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

	return nil
}
