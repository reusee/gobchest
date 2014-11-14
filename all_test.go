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
	time.Sleep(time.Second * 2)
}

func TestMultipleClient(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	for i := 0; i < 16; i++ {
		client := setupTestClient(t, addr)
		defer client.Close()
		go func() {
			for {
				client.Set("foo", "foo")
				client.Get("foo")
			}
		}()
	}
	time.Sleep(time.Second * 1)
}

func TestListAppend(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()

	err := client.ListAppend("foo", 1, 2, 3, 4, 5)
	if err != nil {
		t.Fatal(err)
	}
	v, err := client.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := v.([]int); !ok {
		t.Fatal("type not match")
	} else {
		for i := 1; i <= 5; i++ {
			if v[i-1] != i {
				t.Fatal("list not match")
			}
		}
	}

	err = client.ListAppend("foo", 6, 7, 8, 9, 10)
	if err != nil {
		t.Fatal(err)
	}
	v, err = client.Get("foo")
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := v.([]int); !ok {
		t.Fatal("type not match")
	} else {
		for i := 1; i <= 10; i++ {
			if v[i-1] != i {
				t.Fatal("list not match")
			}
		}
	}
}

func TestSetAdd(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()

	err := client.SetAdd("foo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := server.chest.Data["foo"]; !ok {
		t.Fatalf("set not set")
	} else {
		if _, ok := v.(map[int]struct{})[1]; !ok {
			t.Fatalf("key not set")
		}
	}

	err = client.SetAdd("foo", 42)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := server.chest.Data["foo"]; !ok {
		t.Fatalf("set not set")
	} else {
		if _, ok := v.(map[int]struct{})[42]; !ok {
			t.Fatalf("key not set")
		}
	}
}

func TestSetExists(t *testing.T) {
	server, addr := setupTestServer(t)
	defer server.Close()
	client := setupTestClient(t, addr)
	defer client.Close()

	err := client.SetAdd("foo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !client.SetExists("foo", 1) {
		t.Fatal("should be true")
	}
	if client.SetExists("foo", 42) {
		t.Fatal("should be false")
	}
	if client.SetExists("bar", 42) {
		t.Fatal("should be false")
	}
	client.SetAdd("bar", 42)
	if !client.SetExists("bar", 42) {
		t.Fatal("should be true")
	}
}

func TestRegister(t *testing.T) {
	Register(new(struct{}))
}
