package gobchest

import (
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Server struct {
	*rpc.Server
	ln        net.Listener
	filePath  string
	addr      string
	chest     *Chest
	stop      chan struct{}
	closeOnce sync.Once
}

func NewServer(addr string, filePath string) (*Server, error) {
	chest, err := NewChest(filePath)
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
		chest:    chest,
		stop:     make(chan struct{}),
	}
	server.Register(chest)
	go server.saver()
	go server.accept()
	return server, nil
}

func (s *Server) SetErrorHandler(fn func(error)) {
	s.chest.handleError = fn
}

func (s *Server) saver() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-s.stop:
			s.chest.save()
			return
		case <-s.chest.sigSave:
			if time.Now().Sub(s.chest.saveTime) > time.Second {
				s.chest.save()
			}
		case <-ticker.C:
			if s.chest.dirty {
				s.chest.save()
			}
		}
	}
}

func (s *Server) Save() {
	s.chest.save()
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
