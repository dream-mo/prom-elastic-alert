package model

import (
	"testing"
	"time"

	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logger.SetLogLevel(log.InfoLevel)
}

func TestBooleanQueryGetDSL(t *testing.T) {
	end := time.Unix(1701187340, 0)
	start := end.Add(-time.Minute * 5)
	from := 1
	size := 10
	t.Run("test build boolean query", func(t *testing.T) {
		b := BooleanQuery(`{"bool":{"filter":[{"query_string":{"query":"log:error"}}]}}`)
		expectDSL := `{"from":1,"query":{"bool":{"filter":[{"query_string":{"query":"log:error"}},{"range":{"time":{"format":"strict_date_optional_time","gte":"2023-11-28T15:57:20Z","lte":"2023-11-28T16:02:20Z"}}}]}},"size":10,"sort":[{"time":{"order":"asc"}}]}`

		dsl := b.GetDSL(from, size, start, end)
		require.Equal(t, expectDSL, dsl)
	})
	t.Run("test empty BooleanQuery", func(t *testing.T) {
		b := BooleanQuery("")
		expectDSL := `{"from":1,"query":{},"size":10,"sort":[{"time":{"order":"asc"}}]}`
		dsl := b.GetDSL(from, size, start, end)
		require.Equal(t, expectDSL, dsl)
	})
	t.Run("test miss bool clause in BooleanQuery", func(t *testing.T) {
		b := BooleanQuery(`{"match":[{"query_string":{"query":"log:error"}}]}`)
		expectDSL := `{"from":1,"query":{"match":[{"query_string":{"query":"log:error"}}]},"size":10,"sort":[{"time":{"order":"asc"}}]}`
		dsl := b.GetDSL(from, size, start, end)
		require.Equal(t, expectDSL, dsl)
	})
	t.Run("test miss filter clause in bool clause", func(t *testing.T) {
		b := BooleanQuery(`{"bool":{"match":[{"query_string":{"query":"log:error"}}]}}`)
		expectDSL := `{"from":1,"query":{"bool":{"filter":[{"range":{"time":{"format":"strict_date_optional_time","gte":"2023-11-28T15:57:20Z","lte":"2023-11-28T16:02:20Z"}}}],"match":[{"query_string":{"query":"log:error"}}]}},"size":10,"sort":[{"time":{"order":"asc"}}]}`
		dsl := b.GetDSL(from, size, start, end)
		require.Equal(t, expectDSL, dsl)
	})
	t.Run("test exist time range in filter clause", func(t *testing.T) {
		b := BooleanQuery(`{"bool":{"filter":[{"range":{"time":{"format":"strict_date_optional_time","gte":"2022-11-28T15:57:20Z","lte":"2022-11-28T16:02:20Z"}}},{"query_string":{"query":"log:error"}}]}}`)
		expectDSL := `{"from":1,"query":{"bool":{"filter":[{"query_string":{"query":"log:error"}},{"range":{"time":{"format":"strict_date_optional_time","gte":"2023-11-28T15:57:20Z","lte":"2023-11-28T16:02:20Z"}}}]}},"size":10,"sort":[{"time":{"order":"asc"}}]}`
		dsl := b.GetDSL(from, size, start, end)
		require.Equal(t, expectDSL, dsl)
	})
}

func TestBooleanQueryGetCountDSL(t *testing.T) {
	end := time.Unix(1701187340, 0)
	start := end.Add(-time.Minute * 5)

	b := BooleanQuery(`{"bool":{"filter":[{"query_string":{"query":"log:error"}}]}}`)
	expectDSL := `{"query":{"bool":{"filter":[{"query_string":{"query":"log:error"}},{"range":{"time":{"format":"strict_date_optional_time","gte":"2023-11-28T15:57:20Z","lte":"2023-11-28T16:02:20Z"}}}]}}}`

	dsl := b.GetCountDSL(start, end)
	require.Equal(t, expectDSL, dsl)

}
