package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"code.cloudfoundry.org/test-log-emitter/client"
	"code.cloudfoundry.org/test-log-emitter/config"
	"code.cloudfoundry.org/test-log-emitter/emitters"
	flag "github.com/spf13/pflag"
)

func main() {
	var configFilePath *string = flag.String("config", "", "path to config file")
	flag.Parse()

	if *configFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	confContents, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	conf := new(config.Config)
	err = yaml.Unmarshal(confContents, conf)
	if err != nil {
		log.Fatal(err)
	}

	loggregatorClient, err := client.NewLoggregatorIngressClient(conf.Loggregator)
	if err != nil {
		log.Fatal(err)
	}
	gaugeEmitter := emitters.NewGaugeEmitter(loggregatorClient)
	counterEmitter := emitters.NewCounterEmitter(loggregatorClient)
	timerEmitter := emitters.NewTimerEmitter(loggregatorClient)

	http.HandleFunc("/", ping)
	http.Handle("/gauge", gaugeEmitter.EmitGauge())
	http.Handle("/timer", timerEmitter.EmitTimer())
	http.Handle("/counter", counterEmitter.EmitCounter())

	fmt.Printf("Starting cpu usage logger on port %d...", conf.ListenPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.ListenPort), nil); err != nil {
		log.Fatal(err)
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	message := "What do you want to emit today?\n"
	message = message + "* POST /gauge - posts a gauge metric\n"
	message = message + "* POST /counter - posts an counter metric\n"

	if _, err := io.WriteString(w, message); err != nil {
		http.Error(w, fmt.Sprintf("Failed to resond to ping request: %v", err), http.StatusInternalServerError)
		return
	}
}
