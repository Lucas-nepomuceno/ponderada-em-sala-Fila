package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	dbConfig "middleware/db_config"
	"time"
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

func isValidTimestamp(value string) bool {
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

type telemetria struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	TipoSensor  string  `json:"tipo-sensor"`
	TipoLeitura string  `json:"tipo-leitura"`
	Valor       float64 `json:"valor"`
}

func verificaCampos(dados telemetria) (telemetria, error) {
	if dados.ID == "" ||
		dados.Timestamp == "" ||
		dados.TipoSensor == "" ||
		dados.TipoLeitura == "" ||
		dados.Valor == 0 {
		return dados, errors.New("todos os campos devem ser preenchidos")
	}

	if dados.TipoLeitura != "analogico" && dados.TipoLeitura != "discreto" {
		return dados, errors.New("o tipo de sensor deve ser analogico ou discreto")
	}

	if !isValidTimestamp(dados.Timestamp) {
		return dados, errors.New("A data enviada não está formatada corretamente")
	}

	return dados, nil
}

func sqlInsert(data telemetria) {

	sqlStatement := fmt.Sprintf(`INSERT INTO %s (id_sensor, hora_leitura, tipo_sensor, tipo_leitura, valor) VALUES ($1, $2, $3, $4, $5)`, cfg.TableName)
	insert, err := db.Prepare(sqlStatement)
	checkErr(err)

	dados_verificados, err := verificaCampos(data)
	checkErr(err)

	if err == nil {
		_, err = insert.Exec(dados_verificados.ID, dados_verificados.Timestamp, dados_verificados.TipoLeitura, dados_verificados.TipoSensor, dados_verificados.Valor)
		checkErr(err)
		fmt.Println("Adicionado!")
	}

}

func main() {
	var dados telemetria
	var err error

	db, err = sql.Open(cfg.PostgresDriver, cfg.DataSourceName)

	if err != nil {
		panic(err.Error())
	} else {
		fmt.Println("Conectado!")
	}

	defer db.Close()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Falha em conectar com o  RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Falha em abrir o canal")
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
	failOnError(err, "Falha em declarar a fila")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Falha em registrar o consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			err = json.Unmarshal(d.Body, &dados)
			sqlInsert(dados)
		}
	}()

	log.Printf(" [*] Aguardando mensagens.")
	<-forever
}
