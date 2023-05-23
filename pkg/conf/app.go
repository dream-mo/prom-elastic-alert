package conf

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	jsonYaml "github.com/ghodss/yaml"
	"github.com/openinsight-proj/elastic-alert/pkg/utils"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

var AppConf *AppConfig

type EsConfig struct {
	Addresses   []string `yaml:"addresses"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	ConnTimeout uint     `yaml:"conn_timeout" default:"10"`
	Version     string   `yaml:"version" default:"v7"`
}

type Datasource struct {
	Type   string         `yaml:"type" default:"file"`
	Config map[string]any `yaml:"config"`
}

// AppConfig is application global configure
type AppConfig struct {
	Server struct {
		ListenAddr string     `yaml:"listen_addr" default:":8080"`
		Enabled    bool       `yaml:"enabled" default:"true"`
		DB         Datasource `yaml:"datasource"`
	} `yaml:"server"`

	Exporter struct {
		ListenAddr string `yaml:"listen_addr" default:":9003"`
		Enabled    bool   `yaml:"enabled" default:"true"`
	} `yaml:"exporter"`
	Loader struct {
		Type   string         `yaml:"type" default:"FileLoader"`
		Config map[string]any `yaml:"config"`
	} `yaml:"loader"`
	Alert struct {
		Alertmanager struct {
			Enabled   bool   `yaml:"enabled" default:"false"`
			Url       string `yaml:"url"`
			BasicAuth struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"basic_auth"`
		} `yaml:"alertmanager"`
		Generator struct {
			BaseUrl string          `yaml:"base_url"`
			Expire  xtime.TimeLimit `yaml:"expire"`
		} `yaml:"generator"`
	} `yaml:"alert"`
	Redis struct {
		Addr         string `yaml:"addr" default:"127.0.0.1"`
		Port         int    `yaml:"port" default:"6379"`
		Password     string `yaml:"password"`
		Db           int    `yaml:"db" default:"0"`
		PoolSize     int    `yaml:"pool_size" default:"512"`
		ReadTimeout  int    `yaml:"read_timeout" default:"30"`
		WriteTimeout int    `yaml:"write_timeout" default:"30"`
		DialTimeout  int    `yaml:"dial_timeout" default:"5"`
	} `yaml:"redis"`
	RunEvery          xtime.TimeLimit `yaml:"run_every"`
	BufferTime        xtime.TimeLimit `yaml:"buffer_time"`
	AlertTimeLimit    xtime.TimeLimit `yaml:"alert_time_limit"`
	MaxScrollingCount uint            `yaml:"max_scrolling_count" default:"5"`
}

// FlagOption is application run args
type FlagOption struct {
	ConfigPath string `short:"c" long:"config" description:"config.yaml path" default:"./config.yaml"`
	Debug      bool   `long:"debug" description:"debug log level"`
	Verbose    string `short:"v" long:"verbose" description:"log level: debug、info、warn、error" default:"info"`
	Rule       string `long:"rule" description:"will only run the given single rule. The rule file may be a complete file path"`
	Zone       string `long:"zone" description:"time zone, e.g like PRC、UTC" default:"PRC"`
}

// GetLogLevel can get application log level
func (f FlagOption) GetLogLevel() log.Level {
	if f.Debug {
		return log.DebugLevel
	}
	switch f.Verbose {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn", "warning":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		panic("log level is not validated!")
	}
}

func (f FlagOption) IsDebug() bool {
	return f.GetLogLevel() == log.DebugLevel
}

func GetAppConfig(path string) *AppConfig {
	ok, _ := utils.PathExists(path)
	if !ok {
		t := fmt.Sprintf("%s is not exists", path)
		panic(t)
	}
	c := &AppConfig{}
	confBytes, e := os.ReadFile(path)
	if e != nil {
		t := fmt.Sprintf("read %s error %s", path, e)
		panic(t)
	}
	// set struct object default value
	e = defaults.Set(c)
	if e != nil {
		t := fmt.Sprintf("defaults set object error %s", e)
		panic(t)
	}
	appSchemaJson, _ := jsonYaml.YAMLToJSON([]byte(AppYamlSchema))
	appSchemaLoader := gojsonschema.NewBytesLoader(appSchemaJson)
	appConfJson, _ := jsonYaml.YAMLToJSON(confBytes)
	appConfLoader := gojsonschema.NewBytesLoader(appConfJson)
	res, e := gojsonschema.Validate(appSchemaLoader, appConfLoader)
	if e != nil {
		panic("config file schema error: " + e.Error())
	}
	if !res.Valid() {
		panic("config file schema error: " + res.Errors()[0].String())
	}
	e = yaml.Unmarshal(confBytes, &c)
	if c.Loader.Type == "FileLoader" && c.Loader.Config == nil {
		c.Loader.Config = map[string]any{
			"rules_folder":           "rules",
			"rules_folder_recursion": false,
		}
	}
	if e != nil {
		t := fmt.Sprintf("yaml Unmarshal %s error %s", path, e)
		panic(t)
	}
	AppConf = c
	return c
}
