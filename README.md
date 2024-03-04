# Namadexer-go

Namadexer-go is a Golang implementation of indexer for [Namada](https://github.com/anoma/namada).
It supports all endpoints from original [namadexer](https://github.com/Zondax/namadexer) and has additional endpoints:
 - `/txs?hash=<hash-id-1>&hash=<hash-id-2>...` - fetch list of transactions by specified hashes
 - `/txs/memo/{memo}` - fetch list of transactions by specified memo with limit and offset
 - `/txs/memo/{memo}/total` - total number of transactions by specified memo
 - `/account/txs/{account_id}` - fetch list of transactions associated with specified account, limit and offset
 - `/account/txs/{account_id}/total` - total number of transactions associated with specified account

## Overview

The project is composed of 2 entities : the `indexer` and the `server`. They are both written in Golang.

- the `indexer`: This component establishes a connection to the Namada node via RPC and collects blocks and transactions. It then stores this data in a PostgreSQL database. The indexer operates independently of the server and can be initiated on its own.

- the `server`: This is a JSON-based server that facilitates querying of blocks and transactions using unique identifiers.

These services require a connection to a [postgres](https://www.postgresql.org/) database.

Overall, the structure is pretty similar to [namadexer](https://github.com/Zondax/namadexer).

## Development

You will need access to a namada node and specify its tendermint rpc host and port in the `config/config.toml` file. You can use the `config.example.toml` as a template.

### Dev dependencies

To proceed, you must have Docker installed and a Namada node.

### Run

Start all the components:
```
$ make compose
```

To start components separately you can use `make run-postgres`, `make run-indexer` and `make run-server` commands.
