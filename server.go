package store

import (
	"net"
	"net/rpc"
)

type Server struct {
	*rpc.Server
	ln net.Listener
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		Server: rpc.NewServer(),
		ln:     ln,
	}
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
