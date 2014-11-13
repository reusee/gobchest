package store

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func randomFilePath() string {
	return filepath.Join(os.TempDir(), "golang-store-test-file-"+fmt.Sprintf("%d", rand.Uint32()))
}

func randomAddr() string {
	return fmt.Sprintf("localhost:%d", 20000+rand.Intn(10000))
}

func setupTestServer(t *testing.T) (*Server, string) {
	retries := 3
retry:
	addr := randomAddr()
	server, err := NewServer(addr, randomFilePath())
	if err != nil {
		if retries > 0 {
			retries--
			goto retry
		} else {
			t.Fatal(err)
		}
	}
	return server, addr
}

func setupTestClient(t *testing.T, addr string) *Client {
	client, err := NewClient(addr)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestNew(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()
}

func TestBadAddr(t *testing.T) {
	_, err := NewClient("nohost:999")
	if err == nil {
		t.Fatal("should fail")
	}

	_, err = NewServer("nohost:999", randomFilePath())
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestInvalidFilePath(t *testing.T) {
	_, err := NewServer(randomAddr(), os.TempDir())
	if err == nil {
		t.Fatal("should fail")
	}

	_, err = NewServer(randomAddr(), "\000\000\000")
	if err == nil {
		t.Fatal("should fail")
	}

	dirPath := randomFilePath()
	err = os.Mkdir(dirPath, 0500) // read-only dir
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dirPath, "foo")
	_, err = NewServer(randomAddr(), filePath)
	if err == nil {
		t.Fatal("should fail")
	}

	server, err := NewServer(randomAddr(), randomFilePath())
	if err != nil {
		t.Fatal(err)
	}
	filePath = server.filePath
	err = os.Chmod(filePath, 0000) // make it non-readable
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewServer(randomAddr(), filePath)
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestReopen(t *testing.T) {
	server, err := NewServer(randomAddr(), randomFilePath())
	if err != nil {
		t.Fatal(err)
	}
	filePath := server.filePath
	server.Close()
	server, err = NewServer(randomAddr(), filePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSet(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()

	err := client.Set("foo", "foo")
	if err != nil {
		t.Fatal(err)
	}
}
