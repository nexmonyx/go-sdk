package nexmonyx

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

// BenchmarkClientCreation benchmarks client initialization
func BenchmarkClientCreation(b *testing.B) {
	b.Run("WithToken", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewClient(&Config{
				BaseURL: "https://api.example.com",
				Auth: AuthConfig{
					Token: "test-token",
				},
			})
		}
	})

	b.Run("WithAPIKey", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewClient(&Config{
				BaseURL: "https://api.example.com",
				Auth: AuthConfig{
					APIKey:    "key",
					APISecret: "secret",
				},
			})
		}
	})

	b.Run("WithServerCredentials", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = NewClient(&Config{
				BaseURL: "https://api.example.com",
				Auth: AuthConfig{
					ServerUUID:   "uuid",
					ServerSecret: "secret",
				},
			})
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = NewClient(&Config{
					BaseURL: "https://api.example.com",
					Auth: AuthConfig{
						Token: "test-token",
					},
				})
			}
		})
	})
}

// BenchmarkClientAuthMethods benchmarks client auth method changes
func BenchmarkClientAuthMethods(b *testing.B) {
	client, _ := NewClient(&Config{
		BaseURL: "https://api.example.com",
		Auth: AuthConfig{
			Token: "test-token",
		},
	})

	b.Run("WithToken", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.WithToken("new-token")
		}
	})

	b.Run("WithAPIKey", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.WithAPIKey("key", "secret")
		}
	})

	b.Run("WithServerCredentials", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.WithServerCredentials("uuid", "secret")
		}
	})
}

// BenchmarkJSONSerialization benchmarks JSON operations
func BenchmarkJSONSerialization(b *testing.B) {
	org := &Organization{
		UUID:     "org-123",
		Name:     "Test Organization",
		Industry: "Technology",
	}

	b.Run("OrganizationMarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(org)
		}
	})

	orgJSON, _ := json.Marshal(org)

	b.Run("OrganizationUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var o Organization
			_ = json.Unmarshal(orgJSON, &o)
		}
	})

	// Benchmark large payload
	largeServers := make([]*Server, 100)
	for i := 0; i < 100; i++ {
		largeServers[i] = &Server{
			ServerUUID: fmt.Sprintf("server-%d", i),
			Hostname:   fmt.Sprintf("host-%d.example.com", i),
		}
	}

	b.Run("LargePayloadMarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(largeServers)
		}
	})

	largeJSON, _ := json.Marshal(largeServers)

	b.Run("LargePayloadUnmarshal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var servers []*Server
			_ = json.Unmarshal(largeJSON, &servers)
		}
	})
}

// BenchmarkModelAllocation benchmarks model creation
func BenchmarkModelAllocation(b *testing.B) {
	b.Run("ServerAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Server{
				ServerUUID: "server-123",
				Hostname:   "test-server",
			}
		}
	})

	b.Run("OrganizationAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Organization{
				UUID: "org-123",
				Name: "Test Org",
			}
		}
	})

	b.Run("UserAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &User{
				Email:     "user@example.com",
				FirstName: "Test",
			}
		}
	})

	b.Run("AlertAllocation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &Alert{
				Name:     "High CPU",
				Severity: "critical",
			}
		}
	})
}

// BenchmarkConcurrentOperations benchmarks parallel execution
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("ConcurrentClientCreation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = NewClient(&Config{
					BaseURL: "https://api.example.com",
					Auth: AuthConfig{
						Token: "test-token",
					},
				})
			}
		})
	})

	b.Run("ConcurrentModelMarshal", func(b *testing.B) {
		org := &Organization{
			UUID: "org-123",
			Name: "Test Org",
		}

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = json.Marshal(org)
			}
		})
	})

	b.Run("ConcurrentModelUnmarshal", func(b *testing.B) {
		orgJSON := []byte(`{"uuid":"org-123","name":"Test Org"}`)

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				var org Organization
				_ = json.Unmarshal(orgJSON, &org)
			}
		})
	})

	b.Run("ConcurrentAuthChange", func(b *testing.B) {
		client, _ := NewClient(&Config{
			BaseURL: "https://api.example.com",
			Auth: AuthConfig{
				Token: "test-token",
			},
		})

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = client.WithToken("new-token")
			}
		})
	})
}

// BenchmarkConcurrentMemory benchmarks memory behavior under concurrent load
func BenchmarkConcurrentMemory(b *testing.B) {
	b.Run("100Goroutines", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 100; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/100; i++ {
					_ = &Server{
						ServerUUID: "server-123",
						Hostname:   "test-server",
					}
				}
			}()
		}
		wg.Wait()
	})

	b.Run("1000Goroutines", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 1000; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/1000; i++ {
					_ = &Organization{
						UUID: "org-123",
						Name: "Test Org",
					}
				}
			}()
		}
		wg.Wait()
	})
}

// BenchmarkSynchronizationPrimitives benchmarks lock and sync operations
func BenchmarkSynchronizationPrimitives(b *testing.B) {
	b.Run("MutexLocking", func(b *testing.B) {
		var mu sync.Mutex
		counter := 0

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		})
	})

	b.Run("RWMutexRead", func(b *testing.B) {
		var mu sync.RWMutex
		data := make(map[string]string)
		data["key"] = "value"

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.RLock()
				_ = data["key"]
				mu.RUnlock()
			}
		})
	})

	b.Run("ChannelSending", func(b *testing.B) {
		ch := make(chan int, 1000)
		defer close(ch)

		go func() {
			for range ch {
			}
		}()

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ch <- 1
			}
		})
	})
}

// BenchmarkConcurrentDataStructures benchmarks concurrent data structure access
func BenchmarkConcurrentDataStructures(b *testing.B) {
	b.Run("SliceAppend", func(b *testing.B) {
		var mu sync.Mutex
		servers := make([]*Server, 0)

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				servers = append(servers, &Server{
					ServerUUID: "server-123",
					Hostname:   "test-server",
				})
				mu.Unlock()
			}
		})
	})

	b.Run("MapAccess", func(b *testing.B) {
		var mu sync.RWMutex
		cache := make(map[string]*Organization)
		cache["test"] = &Organization{UUID: "org-123"}

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.RLock()
				_ = cache["test"]
				mu.RUnlock()
			}
		})
	})

	b.Run("MapInsertion", func(b *testing.B) {
		var mu sync.Mutex
		cache := make(map[string]*Organization)

		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				mu.Lock()
				cache[fmt.Sprintf("key-%d", i)] = &Organization{
					UUID: "org-123",
					Name: "Test Org",
				}
				mu.Unlock()
				i++
			}
		})
	})
}

// BenchmarkRealisticLoadPatterns benchmarks realistic usage patterns
func BenchmarkRealisticLoadPatterns(b *testing.B) {
	b.Run("LightLoad10Concurrent", func(b *testing.B) {
		client, _ := NewClient(&Config{
			BaseURL: "https://api.example.com",
			Auth: AuthConfig{
				Token: "test-token",
			},
		})

		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 10; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/10; i++ {
					_ = client.WithToken("token-123")
				}
			}()
		}
		wg.Wait()
	})

	b.Run("MediumLoad100Concurrent", func(b *testing.B) {
		client, _ := NewClient(&Config{
			BaseURL: "https://api.example.com",
			Auth: AuthConfig{
				Token: "test-token",
			},
		})

		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 100; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/100; i++ {
					_ = client.WithToken("token-123")
				}
			}()
		}
		wg.Wait()
	})

	b.Run("HeavyLoad1000Concurrent", func(b *testing.B) {
		client, _ := NewClient(&Config{
			BaseURL: "https://api.example.com",
			Auth: AuthConfig{
				Token: "test-token",
			},
		})

		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 1000; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/1000; i++ {
					_ = client.WithToken("token-123")
				}
			}()
		}
		wg.Wait()
	})

	b.Run("MixedOperations50Concurrent", func(b *testing.B) {
		client, _ := NewClient(&Config{
			BaseURL: "https://api.example.com",
			Auth: AuthConfig{
				Token: "test-token",
			},
		})

		org := &Organization{
			UUID: "org-123",
			Name: "Test Org",
		}

		b.ReportAllocs()
		b.ResetTimer()

		wg := sync.WaitGroup{}
		for g := 0; g < 50; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/50; i++ {
					switch i % 3 {
					case 0:
						_ = client.WithToken("token-123")
					case 1:
						_, _ = json.Marshal(org)
					default:
						var o Organization
						_ = json.Unmarshal([]byte(`{"uuid":"org-123","name":"Test"}`), &o)
					}
				}
			}()
		}
		wg.Wait()
	})
}

// BenchmarkResourceCleanup benchmarks cleanup operations
func BenchmarkResourceCleanup(b *testing.B) {
	b.Run("ClientCreationAndCleanup", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			clients := make([]*Client, 0, 100)
			wg := sync.WaitGroup{}

			for i := 0; i < 100; i++ {
				c, _ := NewClient(&Config{
					BaseURL: "https://api.example.com",
					Auth: AuthConfig{
						Token: "test-token",
					},
				})
				clients = append(clients, c)
				wg.Add(1)
				go func(client *Client) {
					defer wg.Done()
					_ = client.WithToken("new-token")
				}(c)
			}

			wg.Wait()
			clients = nil
		}
	})

	b.Run("WaitGroupCompletion", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for iteration := 0; iteration < b.N; iteration++ {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(1 * time.Microsecond)
				}()
			}
			wg.Wait()
		}
	})
}
