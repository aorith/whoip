package whoip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// fetchGoogleData fetches the Google data and updates the ProviderData struct.
func fetchGoogleData(provider *ProviderData) error {
	provider.Mu.Lock()
	defer provider.Mu.Unlock()

	if provider.LastUpdate.After(time.Now().Add(-24 * time.Hour)) {
		return nil // Data is up to date
	}

	resp, err := http.Get(provider.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	var fetchedData struct {
		SyncToken    string `json:"syncToken"`
		CreationTime string `json:"creationTime"`
		Prefixes     []struct {
			IPv4Prefix string `json:"ipv4Prefix"`
			Service    string `json:"service"`
			Scope      string `json:"scope"`
		} `json:"prefixes"`
	}

	err = json.NewDecoder(resp.Body).Decode(&fetchedData)
	if err != nil {
		return fmt.Errorf("failed to decode json: %v", err)
	}

	var prefixes []Prefix
	for _, p := range fetchedData.Prefixes {
		_, network, err := net.ParseCIDR(p.IPv4Prefix)
		if err != nil {
			continue
		}
		prefixes = append(prefixes, Prefix{
			Network: *network,
			Details: map[string]string{
				"Service": p.Service,
				"Scope":   p.Scope,
			},
		})
	}

	provider.IPRanges.Prefixes = prefixes
	provider.LastUpdate = time.Now()
	return nil
}
