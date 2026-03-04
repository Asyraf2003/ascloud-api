package obs

import (
	"log/slog"
	"sort"
	"time"
)

type MetricValue struct {
	Value float64
	Unit  string
}

func EmitEMF(log *slog.Logger, namespace string, dims map[string]string, metrics map[string]MetricValue) {
	if log == nil {
		return
	}

	keys := make([]string, 0, len(dims))
	for k := range dims {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	metricDefs := make([]map[string]any, 0, len(metrics))
	for name, mv := range metrics {
		def := map[string]any{"Name": name}
		if mv.Unit != "" {
			def["Unit"] = mv.Unit
		}
		metricDefs = append(metricDefs, def)
	}

	aws := map[string]any{
		"Timestamp": time.Now().UTC().UnixMilli(),
		"CloudWatchMetrics": []any{
			map[string]any{
				"Namespace":  namespace,
				"Dimensions": [][]string{keys},
				"Metrics":    metricDefs,
			},
		},
	}

	args := make([]any, 0, 2+len(keys)*2+len(metrics)*2)
	args = append(args, "_aws", aws)

	for _, k := range keys {
		args = append(args, k, dims[k])
	}

	for name, mv := range metrics {
		args = append(args, name, mv.Value)
	}

	log.Info("metric", args...)
}
