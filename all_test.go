package gobchest

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func randomFilePath() string {
	return filepath.Join(os.TempDir(), "golang-gobchest-test-file-"+fmt.Sprintf("%d", rand.Uint32()))
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

	server, err := NewServer(randomAddr(), randomFilePath())
	if err != nil {
		t.Fatal(err)
	}
	filePath := server.filePath
	server.Save()
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
	server.Save()
	server.Close()
	server, err = NewServer(randomAddr(), filePath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidFile(t *testing.T) {
	filePath := randomFilePath()
	err := ioutil.WriteFile(filePath, []byte("garbage"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewServer(randomAddr(), filePath)
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestSaveFail(t *testing.T) {
	dirPath := randomFilePath()
	err := os.Mkdir(dirPath, 0500)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(dirPath, "foo")
	server, err := NewServer(randomAddr(), filePath)
	if err != nil {
		t.Fatal(err)
	}

	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Fatal("should fail")
			}
		}()
		server.Save()
	}()

	errored := false
	server.SetErrorHandler(func(err error) {
		errored = true
	})
	server.Save()
	if !errored {
		t.Fatal("should fail")
	}
}

func TestInvalidValue(t *testing.T) {
	server, _ := setupTestServer(t)
	gob.Register(new(func()))
	server.chest.Data["foo"] = func() {} // not encodable
	errored := false
	server.SetErrorHandler(func(error) {
		errored = true
	})
	server.Save()
	if !errored {
		t.Fatal("should fail")
	}
}

func TestSetGet(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()

	err := client.Set("foo", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if server.chest.Data["foo"] != "foo" {
		t.Fatal("Set: foo is not foo")
	}

	v, err := client.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "foo" {
		t.Fatal("Get: foo is not foo")
	}

	v, err = client.Get("keynotexists")
	if err == nil {
		t.Fatal("Get: should fail")
	}
}

func TestSave(t *testing.T) {
	server, addr := setupTestServer(t)
	client := setupTestClient(t, addr)
	err := client.Set("foo", "foo")
	if err != nil {
		t.Fatal(err)
	}
	server.Save()
	server.Close()
	server, err = NewServer(randomAddr(), server.filePath)
	if err != nil {
		t.Fatal(err)
	}
	client, err = NewClient(server.addr)
	if err != nil {
		t.Fatal(err)
	}
	v, err := client.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "foo" {
		t.Fatal("foo is not foo")
	}
}

func TestPeriodicSave(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()
	for i := 0; i < 128; i++ {
		go func() {
			for {
				client.Set("foo", "foo")
			}
		}()
	}
	time.Sleep(time.Second * 3)
}
