package config

import (
	"fmt"

	"github.com/josnelihurt/mailer-go/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	// Email configuration
	Email          string   `mapstructure:"email"`
	Password       string   `mapstructure:"password"`
	RecipientEmail []string `mapstructure:"recipient_email"`

	// Redis configuration
	RedisHost    string `mapstructure:"redis_host"`
	RedisPort    string `mapstructure:"redis_port"`
	RedisEnabled bool   `mapstructure:"redis_enabled"`

	// IMEI to phone number mapping
	ImeiToPhone map[string]string `mapstructure:"imei_to_phone"`

	// Server communication
	ServerURL string `mapstructure:"server_url"`
	APIKey    string `mapstructure:"api_key"`

	// GSM modem configuration
	ModemDevice      string `mapstructure:"modem_device"`       // e.g. /dev/ttyUSB0
	ModemBaud        int    `mapstructure:"modem_baud"`         // e.g. 115200
	ModemInitTimeout int    `mapstructure:"modem_init_timeout"` // in seconds, default 30
	DeleteAfterRead  bool   `mapstructure:"delete_after_read"`  // true to delete after processing
}

func (c Config) String() string {
	return fmt.Sprintf("{email:%s, password:{hidden}, recipient_email:%s, redis_enabled:%t, redis_host:%s, redis_port:%s, imei_to_phone_count:%d, modem_device:%s, modem_baud:%d}",
		c.Email, c.RecipientEmail, c.RedisEnabled, c.RedisHost, c.RedisPort, len(c.ImeiToPhone), c.ModemDevice, c.ModemBaud)
}

func Read() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("config")
	viper.AutomaticEnv()

	// Set defaults for modem configuration
	viper.SetDefault("modem_device", "/dev/ttyUSB0")
	viper.SetDefault("modem_baud", 115200)
	viper.SetDefault("modem_init_timeout", 30)
	viper.SetDefault("delete_after_read", true)

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("unable to read config file:%w", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("unable to parse config: %s :%w", config.String(), err)
	}

	// Validate required fields for AT mode
	if config.Email == "" || config.Password == "" {
		return Config{}, fmt.Errorf("invalid configuration: email and password are required :%w", errors.ErrApp)
	}

	if config.ModemDevice == "" {
		return Config{}, fmt.Errorf("invalid configuration: modem_device is required :%w", errors.ErrApp)
	}

	return config, nil
}
