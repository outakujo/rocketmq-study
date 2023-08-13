package main

import (
	"context"
	"fmt"
	client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Endpoint                = "localhost:9876"
	Topic                   = "ert"
	Body                    = "content"
	AccessKey               = ""
	AccessSecret            = ""
	awaitDuration           = time.Second * 5
	maxMessageNum     int32 = 10
	invisibleDuration       = time.Second * 20
	group                   = "aa"
)

var ems = []string{"👨", "😂", "🤣", "🤨️", "😎", "😍"}

var wc = make(chan int)

func main() {
	_ = os.Setenv(client.ENABLE_CONSOLE_APPENDER, "true")
	_ = os.Setenv(client.CLIENT_LOG_LEVEL, "error")
	client.ResetLogger()
	//asyns()
	go consu("", group)
	wait()
}

func wait() {
	for i := 0; i < 5; i++ {
		j := <-wc
		if j == 0 {
			i = 0
			fmt.Println("reset")
		}
	}
	fmt.Println("over")
}

func asyns() {
	go consu("1hao", group)
	go consu("2hao", group)
	proc("")
}

func proc(grp string) {
	producer, err := client.NewProducer(&client.Config{
		Endpoint: Endpoint,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: AccessSecret,
		},
	},
		client.WithTopics(Topic),
	)
	if err != nil {
		panic(err)
	}
	err = producer.Start()
	if err != nil {
		panic(err)
	}
	msg := client.Message{
		Topic: Topic,
		Body:  []byte(Body),
	}
	if grp != "" {
		msg.SetMessageGroup(grp)
	}
	defer producer.GracefulStop()
	for i := 0; ; i++ {
		m := msg
		itoa := strconv.Itoa(i)
		m.Body = append(m.Body, []byte(itoa)...)
		m.AddProperty("ha", "是范"+ems[rand.Intn(len(ems))])
		if grp == "" {
			intn := rand.Intn(100)
			dely := time.Now().Add(time.Duration(intn) * time.Second)
			m.SetDelayTimestamp(dely)
			fmt.Printf("content:%s dely:%s\n", string(m.Body), dely)
		}
		resp, err := producer.Send(context.TODO(), &m)
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(resp); i++ {
			//fmt.Printf("send offset:%d\n", resp[i].Offset)
		}
		time.Sleep(2 * time.Second)
	}
}

func consu(num, grp string) {
	simpleConsumer, err := client.NewSimpleConsumer(&client.Config{
		Endpoint:      Endpoint,
		ConsumerGroup: grp,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: AccessSecret,
		},
	},
		client.WithAwaitDuration(awaitDuration),
		client.WithSubscriptionExpressions(map[string]*client.FilterExpression{
			Topic: client.SUB_ALL,
		}),
	)
	if err != nil {
		panic(err)
	}
	err = simpleConsumer.Start()
	if err != nil {
		panic(err)
	}
	defer simpleConsumer.GracefulStop()

	for {
		mvs, err := simpleConsumer.Receive(context.TODO(), maxMessageNum, invisibleDuration)
		if err != nil {
			if strings.Contains(err.Error(), "MESSAGE_NOT_FOUND") {
				wc <- 1
			} else {
				fmt.Println("err:", err)
			}
			continue
		}
		wc <- 0
		for _, mv := range mvs {
			simpleConsumer.Ack(context.TODO(), mv)
			fmt.Println(mv.GetBornTimestamp())
			fmt.Println(num, mv.GetOffset(), string(mv.GetBody()),
				mv.GetProperties()["ha"], mv.GetDeliveryTimestamp())
		}
		time.Sleep(1 * time.Second)
	}
}
