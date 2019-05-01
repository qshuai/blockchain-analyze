package main

import (
	"bytes"
	"encoding/binary"

	"github.com/btcsuite/btcd/wire"
)

type Utxo struct {
	isCoinbase bool
	height     int
	amount     int64
	pkScript   []byte
}

func (u *Utxo) encode() []byte {
	if u == nil {
		return nil
	}

	serializeSize := 4 + 8 + wire.VarIntSerializeSize(uint64(len(u.pkScript))) + len(u.pkScript)
	buf := bytes.NewBuffer(make([]byte, 0, serializeSize))
	buf.Write(u.coinbaseAndHeight())

	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, uint64(u.amount))
	buf.Write(value)

	err := wire.WriteVarInt(buf, 0, uint64(wire.VarIntSerializeSize(uint64(len(u.pkScript)))))
	if err != nil {
		panic(err)
	}

	buf.Write(u.pkScript)

	return buf.Bytes()
}

func (u *Utxo) decode(value []byte) error {
	buf := bytes.NewReader(value)
	coinbaseAndHeight := make([]byte, 4)
	_, err := buf.Read(coinbaseAndHeight)
	if err != nil {
		return err
	}

	v := make([]byte, 8)
	_, err = buf.Read(v)
	if err != nil {
		return err
	}

	length, err := wire.ReadVarInt(buf, 0)
	if err != nil {
		return err
	}

	ps := make([]byte, length)
	_, err = buf.Read(ps)
	if err != nil {
		return err
	}

	info := binary.LittleEndian.Uint32(coinbaseAndHeight)

	u.isCoinbase = true
	u.height = int(info / 2)
	u.amount = int64(binary.LittleEndian.Uint64(v))
	u.pkScript = ps

	return nil
}

func (u *Utxo) coinbaseAndHeight() []byte {
	r := make([]byte, 4)
	if u.isCoinbase {
		binary.LittleEndian.PutUint32(r, uint32((u.height*2)|1))
	} else {
		binary.LittleEndian.PutUint32(r, uint32(u.height*2))
	}

	return r
}
