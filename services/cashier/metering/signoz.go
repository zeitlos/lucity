package metering

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// SigNozClient queries SigNoz's ClickHouse for container metrics.
type SigNozClient struct {
	conn clickhouse.Conn
}

// NewSigNozClient connects to SigNoz's ClickHouse instance.
func NewSigNozClient(dsn string) (*SigNozClient, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clickhouse DSN: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return &SigNozClient{conn: conn}, nil
}

// Close closes the ClickHouse connection.
func (c *SigNozClient) Close() error {
	return c.conn.Close()
}

// namespaceFilter builds a SQL IN clause for namespace filtering.
func namespaceFilter(namespaces []string) string {
	quoted := make([]string, len(namespaces))
	for i, ns := range namespaces {
		quoted[i] = "'" + strings.ReplaceAll(ns, "'", "\\'") + "'"
	}
	return strings.Join(quoted, ",")
}

// CPUByNamespace returns total CPU-seconds consumed per namespace over the given time range.
// Uses container.cpu.time (cumulative counter) — delta per time series, summed per namespace.
func (c *SigNozClient) CPUByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()
	nsFilter := namespaceFilter(namespaces)

	query := fmt.Sprintf(`
		SELECT
			namespace,
			sum(per_series_delta) AS cpu_seconds
		FROM (
			SELECT
				fingerprint,
				max(value) - min(value) AS per_series_delta
			FROM signoz_metrics.distributed_samples_v4
			INNER JOIN (
				SELECT DISTINCT fingerprint, labels
				FROM signoz_metrics.distributed_time_series_v4
				WHERE metric_name = 'container.cpu.time'
				  AND temporality = 'Cumulative'
				  AND unix_milli >= %d
				  AND unix_milli < %d
				  AND JSONExtractString(labels, 'k8s.namespace.name') IN (%s)
			) AS ts USING (fingerprint)
			WHERE metric_name = 'container.cpu.time'
			  AND unix_milli >= %d
			  AND unix_milli < %d
			GROUP BY fingerprint
		) AS deltas
		INNER JOIN (
			SELECT DISTINCT fingerprint, JSONExtractString(labels, 'k8s.namespace.name') AS namespace
			FROM signoz_metrics.distributed_time_series_v4
			WHERE metric_name = 'container.cpu.time'
			  AND temporality = 'Cumulative'
			  AND unix_milli >= %d
			  AND unix_milli < %d
		) AS ts2 USING (fingerprint)
		GROUP BY namespace
	`, startMs, endMs, nsFilter, startMs, endMs, startMs, endMs)

	return c.queryFloat64Map(ctx, query)
}

// MemoryByNamespace returns average memory working set (bytes) per namespace over the given time range.
func (c *SigNozClient) MemoryByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()
	nsFilter := namespaceFilter(namespaces)

	query := fmt.Sprintf(`
		SELECT
			JSONExtractString(labels, 'k8s.namespace.name') AS namespace,
			avg(value) AS avg_memory_bytes
		FROM signoz_metrics.distributed_samples_v4
		INNER JOIN (
			SELECT DISTINCT fingerprint, labels
			FROM signoz_metrics.distributed_time_series_v4
			WHERE metric_name = 'container.memory.working_set'
			  AND temporality = 'Unspecified'
			  AND unix_milli >= %d
			  AND unix_milli < %d
			  AND JSONExtractString(labels, 'k8s.namespace.name') IN (%s)
		) AS ts USING (fingerprint)
		WHERE metric_name = 'container.memory.working_set'
		  AND unix_milli >= %d
		  AND unix_milli < %d
		GROUP BY namespace
	`, startMs, endMs, nsFilter, startMs, endMs)

	return c.queryFloat64Map(ctx, query)
}

// DiskByNamespace returns total PVC capacity (bytes) per namespace (latest value per volume).
func (c *SigNozClient) DiskByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}

	startMs := start.UnixMilli()
	endMs := end.UnixMilli()
	nsFilter := namespaceFilter(namespaces)

	query := fmt.Sprintf(`
		SELECT
			namespace,
			sum(latest_cap) AS total_capacity_bytes
		FROM (
			SELECT fingerprint, argMax(value, unix_milli) AS latest_cap
			FROM signoz_metrics.distributed_samples_v4
			INNER JOIN (
				SELECT DISTINCT fingerprint, labels
				FROM signoz_metrics.distributed_time_series_v4
				WHERE metric_name = 'k8s.volume.capacity'
				  AND temporality = 'Unspecified'
				  AND unix_milli >= %d
				  AND unix_milli < %d
				  AND JSONExtractString(labels, 'k8s.namespace.name') IN (%s)
			) AS ts USING (fingerprint)
			WHERE metric_name = 'k8s.volume.capacity'
			  AND unix_milli >= %d
			  AND unix_milli < %d
			GROUP BY fingerprint
		) AS latest
		INNER JOIN (
			SELECT DISTINCT fingerprint, JSONExtractString(labels, 'k8s.namespace.name') AS namespace
			FROM signoz_metrics.distributed_time_series_v4
			WHERE metric_name = 'k8s.volume.capacity'
			  AND temporality = 'Unspecified'
			  AND unix_milli >= %d
			  AND unix_milli < %d
		) AS ts2 USING (fingerprint)
		GROUP BY namespace
	`, startMs, endMs, nsFilter, startMs, endMs, startMs, endMs)

	return c.queryFloat64Map(ctx, query)
}

// queryFloat64Map executes a query that returns (string, float64) rows and returns a map.
func (c *SigNozClient) queryFloat64Map(ctx context.Context, query string) (map[string]float64, error) {
	rows, err := c.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var key string
		var val float64
		if err := rows.Scan(&key, &val); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[key] = val
	}
	return result, rows.Err()
}
