package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Failed to read config file:", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Failed to parse config:", err)
	}

	if config.Email == "" || config.Password == "" {
		log.Fatal("Email and password are required")
	}

	fmt.Println("Email:", config.Email)
	fmt.Println("Password:", config.Password)
}
