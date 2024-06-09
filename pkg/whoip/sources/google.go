package sources

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var googleMu sync.Mutex

// fetchGoogleData fetches the Google data and updates the MetaData.
func fetchGoogleData(src *IPSource) error {
	googleMu.Lock()
	defer googleMu.Unlock()

	if time.Since(src.MetaData.LastUpdate) < src.RefreshInterval {
		return nil // Data is up to date
	}

	if err := src.load(); err != nil {
		log.Printf("Failed to load saved data: '%s'.", err)
	} else {
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

	src.MetaData.Prefixes = prefixes
	src.MetaData.LastUpdate = time.Now()
	src.mustSave()

	return nil
}
