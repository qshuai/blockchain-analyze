package main

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

const SerializeSize = 32

type AddressBalanceInfo struct {
	received  uint64
	send      uint64
	txes      uint32
	unspentTx uint32
	created   uint32
	updated   uint32
}

func (info *AddressBalanceInfo) encode() []byte {
	r := make([]byte, 32)
	binary.LittleEndian.PutUint64(r[:8], info.received)
	binary.LittleEndian.PutUint64(r[8:16], info.send)
	binary.LittleEndian.PutUint32(r[16:20], info.txes)
	binary.LittleEndian.PutUint32(r[20:24], info.unspentTx)
	binary.LittleEndian.PutUint32(r[24:28], info.created)
	binary.LittleEndian.PutUint32(r[28:32], info.updated)

	return r
}

func decode(value []byte) (*AddressBalanceInfo, error) {
	if value == nil || len(value) != SerializeSize {
		return nil, errors.New("invalid size")
	}

	info := AddressBalanceInfo{
		received:  binary.LittleEndian.Uint64(value[:8]),
		send:      binary.LittleEndian.Uint64(value[8:16]),
		txes:      binary.LittleEndian.Uint32(value[16:20]),
		unspentTx: binary.LittleEndian.Uint32(value[20:24]),
		created:   binary.LittleEndian.Uint32(value[24:28]),
		updated:   binary.LittleEndian.Uint32(value[28:32]),
	}

	return &info, nil
}

func (info *AddressBalanceInfo) getBalance() uint64 {
	if info == nil {
		return 0
	}

	return info.received - info.send
}
