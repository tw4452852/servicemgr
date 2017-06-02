package util

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

var InternalErr = errors.New("internal error")

type TLV struct {
	T uint64
	L uint64
	V []byte
}

func (tlv TLV) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[type: %#x, length: %d, ", tlv.T, tlv.L)
	v := tlv.V
	// show at most 32 bytes
	if len(v) > 32 {
		v = v[:32]
	}
	fmt.Fprintf(&buf, "value: %q]", v)
	return buf.String()
}

var lengthMismatchErr = errors.New("length is mismatch")

func WriteTLV(w io.Writer, tlv TLV) error {
	// TODO: use buffer pool
	var b bytes.Buffer

	if int(tlv.L) != binary.Size(tlv.V) {
		log.Printf("[tlv]: length mismatch expect[%d], but got[%d]\n",
			binary.Size(tlv.V), int(tlv.L))
		return lengthMismatchErr
	}

	err := binary.Write(&b, binary.BigEndian, tlv.T)
	if err != nil {
		log.Printf("[tlv]: write type[%#x] error: %s\n", tlv.T, err)
		return err
	}

	err = binary.Write(&b, binary.BigEndian, tlv.L)
	if err != nil {
		log.Printf("[tlv]: write length[%#x] error: %s\n", tlv.L, err)
		return err
	}

	err = binary.Write(&b, binary.BigEndian, tlv.V)
	if err != nil {
		log.Printf("[tlv]: write value[%v] error: %s\n", tlv.V, err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, b.Bytes())
	if err != nil {
		log.Printf("[tlv]: write tlv[%v] error: %s\n", b.Bytes(), err)
		return err
	}

	return nil
}

func ReadTLV(r io.Reader) (tlv TLV, err error) {
	var (
		t uint64
		l uint64
		v []byte
	)
	defer func() {
		if e := recover(); e != nil {
			log.Printf("panic: %v\nt[%#x], l[%#x],  v[%v]\n", e, t, l, v)
			err = InternalErr
		}
	}()

	err = binary.Read(r, binary.BigEndian, &t)
	if err != nil {
		ne, ok := err.(net.Error)
		if !ok || !ne.Temporary() {
			log.Printf("[tlv]: read type error: %s\n", err)
		}
		return
	}

	err = binary.Read(r, binary.BigEndian, &l)
	if err != nil {
		ne, ok := err.(net.Error)
		if !ok || !ne.Temporary() {
			log.Printf("[tlv]: read length error: %s\n", err)
		}
		return
	}

	v = make([]byte, l)
	err = binary.Read(r, binary.BigEndian, &v)
	if err != nil {
		ne, ok := err.(net.Error)
		if !ok || !ne.Temporary() {
			log.Printf("[tlv]: read value error: %s\n", err)
		}
		return
	}

	tlv = TLV{
		T: t,
		L: l,
		V: v,
	}

	return tlv, nil
}
