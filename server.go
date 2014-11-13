package store

import (
	"net"
	"net/rpc"
	"sync"
)

type Server struct {
	*rpc.Server
	ln        net.Listener
	filePath  string
	addr      string
	store     *Store
	stop      chan struct{}
	closeOnce sync.Once
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
		addr:     addr,
		store:    store,
		stop:     make(chan struct{}),
	}
	server.Register(store)
	go server.saver()
	go server.accept()
	return server, nil
}

func (s *Server) SetErrorHandler(fn func(error)) {
	s.store.handleError = fn
}

func (s *Server) saver() {
}

func (s *Server) Save() {
	s.store.save()
}

func (s *Server) accept() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			break
		}
		go s.ServeConn(conn)
	}
	s.Close()
}

func (s *Server) Close() {
	s.ln.Close()
	s.closeOnce.Do(func() {
		close(s.stop)
	})
}
