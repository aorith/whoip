package sources

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// fetchBingBotData fetches the Google data and updates the MetaData.
func fetchBingBotData(src *IPSource) error {
	src.Mu.Lock()
	defer src.Mu.Unlock()

	if time.Since(src.MetaData.LastUpdate) < src.RefreshInterval {
		return nil // Data is up to date
	}

	if src.load() {
		// Data is still valid
		return nil
	}

	resp, err := http.Get(src.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	var fetchedData struct {
		Prefixes []struct {
			IPv4Prefix string `json:"ipv4Prefix,omitempty"`
			IPv6Prefix string `json:"ipv6Prefix,omitempty"`
		} `json:"prefixes"`
	}

	err = json.NewDecoder(resp.Body).Decode(&fetchedData)
	if err != nil {
		return fmt.Errorf("failed to decode json: %v", err)
	}

	var prefixes []Prefix
	for _, p := range fetchedData.Prefixes {
		var network *net.IPNet
		_, network, err = net.ParseCIDR(p.IPv4Prefix)
		if err != nil {
			_, network, err = net.ParseCIDR(p.IPv6Prefix)
			if err != nil {
				continue
			}
		}
		prefixes = append(prefixes, Prefix{
			Network: *network,
		})
	}

	src.MetaData.Prefixes = prefixes
	src.MetaData.LastUpdate = time.Now()
	src.mustSave()

	return nil
}
