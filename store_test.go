package store

import "testing"

var testAddr = "localhost:2888"

func setupTestServer(t *testing.T) *Server {
	server, err := NewServer(testAddr)
	if err != nil {
		t.Fatal(err)
	}
	return server
}

func setupTestClient(t *testing.T) *Client {
	client, err := NewClient(testAddr)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestNew(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t)
	defer client.Close()
}

func TestBadAddr(t *testing.T) {
	_, err := NewClient("nohost:999")
	if err == nil {
		t.Fatal("should fail")
	}

	_, err = NewServer("nohost:999")
	if err == nil {
		t.Fatal("should fail")
	}
}
