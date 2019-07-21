package internal

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"strings"
	"sync"
)

var (
	confOnce sync.Once
	config   *Config
)

type Config struct {
	ConfYaml
}

func NewConfig() *Config {
	return &Config{}
}

func GetConfig() *Config {
	confOnce.Do(func() {
		config = NewConfig()
	})
	return config
}

var defaultConf = []byte(`
core:
  mode: "release" # release, debug, test
  work_pool_size: 1000
{{ if Postgres }}
postgres:
  host: ""
  port: 5432
  db: ""
  user: ""
  pass: ""
  batch_count: 5
{{ end }}
{{ if Kafka }}
kafka:
  bootstrap_servers: ""
  group_id: "random"
  auto_offset_reset: "earliest"
  topic: "{{ProjectName}}"
{{ end }}
prometheus:
  port: 8080
log:
  level: debug
endpoints:
{{ if Grpc }}
  grpc:
    address: "127.0.0.1:5050"
{{ end }}
  http:
    address: ":4040"
    user: "test"
    pass: "test"
`)

type ConfYaml struct {
	Core       SectionCore       `yaml:"core"`
	{{ if Postgres }}
	Postgres   SectionPostgres   `yaml:"postgres"`
	{{ end }}
    {{ if Kafka }}
	Kafka      SectionKafka      `yaml:"kafka"`
	{{ end }}
	Prometheus SectionPrometheus `yaml:"prometheus"`
	Log        SectionLog        `yaml:"log"`
	Endpoints  SectionEndpoints  `yaml:"endpoints"`
}

// SectionCore is sub section of config.
type SectionCore struct {
	Mode         string `yaml:"mode"`
	WorkPoolSize int    `yaml:"work_pool_size"`
}

{{ if Postgres }}
// SectionPostgres is sub section of config.
type SectionPostgres struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	DB         string `yaml:"db"`
	User       string `yaml:"user"`
	Pass       string `yaml:"pass"`
	BatchCount int    `yaml:"batch_count"`
}
{{ end }}

{{ if Kafka }}
// SectionKafka is sub section of config.
type SectionKafka struct {
	BootstrapServers string `yaml:"bootstrap_servers"`
	GroupId          string `yaml:"group_id"`
	AutoOffsetReset  string `yaml:"auto_offset_reset"`
	Topic            string `yaml:"topic"`
}
{{ end }}

type SectionPrometheus struct {
	Port int `yaml:"port"`
}

type SectionLog struct {
	Level string `yaml:"level"`
}

type SectionEndpoints struct {
	{{ if Grpc }}
	Grpc SectionGrpc `yaml:"grpc"`
	{{ end }}
	Http SectionHttp `yaml:"http"`
}

{{ if Grpc }}
type SectionGrpc struct {
	Address string `yaml:"address"`
}
{{ end }}

type SectionHttp struct {
	Address string `yaml:"address"`
	User    string `yaml:"user"`
	Pass    string `yaml:"pass"`
}

// LoadConf load config from file and read in environment variables that match
func (config *Config) LoadConf(confPath string) (ConfYaml, error) {
	var conf ConfYaml

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()             // read in environment variables that match
	viper.SetEnvPrefix("{{ProjectName}}") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if confPath != "" {
		content, err := ioutil.ReadFile(confPath)

		if err != nil {
			log.Errorf("File does not exist : %s", confPath)
			return conf, err
		}

		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return conf, err
		}
	} else {
		// Search config in home directory with name ".pkg" (without extension).
		viper.AddConfigPath("/etc/{{ProjectName}}/")
		viper.AddConfigPath("$HOME/.{{ProjectName}}")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")

		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			// load default config
			if err := viper.ReadConfig(bytes.NewBuffer(defaultConf)); err != nil {
				return conf, err
			}
		}
	}

	// Core
	conf.Core.Mode = viper.GetString("core.mode")
	conf.Core.WorkPoolSize = viper.GetInt("core.work_pool_size")

	{{ if Postgres }}
	// Postgres
	conf.Postgres.Host = viper.GetString("postgres.host")
	conf.Postgres.Port = viper.GetInt("postgres.port")
	conf.Postgres.DB = viper.GetString("postgres.db")
	conf.Postgres.User = viper.GetString("postgres.user")
	conf.Postgres.Pass = viper.GetString("postgres.pass")
	conf.Postgres.BatchCount = viper.GetInt("postgres.batch_count")
	{{ end }}

	{{ if Kafka }}
	// Kafka
	conf.Kafka.BootstrapServers = viper.GetString("kafka.bootstrap_servers")
	conf.Kafka.GroupId = viper.GetString("kafka.group_id")
	conf.Kafka.AutoOffsetReset = viper.GetString("kafka.auto_offset_reset")
	conf.Kafka.Topic = viper.GetString("kafka.topic")
	{{ end }}

	// Prometheus
	conf.Prometheus.Port = viper.GetInt("prometheus.port")

	//Log
	conf.Log.Level = viper.GetString("log.level")

	//Endpoints
	{{ if Grpc }}
	conf.Endpoints.Grpc.Address = viper.GetString("endpoints.grpc.address")
	{{ end }}
	conf.Endpoints.Http.Address = viper.GetString("endpoints.http.address")
	conf.Endpoints.Http.User = viper.GetString("endpoints.http.user")
	conf.Endpoints.Http.Pass = viper.GetString("endpoints.http.pass")

	return conf, nil
}

func (config *Config) Initialize(path string) {
	var err error
	config.ConfYaml, err = config.LoadConf(path)
	if err != nil {
		log.Fatalf("Load yaml config file error: '%v'", err)
		return
	}
}