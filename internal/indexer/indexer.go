package indexer

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/bytes"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/coretypes"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/internal/types"
	ptypes "github.com/the-laziest/namadexer-go/internal/types/proto"
	"github.com/the-laziest/namadexer-go/pkg/borsh"
	"github.com/the-laziest/namadexer-go/pkg/errors"
	"github.com/the-laziest/namadexer-go/pkg/logger"
)

type blockInfo struct {
	resultBlock        *coretypes.ResultBlock
	resultBlockResults *coretypes.ResultBlockResults
}

type processedBlock struct {
	height int64
	txs    []repository.Transaction
}

type Indexer struct {
	config Config

	client *rpchttp.HTTP

	blockChan chan blockInfo
	lastBlock processedBlock

	repository repository.Repository
}

func New(config Config, repository repository.Repository) (*Indexer, error) {
	client, err := rpchttp.New(config.RpcURL)
	if err != nil {
		return nil, errors.New(err, "Init client")
	}
	return &Indexer{
		config:     config,
		client:     client,
		blockChan:  make(chan blockInfo, config.MaxBlocksInChannel),
		repository: repository,
	}, nil
}

func (i *Indexer) Close() {
	close(i.blockChan)
}

var ErrBlockNotFound = errors.Create("Block not found")

func (i *Indexer) Start(ctx context.Context) error {
	lastSavedHeight, err := i.repository.GetLastHeight(ctx)
	if err != nil {
		return errors.New(err, "Get last height")
	}

	go i.blockFetcher(ctx, lastSavedHeight+1)

	err = i.startBlockProcessor(ctx)
	if err != nil {
		return errors.New(err, "Block processor")
	}

	return nil
}

func (i *Indexer) blockFetcher(ctx context.Context, startHeight int64) {
	for height := startHeight; ; height++ {
		select {
		case <-ctx.Done():
			logger.Error("Stopping block fecher", zap.Error(ctx.Err()))
			return
		default:
			blockInfo, err := i.getBlock(ctx, height)
			if err != nil {
				logger.Error("Get block info failed", zap.Error(err))
				height--
				time.Sleep(time.Second * time.Duration(i.config.WaitForBlock))
				continue
			}
			i.blockChan <- blockInfo
		}
	}
}

func (i *Indexer) startBlockProcessor(ctx context.Context) error {
	latestBlock, err := i.client.Block(ctx, nil)
	if err != nil {
		return errors.New(err, "Get latest block")
	}

	latestHeight := latestBlock.Block.Height
	logger.Info("Latest block height on start: " + strconv.FormatInt(latestHeight, 10))

	hasIndexes, err := i.repository.HasIndexes(ctx)
	if err != nil {
		return err
	}

	for blockInfo := range i.blockChan {

		logger.Info("Processing block", zap.Int64("height", blockInfo.resultBlock.Block.Height))

		err := i.processBlock(ctx, blockInfo.resultBlock, blockInfo.resultBlockResults)
		if err != nil {
			return errors.New(err, "Process block failed")
		}

		logger.Info("Block saved", zap.Int64("height", blockInfo.resultBlock.Block.Height))

		if blockInfo.resultBlock.Block.Height == latestHeight {
			logger.Info("Indexer synced")
			if !hasIndexes {
				err = i.repository.CreateIndexes(ctx)
				if err != nil {
					return err
				}
				logger.Info("Database indexes created")
			}
		}
	}
	return nil
}

func (i *Indexer) checkNotFoundError(err error) error {
	var rpcErr *rpctypes.RPCError
	if errors.As(err, &rpcErr) {
		if rpcErr.Code == -32603 {
			return ErrBlockNotFound
		}
	}
	return err
}

func (i *Indexer) getBlock(ctx context.Context, height int64) (blockInfo, error) {

	logger.Info("Requesting block", zap.Int64("height", height))

	resultBlock, err := i.client.Block(ctx, &height)
	if err != nil {
		return blockInfo{}, i.checkNotFoundError(errors.New(err, "Get block"))
	}
	resultBlockResults, err := i.client.BlockResults(ctx, &height)
	if err != nil {
		return blockInfo{}, i.checkNotFoundError(errors.New(err, "Get block results"))
	}

	logger.Info("Block info received", zap.Int64("height", height))

	return blockInfo{resultBlock, resultBlockResults}, nil
}

func (i *Indexer) processBlock(ctx context.Context, resultBlock *coretypes.ResultBlock, resultBlockResults *coretypes.ResultBlockResults) error {

	height := resultBlock.Block.Height

	block := resultBlock.Block
	blockID := resultBlock.BlockID.Hash

	rBlock := repository.Block{
		BlockID:                           blockID,
		HeaderVersionApp:                  block.Header.Version.App,
		HeaderVersionBlock:                block.Header.Version.Block,
		HeaderChainID:                     block.Header.ChainID,
		HeaderHeight:                      height,
		HeaderTime:                        block.Header.Time,
		HeaderLastBlockIDHash:             block.Header.LastBlockID.Hash,
		HeaderLastBlockIDPartsHeaderTotal: block.Header.LastBlockID.PartSetHeader.Total,
		HeaderLastBlockIDPartsHeaderHash:  block.Header.LastBlockID.PartSetHeader.Hash,
		HeaderLastCommitHash:              block.Header.LastCommitHash,
		HeaderDataHash:                    block.Header.DataHash,
		HeaderValidatorsHash:              block.Header.ValidatorsHash,
		HeaderNextValidatorsHash:          block.Header.NextValidatorsHash,
		HeaderConsensusHash:               block.Header.ConsensusHash,
		HeaderAppHash:                     block.Header.AppHash,
		HeaderLastResultsHash:             block.Header.LastResultsHash,
		HeaderEvidenceHash:                block.Header.EvidenceHash,
		HeaderProposerAddress:             block.Header.ProposerAddress,
		CommitHeight:                      block.LastCommit.Height,
		CommitRound:                       block.LastCommit.Round,
		CommitBlockIDHash:                 block.LastCommit.BlockID.Hash,
		CommitBlockIDPartsHeaderTotal:     block.LastCommit.BlockID.PartSetHeader.Total,
		CommitBlockIDPartsHeaderHash:      block.LastCommit.BlockID.PartSetHeader.Hash,
	}

	commitSignatures := i.getCommitSignatures(blockID, block.LastCommit.Signatures)
	evidences := i.getEvidences(blockID, block.Evidence.Evidence)

	txs := make([]repository.Transaction, 0, len(block.Data.Txs))
	accTxs := make([]repository.AccountTransaction, 0)
	decryptedID := 0

	for id, tx := range block.Data.Txs {

		logger.Info("Processing tx", zap.Int64("height", height), zap.Int("tx_id", id))

		tx, accTx, err := i.processTx(blockID, height, int64(id), &decryptedID, tx, resultBlockResults)
		if err != nil {
			logger.Error("Process tx failed", zap.Int64("height", height), zap.Int("tx_id", id), zap.Error(err))
			return errors.New(err, "Process tx failed")
		}

		txs = append(txs, tx)
		if accTx != nil {
			accTxs = append(accTxs, *accTx)
		}
	}

	err := i.repository.RunInTransaction(ctx, func(txCtx context.Context, repo repository.Repository) error {
		err := repo.AddBlock(txCtx, rBlock)
		if err != nil {
			return err
		}

		err = repo.AddCommitSignatures(txCtx, commitSignatures...)
		if err != nil {
			return err
		}

		err = repo.AddEvidences(txCtx, evidences...)
		if err != nil {
			return err
		}

		err = repo.AddTransactions(txCtx, txs...)
		if err != nil {
			return err
		}

		return repo.AddAccountTransactions(txCtx, accTxs...)
	})
	if err != nil {
		return errors.New(err, "Save block info")
	}

	i.lastBlock = processedBlock{
		height: height,
		txs:    txs,
	}

	return nil
}

func (i *Indexer) getCommitSignatures(blockID bytes.HexBytes, signatures []tmtypes.CommitSig) []repository.CommitSignature {
	commitSignatures := make([]repository.CommitSignature, 0, len(signatures))
	for _, signature := range signatures {
		commitSignatures = append(commitSignatures, repository.CommitSignature{
			BlockID:          blockID,
			BlockIDFlag:      int(signature.BlockIDFlag),
			ValidatorAddress: signature.ValidatorAddress,
			Timestamp:        signature.Timestamp.Unix(),
			Signature:        signature.Signature,
		})
	}
	return commitSignatures
}

func (i *Indexer) getEvidences(blockID bytes.HexBytes, blockEvidences []tmtypes.Evidence) []repository.Evidence {
	evidences := make([]repository.Evidence, 0, len(blockEvidences))
	for _, evidence := range blockEvidences {
		dve, ok := evidence.(*tmtypes.DuplicateVoteEvidence)
		if !ok {
			continue
		}
		evidences = append(evidences, repository.Evidence{
			BlockID:          blockID,
			Height:           dve.VoteA.Height,
			Time:             dve.VoteA.Timestamp.Unix(),
			Address:          dve.VoteA.ValidatorAddress,
			TotalVotingPower: dve.TotalVotingPower,
			ValidatorPower:   dve.ValidatorPower,
		})
	}
	return evidences
}

func (i *Indexer) processTx(blockID bytes.HexBytes, height, txID int64, decryptedID *int, txRawData tmtypes.Tx, resultBlockResults *coretypes.ResultBlockResults) (repository.Transaction, *repository.AccountTransaction, error) {
	tx, err := i.decodeTxRawData(txRawData)
	if err != nil {
		return repository.Transaction{}, nil, errors.New(err, "Decode tx raw data")
	}

	tx.BlockHeight = height
	tx.TxPos = txID

	txHash, err := tx.GetHash()
	if err != nil {
		return repository.Transaction{}, nil, errors.New(err, "Get tx hash")
	}
	tx.TxHash = txHash

	logger.Info("Decoded tx", zap.Int64("height", height), zap.Int64("tx_id", txID), zap.String("tx_hash", txHash.String()), zap.String("tx_type", tx.Type()))

	var (
		returnCode                    *int64
		wrapper                       []byte
		code                          []byte
		feeAmountPerGasUnit, feeToken string
		gasLimitMultiplier            *uint64
		accountTx                     *repository.AccountTransaction
	)
	data := []byte("null")

	if tx.Header.TxType.IsDecrypted() {

		if i.lastBlock.height == height-1 && *decryptedID < len(i.lastBlock.txs) {
			wrapper = i.lastBlock.txs[*decryptedID].Hash
		}
		*decryptedID++

		codeHash, err := tx.GetCodeHash()
		if err != nil {
			return repository.Transaction{}, nil, errors.New(err, "Get code hash")
		}
		if !codeHash.IsEmpty() {
			code = codeHash[:]
		}

		txType, ok := i.config.Checksums[codeHash.String()]
		if !ok {
			txType = "undefined"
		}
		tx.DecryptedTxType = txType

		returnCodeFound := i.findTxReturnCode(txHash, resultBlockResults)
		returnCode = &returnCodeFound

		logger.Info("Decrypted tx", zap.Int64("height", height), zap.Int64("tx_id", txID), zap.String("decrypted_tx_type", txType), zap.Int64p("return_code", returnCode))

		if returnCodeFound == 0 {
			data, accountTx, err = i.processSuccessTx(tx)
			if err != nil {
				return repository.Transaction{}, nil, errors.New(err, "Process success tx")
			}
		}
	} else if tx.Header.TxType.IsWrapper() {
		feeAmountPerGasUnit = tx.Header.TxType.Wrapper.Fee.AmountPerGasUnit.String()
		feeToken = tx.Header.TxType.Wrapper.Fee.Token.String()
		gasLimitMultiplier = &tx.Header.TxType.Wrapper.GasLimit
	}

	memo, err := tx.GetMemo()
	if err != nil {
		return repository.Transaction{}, nil, err
	}

	rTx := repository.Transaction{
		Hash:                txHash[:],
		BlockID:             blockID,
		TxType:              tx.Type(),
		WrapperID:           wrapper,
		Memo:                memo,
		FeeAmountPerGasUnit: feeAmountPerGasUnit,
		FeeToken:            feeToken,
		GasLimitMultiplier:  gasLimitMultiplier,
		Code:                code,
		Data:                data,
		ReturnCode:          returnCode,
		PosInBlock:          txID,
	}

	return rTx, accountTx, nil
}

func (i *Indexer) decodeTxRawData(txRawData tmtypes.Tx) (types.Tx, error) {
	var pTx ptypes.Tx
	err := proto.Unmarshal(txRawData, &pTx)
	if err != nil {
		return types.Tx{}, errors.New(err, "Proto unmarshal")
	}
	var tx types.Tx
	err = borsh.Deserialize(&tx, pTx.Data)
	if err != nil {
		return types.Tx{}, errors.New(err, "Borsh deserialize")
	}
	return tx, nil
}

func (i *Indexer) findTxReturnCode(txHash types.Hash, resultBlockResults *coretypes.ResultBlockResults) int64 {
	txHashS := strings.ToUpper(txHash.String())
	for _, event := range resultBlockResults.EndBlockEvents {
		correctEvent := false
		for _, attr := range event.Attributes {
			if string(attr.Key) == "hash" && string(attr.Value) == txHashS {
				correctEvent = true
				break
			}
		}
		if !correctEvent {
			continue
		}
		for _, attr := range event.Attributes {
			if string(attr.Key) == "code" {
				v, err := strconv.ParseInt(string(attr.Value), 10, 64)
				if err != nil {
					return -1
				}
				return v
			}
		}
	}
	return -1
}

func (i *Indexer) processSuccessTx(tx types.Tx) (json.RawMessage, *repository.AccountTransaction, error) {
	dataSection, err := tx.GetSection(tx.Header.DataHash)
	if err != nil {
		return nil, nil, errors.New(err, "Get data section")
	}
	if dataSection == nil {
		return nil, nil, nil
	}

	newAccountTransaction := func(secondaryAddress types.Address, defaultAddress ...*types.Address) *repository.AccountTransaction {
		return newAccountTransactionWithTxHash(secondaryAddress, tx.TxHash, tx.BlockHeight, tx.TxPos, defaultAddress...)
	}

	var data interface{}
	var accTx *repository.AccountTransaction

	switch tx.DecryptedTxType {
	case "tx_transfer":
		var elem types.Transfer
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Source)
		data = elem
	case "tx_bond":
		var elem types.Bond
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator, elem.Source)
		data = elem
	case "tx_unbond":
		var elem types.Unbond
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator, elem.Source)
		data = elem
	case "tx_bridge_pool":
		var elem types.PendingTransfer
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Transfer.Sender)
		data = elem
	case "tx_vote_proposal":
		var elem types.VoteProposalData
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Voter)
		data = elem
	case "tx_reveal_pk":
		var elem types.RevealPK
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		data = elem
	case "tx_resign_steward":
		var elem types.ResignSteward
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		data = elem
	case "tx_update_steward_commission":
		var elem types.UpdateStewardCommission
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		data = elem
	case "tx_init_account":
		var elem types.InitAccount
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		data = elem
	case "tx_update_account":
		var elem types.UpdateAccount
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Addr)
		data = elem
	case "tx_ibc":
		data = bytes.HexBytes(dataSection.Data.Data)
	case "tx_become_validator":
		var elem types.BecomeValidator
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Address)
		data = elem
	case "tx_change_consensus_key":
		var elem types.ConsensusKeyChange
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator)
		data = elem
	case "tx_change_validator_commission":
		var elem types.CommissionChange
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator)
		data = elem
	case "tx_change_validator_metadata":
		var elem types.MetaDataChange
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator)
		data = elem
	case "tx_claim_rewards":
		var elem types.ClaimRewards
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator, elem.Source)
		data = elem
	case "tx_deactivate_validator":
		var elem types.Address
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem)
		data = elem
	case "tx_init_proposal":
		var elem types.InitProposalData
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Author)
		data = elem
	case "tx_reactivate_validator":
		var elem types.Address
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem)
		data = elem
	case "tx_unjail_validator":
		var elem types.Address
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem)
		data = elem
	case "tx_redelegate":
		var elem types.Redelegation
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Owner)
		data = elem
	case "tx_withdraw":
		var elem types.Withdraw
		err = borsh.Deserialize(&elem, dataSection.Data.Data)
		accTx = newAccountTransaction(elem.Validator, elem.Source)
		data = elem
	default:
		data = bytes.HexBytes(dataSection.Data.Data)
	}

	if err != nil {
		return nil, nil, err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	return jsonData, accTx, nil
}

func newAccountTransactionWithTxHash(secondaryAddress types.Address, txHash types.Hash, blockHeight, txPos int64, defaultAddress ...*types.Address) *repository.AccountTransaction {
	address := secondaryAddress
	if len(defaultAddress) != 0 && defaultAddress[0] != nil {
		address = *defaultAddress[0]
	}
	return &repository.AccountTransaction{Address: address.String(), TxHash: txHash[:], BlockHeight: blockHeight, TxPos: txPos}
}
