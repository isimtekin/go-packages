//go:build tenantstores

package redisclient

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestLive_ACLUserIsConfinedToItsWorkspace(t *testing.T) {
	addr := os.Getenv("TENANT_REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("TENANT_REDIS_TEST_ADDR is unset")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	admin := redis.NewClient(&redis.Options{Addr: addr, Username: os.Getenv("TENANT_REDIS_TEST_ADMIN_USER"), Password: os.Getenv("TENANT_REDIS_TEST_ADMIN_PASSWORD")})
	t.Cleanup(func() {
		verifier := redis.NewClient(&redis.Options{Addr: addr, Username: os.Getenv("TENANT_REDIS_TEST_ADMIN_USER"), Password: os.Getenv("TENANT_REDIS_TEST_ADMIN_PASSWORD")})
		defer verifier.Close()
		result, err := verifier.Do(context.Background(), "ACL", "GETUSER", "t_live1").Result()
		if err != redis.Nil && (err != nil || result != nil) {
			t.Errorf("ACL user remained after test cleanup: result=%v err=%v", result, err)
		}
	})
	const user, prefix, password = "t_live1", "t_live1", "tenantpw"
	const allowList = "get set del exists expire incrby decrby hset hget hgetall lpush rpush lrange sadd smembers publish subscribe psubscribe unsubscribe punsubscribe ping auth hello client|setinfo"
	args := []interface{}{"ACL", "SETUSER", user, "on", "resetpass", ">" + password, "resetkeys", "~" + prefix + ":*", "resetchannels", "&" + prefix + ":*", "-@all"}
	for _, command := range strings.Fields(allowList) {
		args = append(args, "+"+command)
	}
	if err := admin.Do(ctx, args...).Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		admin.Del(cleanupCtx, prefix+":k")
		admin.Do(cleanupCtx, "ACL", "DELUSER", user)
		admin.Do(cleanupCtx, "ACL", "SAVE")
		admin.Close()
	})
	if err := admin.Do(ctx, "ACL", "SAVE").Err(); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	cfg.Addr, cfg.Username, cfg.Password, cfg.Workspace = addr, user, password, prefix
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Set(ctx, "k", "v", 0); err != nil {
		t.Fatal(err)
	}
	if got, err := client.Get(ctx, "k"); err != nil || got != "v" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	sub := client.Subscribe(ctx, "room")
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Publish(ctx, "room", "hello"); err != nil {
		t.Fatal(err)
	}
	if message, err := sub.ReceiveMessage(ctx); err != nil || message.Payload != "hello" {
		t.Fatalf("Subscribe received %v, %v", message, err)
	}
	other := *cfg
	other.Workspace = "t_live2"
	forbidden, err := New(&other)
	if err != nil {
		t.Fatal(err)
	}
	defer forbidden.Close()
	if err := forbidden.Set(ctx, "k", "v", 0); err == nil {
		t.Fatal("Set outside the ACL workspace succeeded")
	}
}
