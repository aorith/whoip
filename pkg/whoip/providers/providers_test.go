package whoip

import (
	"net"
	"sync"
	"testing"
	"time"
)

var testProviders = map[string]*ProviderData{
	"fake": {
		URL:         "https://www.example.com/ranges.json",
		Name:        "Fake Provider",
		Description: "A fake provider for testing purposes",
		Fetcher:     fetchFakeData,
	},
}

var fakeMu sync.Mutex

// fetchFakeData fetches fake data for testing purposes.
func fetchFakeData(provider *ProviderData) error {
	fakeMu.Lock()
	defer fakeMu.Unlock()

	if time.Since(provider.LastUpdate) < 24*time.Hour {
		return nil
	}

	fakePrefixes := []Prefix{
		{
			Network: net.IPNet{
				IP:   net.ParseIP("192.168.1.0"),
				Mask: net.CIDRMask(24, 32),
			},
			Details: map[string]string{
				"Service": "FakeService1",
				"Region":  "us-west-1",
			},
		},
		{
			Network: net.IPNet{
				IP:   net.ParseIP("10.0.0.0"),
				Mask: net.CIDRMask(8, 32),
			},
			Details: map[string]string{
				"Service": "FakeService2",
				"Region":  "us-east-1",
			},
		},
	}

	provider.IPRanges.Prefixes = fakePrefixes
	provider.LastUpdate = time.Now()
	return nil
}

func TestFetchProviderDataConcurrency(t *testing.T) {
	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			var wg sync.WaitGroup
			concurrentFetches := 10
			wg.Add(concurrentFetches)

			for i := 0; i < concurrentFetches; i++ {
				go func() {
					defer wg.Done()
					err := provider.Fetcher(provider)
					if err != nil {
						t.Errorf("Failed to fetch data: %v", err)
					}
					t.Logf("OK: %s LastUpdate: %s ago.", provider.Name, (time.Since(provider.LastUpdate)))
				}()
				time.Sleep(time.Millisecond * 200)
			}

			wg.Wait()

			if len(provider.IPRanges.Prefixes) == 0 {
				t.Fatalf("No IP ranges fetched for provider %s", name)
			}
		})
	}
}

func TestFetchFakeProviderData(t *testing.T) {
	provider := testProviders["fake"]

	err := provider.Fetcher(provider)
	if err != nil {
		t.Fatalf("Failed to fetch data for fake provider: %v", err)
	}

	if len(provider.IPRanges.Prefixes) == 0 {
		t.Fatalf("No IP ranges fetched for fake provider")
	}

	if provider.LastUpdate.Before(time.Now().Add(-24 * time.Hour)) {
		t.Fatalf("LastUpdate not updated for fake provider")
	}

	// TODO: check an IP present in the ranges of the fake provider, i.e: 192.168.1.5
}
