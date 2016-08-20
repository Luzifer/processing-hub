package main

import "log"

func getConfigStore() configStore {
	switch cfg.ConfigBackend {
	case "env":
		return &configStoreEnv{}
	default:
		log.Fatalf("Config backend '%s' is not defined.", cfg.ConfigBackend)
	}

	return nil
}
