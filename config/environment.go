package config

import (
	"errors"
	"os"
	"reflect"
)

type Environment struct {
	DISCORD_BOT_TOKEN     string
	TOURNAMENT_CHANNEL_ID string
	DB_HOST               string
	DB_NAME               string
	DB_PORT               string
	DB_USER               string
	DB_PASSWORD           string
}

var environment Environment

func initEnvironment() error {
	envType := reflect.TypeOf(environment)
	envValue := reflect.ValueOf(&environment).Elem()

	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		envVar := field.Name

		value := os.Getenv(envVar)
		if value == "" {
			return errors.New("environment variable " + envVar + " is required")
		}

		envValue.FieldByName(envVar).SetString(value)
	}

	return nil
}

func GetEnv() *Environment {
	return &environment
}
