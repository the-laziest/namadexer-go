package repository

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string

	Schema            string
	CreateIndex       bool
	ConnectionTimeout int
}
