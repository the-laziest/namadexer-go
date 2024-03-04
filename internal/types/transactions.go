package types

import (
	"bytes"
	"encoding/json"

	"github.com/the-laziest/namadexer-go/internal/types/basic"
	"github.com/the-laziest/namadexer-go/pkg/borsh"
)

type Transfer struct {
	Source   Address           `json:"source"`
	Target   Address           `json:"target"`
	Token    Address           `json:"token"`
	Amount   DenominatedAmount `json:"amount"`
	Key      *string           `json:"key,omitempty"`
	Shielded *Hash             `json:"shielded,omitempty"`
}

type BecomeValidator struct {
	Address                 Address            `json:"address"`
	ConsensysKey            PublicKey          `json:"consensys_key"`
	EthColdKey              Secp256k1PublicKey `json:"eth_cold_key"`
	EthHotKey               Secp256k1PublicKey `json:"eth_hot_key"`
	ProtocolKey             PublicKey          `json:"protocol_key"`
	ComissionRate           Dec                `json:"comission_rate"`
	MaxCommissionRateChange Dec                `json:"max_commission_rate_change"`
	Email                   string             `json:"email"`
	Description             *string            `json:"description,omitempty"`
	Website                 *string            `json:"website,omitempty"`
	DiscordHandle           *string            `json:"discord_handle,omitempty"`
	Avatar                  *string            `json:"avatar,omitempty"`
}

type Bond struct {
	Validator Address  `json:"validator"`
	Amount    Amount   `json:"amount"`
	Source    *Address `json:"source,omitempty"`
}

type Unbond Bond

type Withdraw struct {
	Validator Address  `json:"validator"`
	Source    *Address `json:"source,omitempty"`
}

type ClaimRewards struct {
	Validator Address  `json:"validator"`
	Source    *Address `json:"source,omitempty"`
}

type Redelegation struct {
	SrcValidator  Address `json:"src_validator"`
	DestValidator Address `json:"dest_validator"`
	Owner         Address `json:"owner"`
	Amount        Amount  `json:"amount"`
}

type CommissionChange struct {
	Validator Address `json:"validator"`
	NewRate   Dec     `json:"new_rate"`
}

type MetaDataChange struct {
	Validator     Address `json:"validator"`
	Email         *string `json:"email,omitempty"`
	Description   *string `json:"description,omitempty"`
	Website       *string `json:"website,omitempty"`
	DiscordHandle *string `json:"discord_handle,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	ComissionRate *Dec    `json:"commission_rate,omitempty"`
}

type ConsensusKeyChange struct {
	Validator    Address   `json:"validator"`
	ConsensusKey PublicKey `json:"consensus_key"`
}

type AddRemove[T any] struct {
	Enum   borsh.Enum `borsh_enum:"true"`
	Add    T
	Remove T
}

func (ar AddRemove[T]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{\"")
	if ar.Enum == 0 {
		buf.WriteString("add")
	} else {
		buf.WriteString("remove")
	}
	buf.WriteString("\":")
	var bs []byte
	var err error
	if ar.Enum == 0 {
		bs, err = json.Marshal(ar.Add)
	} else {
		bs, err = json.Marshal(ar.Remove)
	}
	if err != nil {
		return nil, err
	}
	buf.Write(bs)
	buf.WriteString("}")
	return buf.Bytes(), nil
}

type PGFInternalTarget struct {
	Target Address `json:"target"`
	Amount Amount  `json:"amount"`
}

type PGFIbcTarget struct {
	Target    string `json:"target"`
	Amount    Amount `json:"amount"`
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

type PGFTarget struct {
	Enum     borsh.Enum `borsh_enum:"true"`
	Internal PGFInternalTarget
	Ibc      PGFIbcTarget
}

func (pt PGFTarget) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{\"")
	if pt.Enum == 0 {
		buf.WriteString("Internal")
	} else {
		buf.WriteString("Ibc")
	}
	buf.WriteString("\":")
	var bs []byte
	var err error
	if pt.Enum == 0 {
		bs, err = json.Marshal(pt.Internal)
	} else {
		bs, err = json.Marshal(pt.Ibc)
	}
	if err != nil {
		return nil, err
	}
	buf.Write(bs)
	buf.WriteString("}")
	return buf.Bytes(), nil
}

type PGFAction struct {
	Enum       borsh.Enum `borsh_enum:"true"`
	Continuous AddRemove[PGFTarget]
	Retro      PGFTarget
}

func (pa PGFAction) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{\"")
	if pa.Enum == 0 {
		buf.WriteString("Continuous")
	} else {
		buf.WriteString("Retro")
	}
	buf.WriteString("\":")
	var bs []byte
	var err error
	if pa.Enum == 0 {
		bs, err = json.Marshal(pa.Continuous)
	} else {
		bs, err = json.Marshal(pa.Retro)
	}
	if err != nil {
		return nil, err
	}
	buf.Write(bs)
	buf.WriteString("}")
	return buf.Bytes(), nil
}

type ProposalType struct {
	Enum       borsh.Enum `borsh_enum:"true"`
	Default    *Hash
	PGFSteward []AddRemove[Address]
	PGFPayment []PGFAction
}

func (pt ProposalType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{\"")
	if pt.Enum == 0 {
		buf.WriteString("Default")
	} else if pt.Enum == 1 {
		buf.WriteString("PGFSteward")
	} else {
		buf.WriteString("PGFPayment")
	}
	buf.WriteString("\":")
	var bs []byte
	var err error
	if pt.Enum == 0 {
		bs, err = json.Marshal(pt.Default)
	} else if pt.Enum == 1 {
		bs, err = json.Marshal(pt.PGFSteward)
	} else {
		bs, err = json.Marshal(pt.PGFPayment)
	}
	if err != nil {
		return nil, err
	}
	buf.Write(bs)
	buf.WriteString("}")
	return buf.Bytes(), nil
}

type InitProposalData struct {
	ID               uint64       `json:"id"`
	Content          Hash         `json:"content"`
	Author           Address      `json:"author"`
	Type             ProposalType `json:"type"`
	VotingStartEpoch uint64       `json:"voting_start_epoch"`
	VotingEndEpoch   uint64       `json:"voting_end_epoch"`
	GraceEpoch       uint64       `json:"grace_epoch"`
}

type ProposalVote borsh.Enum

const (
	ProposalVoteYay ProposalVote = iota
	ProposalVoteNay
	ProposalVoteAbstain
)

func (pv ProposalVote) String() string {
	switch pv {
	case ProposalVoteYay:
		return "Yay"
	case ProposalVoteNay:
		return "Nay"
	case ProposalVoteAbstain:
		return "Abstain"
	default:
		return ""
	}
}

func (pb ProposalVote) MarshalJSON() ([]byte, error) {
	return json.Marshal(pb.String())
}

type VoteProposalData struct {
	ID          uint64       `json:"id"`
	Vote        ProposalVote `json:"vote"`
	Voter       Address      `json:"voter"`
	Delegations []Address    `json:"delegations"`
}

type RevealPK PublicKey

type ResignSteward Address

type UpdateStewardCommission struct {
	Steward    Address                      `json:"steward"`
	Commission basic.BTreeMap[Address, Dec] `json:"commission"`
}

type InitAccount struct {
	PublicKeys []PublicKey `json:"public_keys"`
	VpCodeHash Hash        `json:"vp_code_hash"`
	Threshold  uint8       `json:"threshold"`
}

type UpdateAccount struct {
	Addr       Address     `json:"address"`
	VpCodeHash *Hash       `json:"vp_code_hash,omitempty"`
	PublicKeys []PublicKey `json:"public_keys,omitempty"`
	Threshold  *uint8      `json:"threshold,omitempty"`
}

type TransferToEthereumKind borsh.Enum

const (
	TransferToEthereumKindErc20 TransferToEthereumKind = iota
	TransferToEthereumKindNut
)

func (ttek TransferToEthereumKind) String() string {
	if ttek == TransferToEthereumKindErc20 {
		return "ERC20"
	}
	return "NUT"
}

func (ttek TransferToEthereumKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(ttek.String())
}

type TransferToEthereum struct {
	Kind      TransferToEthereumKind `json:"kind"`
	Asset     EthAddress             `json:"asset"`
	Recipient EthAddress             `json:"recipient"`
	Sender    Address                `json:"sender"`
	Amount    Amount                 `json:"amount"`
}

type PendingTransfer struct {
	Transfer TransferToEthereum `json:"transfer"`
	GasFee   GasFee             `json:"gas_fee"`
}
