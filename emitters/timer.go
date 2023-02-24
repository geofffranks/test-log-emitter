package emitters

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/go-loggregator/v9"
)

type TimerValue struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type TimerMetric struct {
	SourceId   string
	InstanceId string
	Tags       map[string]string
	Value      TimerValue
}

type TimerEmitter struct {
	client *loggregator.IngressClient
}

func NewTimerEmitter(client *loggregator.IngressClient) *TimerEmitter {
	return &TimerEmitter{client: client}
}

func (e TimerEmitter) SendTimer(timer TimerMetric) {
	opts := []loggregator.EmitTimerOption{
		loggregator.WithTimerSourceInfo(timer.SourceId, timer.InstanceId),
		loggregator.WithEnvelopeTags(timer.Tags),
	}
	e.client.EmitTimer(timer.Value.Name, timer.Value.Start, timer.Value.End, opts...)
}

func (e TimerEmitter) EmitTimer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Sorry, only POST methods are supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read body: %v", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		timerMetric := TimerMetric{}
		if err := json.Unmarshal(body, &timerMetric); err != nil {
			http.Error(w, fmt.Sprintf("Failed to unmarshal body: %v", err), http.StatusInternalServerError)
			return
		}

		e.SendTimer(timerMetric)
	}
}
