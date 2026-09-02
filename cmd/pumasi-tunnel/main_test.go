package main

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestNewTLSDialerUsesRelayHostAndSafeDefaults(t *testing.T) {
	d, err := newTLSDialer("pumasi.link:7001", "")
	if err != nil {
		t.Fatal(err)
	}
	if d.Config.ServerName != "pumasi.link" {
		t.Errorf("server name = %q", d.Config.ServerName)
	}
	if d.Config.MinVersion != tls.VersionTLS12 {
		t.Errorf("minimum TLS version = %x", d.Config.MinVersion)
	}
	if d.NetDialer.Timeout != 10*time.Second {
		t.Errorf("dial timeout = %s", d.NetDialer.Timeout)
	}
}

func TestNewTLSDialerHonorsNameOverride(t *testing.T) {
	d, err := newTLSDialer("127.0.0.1:7001", "relay.example")
	if err != nil {
		t.Fatal(err)
	}
	if d.Config.ServerName != "relay.example" {
		t.Errorf("server name = %q", d.Config.ServerName)
	}
}

func TestNewTLSDialerRejectsAddressWithoutPort(t *testing.T) {
	if _, err := newTLSDialer("pumasi.link", ""); err == nil {
		t.Fatal("expected host:port validation error")
	}
}
