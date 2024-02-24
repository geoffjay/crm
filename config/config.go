package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Env     string        `mapstructure:"env"`
	Log     logConfig     `mapstructure:"log"`
	Cors    corsConfig    `mapstructure:"cors"`
	Session sessionConfig `mapstructure:"session"`
}

var lock = &sync.Mutex{}
var instance *Config

var defaults = map[string]interface{}{
	"env":                      "development",
	"client-endpoint":          "tcp://localhost:9797",
	"log.formatter":            "text",
	"log.level":                "info",
	"log.loki.address":         "http://localhost:3100",
	"log.loki.labels":          map[string]string{"app": "app", "environment": "development"},
	"cors.allow-credentials":   true,
	"cors.allow-origins":       "*",
	"cors.allow-headers":       "Origin, Content-Type, Accept, Content-Length, Accept-Language, Accept-Encoding, Connection, Authorization, Access-Control-Allow-Origin, Access-Control-Allow-Methods, Access-Control-Allow-Headers, Access-Control-Allow-Origin",
	"cors.allow-methods":       "GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS",
	"session.expiration":       "2h",
	"session.key-lookup":       "cookie:__Host-session",
	"session.cookie-secure":    true,
	"session.cookie-http-only": true,
	"session.cookie-same-site": "Lax",
}

// GetConfig returns the application configuration singleton.
func GetConfig() *Config {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			if err := loadConfig(&instance, defaults); err != nil {
				log.Fatalf("error reading config file: %s\n", err)
			}
		}
	}

	log.Tracef("config: %+v", instance)

	return instance
}

func prepare() (*viper.Viper, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	config := viper.New()

	file := os.Getenv("CRM_CONFIG")
	if file == "" {
		config.SetConfigName("crm")
		config.AddConfigPath(".")
		config.AddConfigPath(fmt.Sprintf("%s/.config/crm", home))
		config.AddConfigPath("/etc/crm")
	} else {
		var extension string
		regex := regexp.MustCompile("((y(a)?ml)|json|toml)$")
		base := filepath.Base(file)
		if regex.Match([]byte(base)) {
			// strip the file type for viper
			parts := strings.Split(filepath.Base(file), ".")
			base = strings.Join(parts[:len(parts)-1], ".")
			extension = parts[len(parts)-1]
		} else {
			return nil, errors.New("configuration does not support that extension type")
		}
		config.SetConfigName(base)
		config.SetConfigType(extension)
		config.SetConfigFile(file)
		config.AddConfigPath(filepath.Dir(file))
	}

	return config, nil
}

func loadConfig(c interface{}, defaults map[string]interface{}) error {
	config, err := prepare()
	if err != nil {
		return err
	}

	err = config.ReadInConfig()
	if err != nil {
		return err
	}

	config.SetEnvPrefix("CRM")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	for key, value := range defaults {
		fmt.Printf("Setting default: %s = %s\n", key, value)
		config.SetDefault(key, value)
	}

	err = config.Unmarshal(&c)
	if err != nil {
		return err
	}

	return nil
}
