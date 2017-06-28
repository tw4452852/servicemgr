package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/tw4452852/servicemgr/util"
)

var test = false

const (
	audioFormat  = 2     // pcm 16 bit
	audioRate    = 44100 // sample rate
	audioChannel = 2     // channel count
)

type Connection struct {
	disableAudio bool
	net.Conn
}

func CreateConnection(c net.Conn) (*Connection, error) {
	conn := &Connection{
		disableAudio: true,
		Conn:         c,
	}

	if test {
		return conn, nil
	}

	err := conn.initAudio()
	if err != nil {
		return conn, err
	}

	return conn, nil
}

func (conn *Connection) initAudio() error {
	req, err := json.Marshal(struct {
		Format  int `json:"format"`
		Rate    int `json:"rate"`
		Channel int `json:"channel"`
	}{
		Format:  audioFormat,
		Rate:    audioRate,
		Channel: audioChannel,
	})
	if err != nil {
		return err
	}

	tlv := util.TLV{
		T: uint64(TypeOpenSound),
		L: uint64(len(req)),
		V: req,
	}
	err = util.WriteTLV(conn, tlv)
	if err != nil {
		return err
	}
	tlv, err = util.ReadTLV(conn)
	if Type(tlv.T) != TypeOpenSound {
		log.Printf("[audio]: received a unmatched type[%#v], want %v", tlv, TypeOpenSound)
		return dataInvalidErr
	}

	// enable audio
	conn.disableAudio = false

	return nil
}

func (conn *Connection) WriteTLV(tlv util.TLV) error {
	t := Type(tlv.T & 0x00000000ffffffff)

	if t == TypeSoundData && conn.disableAudio {
		log.Printf("[connection]: audio is disable, skip audio data\n")
		return nil
	}

	return util.WriteTLV(conn, tlv)
}
