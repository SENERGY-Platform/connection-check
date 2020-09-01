package docker

import (
	"context"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/ory/dockertest/v3"
	uuid "github.com/satori/go.uuid"
	"log"
	"os"
	"sync"
)

func VernemqWithManagementApi(pool *dockertest.Pool, ctx context.Context, wg *sync.WaitGroup) (brokerUrl string, managementUrl string, err error) {
	log.Println("start mqtt")
	container, err := pool.Run("erlio/docker-vernemq", "1.9.1-alpine", []string{
		"DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on",
		"DOCKER_VERNEMQ_LOG__CONSOLE__LEVEL=debug",
		"DOCKER_VERNEMQ_SHARED_SUBSCRIPTION_POLICY=random",
	})
	if err != nil {
		return "", "", err
	}
	go Dockerlog(pool, ctx, container, "VERNEMQ")
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	err = pool.Retry(func() error {
		log.Println("DEBUG: try to connection to broker")
		options := paho.NewClientOptions().
			SetAutoReconnect(true).
			SetCleanSession(false).
			SetClientID(uuid.NewV4().String()).
			AddBroker("tcp://" + container.Container.NetworkSettings.IPAddress + ":1883")

		client := paho.NewClient(options)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Println("Error on Mqtt.Connect(): ", token.Error())
			return token.Error()
		}
		defer client.Disconnect(0)
		return nil
	})
	execLog := &LogWriter{logger: log.New(os.Stdout, "[VERNEMQ-EXEC]", 0)}
	code, err := container.Exec([]string{"vmq-admin", "api-key", "add", "key=testkey"}, dockertest.ExecOptions{StdErr: execLog, StdOut: execLog})
	if err != nil {
		log.Panic("ERROR", code, err)
		return "", "", err
	}
	return "tcp://" + container.Container.NetworkSettings.IPAddress + ":1883", "http://testkey@" + container.Container.NetworkSettings.IPAddress + ":8888", err
}
