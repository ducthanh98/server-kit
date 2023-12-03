package kafka_driver

type KafkaConfig struct {
	Brokers  string `json:"brokers"`
	Group    string `json:"group"`
	Version  string `json:"version"`
	Topics   string `json:"topics"`
	Assignor string `json:"assignor"`
	Oldest   bool   `json:"oldest"`
	Verbose  bool   `json:"verbose"`
}
