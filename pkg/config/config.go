package config

import (
	"fmt"

	"github.com/josnelihurt/mailer-go/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Email          string   `mapstructure:"email"`
	Password       string   `mapstructure:"password"`
	RecipientEmail []string `mapstructure:"recipient_email"`
	Inbox          string   `mapstructure:"inbox_folder"`
	ErrBox         string   `mapstructure:"err_folder"`
	DoneBox        string   `mapstructure:"done_folder"`
}

func (c Config) String() string {
	return fmt.Sprintf("{email:%s, password:{hidden}, recipient_email:%s, inbox_folder:%s, err_folder:%s, done_folder:%s}",
		c.Email, c.RecipientEmail, c.Inbox, c.ErrBox, c.DoneBox)
}

func Read() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("unable to read config file:%w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %s :%w", config.String(), err)
	}

	if config.Email == "" || config.Password == "" || config.Inbox == "" || config.ErrBox == "" || config.DoneBox == "" {
		return Config{}, fmt.Errorf("invalid configuration: %s :%w", config.String(), errors.ErrApp)
	}
	return config, nil
}
