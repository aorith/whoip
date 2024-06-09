package whoip

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/aorith/whoip/pkg/sources"
)

type WhoIPInfo struct {
	URL         string             `json:"url"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Categories  []sources.Category `json:"categories"`
	Prefix      Prefix             `json:"prefix"`
}

type Prefix struct {
	Network string            `json:"network"`
	Details map[string]string `json:"details"`
}

func FindIP(ip net.IP) string {
	UpdateSources()

	var info []WhoIPInfo
	for _, src := range sources.IPRangeSources {
		prefix := src.ContainsIP(ip)
		if prefix != nil {
			newInfo := WhoIPInfo{
				URL:         src.URL,
				Name:        src.Name,
				Description: src.Description,
				Prefix:      Prefix{Network: prefix.Network.String(), Details: prefix.Details},
			}
			if len(prefix.Categories) > 0 {
				newInfo.Categories = prefix.Categories
			} else {
				newInfo.Categories = src.Categories
			}
			info = append(info, newInfo)
		}
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	return fmt.Sprintf("%s\n", jsonData)
}

func UpdateSources() {
	var wg sync.WaitGroup
	numSources := len(sources.IPRangeSources)
	wg.Add(numSources)

	for _, src := range sources.IPRangeSources {
		go func(src *sources.IPSource) {
			defer wg.Done()
			err := src.Fetcher(src)
			if err != nil {
				log.Printf("Failure updating source '%s'.", src.Name)
			}
		}(src)
	}

	wg.Wait()
}

func Categories() string {
	var categories []sources.Category
	for _, cat := range sources.Categories {
		categories = append(categories, cat)
	}

	jsonData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	return fmt.Sprintf("%s\n", jsonData)
}
