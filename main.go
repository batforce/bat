package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"com.batforce.bat/internal"
	"github.com/streadway/amqp"
)

func rabbitMQConnectionString() string {

	host := os.Getenv("BAT_RABBIT_MQ_HOST")
	if len(host) == 0 {
		host = "localhost"
	}
	port := os.Getenv("BAT_RABBIT_MQ_PORT")
	if len(port) == 0 {
		port = "5672"
	}
	user := os.Getenv("BAT_RABBIT_MQ_USER")
	if len(user) == 0 {
		user = "guest"
	}
	password := os.Getenv("BAT_RABBIT_MQ_PASSWORD")
	if len(password) == 0 {
		password = "guest"
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
}

//go:embed script
var script string

func main() {

	connectionString := rabbitMQConnectionString()
	conn, err := amqp.Dial(connectionString)
	internal.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	internal.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"bat_worker_queue", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	internal.FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	//msgs, err := internal.ConsumeQueue()
	internal.FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			// dotCount := bytes.Count(d.Body, []byte("."))
			// t := time.Duration(dotCount)
			request := &internal.WorkRequest{}
			json.Unmarshal(d.Body, request)
			out, err := json.MarshalIndent(request, "", "  ")
			log.Printf("Received a message: \n%s\n", out)
			log.Printf("Downloading build kit")
			err = internal.CreateWorkspace()
			if err != nil {
				log.Fatalln(err)
			}

			err = RunCompile(request)

			if err != nil {
				internal.CleanWorkspace()
				d.Ack(false)
				log.Println(err)
				return
			}
			internal.CleanWorkspace()
			d.Ack(false)

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func RunCompile(request *internal.WorkRequest) error {
	buildkit, err := internal.UseBuildKit(request.Kit)
	if err != nil {
		return err
	}

	internal.CheckoutGit(script, *request)

	isSupported, framework := buildkit.Detect(*request)

	if !isSupported {
		return errors.New("Unsupported framework")
	}
	fmt.Printf("Supported framework: %s\n", framework)
	_, err = buildkit.Compile(*request)
	if err != nil {
		return err
	}
	_, err = buildkit.Release(*request)
	if err != nil {
		return err
	}

	// hehe, _ := json.Marshal(request)
	// log.Println(string(hehe))
	return nil
}

func RunDeploy(request *internal.WorkRequest) error {
	deploykit, err := internal.UseDeployKit(request.Kit)
	if err != nil {
		return err
	}

	internal.CheckoutGit(script, *request)

	isSupported, framework := deploykit.Detect(*request)

	if !isSupported {
		return errors.New("Unsupported framework")
	}
	fmt.Printf("Supported framework: %s\n", framework)
	_, err = deploykit.Deploy(*request)
	if err != nil {
		return err
	}
	// hehe, _ := json.Marshal(request)
	// log.Println(string(hehe))
	return nil
}
