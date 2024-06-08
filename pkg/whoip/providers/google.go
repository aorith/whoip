package whoip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var googleMu sync.Mutex

// fetchGoogleData fetches the Google data and updates the ProviderData struct.
func fetchGoogleData(provider *ProviderData) error {
	googleMu.Lock()
	defer googleMu.Unlock()

	if time.Since(provider.LastUpdate) < provider.RefreshInterval {
		return nil // Data is up to date
	}

	pd, err := loadFromGob(provider.DataFilename)
	if err == nil && time.Since(pd.LastUpdate) < provider.RefreshInterval {
		provider.URL = pd.URL
		provider.Name = pd.Name
		provider.Description = pd.Description
		provider.LastUpdate = pd.LastUpdate
		provider.IPRanges = pd.IPRanges
		return nil
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

	if err = saveToGob(provider.DataFilename, *provider); err != nil {
		return fmt.Errorf("failed to save data to Gob: %v", err)
	}

	return nil
}
