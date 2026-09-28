package redisclient

import "testing"

func TestUsernameReachesGoRedis(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = "127.0.0.1:1"
	cfg.Username = "t_x"
	cfg.Password = "p"
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if got := client.Client().Options().Username; got != "t_x" {
		t.Errorf("username = %q, want t_x", got)
	}
	if got := client.Client().Options().Password; got != "p" {
		t.Errorf("password = %q, want p", got)
	}
}

func TestUsernameEmptyByDefault(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Username != "" {
		t.Fatalf("default username = %q, want empty", cfg.Username)
	}
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if got := client.Client().Options().Username; got != "" {
		t.Errorf("go-redis username = %q, want empty", got)
	}
}

func TestWithUsername(t *testing.T) {
	client, err := NewWithOptions(WithUsername("t_x"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if got := client.Client().Options().Username; got != "t_x" {
		t.Errorf("username = %q, want t_x", got)
	}
}
