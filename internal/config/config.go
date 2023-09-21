package config

import (
	"github.com/spf13/viper"
	"github.com/xtclalala/infoK1t/internal/output"
	"strings"
	"sync"
)

type Options struct {
	Ping    Ping   `yaml:"ping"`
	Probe   Probe  `yaml:"probe"`
	Logger  Logger `yaml:"logger"`
	Dns     Dns    `yaml:"dns"`
	Gateway int    `yaml:"gateway"`
	Active  string `yaml:"active"`
}

type Ping struct {
}

type Logger struct {
	File []string `yaml:"file"`
	Mode string   `yaml:"mode"`
}

type Probe struct {
}

type Dns struct {
	Servers []string `yaml:"servers"`
}

var options Options
var once sync.Once

func GetOptions() *Options {
	once.Do(func() {
		InitConfig()
	})
	return &options
}

func InitConfig() {
	viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.infoK1t")
	viper.ReadInConfig()
	err := viper.Unmarshal(&options)
	if err != nil {
		output.GetLogger().Panic(err.Error())
	}
}

func IsDev() bool {
	return strings.EqualFold(options.Active, "dev")
}
