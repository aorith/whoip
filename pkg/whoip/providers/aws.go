package whoip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// fetchAWSData fetches the AWS data and updates the ProviderData struct.
func fetchAWSData(provider *ProviderData) error {
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
		SyncToken  string `json:"syncToken"`
		CreateDate string `json:"createDate"`
		Prefixes   []struct {
			IPPrefix           string `json:"ip_prefix"`
			Region             string `json:"region"`
			Service            string `json:"service"`
			NetworkBorderGroup string `json:"network_border_group"`
		} `json:"prefixes"`
	}

	err = json.NewDecoder(resp.Body).Decode(&fetchedData)
	if err != nil {
		return fmt.Errorf("failed to decode json: %v", err)
	}

	var prefixes []Prefix
	for _, p := range fetchedData.Prefixes {
		_, network, err := net.ParseCIDR(p.IPPrefix)
		if err != nil {
			continue
		}
		prefixes = append(prefixes, Prefix{
			Network: *network,
			Details: map[string]string{
				"Region":             p.Region,
				"Service":            p.Service,
				"NetworkBorderGroup": p.NetworkBorderGroup,
			},
		})
	}

	provider.IPRanges.Prefixes = prefixes
	provider.LastUpdate = time.Now()
	return nil
}
