package mq

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/nsqio/go-nsq"
	log "github.com/sirupsen/logrus"

	"zuccacm-server/config"
)

var Instance *nsq.Producer

func init() {
	c := nsq.NewConfig()
	var err error
	Instance, err = nsq.NewProducer(config.Instance.MessageQueue, c)
	if err != nil {
		log.Fatal(err)
	}
	Instance.SetLoggerLevel(nsq.LogLevelWarning)
}

func Topic(ojId int) string {
	return fmt.Sprintf("zuccacm-%02d", ojId)
}

type Task gabs.Container

func newTask() *Task {
	return (*Task)(gabs.New())
}

func (t *Task) mustSet(v interface{}, path string) {
	_, err := (*gabs.Container)(t).SetP(v, path)
	if err != nil {
		panic(err)
	}
}

func (t *Task) String() string {
	return (*gabs.Container)(t).String()
}

func ExecTask(topic string, task *Task) {
	err := Instance.Publish(topic, []byte(task.String()))
	if err != nil {
		panic(err)
	}
	log.WithFields(log.Fields{
		"topic": topic,
		"task":  task.String(),
	}).Info("Task has been created")
}
