package envutil_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	envutil "github.com/isimtekin/go-packages/env-util"
)

func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_STRING", "benchmark_value")
	defer os.Unsetenv("BENCH_STRING")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetEnv("BENCH_STRING", "default")
	}
}

func BenchmarkGetEnvInt(b *testing.B) {
	os.Setenv("BENCH_INT", "12345")
	defer os.Unsetenv("BENCH_INT")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetEnvInt("BENCH_INT", 0)
	}
}

func BenchmarkGetEnvBool(b *testing.B) {
	os.Setenv("BENCH_BOOL", "true")
	defer os.Unsetenv("BENCH_BOOL")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetEnvBool("BENCH_BOOL", false)
	}
}

func BenchmarkGetEnvDuration(b *testing.B) {
	os.Setenv("BENCH_DURATION", "30s")
	defer os.Unsetenv("BENCH_DURATION")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetEnvDuration("BENCH_DURATION", time.Second)
	}
}

func BenchmarkGetEnvStringSlice(b *testing.B) {
	os.Setenv("BENCH_SLICE", "a,b,c,d,e")
	defer os.Unsetenv("BENCH_SLICE")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetEnvStringSlice("BENCH_SLICE", nil)
	}
}

func BenchmarkClientGetString(b *testing.B) {
	client := envutil.NewDefault()
	os.Setenv("BENCH_CLIENT_STRING", "value")
	defer os.Unsetenv("BENCH_CLIENT_STRING")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetString("BENCH_CLIENT_STRING", "default")
	}
}

func BenchmarkClientGetStringWithCache(b *testing.B) {
	client := envutil.NewDefault()
	os.Setenv("BENCH_CACHE_STRING", "cached_value")
	defer os.Unsetenv("BENCH_CACHE_STRING")
	
	// Prime the cache
	_ = client.GetString("BENCH_CACHE_STRING", "default")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetString("BENCH_CACHE_STRING", "default")
	}
}

func BenchmarkClientWithPrefix(b *testing.B) {
	client := envutil.NewWithOptions(
		envutil.WithPrefix("APP_"),
	)
	os.Setenv("APP_BENCH_PREFIX", "prefixed_value")
	defer os.Unsetenv("APP_BENCH_PREFIX")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.GetString("BENCH_PREFIX", "default")
	}
}

func BenchmarkIsEnvSet(b *testing.B) {
	os.Setenv("BENCH_SET", "value")
	defer os.Unsetenv("BENCH_SET")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.IsEnvSet("BENCH_SET")
	}
}

func BenchmarkGetAllEnvWithPrefix(b *testing.B) {
	// Set up multiple env vars with prefix
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("BENCH_PREFIX_%d", i)
		os.Setenv(key, "value")
		defer os.Unsetenv(key)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = envutil.GetAllEnvWithPrefix("BENCH_PREFIX_")
	}
}