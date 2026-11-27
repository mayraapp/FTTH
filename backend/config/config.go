package config

import "os"

// Config reúne las rutas y parámetros de ejecución de la aplicación.
type Config struct {
	Port            string
	RequestsFile    string
	ModuleDataFile  string
	PublishedWebDir string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return Config{
		Port:            port,
		RequestsFile:    "data/solicitudes.json",
		ModuleDataFile:  "data/modulos.json",
		PublishedWebDir: "internal/web/dist",
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisAddr:       os.Getenv("REDIS_ADDR"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
	}
}
