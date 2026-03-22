package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	dbConfig "middleware/db_config"
)

var db *sql.DB
var cfg = dbConfig.LoadConfig()

func checkErr(err error) {
	if err != nil {
		log.Println(err.Error())
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

type telemetria struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	TipoSensor  string  `json:"tipo-sensor"`
	TipoLeitura string  `json:"tipo-leitura"`
	Valor       float64 `json:"valor"`
}

func sqlInsert(data telemetria) {

	sqlStatement := fmt.Sprintf(`INSERT INTO %s (id_sensor, hora_leitura, tipo_sensor, tipo_leitura, valor) VALUES ($1, $2, $3, $4, $5)`, cfg.TableName)
	insert, err := db.Prepare(sqlStatement)
	checkErr(err)

	_, err = insert.Exec(data.ID, data.Timestamp, data.TipoLeitura, data.TipoSensor, data.Valor)
	checkErr(err)

	fmt.Println("adicionado!")
}

func main() {
	var dados telemetria
	var err error

	db, err = sql.Open(cfg.PostgresDriver, cfg.DataSourceName)

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Connected!")
	}

	defer db.Close()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			err = json.Unmarshal(d.Body, &dados)
			sqlInsert(dados)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
