package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/tw4452852/servicemgr/util"
)

var fakeTest = false

type cmdType int

const (
	addClient cmdType = iota
)

type cmd struct {
	typ  cmdType
	data interface{}
	err  chan error
}

type Server struct {
	ln net.Listener

	connMu sync.RWMutex
	conn   *Connection

	clients sync.Map

	cmds chan *cmd
	exit chan struct{}
}

func NewServer(listenAddr string) (*Server, error) {
	s := &Server{
		cmds: make(chan *cmd, 16),
		exit: make(chan struct{}),
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	s.ln = ln

	go s.makeConnection()
	go s.loop()
	if fakeTest {
		go s.fakeTest()
	}

	return s, nil
}

func (s *Server) fakeTest() {
	for range time.Tick(1 * time.Second) {
		s.clients.Range(func(k, v interface{}) bool {
			id := k.(uint32)
			client := v.(*Client)
			const content = `{"type":"scanRes", "result":"0", "scanData":"xxxxx"}`
			err := util.WriteTLV(client, util.TLV{T: 4, L: uint64(len(content)), V: []byte(content)})
			if err != nil {
				log.Printf("[server]: fakeTest send to client %d failed: %v\n", id, err)
			}
			return true
		})
	}
}

func (s *Server) Close() {
	close(s.exit)

	if s.ln != nil {
		s.ln.Close()
	}

	s.clients.Range(func(_, v interface{}) bool {
		client := v.(*Client)
		client.Close()
		return true
	})
}

func (s *Server) makeConnection() {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			log.Printf("[server]: accept failed with %s\n", err)
			return
		}

		conn, err := CreateConnection(c)
		if err != nil {
			log.Printf("[server]: create connection failed with %s, close it\n", err)
			conn.Close()
			continue
		}

		s.connMu.Lock()
		// close previous connection if any
		if s.conn != nil {
			log.Printf("[server]: a new connection accepted, cleanup previous old one\n")
			s.conn.Close()
		}
		log.Printf("[server]: a new connection establish\n")
		s.conn = conn
		go s.pollConnection()
		s.connMu.Unlock()
	}
}

func (s *Server) pollConnection() {
	for {
		s.connMu.RLock()
		conn := s.conn
		s.connMu.RUnlock()

		tlv, err := util.ReadTLV(conn)
		if err == util.InternalErr {
			log.Println("[server]: internal error happend when reading from connection, try again")
			continue
		}
		if err != nil {
			log.Printf("[server]: read from connection failed with [%s], exit polling\n", err)
			return
		}

		Log("[server]: get %v from connection\n", tlv)
		id := uint32(tlv.T >> 32)
		v, ok := s.clients.Load(id)
		if !ok {
			log.Printf("[server]: client %d doesn't exist, skip forwarding %v to client\n", id, tlv)
			continue
		}

		// clear high 32 bits
		t := tlv.T & 0x00000000ffffffff
		if !Type(t).IsValid() {
			log.Printf("[server]: type[%d] is invalid, skip forwarding %v to client\n", t, tlv)
			continue
		}

		tlv.T = t
		err = util.WriteTLV(v.(*Client), tlv)
		if err != nil {
			log.Printf("[server]: forwarding to client %d failed with [%s]\n", id, err)
			continue
		}
	}
}

var dataInvalidErr = errors.New("data invalid")

func (s *Server) loop() {
	for {
		select {
		case cmd := <-s.cmds:
			switch cmd.typ {
			case addClient:
				cmd.err <- s.addClient(cmd.data)
			default:
				log.Printf("[server]: unknown cmd type[%d]\n", cmd.typ)
			}
		case <-s.exit:
			return
		}
	}
}

func (s *Server) AddClient(client *Client) error {
	cmd := &cmd{
		typ:  addClient,
		err:  make(chan error),
		data: client,
	}
	s.cmds <- cmd
	return <-cmd.err
}

func (s *Server) addClient(data interface{}) error {
	client, ok := data.(*Client)
	if !ok {
		return dataInvalidErr
	}

	id := client.Id()
	if _, exist := s.clients.LoadOrStore(id, client); exist {
		return fmt.Errorf("client id[%d] already exist", id)
	}
	go s.pollClient(client)
	return nil
}

func (s *Server) pollClient(client *Client) {
	id := client.Id()
	Log("[server]: add a new client %d\n", id)

	defer func() {
		log.Printf("[server]: client %d exit\n", id)
		client.Close()
		s.clients.Delete(id)
	}()

	returnErr := func(kind Type) {
		err := util.WriteTLV(client, util.TLV{T: uint64(kind)})
		if err != nil {
			log.Printf("[server]: write error type[%#x] failed with %v\n", kind, err)
		}
	}

	for {
		tlv, err := util.ReadTLV(client)
		if err == util.InternalErr {
			log.Println("[server]: internal error happend when reading from client, try again")
			returnErr(ErrorInternal)
			continue
		}
		if err != nil {
			log.Printf("[server]: read from client %d failed with [%s]\n", id, err)
			return
		}
		Log("[server]: get %v from client %d\n", tlv, id)

		if !Type(tlv.T).IsValid() {
			log.Printf("[server]: type[%d] is invalid, skip forwarding %v to connection\n", tlv.T, tlv)
			returnErr(ErrorInvalidType)
			continue
		}

		s.connMu.RLock()
		conn := s.conn
		s.connMu.RUnlock()
		if conn == nil {
			log.Printf("[server]: connection doesn't establish, skip forwarding %v to connection\n", tlv)
			returnErr(ErrorConnectionGone)
			continue
		}

		tlv.T |= uint64(id) << 32
		err = conn.WriteTLV(tlv)
		if err != nil {
			log.Printf("[server]: write %v to connection failed with [%s]\n", tlv, err)
			returnErr(ErrorSend)
			continue
		}
	}
}
