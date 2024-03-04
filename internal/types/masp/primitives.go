// Currently this package can be used ONLY for deserializing!

package masp

import (
	"hash"
	"io"

	"github.com/the-laziest/namadexer-go/internal/types/basic"
	"github.com/the-laziest/namadexer-go/pkg/borsh"
)

const MASPv5VersionGroupID uint32 = 0x26A7270A

type TxVersion struct {
	Header         uint32
	VersionGroupID uint32
}

type BranchID uint32

const BranchIDMASP BranchID = 0xe9ff75a6

type AssetType [32]byte

type AddressHash [20]byte

type TransparentAddress AddressHash

type TxInOut struct {
	AssetType AssetType
	Value     uint64
	Address   TransparentAddress
}

type TxsInOut []TxInOut

func (ti TxsInOut) DecodeBorsh(r io.Reader, read borsh.ReadFunc) (interface{}, error) {
	sz, err := ReadCompactSize(r)
	if err != nil {
		return nil, err
	}
	txs := make([]TxInOut, sz)
	for i := range sz {
		bs, err := read(32)
		if err != nil {
			return nil, err
		}
		copy(txs[i].AssetType[:], bs)
		txs[i].Value, err = borsh.ReadUint64(r)
		if err != nil {
			return nil, err
		}
		bs, err = read(20)
		if err != nil {
			return nil, err
		}
		copy(txs[i].Address[:], bs)
	}
	return TxsInOut(txs), nil
}

type TransparentBundle struct {
	Vin  TxsInOut
	Vout TxsInOut
}

const spendDescriptionV5Len = 32 + 32 + 32

type SpendDescriptionV5Bytes [spendDescriptionV5Len]byte

const convertDescriptionV5Len = 32

type ConvertDescriptionV5Bytes [convertDescriptionV5Len]byte

const outputDescriptionV5Len = 32 + 32 + 32 + 580 + 32 + 80

type OutputDescriptionV5Bytes [outputDescriptionV5Len]byte

const (
	assetTypeLen    = 32
	i128SumLen      = assetTypeLen + 16
	zkProofLen      = 48 + 96 + 48
	spendAuthSigLen = 64
	bindingSigLen   = 64
)

type SaplingBundle struct {
}

func readI128Sum(r io.Reader) ([][]byte, error) {
	sz, err := ReadCompactSize(r)
	if err != nil {
		return nil, err
	}
	bs, err := borsh.ReadBytesArray(r, sz, i128SumLen)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (sb SaplingBundle) DecodeBorsh(r io.Reader, read borsh.ReadFunc) (interface{}, error) {
	sdl, err := ReadCompactSize(r)
	if err != nil {
		return nil, err
	}
	sds, err := borsh.ReadBytesArray(r, sdl, spendDescriptionV5Len)
	if err != nil {
		return nil, err
	}
	cdl, err := ReadCompactSize(r)
	if err != nil {
		return nil, err
	}
	cds, err := borsh.ReadBytesArray(r, cdl, convertDescriptionV5Len)
	if err != nil {
		return nil, err
	}
	odl, err := ReadCompactSize(r)
	if err != nil {
		return nil, err
	}
	ods, err := borsh.ReadBytesArray(r, odl, outputDescriptionV5Len)
	if err != nil {
		return nil, err
	}
	if len(sds) > 0 || len(ods) > 0 {
		_, err = readI128Sum(r)
		if err != nil {
			return nil, err
		}
	}
	if len(sds) > 0 {
		_, err = read(32)
		if err != nil {
			return nil, err
		}
	}
	if len(cds) > 0 {
		_, err = read(32)
		if err != nil {
			return nil, err
		}
	}
	_, err = borsh.ReadBytesArray(r, len(sds), zkProofLen)
	if err != nil {
		return nil, err
	}
	_, err = borsh.ReadBytesArray(r, len(sds), spendAuthSigLen)
	if err != nil {
		return nil, err
	}
	_, err = borsh.ReadBytesArray(r, len(cds), zkProofLen)
	if err != nil {
		return nil, err
	}
	_, err = borsh.ReadBytesArray(r, len(ods), zkProofLen)
	if err != nil {
		return nil, err
	}
	if len(sds) > 0 || len(ods) > 0 {
		_, err = read(bindingSigLen)
		if err != nil {
			return nil, err
		}
	}
	return SaplingBundle{}, nil
}

type SaplingMetadata struct {
	SpendIndices   []uint
	ConvertIndices []uint
	OutputIndices  []uint
}

type TransparentInputInfo struct {
	Coin TxInOut
}

type TransparentBuilder struct {
	Inputs []TransparentInputInfo
	Vout   []TxInOut
}

type ExpandedSpendingKey [96]byte

type ExtendedSpendingKey struct {
	Depth      uint8
	Tag        [4]byte
	ChildIndex uint32
	ChainCode  [32]byte
	ExpSK      ExpandedSpendingKey
	Dk         [32]byte
}

type Note struct {
	AssetType AssetType
	Value     uint64
	Gd        [32]byte
	Pkd       [32]byte
	RSeedType byte
	RSeed     [32]byte
}

type MerklePath struct {
	AuthPath [][33]byte
	Position uint64
}

func (mp MerklePath) DecodeBorsh(r io.Reader, read borsh.ReadFunc) (interface{}, error) {
	depth, err := read(1)
	if err != nil {
		return nil, err
	}
	path, err := borsh.ReadBytesArray(r, int(depth[0]), 33)
	if err != nil {
		return nil, err
	}
	position, err := borsh.ReadUint64(r)
	if err != nil {
		return nil, err
	}
	authPath := make([][33]byte, len(path))
	for i, p := range path {
		copy(authPath[i][:], p)
	}
	return MerklePath{
		AuthPath: authPath,
		Position: position,
	}, nil
}

type SpendDescriptionInfo struct {
	ExpSK       ExpandedSpendingKey
	Diversifier [11]byte
	Note        Note
	Alpha       [32]byte
	MerklePath  MerklePath
}

type AllowedConversion struct {
	Assets    [][]byte
	Generator [32]byte
}

func (ac AllowedConversion) DecodeBorsh(r io.Reader, read borsh.ReadFunc) (interface{}, error) {
	assets, err := readI128Sum(r)
	if err != nil {
		return nil, err
	}
	bs, err := read(32)
	if err != nil {
		return nil, err
	}
	var generator [32]byte
	copy(generator[:], bs)
	return AllowedConversion{
		Assets:    assets,
		Generator: generator,
	}, nil
}

type ConvertDescriptionInfo struct {
	Allowed    AllowedConversion
	Value      uint64
	MerklePath MerklePath
}

type OutgoingViewingKey [32]byte

type PaymentAddress [43]byte

type MemoBytes [512]byte

type SaplingOutputInfo struct {
	Ovk  *OutgoingViewingKey
	To   PaymentAddress
	Note Note
	Memo MemoBytes
}

type i128 [16]byte

type I128Sum basic.BTreeMap[AssetType, i128]

type SaplingBuilder struct {
	Params        struct{}
	SpendAnchor   *([32]byte)
	TargetHeight  uint32
	ValueBalance  I128Sum
	ConvertAnchor *([32]byte)
	Spends        []SpendDescriptionInfo
	Converts      []ConvertDescriptionInfo
	Outputs       []SaplingOutputInfo
}

type Builder struct {
	TargetHeight       uint32
	ExpiryHeight       uint32
	TransparentBuilder TransparentBuilder
	SaplingBuilder     SaplingBuilder
}

type TransactionData struct {
	Version           TxVersion
	ConsensusBranchID BranchID
	LockTime          uint32
	ExpiryHeight      uint32
	TransparentBundle TransparentBundle
	SaplingBundle     SaplingBundle
}

type Transaction struct {
	TxID [32]byte `borsh_skip:"true"`
	Data TransactionData
}

func (t Transaction) Hash(hasher hash.Hash) error {
	hasher.Write(t.TxID[:])
	return nil
}
