package postgres

import "fmt"

func createBlocksTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		block_id BYTEA NOT NULL,
		header_version_app BIGINT NOT NULL,
		header_version_block BIGINT NOT NULL,
		header_chain_id TEXT NOT NULL,
		header_height BIGINT NOT NULL,
		header_time TIMESTAMP NOT NULL,
		header_last_block_id_hash BYTEA,
		header_last_block_id_parts_header_total BIGINT,
		header_last_block_id_parts_header_hash BYTEA,
		header_last_commit_hash BYTEA,
		header_data_hash BYTEA,
		header_validators_hash BYTEA NOT NULL,
		header_next_validators_hash BYTEA NOT NULL,
		header_consensus_hash BYTEA NOT NULL,
		header_app_hash BYTEA NOT NULL,
		header_last_results_hash BYTEA,
		header_evidence_hash BYTEA,
		header_proposer_address BYTEA NOT NULL,
		commit_height BIGINT,
		commit_round BIGINT,
		commit_block_id_hash BYTEA,
		commit_block_id_parts_header_total BIGINT,
		commit_block_id_parts_header_hash BYTEA
	);`, blocksTable)
}

func createTransactionsTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		hash BYTEA NOT NULL,
		block_id BYTEA NOT NULL,
		tx_type TEXT NOT NULL,
		wrapper_id BYTEA,
		memo TEXT,
		fee_amount_per_gas_unit TEXT,
		fee_token TEXT,
		gas_limit_multiplier BIGINT,
		code BYTEA,
		data JSONB,
		return_code BIGINT,
		pos_in_block BIGINT NOT NULL
	);`, transactionsTable)
}

func createEvidencesTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
        block_id BYTEA NOT NULL,
        height BIGINT NOT NULL,
        time BIGINT NOT NULL,
        address BYTEA,
        total_voting_power BIGINT NOT NULL,
        validator_power BIGINT NOT NULL
    );`, evidencesTable)
}

func createCommitSignaturesTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
        block_id BYTEA NOT NULL,
        block_id_flag INTEGER NOT NULL,
        validator_address BYTEA NOT NULL,
        timestamp BIGINT NOT NULL,
        signature BYTEA NOT NULL
    );`, commitSignaturesTable)
}

func createAccountTransactionsTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		address BYTEA NOT NULL,
        tx_hash BYTEA NOT NULL,
		block_height BIGINT NOT NULL,
		tx_pos BIGINT NOT NULL
	);`, accountTransactionsTable)
}
