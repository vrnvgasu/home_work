package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/vrnvgasu/home_work/hw12_13_14_15_calendar/internal/logger"
)

type DBType string

const (
	DBTypeSQL    DBType = "sql"
	DBTypeMemory DBType = "memory"
)

var Cfg *Config

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger     LoggerConf `yaml:"logger"`
	PSQL       PSQLConfig `json:"psql"`
	DBType     `yaml:"dbType"`
	Server     ServerConf     `yaml:"server"`
	GRPSServer GRPSServerConf `yaml:"grpsServer"`
	Rabbitmq   `yaml:"rabbitmq"`
	Scheduler  `yaml:"scheduler"`
}

type ServerConf struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type GRPSServerConf struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type LoggerConf struct {
	Level logger.LogLevel `yaml:"level"`
}

type PSQLConfig struct {
	DSN       string `yaml:"dsn"`
	Migration string `json:"migration"`
}

type Rabbitmq struct {
	URI          string `yaml:"uri"`
	ExchangeName string `yaml:"exchangeName"`
	ExchangeType string `yaml:"exchangeType"`
	RoutingKey   string `yaml:"routingKey"`
	Queue        string `yaml:"queue"`
	ConsumerTag  string `yaml:"consumerTag"`
}

type Scheduler struct {
	Ticker         int64 `yaml:"ticker"`
	EventsLifeTime int64 `yaml:"eventsLifeTime"`
}

func NewConfig(configFile string) *Config {
	c := Config{}

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Error parsing config file: %s", err.Error())
	}

	return &c
}
