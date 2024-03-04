package repository

import "time"

type Block struct {
	BlockID                           []byte
	HeaderVersionApp                  uint64
	HeaderVersionBlock                uint64
	HeaderChainID                     string
	HeaderHeight                      int64
	HeaderTime                        time.Time
	HeaderLastBlockIDHash             []byte
	HeaderLastBlockIDPartsHeaderTotal uint32
	HeaderLastBlockIDPartsHeaderHash  []byte
	HeaderLastCommitHash              []byte
	HeaderDataHash                    []byte
	HeaderValidatorsHash              []byte
	HeaderNextValidatorsHash          []byte
	HeaderConsensusHash               []byte
	HeaderAppHash                     []byte
	HeaderLastResultsHash             []byte
	HeaderEvidenceHash                []byte
	HeaderProposerAddress             []byte
	CommitHeight                      int64
	CommitRound                       int32
	CommitBlockIDHash                 []byte
	CommitBlockIDPartsHeaderTotal     uint32
	CommitBlockIDPartsHeaderHash      []byte
}

type BlockFilter struct {
	Height  int64
	BlockID []byte
}

type Transaction struct {
	Hash                []byte
	BlockID             []byte
	TxType              string
	WrapperID           []byte
	Memo                string
	FeeAmountPerGasUnit string
	FeeToken            string
	GasLimitMultiplier  *uint64
	Code                []byte
	Data                []byte
	ReturnCode          *int64
	PosInBlock          int64
}

type TxFilter struct {
	Hashes  [][]byte
	BlockID []byte
	Memo    string
	Offset  uint64
	Limit   uint64
}

type AccountTransaction struct {
	Address     string
	TxHash      []byte
	BlockHeight int64
	TxPos       int64
}

type Evidence struct {
	BlockID          []byte
	Height           int64
	Time             int64
	Address          []byte
	TotalVotingPower int64
	ValidatorPower   int64
}

type CommitSignature struct {
	BlockID          []byte
	BlockIDFlag      int
	ValidatorAddress []byte
	Timestamp        int64
	Signature        []byte
}
