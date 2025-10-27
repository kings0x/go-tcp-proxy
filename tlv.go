package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

//we are going to set the type binary type
//we set the max payload size which would be a bytes 32
//set a type of binary which would be a bytes array
//have a set of methods the binary array would implement

const (
	BinaryType uint8 = iota + 1

	MaxSize = 10 << 20
)

var ErrMaxSizeReached = errors.New("size is greater than max size")

type Binary []byte

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

func (b Binary) String() string {
	return string(b)
}

func (b Binary) Bytes() []byte {
	return b
}

func (b Binary) WriteTo(w io.Writer) (int64, error) {
	//send the type first
	//send the size
	//send the message
	err := binary.Write(w, binary.BigEndian, BinaryType)

	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(b)))

	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write(b)

	if err != nil {
		return n, err
	}

	return n + int64(o), nil
}

func (b *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8

	err := binary.Read(r, binary.BigEndian, &typ)

	if err != nil {
		return 0, err
	}

	var n int64 = 1

	if typ != BinaryType {
		return n, errors.New("invalid binary")
	}

	var size int32

	err = binary.Read(r, binary.BigEndian, &size)

	if err != nil {
		return n, err
	}

	n += 4

	if size > MaxSize {
		return n, ErrMaxSizeReached
	}

	*b = make([]byte, size)

	o, err := r.Read(*b)

	if err != nil {
		return n, err
	}
	return n + int64(o), nil
}

func decode(r io.Reader) (Payload, error) {
	var typ uint8

	err := binary.Read(r, binary.BigEndian, &typ)

	if err != nil {
		return nil, err
	}

	var payload Payload

	switch typ {
	case BinaryType:
		payload = new(Binary)

	default:
		return nil, errors.New("invalid type")
	}

	_, err = payload.ReadFrom(io.MultiReader(bytes.NewReader([]byte{typ}), r))

	if err != nil {
		return nil, err
	}

	return payload, nil
}
