package dbConfig

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Telemetria struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	TipoSensor  string  `json:"tipo-sensor"`
	TipoLeitura string  `json:"tipo-leitura"`
	Valor       float64 `json:"valor"`
}

type Config struct {
	PostgresDriver string
	User           string
	Host           string
	Port           string
	Password       string
	DbName         string
	TableName      string
	DataSourceName string
}

func LoadConfig() *Config {
	// carrega .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: .env não encontrado, usando variáveis do sistema")
	}

	cfg := &Config{
		PostgresDriver: os.Getenv("POSTGRESDRIVER"),
		User:           os.Getenv("DBUSER"),
		Host:           os.Getenv("DBHOST"),
		Port:           os.Getenv("DBPORT"),
		Password:       os.Getenv("DBPASSWORD"),
		DbName:         os.Getenv("DBNAME"),
		TableName:      os.Getenv("DBTABLENAME"),
	}

	cfg.DataSourceName = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DbName,
	)

	return cfg
}