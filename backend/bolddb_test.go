package store

import (
	"log"
	"os"
	"testing"
)

func TestNewAndClose(t *testing.T) {
	b, err := newBoltBackend(os.TempDir())
	if err != nil {
		t.Error(err)
		return
	}

	if err := b.Close(); err != nil {
		t.Error(err)
	}
}

func TestCrud(t *testing.T) {
	b, err := newBoltBackend(os.TempDir())
	if err != nil {
		t.Error(err)
		return
	}

	log.Println("put test")
	if err := b.Put("testKey1", []byte("test data")); err != nil {
		t.Error(err)
	}

	log.Println("get test")
	data, err := b.Get("testKey1")
	if err != nil {
		t.Error(err)
	}

	if string(data) != "test data" {
		t.Error("expected test data - actual", string(data))
	}

	log.Println("delete test")
	if err := b.Delete("testKey1"); err != nil {
		t.Error(err)
	}

	log.Println("get after delete test")
	if _, err := b.Get("testKey1"); err == nil {
		t.Error(err)
	}

	if err := b.Close(); err != nil {
		t.Error(err)
	}

}
