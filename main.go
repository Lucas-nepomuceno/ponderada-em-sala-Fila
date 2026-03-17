package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
  "fmt"
)

type telemetria struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	TipoSensor  string  `json:"tipo-sensor"`
	TipoLeitura string  `json:"tipo-leitura"`
	Valor       float64 `json:"valor"`
}

func addInQueue(c *gin.Context) {
	var dados telemetria

	if err := c.BindJSON(&dados); err != nil {
		return
	}
  message := fmt.Sprintf("Sua requisição %s será processada em breve", dados.ID)

	c.IndentedJSON(http.StatusCreated, message)
}

func main() {
	router := gin.Default()

	router.POST("/dados-sensores", addInQueue)

	if err := router.Run(); err != nil {
		log.Fatalf("Falha ao ler o arquivo: %v", err)
	}
}
