package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefixArchivelogs = "archivelog_stats"
	archivelogQuery   = `
          SELECT *
          FROM   (SELECT to_char(inst_id) inst_id,
                         min(sequence#) min_sequence_unpivotme,
                         max(sequence#) max_sequence_unpivotme,
                         sum(sequence#) sequences_total_unpivotme,
                         sum(CASE WHEN archived='YES' THEN 1 ELSE 0 END) archived_total_unpivotme,
                         sum(CASE WHEN applied='YES' THEN 1 ELSE 0 END) applied_total_unpivotme,
                         round((min(first_time) - to_date('1970-01-01','YYYY-MM-DD'))*60*60*24) min_first_time_unpivotme,
                         round((min(next_time) - to_date('1970-01-01','YYYY-MM-DD'))*60*60*24) max_next_time_unpivotme,
                         sum(blocks*block_size) total_size_bytes_unpivotme,
                         round(avg(blocks*block_size), 0) avg_size_bytes_unpivotme
                  FROM   gv$archived_log
                  WHERE  first_time > SYSDATE-1/24
                  GROUP BY inst_id
                  )
          UNPIVOT (metr FOR unpivotme
                        IN (min_sequence_unpivotme AS 'min_sequence',
                            max_sequence_unpivotme AS 'max_sequence',
                            sequences_total_unpivotme AS 'sequences_total',
                            archived_total_unpivotme AS 'archived_total',
                            applied_total_unpivotme AS 'applied_total',
                            min_first_time_unpivotme AS 'min_first_time',
                            max_next_time_unpivotme AS 'max_next_time',
                            total_size_bytes_unpivotme AS 'total_size_bytes',
                            avg_size_bytes_unpivotme AS 'avg_size_bytes'
                           )
                  )`
	archCompletionTimeQuery = `
          SELECT to_char(inst_id) inst_id,
    	         round(avg((completion_time-next_time)*24*60*60*1000), 0) completion_avg_time_ms
          FROM   (SELECT * FROM gv$archived_log WHERE first_time > SYSDATE-0.25)
          WHERE  ROWNUM <= 5
          GROUP BY inst_id`
)

var (
	archCompletionAvgTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, prefixArchivelogs, "completion_avg_time_ms"),
		"Average archive completion time of last 5 archivelogs.",
		[]string{}, nil,
	)
)

// ScrapeArchivelogs gathers metrics about archivelogs.
func ScrapeArchivelogs(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		//		err          error
		instID, metricName string
		metricValue        float64
	)

	// Details about archivelogs in the last hour
	archRows, err := db.Query(archivelogQuery)
	if err != nil {
		return err
	}
	defer archRows.Close()

	for archRows.Next() {
		if err := archRows.Scan(
			&instID, &metricName, &metricValue,
		); err != nil {
			return err
		}

		var (
			archDetails = prometheus.NewDesc(
				prometheus.BuildFQName(namespace, prefixArchivelogs, metricName),
				"Details about archivelogs in the last hour.",
				[]string{"inst_id"}, nil,
			)
		)
		ch <- prometheus.MustNewConstMetric(
			archDetails,
			prometheus.GaugeValue,
			metricValue, instID,
		)
	}

	// Average completion time in ms
	err = db.QueryRow(archCompletionTimeQuery).Scan(&instID, &metricValue)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		archCompletionAvgTime,
		prometheus.GaugeValue,
		metricValue,
	)
	return nil
}
