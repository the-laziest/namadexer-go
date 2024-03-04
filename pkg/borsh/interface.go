package borsh

import "io"

type BorshDeserializer interface {
	DecodeBorsh(r io.Reader, read ReadFunc) (interface{}, error)
}
