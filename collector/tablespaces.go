package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixTbs = "tablespaces"
	// Scrape querys
	// Query tbsUsageQuery is based on Tim Hall's ts_fee_space.sql script.
	//   Detailed information about UNDO tbs was added by @odbaeu.
	tbsUsageQuery = `
      SELECT a.tablespace_name,
             b.tbs_size size_bytes,
             a.free free_bytes,
             b.max_size max_size_bytes,
             a.free + (b.max_size - b.tbs_size) AS max_free_bytes
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
              FROM   dba_tablespaces
              WHERE  contents = 'PERMANENT') c
      WHERE  a.tablespace_name = b.tablespace_name
      AND    b.tablespace_name = c.tablespace_name
      UNION ALL
      SELECT par.tablespace_name,
               fs.tbs_size size_bytes,
               fs.tbs_size - (SELECT VALUE db_block_size FROM v$parameter WHERE NAME LIKE 'db_block_size')*(ext.activeblks + ext.unexpiredblks) free_bytes,
               fs.tbs_max_size max_size_bytes,
               fs.tbs_max_size - (SELECT VALUE db_block_size FROM v$parameter WHERE NAME LIKE 'db_block_size')*(ext.activeblks + ext.unexpiredblks) max_free_bytes
          FROM (SELECT SUM(activeblks) activeblks, SUM(unexpiredblks) unexpiredblks FROM v$undostat) ext
              ,(SELECT VALUE tablespace_name FROM v$parameter WHERE NAME = 'undo_tablespace') par
              ,(SELECT tablespace_name, SUM(greatest(bytes,maxbytes)) tbs_max_size, SUM(bytes) tbs_size FROM dba_data_files GROUP BY tablespace_name) fs
         WHERE par.tablespace_name=fs.tablespace_name`
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
