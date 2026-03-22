package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/text/unicode/norm"
	"log"
	"net/http"
	"time"
	"unicode"
)

var (
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
)

type telemetria struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	TipoSensor  string  `json:"tipo-sensor"`
	TipoLeitura string  `json:"tipo-leitura"`
	Valor       float64 `json:"valor"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func removerAcentos(s string) string {
	t := norm.NFD.String(s) // decompõe caracteres (ex: é → e + ´)

	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue // remove acentos (marks)
		}
		result = append(result, r)
	}
	return string(result)
}

func addInQueue(c *gin.Context) {
	var dados telemetria
	var err error

	if err := c.BindJSON(&dados); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dados.TipoSensor = removerAcentos(dados.TipoSensor)

	body, err := json.Marshal(dados)

	if err != nil {
		log.Println("Erro ao marshalling:", err)
		c.JSON(http.StatusInternalServerError, "erro ao enviar mensagem")
		return
	}

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})

	if err != nil {
		log.Println("Erro ao publicar:", err)
		c.JSON(http.StatusInternalServerError, "erro ao enviar mensagem")
		return
	}

	message := fmt.Sprintf("Sua requisição %s será processada em breve", dados.ID)

	c.IndentedJSON(http.StatusCreated, message)
}

func main() {
	router := gin.Default()
	var err error

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err = ch.QueueDeclare(
		"dados-sensores", // name
		true,             // durability
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	failOnError(err, "Failed to declare a queue")

	router.POST("/dados-sensores", addInQueue)

	if err := router.Run(); err != nil {
		log.Fatalf("Falha ao ler o arquivo: %v", err)
	}
}
