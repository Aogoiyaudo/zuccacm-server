package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	LogConfig
	ServerConfig
	Secret
}

type LogConfig struct {
	Level string
	Path  string
}

type ServerConfig struct {
	Port int
}

type Secret struct {
	SessionKey string
	SSO_URL    string
	DBConfig
	MessageQueue string
}

type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Pwd      string
}

var Instance *Config

const (
	RootDir           = "zuccacm-server"
	DefaultConfigFile = "./.zuccacm-server.yaml"
)

// Have to do this to get constants from config file
func init() {
	Init(DefaultConfigFile)
}

func Init(cfgFile string) {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.WithField("error", err).Fatal("Read config file failed!")
	}
	if err := viper.Unmarshal(&Instance); err != nil {
		log.WithField("error", err).Fatal("Unmarshal config file failed!")
	}
	path, level := initLog()
	log.WithField("File", cfgFile).Info("Read config file succeed!")
	log.WithFields(log.Fields{
		"Path":  path,
		"Level": level,
	}).Info("Set log succeed")
}
