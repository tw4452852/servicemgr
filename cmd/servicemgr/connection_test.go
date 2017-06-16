package main

import (
	"encoding/json"
	"net"
	"reflect"
	"testing"

	"github.com/tw4452852/servicemgr/util"
)

func TestInitAudio(t *testing.T) {
	old := test
	test = false
	defer func() {
		test = old
	}()

	c1, c2 := net.Pipe()
	defer func() {
		c1.Close()
		c2.Close()
	}()

	done := make(chan struct{})
	defer close(done)

	go func() {
		defer func() {
			done <- struct{}{}
		}()
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
			t.Fatal(err)
		}

		expect := util.TLV{
			T: uint64(TypeOpenSound),
			L: uint64(len(req)),
			V: req,
		}
		got, err := util.ReadTLV(c2)
		if err != nil {
			t.Fatal(err)
			return
		}
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %v but got %v", expect, got)
			return
		}

		// mock success at first
		err = util.WriteTLV(c2, util.TLV{T: uint64(TypeOpenSound)})
		if err != nil {
			t.Fatal(err)
			return
		}

		got, err = util.ReadTLV(c2)
		if err != nil {
			t.Fatal(err)
			return
		}
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %v but got %v", expect, got)
			return
		}

		// mock failure then
		err = util.WriteTLV(c2, util.TLV{T: 0xdead})
		if err != nil {
			t.Fatal(err)
			return
		}
	}()

	conn, err := CreateConnection(c1)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if conn.disableAudio {
		t.Errorf("expect audio work, but not")
	}

	conn, err = CreateConnection(c1)
	if err != dataInvalidErr {
		t.Errorf("not got expected error: %v", dataInvalidErr)
	}
	if !conn.disableAudio {
		t.Errorf("expect audio doesn't work, but it does")
	}

	// wait goroutine exit
	<-done
}
