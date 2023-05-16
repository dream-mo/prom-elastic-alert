package boot

import (
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
)

type Loader interface {
	InjectConfig(config map[string]any)
	GetRules() map[string]*conf.Rule
	ReloadSchedulerJob(*ElasticAlert)
}

var (
	maps map[string]reflect.Type
)

func NewLoaderInstance(t string) Loader {
	T, ok := maps[t]
	if !ok {
		s := fmt.Sprintf("%s not exists", t)
		panic(s)
	}
	v := reflect.New(T)
	i := v.Interface()
	instance := i.(Loader)
	_ = defaults.Set(instance)
	return instance
}

func init() {
	maps = map[string]reflect.Type{
		"FileLoader": reflect.TypeOf(FileLoader{}),
	}
}
