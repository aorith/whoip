package sources

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// fetchAWSData fetches the AWS data and updates the MetaData.
func fetchAWSData(src *IPSource) error {
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

	src.MetaData.Prefixes = prefixes
	src.MetaData.LastUpdate = time.Now()
	src.mustSave()

	return nil
}
