package config

type Config struct {
	ChainName  string           `toml:"chain_name"`
	Database   DatabaseConfig   `toml:"database"`
	Server     ServerConfig     `toml:"server"`
	Indexer    IndexerConfig    `toml:"indexer"`
	Prometheus PrometheusConfig `toml:"prometheus"`
}

type DatabaseConfig struct {
	Host              string `toml:"host"`
	Port              string `toml:"port"`
	User              string `toml:"user"`
	Password          string `toml:"password"`
	DbName            string `toml:"db_name"`
	CreateIndex       bool   `toml:"create_index"`
	ConnectionTimeout int    `toml:"connection_timeout"`
}

type ServerConfig struct {
	Port string `toml:"port"`
}

type IndexerConfig struct {
	RPC                string `toml:"rpc"`
	WaitForBlock       int64  `toml:"wait_for_block"`
	MaxBlocksInChannel int64  `toml:"max_blocks_in_channel"`
}

type PrometheusConfig struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}
