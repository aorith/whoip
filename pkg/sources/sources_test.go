package sources

import (
	"net"
	"sync"
	"testing"
	"time"
)

var testSources = map[string]*IPSource{
	"fake": {
		URL:             "https://www.example.com/ranges.json",
		Name:            "Fake Source",
		Description:     "A fake source for testing purposes",
		Categories:      []Category{Categories["datacenter"]},
		RefreshInterval: 1 * time.Minute,
		Fetcher:         fetchFakeData,
	},
}

// fetchFakeData fetches fake data for testing purposes.
func fetchFakeData(src *IPSource) error {
	src.Mu.Lock()
	defer src.Mu.Unlock()

	if time.Since(src.MetaData.LastUpdate) < src.RefreshInterval {
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

	src.MetaData.Prefixes = fakePrefixes
	src.MetaData.LastUpdate = time.Now()
	return nil
}

func TestNoDuplicateDataFilename(t *testing.T) {
	seen := make(map[string]bool)
	for _, src := range IPRangeSources {
		if seen[src.DataFilename] {
			t.Errorf("Duplicate data filename: '%s' on the source '%s'.", src.DataFilename, src.Name)
		}
		seen[src.DataFilename] = true
	}
}

func TestNoDuplicateURL(t *testing.T) {
	seen := make(map[string]bool)
	for _, src := range IPRangeSources {
		if seen[src.URL] {
			t.Errorf("Duplicate URL: '%s' on the source '%s'.", src.URL, src.Name)
		}
		seen[src.URL] = true
	}
}

func TestTypoInCategories(t *testing.T) {
	for _, src := range IPRangeSources {
		for _, cat := range src.Categories {
			if cat.ID == "" {
				t.Errorf("Invalid category on source '%s'.", src.Name)
			}
		}
	}
}

func TestFetchSourceDataConcurrency(t *testing.T) {
	for name, source := range IPRangeSources {
		t.Run(name, func(t *testing.T) {
			var wg sync.WaitGroup
			concurrentFetches := 5
			wg.Add(concurrentFetches)

			for i := 0; i < concurrentFetches; i++ {
				go func() {
					defer wg.Done()
					err := source.Fetcher(source)
					if err != nil {
						t.Errorf("Failed to fetch data: %v", err)
					}
					t.Logf("OK: %s LastUpdate: %s ago.", source.Name, (time.Since(source.MetaData.LastUpdate)))
				}()
				time.Sleep(time.Millisecond * 25)
			}

			wg.Wait()

			if len(source.MetaData.Prefixes) == 0 {
				t.Errorf("No IP ranges fetched for source %s", name)
			}
		})
	}
}

func TestFetchFakeSourceData(t *testing.T) {
	source := testSources["fake"]

	err := source.Fetcher(source)
	if err != nil {
		t.Errorf("Failed to fetch data for fake source: %v", err)
	}

	if len(source.MetaData.Prefixes) == 0 {
		t.Errorf("No IP ranges fetched for fake source")
	}

	if source.MetaData.LastUpdate.Before(time.Now().Add(-24 * time.Hour)) {
		t.Errorf("LastUpdate not updated for fake source")
	}

	ip := net.ParseIP("192.168.1.5")
	if prefix := source.ContainsIP(ip); prefix != nil {
		expected := "FakeService1"
		if prefix.Details["Service"] != expected {
			t.Errorf("Wrong prefix values. Expected output:\n%s\n\nGot:\n%s\n", expected, prefix.Details["Service"])
		}
	} else {
		t.Errorf("IP not found in prefixes: '%s'", ip)
	}
}
