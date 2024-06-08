package whoip

import (
	"encoding/gob"
	"net"
	"os"
	"path/filepath"
	"time"

	utils "github.com/aorith/whoip/internal"
)

// ProviderData holds the data for a specific provider.
type ProviderData struct {
	URL             string
	Name            string
	Description     string
	DataFilename    string
	RefreshInterval time.Duration
	LastUpdate      time.Time
	IPRanges        ProviderIPRanges
	Fetcher         func(*ProviderData) error
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

func saveToGob(filename string, data ProviderData) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(data)
}

func loadFromGob(filename string) (ProviderData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return ProviderData{}, err
	}
	defer file.Close()

	var data ProviderData
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)
	return data, err
}

var providers = map[string]*ProviderData{
	"aws": {
		URL:             "https://ip-ranges.amazonaws.com/ip-ranges.json",
		Name:            "Amazon AWS",
		Description:     "Amazon AWS IPRanges",
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "aws.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchAWSData,
	},
	"google": {
		URL:             "https://www.gstatic.com/ipranges/cloud.json",
		Name:            "Google Cloud",
		Description:     "Google Cloud IPRanges",
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "google.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchGoogleData,
	},
}
