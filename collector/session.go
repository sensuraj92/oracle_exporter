package collector

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	sess = "session"
	// Scrape querys
	userSessionCountQuery = `
      SELECT (SELECT 
               COUNT(*) 
                FROM v$session
               WHERE type = 'USER'
                 AND status <> 'ACTIVE'
             ) inactive_usr_cnt,
             (SELECT 
               COUNT(*)
                FROM v$session
               WHERE type = 'USER'
                 AND status = 'ACTIVE'
             ) active_usr_cnt
        FROM dual`
	systemSessionCountQuery = `
      SELECT 
       COUNT(*) total_sys_cnt
        FROM v$session
       WHERE type = 'BACKGROUND'`
)

var (
	userSessionInactiveCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sess, "inactive_user_session"),
		"Number of inactive user sessions connected.",
		[]string{}, nil,
	)
	userSessionActiveCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sess, "active_user_session"),
		"Number of active user sessions connected.",
		[]string{}, nil,
	)
	systemSessionCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, sess, "system_session"),
		"Number of system sessions connected.",
		[]string{}, nil,
	)
)

// ScrapeUserSessionCount gathers session metrics from v$session.
func ScrapeUserSessionCount(db *sql.DB, ch chan<- prometheus.Metric) error {
	var (
		err             error
		sessInactiveCnt float64
		sessActiveCnt   float64
	)

	// User sessions
	err = db.QueryRow(userSessionCountQuery).Scan(&sessInactiveCnt, &sessActiveCnt)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		userSessionInactiveCount,
		prometheus.GaugeValue,
		sessInactiveCnt,
	)
	ch <- prometheus.MustNewConstMetric(
		userSessionActiveCount,
		prometheus.GaugeValue,
		sessActiveCnt,
	)

	// System sessions
	err = db.QueryRow(systemSessionCountQuery).Scan(&sessActiveCnt)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		systemSessionCount,
		prometheus.GaugeValue,
		sessActiveCnt,
	)

	return nil
}
