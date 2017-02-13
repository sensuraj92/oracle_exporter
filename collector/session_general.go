package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	sess = "sessions"
	// Scrape querys
	sessionCountQuery = `
      SELECT
       'user_sessions' a,
       COUNT(*) b
        FROM v$session
       WHERE type = 'USER'
      UNION ALL
      SELECT 
       'background_sessions' a,
       COUNT(*) b
        FROM v$session
       WHERE type <> 'USER'`
	activeUserSessionCountQuery = `
      SELECT 
       COUNT(*) active_user_sessions
        FROM v$session
       WHERE type = 'USER'
         AND status = 'ACTIVE'`
)

var (
	sessionCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sess, "total"),
		"Number of sessions currently connected.",
		[]string{"session_type"}, nil,
	)
	activeUserSessionCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sess, "active_user_total"),
		"Number of active user sessions.",
		[]string{}, nil,
	)
)

// ScrapeUserSessionCount gathers session metrics from v$session.
func ScrapeUserSessionCount(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		//		err          error
		sessCnt      float64
		sessCntLable string
	)

	// Sessions count
	sessCntRows, err := db.Query(sessionCountQuery)
	if err != nil {
		return err
	}
	defer sessCntRows.Close()

	for sessCntRows.Next() {
		if err := sessCntRows.Scan(
			&sessCntLable, &sessCnt,
		); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			sessionCount,
			prometheus.GaugeValue,
			sessCnt, sessCntLable,
		)
	}

	// Active sessions count
	err = db.QueryRow(activeUserSessionCountQuery).Scan(&sessCnt)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		activeUserSessionCount,
		prometheus.GaugeValue,
		sessCnt,
	)
	return nil
}
