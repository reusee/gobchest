package gobchest

import "testing"

func BenchmarkSet(b *testing.B) {
	server, err := NewServer(randomAddr(), randomFilePath())
	if err != nil {
		b.Fatal(err)
	}
	defer server.Close()
	client, err := NewClient(server.addr)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Set("foo", i)
	}
}

func BenchmarkGet(b *testing.B) {
	server, err := NewServer(randomAddr(), randomFilePath())
	if err != nil {
		b.Fatal(err)
	}
	defer server.Close()
	client, err := NewClient(server.addr)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()
	client.Set("foo", "foo")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Get("foo")
	}
}
