package config

type Config struct {
	Agents []Agent `yaml:"agents"`
}

type Agent struct {
	Protocol       string   `yaml:"protocol"`
	Bind           string   `yaml:"bind"`
	TransactionTtl int      `yaml:"transaction_ttl"`
	Include        []string `yaml:"include"`
}
