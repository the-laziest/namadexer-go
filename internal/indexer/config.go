package indexer

type Config struct {
	RpcURL    string
	Checksums map[string]string

	WaitForBlock       int64
	MaxBlocksInChannel int64
}
