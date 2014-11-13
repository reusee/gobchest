package store

import (
	"net"
	"net/rpc"
)

type Server struct {
	*rpc.Server
	ln       net.Listener
	filePath string
	store    *Store
}

func NewServer(addr string, filePath string) (*Server, error) {
	store, err := NewStore(filePath)
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		Server:   rpc.NewServer(),
		ln:       ln,
		filePath: filePath,
		store:    store,
	}
	server.Register(store)
	go server.start()
	return server, nil
}

func (s *Server) start() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			break
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) Close() {
	s.ln.Close()
}
