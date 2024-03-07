package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"math/big"
	"strings"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/the-laziest/namadexer-go/internal/types/basic"
	"github.com/the-laziest/namadexer-go/internal/types/masp"
	"github.com/the-laziest/namadexer-go/pkg/bech32m"
	"github.com/the-laziest/namadexer-go/pkg/borsh"
	"github.com/the-laziest/namadexer-go/pkg/errors"
)

type ChainID string

type Hash [32]byte
type AddressHash [20]byte

var emptyHash Hash

func (h Hash) IsEmpty() bool {
	return h == emptyHash
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Equal(other Hash) bool {
	return bytes.Equal(h[:], other[:])
}

func (h *Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func encodeBytes(hrp string, bs []byte) string {
	result, err := bech32m.EncodeFromBase256(hrp, bs)
	if err != nil {
		return err.Error()
	}
	return result
}

type Uint [4]uint64

func (u Uint) BigInt() *big.Int {
	result := new(big.Int).SetUint64(u[3])
	for i := 2; i >= 0; i-- {
		result = new(big.Int).Lsh(result, 64)
		result = new(big.Int).Add(result, new(big.Int).SetUint64(u[i]))
	}
	return result
}

func (u Uint) String() string {
	return u.BigInt().String()
}

type Dec struct {
	Raw Uint
}

const decPrecision = 12

var e12 = new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decPrecision), nil))

func (d Dec) String() string {
	dec := new(big.Float).SetInt(d.Raw.BigInt())
	dec = new(big.Float).Quo(dec, e12)
	return dec.SetPrec(decPrecision).Text('f', -1)
}

func (d Dec) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type Amount struct {
	Raw Uint
}

func (a Amount) String() string {
	return a.Raw.String()
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

type DenominatedAmount struct {
	Amount Amount
	Denom  uint8
}

func (dn DenominatedAmount) String() string {
	amount := dn.Amount.String()
	if dn.Denom == 0 {
		return amount
	}
	if len(amount) > int(dn.Denom) {
		pos := len(amount) - int(dn.Denom)
		return amount[:pos] + "." + amount[pos:]
	}
	var result strings.Builder
	result.WriteString("0.")
	for range int(dn.Denom) - len(amount) {
		result.WriteRune('0')
	}
	result.WriteString(amount)
	return result.String()
}

func (dn DenominatedAmount) MarshalJSON() ([]byte, error) {
	return json.Marshal(dn.String())
}

type Fee struct {
	AmountPerGasUnit DenominatedAmount
	Token            Address
}

type GasFee struct {
	Amount Amount
	Payer  Address
	Token  Address
}

type WrapperTx struct {
	Fee                 Fee
	Pk                  PublicKey
	Epoch               uint64
	GasLimit            uint64
	UnshieldSectionHash *Hash
}

type DecryptedTx borsh.Enum

const (
	Decrypted DecryptedTx = iota
	Undecryptable
)

type ProtocolTxType borsh.Enum

const (
	EthereumEvents ProtocolTxType = iota
	BridgePool
	ValidatorSetUpdate
	EthEventsVext
	BridgePoolVext
	ValSetUpdateVext
)

type ProtocolTx struct {
	Pk PublicKey
	Tx ProtocolTxType
}

type TxType struct {
	Enum      borsh.Enum `borsh_enum:"true"`
	Raw       struct{}
	Wrapper   WrapperTx
	Decrypted DecryptedTx
	Protocol  ProtocolTx
}

func (tt TxType) IsRaw() bool {
	return tt.Enum == 0
}

func (tt TxType) IsWrapper() bool {
	return tt.Enum == 1
}

func (tt TxType) IsDecrypted() bool {
	return tt.Enum == 2
}

func (tt TxType) IsProtocol() bool {
	return tt.Enum == 3
}

func (tt TxType) Type() string {
	switch tt.Enum {
	case 0:
		return "Raw"
	case 1:
		return "Wrapper"
	case 2:
		return "Decrypted"
	case 3:
		return "Protocol"
	}
	return "Undefined"
}

type Header struct {
	ChainID    ChainID
	Expiration *string
	Timestamp  string
	CodeHash   Hash
	DataHash   Hash
	MemoHash   Hash
	TxType     TxType
}

func (h Header) hash(hasher hash.Hash, txType *TxType) error {
	if txType != nil {
		h.TxType = *txType
	}
	bs, err := borsh.Serialize(h)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

type SectionData struct {
	Salt [8]byte
	Data []byte
}

func (sd SectionData) hash(hasher hash.Hash) error {
	bs, err := borsh.Serialize(sd)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

type Commitment struct {
	Enum borsh.Enum `borsh_enum:"true"`
	Hash Hash
	ID   []byte
}

func (c Commitment) GetHash() Hash {
	if c.Enum == 0 {
		return c.Hash
	}
	sha := sha256.Sum256(c.ID)
	return Hash(sha)
}

func (c Commitment) String() string {
	if c.Enum == 0 {
		if c.Hash.IsEmpty() {
			return ""
		}
		return c.Hash.String()
	}
	return string(c.ID)
}

type SectionCode struct {
	Salt [8]byte
	Code Commitment
	Tag  *string
}

func (sc SectionCode) hash(hasher hash.Hash) error {
	hasher.Write(sc.Salt[:])
	h := sc.Code.GetHash()
	hasher.Write(h[:])
	bs, err := borsh.Serialize(sc.Tag)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

var DEFAULT_ADDRESS AddressHash = [20]byte{}

func getHumanAdress(discriminant byte, address AddressHash) string {
	res := make([]byte, 21)
	res[0] = discriminant
	copy(res[1:], address[:])
	return encodeBytes("tnam", res)
}

type EstablishedAddress struct {
	Hash AddressHash
}

func (ea EstablishedAddress) String() string {
	return getHumanAdress(DiscriminantEstablished, ea.Hash)
}

type ImplicitAddress struct {
	AddressHash
}

func (ia ImplicitAddress) String() string {
	return getHumanAdress(DiscriminantImplicit, ia.AddressHash)
}

type IbcTokenHash AddressHash

func (ith IbcTokenHash) String() string {
	return hex.EncodeToString(ith[:])
}

type EthAddress AddressHash

func (ea EthAddress) String() string {
	return "0x" + hex.EncodeToString(ea[:])
}

type InternalAddress struct {
	Enum          borsh.Enum `borsh_enum:"true"`
	PoS           struct{}
	PosSlashPool  struct{}
	Parameters    struct{}
	Ibc           struct{}
	IbcToken      IbcTokenHash
	Governance    struct{}
	EthBridge     struct{}
	EthBridgePool struct{}
	Erc20         EthAddress
	Nut           EthAddress
	Multitoken    struct{}
	Pgf           struct{}
	Masp          struct{}
}

const (
	DiscriminantImplicit byte = iota
	DiscriminantEstablished
	DiscriminantPos
	DiscriminantSlashPool
	DiscriminantParameters
	DiscriminantGovernance
	DiscriminantIbc
	DiscriminantEthBridge
	DiscriminantBridgePool
	DiscriminantMultitoken
	DiscriminantPgf
	DiscriminantErc20
	DiscriminantNut
	DiscriminantIbcToken
	DiscriminantMasp
)

func (ia InternalAddress) String() string {
	switch ia.Enum {
	case 0:
		return getHumanAdress(DiscriminantPos, DEFAULT_ADDRESS)
	case 1:
		return getHumanAdress(DiscriminantSlashPool, DEFAULT_ADDRESS)
	case 2:
		return getHumanAdress(DiscriminantParameters, DEFAULT_ADDRESS)
	case 3:
		return getHumanAdress(DiscriminantIbc, DEFAULT_ADDRESS)
	case 4:
		return getHumanAdress(DiscriminantIbcToken, AddressHash(ia.IbcToken))
	case 5:
		return getHumanAdress(DiscriminantGovernance, DEFAULT_ADDRESS)
	case 6:
		return getHumanAdress(DiscriminantEthBridge, DEFAULT_ADDRESS)
	case 7:
		return getHumanAdress(DiscriminantBridgePool, DEFAULT_ADDRESS)
	case 8:
		return getHumanAdress(DiscriminantErc20, AddressHash(ia.Erc20))
	case 9:
		return getHumanAdress(DiscriminantNut, AddressHash(ia.Nut))
	case 10:
		return getHumanAdress(DiscriminantMultitoken, DEFAULT_ADDRESS)
	case 11:
		return getHumanAdress(DiscriminantPgf, DEFAULT_ADDRESS)
	case 12:
		return getHumanAdress(DiscriminantMasp, DEFAULT_ADDRESS)
	}
	return ""
}

type Address struct {
	Enum        borsh.Enum `borsh_enum:"true"`
	Established EstablishedAddress
	Implicit    ImplicitAddress
	Internal    InternalAddress
}

func (a Address) String() string {
	switch a.Enum {
	case 0:
		return a.Established.String()
	case 1:
		return a.Implicit.String()
	case 2:
		return a.Internal.String()
	}
	return ""
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a Address) NotInternal() bool {
	return a.Enum != 2
}

type Ed25519PublicKey [32]byte

func (epk Ed25519PublicKey) String() string {
	return ed25519.PubKey(epk[:]).Address().String()
}

func (epk Ed25519PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(epk.String())
}

type Secp256k1PublicKey [33]byte

func (spk Secp256k1PublicKey) String() string {
	return secp256k1.PubKey(spk[:]).Address().String()
}

func (spk Secp256k1PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(spk.String())
}

type PublicKey struct {
	Enum      borsh.Enum `borsh_enum:"true"`
	Ed25519   Ed25519PublicKey
	Secp256k1 Secp256k1PublicKey
}

func (pk PublicKey) String() string {
	bs, _ := borsh.Serialize(pk)
	return encodeBytes("tpknam", bs)
}

func (pk PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pk.String())
}

type Signer struct {
	Enum    borsh.Enum `borsh_enum:"true"`
	Address Address
	PubKeys []PublicKey
}

type Ed25519Signature [64]byte

type Secp256k1Signature [65]byte

type Signature struct {
	Enum      borsh.Enum `borsh_enum:"true"`
	Ed25519   Ed25519Signature
	Secp256k1 Secp256k1Signature
}

func (s Signature) String() string {
	bs, _ := borsh.Serialize(s)
	return encodeBytes("signam", bs)
}

type SectionSignature struct {
	Targets    []Hash
	Signer     Signer
	Signatures basic.BTreeMap[byte, Signature]
}

func (ss SectionSignature) hash(hasher hash.Hash) error {
	bs, err := borsh.Serialize(ss)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

type Ciphertext struct {
	Opaque []byte
}

func (ct Ciphertext) hash(hasher hash.Hash) error {
	bs, err := borsh.Serialize(ct)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

type MaspDigitPos borsh.Enum

const (
	MaspDigitPosZero MaspDigitPos = iota
	MaspDigitPosOne
	MaspDigitPosTwo
	MaspDigitPosThree
)

type AssetData struct {
	Token    Address
	Denom    uint8
	Position MaspDigitPos
	Epoch    *uint64
}

type MaspBuilder struct {
	Target     Hash
	AssetTypes []AssetData
	Metadata   masp.SaplingMetadata
	Builder    masp.Builder
}

func (mb MaspBuilder) hash(hasher hash.Hash) error {
	bs, err := borsh.Serialize(mb)
	if err != nil {
		return err
	}
	hasher.Write(bs)
	return nil
}

type Section struct {
	Enum        borsh.Enum `borsh_enum:"true"`
	Data        SectionData
	ExtraData   SectionCode
	Code        SectionCode
	Signature   SectionSignature
	Ciphertext  Ciphertext
	MaspTx      masp.Transaction
	MaspBuilder MaspBuilder
	Header      Header
}

func (s Section) GetHash() (Hash, error) {
	hasher := sha256.New()

	hasher.Write([]byte{byte(s.Enum)})

	var err error

	switch s.Enum {
	case 0:
		err = s.Data.hash(hasher)
	case 1:
		err = s.ExtraData.hash(hasher)
	case 2:
		err = s.Code.hash(hasher)
	case 3:
		err = s.Signature.hash(hasher)
	case 4:
		err = s.Ciphertext.hash(hasher)
	case 5:
		err = s.MaspTx.Hash(hasher)
	case 6:
		err = s.MaspBuilder.hash(hasher)
	case 7:
		err = s.Header.hash(hasher, nil)
	}

	if err != nil {
		return Hash{}, err
	}

	return Hash(hasher.Sum(nil)), nil
}

type Tx struct {
	Header          Header
	Sections        []Section
	DecryptedTxType string `borsh_skip:"true"`
	TxHash          Hash   `borsh_skip:"true"`
	BlockHeight     int64  `borsh_skip:"true"`
	TxPos           int64  `borsh_skip:"true"`
}

func (t Tx) Type() string {
	return t.Header.TxType.Type()
}

func (t Tx) GetHash() (Hash, error) {
	hasher := sha256.New()

	hasher.Write([]byte{7})

	var replaceTxType *TxType
	if t.Header.TxType.IsDecrypted() {
		replaceTxType = &TxType{Enum: 0}
	}

	err := t.Header.hash(hasher, replaceTxType)
	if err != nil {
		return Hash{}, err
	}

	return Hash(hasher.Sum(nil)), nil
}

func (t Tx) GetSection(h Hash) (*Section, error) {
	if h.IsEmpty() {
		return nil, nil
	}
	for _, s := range t.Sections {
		secHash, err := s.GetHash()
		if err != nil {
			return nil, err
		}
		if h.Equal(secHash) {
			return &s, nil
		}
	}
	return nil, nil
}

func (t Tx) GetCodeHash() (Hash, error) {
	codeSection, err := t.GetSection(t.Header.CodeHash)
	if err != nil {
		return Hash{}, errors.New(err, "Get code section")
	}
	if codeSection == nil {
		return Hash{}, nil
	}

	return codeSection.Code.Code.GetHash(), nil
}

func (t Tx) GetMemo() (string, error) {
	memeSection, err := t.GetSection(t.Header.MemoHash)
	if err != nil {
		return "", errors.New(err, "Get memo section")
	}
	if memeSection == nil {
		return "", nil
	}
	return memeSection.ExtraData.Code.String(), nil
}
