package boot

import (
	"testing"
	"time"

	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/model"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
	"github.com/stretchr/testify/assert"
)

func TestGetBufferTime(t *testing.T) {
	t.Run("test default timeframe", func(t *testing.T) {
		d := getBufferTime(&model.Rule{}, &conf.AppConfig{})
		assert.Equal(t, time.Minute, d)
	})
	t.Run("test empty model.rule timeframe", func(t *testing.T) {
		d := getBufferTime(&model.Rule{}, &conf.AppConfig{
			BufferTime: xtime.TimeLimit{
				Seconds: 0,
				Minutes: 5,
				Days:    0,
			},
		})
		assert.Equal(t, time.Minute*5, d)
	})
	t.Run("test empty AppConfig.BufferTime", func(t *testing.T) {
		d := getBufferTime(&model.Rule{
			Query: model.Query{
				Config: model.QueryConfig{
					Timeframe: xtime.TimeLimit{
						Seconds: 5,
					},
				},
			},
			RawContent: "",
			FilePath:   "",
		}, &conf.AppConfig{})
		assert.Equal(t, time.Second*5, d)
	})
	t.Run("test model.Rule is priority of AppConfig.BufferTime", func(t *testing.T) {
		d := getBufferTime(&model.Rule{
			Query: model.Query{
				Config: model.QueryConfig{
					Timeframe: xtime.TimeLimit{
						Seconds: 5,
					},
				},
			},
			RawContent: "",
			FilePath:   "",
		}, &conf.AppConfig{
			BufferTime: xtime.TimeLimit{
				Seconds: 0,
				Minutes: 5,
				Days:    0,
			},
		})
		assert.Equal(t, time.Second*5, d)
	})
}
