package whoip

import (
	"net"
	"sync"
	"time"
)

// ProviderData holds the data for a specific provider.
type ProviderData struct {
	URL         string
	Name        string
	Description string
	LastUpdate  time.Time
	IPRanges    ProviderIPRanges
	Mu          sync.Mutex
	Fetcher     func(*ProviderData) error
}

// ProviderIPRanges holds the IP ranges for a provider.
type ProviderIPRanges struct {
	Prefixes []Prefix
}

// Prefix holds the network and details for a prefix.
type Prefix struct {
	Network net.IPNet
	Details map[string]string
}

var providers = map[string]*ProviderData{
	"aws": {
		URL:         "https://ip-ranges.amazonaws.com/ip-ranges.json",
		Name:        "Amazon AWS",
		Description: "Amazon AWS IPRanges",
		Fetcher:     fetchAWSData,
	},
	"google": {
		URL:         "https://www.gstatic.com/ipranges/cloud.json",
		Name:        "Google Cloud",
		Description: "Google Cloud IPRanges",
		Fetcher:     fetchGoogleData,
	},
}
