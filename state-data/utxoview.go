package main

import (
	"bytes"
	"encoding/binary"

	"github.com/btcsuite/btcd/wire"
)

const BaseSerializeSize = 12

type utxoView struct {
	isCoinbase bool
	height     int
	amount     int64
	pkScript   []byte
}

func (view *utxoView) encode() []byte {
	r := make([]byte, 12+wire.VarIntSerializeSize(uint64(len(view.pkScript)))+len(view.pkScript))
	if view.isCoinbase {
		binary.LittleEndian.PutUint32(r[:4], uint32((view.height*2)|1))
	} else {
		binary.LittleEndian.PutUint32(r[:4], uint32(view.height*2))
	}

	binary.LittleEndian.PutUint64(r[4:12], uint64(view.amount))

	bf := bytes.NewBuffer(make([]byte, 0, 12+wire.VarIntSerializeSize(uint64(len(view.pkScript)))+len(view.pkScript)))
	hb := make([]byte, 4)
	if view.isCoinbase {
		binary.LittleEndian.PutUint32(hb, uint32((view.height*2)|1))
	} else {
		binary.LittleEndian.PutUint32(hb, uint32(view.height*2))
	}
	bf.Write(hb[:])

	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, uint64(view.amount))
	bf.Write(value[:])

	wire.WriteVarInt(bf, 0, uint64(len(view.pkScript)))
}
