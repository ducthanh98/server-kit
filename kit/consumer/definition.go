package consumer

import (
	"encoding/json"
	"github.com/ducthanh98/server-kit/kit/consumer/entity"
	log "github.com/sirupsen/logrus"
)

const DefaultSleepTimeWhenIdle = 100

const (
	// ProcessOKCode marks message as done
	ProcessOKCode int64 = 200
	// ProcessFailRetryCode marks message as fail and retry
	ProcessFailRetryCode = 500
	// ProcessFailDropCode marks message as fail and drop
	ProcessFailDropCode = 400
	// ProcessFailReproduceCode --
	ProcessFailReproduceCode = 302
)

type ConsumerGroup struct {
	Name string `json:"name,omitempty"`
}

// Input --
type Input struct {
	Mode   string
	Config interface{}
}

// Output --
type Output struct {
	Mode   string
	Config interface{}
}

// GConsumerDef --
type GConsumerDef struct {
	Type    string
	Name    string
	Group   *ConsumerGroup
	Inputs  []*Input
	Outputs []*Output
}

func parseTaskDef(def string) *GConsumerDef {
	taskDef := &struct {
		Type    string         `json:"type,omitempty"`
		Name    string         `json:"name,omitempty"`
		Group   *ConsumerGroup `json:"group,omitempty"`
		Inputs  []interface{}  `json:"inputs,omitempty"`
		Outputs []interface{}  `json:"outputs,omitempty"`
	}{}

	if err := json.Unmarshal([]byte(def), &taskDef); err != nil {
		log.Fatal("Cannot parse task info from input", def)
	}

	result := GConsumerDef{Type: taskDef.Type, Name: taskDef.Name, Group: taskDef.Group}
	// input
	for _, inp := range taskDef.Inputs {
		inpM, ok := inp.(map[string]interface{})
		if !ok {
			log.Warn("Bad input. ", inp)
			continue
		}
		t, ok := inpM["mode"]
		if !ok {
			log.Warn("Bad input. Mode not found", inp)
			continue
		}
		var rit interface{}
		switch t {
		case "rabbitmq":
			rit = &entity.RmqInputConf{}
		default:
			log.Warn("Currently not supported. ", t)
			continue
		}

		if fillByTags(inp, rit) != nil {
			continue
		}

		result.Inputs = append(result.Inputs, &Input{Mode: t.(string), Config: rit})
	}

	// output
	for _, out := range taskDef.Outputs {
		outM, ok := out.(map[string]interface{})
		if !ok {
			log.Warn("Bad output. ", out)
			continue
		}
		t, ok := outM["mode"]
		if !ok {
			log.Warn("Bad input. Mode not found", out)
			continue
		}
		var rit interface{}
		switch t {
		case "rabbitmq":
			rit = &entity.RmqOutputConf{}
		default:
			log.Warn("Currently not supported. ", t)
			continue
		}

		if fillByTags(out, rit) != nil {
			continue
		}

		result.Outputs = append(result.Outputs, &Output{Mode: t.(string), Config: rit})
	}

	return &result
}

func fillByTags(i interface{}, o interface{}) error {
	b, _ := json.Marshal(i)
	err := json.Unmarshal(b, o)
	return err
}
