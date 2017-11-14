package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type User struct {
	Username  string
	Password  string
	Email     string
	Name      string
	Telephone string
	Country   string
	City      string
	Address   string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func saveToDB(user *User) int {
	db := getConnection()

	sqlStatement := `
	INSERT INTO users (user_name, email,password, name, telephone, country,city,address)
	VALUES ($1, $2, $3, $4 , $5, $6, $7, $8)
	RETURNING user_id`

	id := 0
	err := db.QueryRow(sqlStatement, user.Username, user.Email, user.Password, user.Name, user.Telephone, user.Country, user.City, user.Address).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}
func validateData(name string, email string) (message string) {
	db := getConnection()

	sqlStatement := ` SELECT COUNT(*) FROM users WHERE TRIM(user_name) = TRIM($1) `
	count := 0
	err := db.QueryRow(sqlStatement, name).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		message = "username already exist"
	}
	//println(email)
	sqlStatement = `
	SELECT COUNT(*) FROM users WHERE email = $1 `
	count = 0
	err = db.QueryRow(sqlStatement, email).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		message = message + "\n email already exist"
	}
	return

}
func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			obj := []byte(d.Body)

			var user User
			json.Unmarshal(obj, &user)

			//println(string(user.Email))
			response := validateData(user.Username, user.Email)
			println(response)
			if response == "" {
				id := saveToDB(&user)
				response = fmt.Sprintf("Created with ID %d", id)
			}
			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			failOnError(err, "Failed to publish a message")

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}
