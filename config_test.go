package goku

import (
	"os"
	"strings"
	"testing"
)

func Test_NewConfiguration(t *testing.T) {
	c := NewConfiguration()

	if c.Debug {
		t.Error("Expected debug mode to be false")
	}

	if c.RPC != ":5127" {
		t.Errorf("expected :5127, actual: %v", c.RPC)
	}

	if c.HTTP != ":8080" {
		t.Errorf("expected :8080, actual: %v", c.HTTP)
	}

	if !strings.Contains(c.Hostname, ".xip.io") {
		t.Errorf("expected *.xip.io, actual: %s", c.Hostname)
	}

	if c.DockerSock != "unix:///var/run/docker.sock" {
		t.Errorf("expected unix:///var/run/docker.sock actual %v", c.DockerSock)
	}

	if c.GitPath != "./repositories" {
		t.Errorf("expected ./repositories actual %v", c.GitPath)
	}

	cases := map[string]string{
		"type": "debug",
	}

	for k, v := range cases {
		vExpected, ok := c.Backend[k]
		if !ok || v != vExpected {
			t.Errorf("expected %v actual %v", v, vExpected)
		}
	}
}

func TestParseDebugFromEnv(t *testing.T) {
	os.Setenv("DEBUG", "")
	if parseDebugFromEnv() {
		t.Error("should be false")
	}

	os.Setenv("DEBUG", "true")
	if !parseDebugFromEnv() {
		t.Error("should be true")
	}

	os.Setenv("DEBUG", "anything at all")
	if !parseDebugFromEnv() {
		t.Error("should be true")
	}

}
