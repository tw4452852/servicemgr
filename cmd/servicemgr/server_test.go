package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"runtime"
	"testing"

	"github.com/tw4452852/servicemgr/client"
	"github.com/tw4452852/servicemgr/util"
)

func init() {
	test = true
	log.SetOutput(ioutil.Discard)
}

func getConnection(s *Server) *Connection {
	s.connMu.RLock()
	conn := s.conn
	s.connMu.RUnlock()
	return conn
}

func createServerEnd(s *Server) (io.ReadWriteCloser, error) {
	serverAddr := s.ln.Addr().String()
	serverEnd, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	// wait server accept us
	conn := getConnection(s)
	for ; conn == nil; conn = getConnection(s) {
	}

	return serverEnd, nil
}

func createClientEnd(s *Server, id int) (io.ReadWriteCloser, error) {
	var err error
	c1, c2 := net.Pipe()
	defer func() {
		if err != nil {
			c1.Close()
			c2.Close()
		}
	}()

	client := client.NewClient(c2)
	if id >= 0 {
		client.SetId(uint32(id))
	}
	err = s.AddClient(client)
	if err != nil {
		return nil, err
	}

	return c1, nil
}

func oneShotRequest(rw io.ReadWriter, req util.TLV) (res util.TLV, err error) {
	err = util.WriteTLV(rw, req)
	if err != nil {
		return
	}
	res, err = util.ReadTLV(rw)
	if err != nil {
		return
	}
	return
}

func TestMakeConnection(t *testing.T) {
	s, err := NewServer("notexistaddr")
	if s != nil || err == nil {
		t.Fatalf("NewServer should return nil server and error, but got server[%v], err[%v]", s, err)
	}

	s, err = NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	if getConnection(s) != nil {
		t.Fatalf("connection should be nil at first")
	}

	serverAddr := s.ln.Addr().String()
	_, err = net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// wait server accept us
	oldConn := getConnection(s)
	for ; oldConn == nil; oldConn = getConnection(s) {
	}

	prevNumGoRoutine := runtime.NumGoroutine()

	_, err = net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// wait server accept us
	for newConn := getConnection(s); newConn == oldConn; newConn = getConnection(s) {
	}

	nowNumGoRoutine := runtime.NumGoroutine()
	if nowNumGoRoutine != prevNumGoRoutine {
		t.Errorf("number of goroutines not equal: previous[%d], now[%d]", prevNumGoRoutine, nowNumGoRoutine)
	}
}

func TestAddClient(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	c1, c2 := net.Pipe()
	defer func() {
		c1.Close()
		c2.Close()
	}()

	client := client.NewClient(c1)
	err = s.addClient(nil)
	if err != dataInvalidErr {
		t.Fatal("AddClient should fail with invalid parameter")
	}
	err = s.AddClient(client)
	if err != nil {
		t.Fatal(err)
	}
	// should error if client exist
	err = s.AddClient(client)
	if err == nil {
		t.Fatal("AddClient should failed with a existed client")
	}
}

func TestSingleDataForward(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	serverEnd, err := createServerEnd(s)
	if err != nil {
		t.Fatal(err)
	}
	defer serverEnd.Close()

	const id = 1
	clientEnd, err := createClientEnd(s, id)
	if err != nil {
		t.Fatal(err)
	}
	defer clientEnd.Close()

	tlvs := [2]util.TLV{
		{T: 1, L: 2, V: []byte{1, 2}},
		{T: uint64(id)<<32 | 1, L: 2, V: []byte{1, 2}},
	}

	// client -> connection
	err = util.WriteTLV(clientEnd, tlvs[0])
	if err != nil {
		t.Fatal(err)
	}
	got, err := util.ReadTLV(serverEnd)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, tlvs[1]) {
		t.Fatalf("send %v from client, expect %v from server, but got %v", tlvs[0], tlvs[1], got)
	}

	// connection -> client
	err = util.WriteTLV(serverEnd, tlvs[1])
	if err != nil {
		t.Fatal(err)
	}
	got, err = util.ReadTLV(clientEnd)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, tlvs[0]) {
		t.Fatalf("send %v from connection, expect %v from client, but got %v", tlvs[1], tlvs[0], got)
	}
}

func TestMultiDataForward(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	serverEnd, err := createServerEnd(s)
	if err != nil {
		t.Fatal(err)
	}
	defer serverEnd.Close()

	const (
		nclients = 10
		ncount   = 30
	)
	tlv := util.TLV{T: 1, L: 2, V: []byte{1, 2}}
	cs := []chan struct{}{}
	for i := 0; i < nclients; i++ {
		ch := make(chan struct{})
		cs = append(cs, ch)
		go func(id int) {
			clientEnd, err := createClientEnd(s, id)
			if err != nil {
				t.Fatal(err)
			}
			defer clientEnd.Close()

			for i := 0; i < ncount; i++ {
				err = util.WriteTLV(clientEnd, tlv)
				if err != nil {
					t.Fatal(err)
				}
			}

			// inform server
			ch <- struct{}{}

			for i := 0; i < ncount; i++ {
				got, err := util.ReadTLV(clientEnd)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, tlv) {
					t.Fatalf("client %d: %d/%d msg %v != %v", id, i, ncount, got, tlv)
				}
			}

			// inform server
			ch <- struct{}{}
		}(i)
	}

	// wait client write done
	for i := 0; i < nclients; i++ {
		<-cs[i]
	}

	// verify what we get
	tlvs := []util.TLV{}
	gets := make(map[string]int)
	for i := 0; i < nclients*ncount; i++ {
		tlv, err := util.ReadTLV(serverEnd)
		if err != nil {
			t.Fatal(err)
		}
		tlvs = append(tlvs, tlv)
		gets[tlv.String()]++
	}
	if len(gets) != nclients {
		t.Fatalf("expect %d from clients, but got %d", nclients, len(gets))
	}
	for k, v := range gets {
		if v != ncount {
			t.Fatalf("expect %d %v, but got %d", ncount, k, v)
		}
	}

	// send back
	for _, tlv := range tlvs {
		err = util.WriteTLV(serverEnd, tlv)
		if err != nil {
			t.Fatal(err)
		}
	}

	// wait client write done
	for i := 0; i < nclients; i++ {
		<-cs[i]
	}
}

func TestConnectionNotEstablish(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	clientEnd, err := createClientEnd(s, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer clientEnd.Close()

	got, err := oneShotRequest(clientEnd, util.TLV{T: 1})
	if err != nil {
		t.Fatal(err)
	}
	expect := util.TLV{T: uint64(ErrorConnectionGone), V: []byte{}}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expect %v, but got %v", expect, got)
	}
}

func TestInvalidType(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	clientEnd, err := createClientEnd(s, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer clientEnd.Close()

	got, err := oneShotRequest(clientEnd, util.TLV{T: uint64(TypeBegin)})
	if err != nil {
		t.Fatal(err)
	}
	expect := util.TLV{T: uint64(ErrorInvalidType), V: []byte{}}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expect %v, but got %v", expect, got)
	}
}

func TestInternalError(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	serverEnd, err := createServerEnd(s)
	if err != nil {
		t.Fatal(err)
	}
	defer serverEnd.Close()

	clientEnd, err := createClientEnd(s, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer clientEnd.Close()

	malform := []byte{0, 0, 0, 0, 0, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0}
	// server should ignore this malform msg
	_, err = serverEnd.Write(malform)
	if err != nil {
		t.Fatal(err)
	}

	// client will receive a error msg for this
	_, err = clientEnd.Write(malform)
	if err != nil {
		t.Fatal(err)
	}
	got, err := util.ReadTLV(clientEnd)
	if err != nil {
		t.Fatal(err)
	}
	expect := util.TLV{T: uint64(ErrorInternal), V: []byte{}}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expect %v, but got %v", expect, got)
	}
}

func TestConnectionGone(t *testing.T) {
	s, err := NewServer(":0")
	if s == nil || err != nil {
		t.Fatalf("NewServer should return success, but got server[%v], err[%v]", s, err)
	}
	defer s.Close()

	clientEnd, err := createClientEnd(s, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer clientEnd.Close()

	serverEnd, err := createServerEnd(s)
	if err != nil {
		t.Fatal(err)
	}
	serverEnd.Close()

	got, err := util.ReadTLV(clientEnd)
	if err != nil {
		t.Fatal(err)
	}
	expect := util.TLV{T: uint64(ErrorConnectionGone), V: []byte{}}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expect %v, but got %v", expect, got)
	}
}
