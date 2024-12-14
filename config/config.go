package config

import "log"

func Init() {
	err := initEnvironment()
	if err != nil {
		log.Fatalf("error initializing environment: %s", err.Error())
	}
}
