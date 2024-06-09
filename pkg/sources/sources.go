package sources

import (
	"encoding/gob"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	utils "github.com/aorith/whoip/internal"
)

// Category represents the type of a category.
type Category struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Predefined Category map.
var Categories = map[string]Category{
	"crawler":     {"crawler", "IP ranges used by web crawlers and bots"},
	"residential": {"residential", "IP ranges assigned to residential users by ISPs"},
	"business":    {"business", "IP ranges assigned to businesses"},
	"mobile":      {"mobile", "IP ranges used by mobile carriers for their data services"},
	"datacenter":  {"datacenter", "IP ranges belonging to data centers and hosting providers"},
	"education":   {"education", "IP ranges assigned to educational institutions"},
	"government":  {"government", "IP ranges used by government agencies"},
	"healthcare":  {"healthcare", "IP ranges used by healthcare providers"},
	"cdn":         {"cdn", "IP ranges used by content delivery networks"},
	"isp":         {"isp", "IP ranges owned by internet service providers"},
	"vpn":         {"vpn", "IP ranges used by VPN and proxy services"},
	"spam":        {"spam", "IP ranges identified as sources of spam activity"},
	"malicious":   {"malicious", "IP ranges identified as sources of malicious activity"},
	"private":     {"private", "Non-routable IP ranges used for private networks"},
	"iot":         {"iot", "IP ranges used by Internet of Things devices"},
	"telecom":     {"telecom", "IP ranges used by telecommunications companies"},
	"rnd":         {"rnd", "IP ranges used by research and development networks"},
	"social":      {"social", "IP ranges belonging to social media platforms"},
	"gaming":      {"gaming", "IP ranges used by online gaming platforms"},
}

// IPSource holds the data for a specific IP ranges source.
type IPSource struct {
	URL             string
	Name            string
	Description     string
	Categories      []Category
	DataFilename    string
	RefreshInterval time.Duration
	MetaData        IPMetaData
	Mu              sync.Mutex
	Fetcher         func(*IPSource) error
}

// IPMetaData holds the IP ranges for a source.
type IPMetaData struct {
	LastUpdate time.Time
	Prefixes   []Prefix
}

// Prefix holds the network and details for a prefix.
type Prefix struct {
	Network    net.IPNet
	Details    map[string]string
	Categories []Category // Overrides main category.
}

// ContainsIP checks if the given IP address is present in the prefixes of the IPSource.
// If present, it returns the prefix that contains the IP address.
func (src *IPSource) ContainsIP(ipAddress net.IP) *Prefix {
	for _, prefix := range src.MetaData.Prefixes {
		if prefix.Network.Contains(ipAddress) {
			return &prefix
		}
	}
	return nil
}

// mustSave serializes and saves the metadata to a file.
func (src *IPSource) mustSave() {
	file, err := os.Create(src.DataFilename)
	if err != nil {
		log.Panicf("Failed to create data file at '%s' for '%s': %v", src.DataFilename, src.Name, err)
	}
	defer file.Close()

	if err := gob.NewEncoder(file).Encode(src.MetaData); err != nil {
		log.Panicf("Failed to save data at '%s' for '%s': %v", src.DataFilename, src.Name, err)
	}
}

// load deserializes and loads the metadata from a file.
// returns true if the current data has been replaced false otherwise.
func (src *IPSource) load() bool {
	file, err := os.Open(src.DataFilename)
	if err != nil {
		log.Printf("Failed to open file '%s': %v", src.DataFilename, err)
		return false
	}
	defer file.Close()

	var data IPMetaData
	if err := gob.NewDecoder(file).Decode(&data); err != nil {
		log.Printf("Failed to decode data file '%s': %v", src.DataFilename, err)
		if removeErr := os.Remove(src.DataFilename); removeErr != nil {
			log.Printf("Failed to remove corrupt data file '%s': %v", src.DataFilename, removeErr)
		}
		return false
	}

	if time.Since(data.LastUpdate) < src.RefreshInterval {
		src.MetaData = data
		return true
	}
	return false
}

// Predefined IP range sources.
var IPRangeSources = map[string]*IPSource{
	"aws": {
		URL:             "https://ip-ranges.amazonaws.com/ip-ranges.json",
		Name:            "Amazon AWS",
		Description:     "Amazon AWS IP Ranges",
		Categories:      []Category{Categories["datacenter"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "aws.bin"),
		RefreshInterval: 48 * time.Hour,
		Fetcher:         fetchAWSData,
	},
	"google": {
		URL:             "https://www.gstatic.com/ipranges/cloud.json",
		Name:            "Google Cloud",
		Description:     "Google Cloud IP Ranges",
		Categories:      []Category{Categories["datacenter"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "google.bin"),
		RefreshInterval: 48 * time.Hour,
		Fetcher:         fetchGoogleData,
	},
	"google-bot": {
		URL:             "https://developers.google.com/static/search/apis/ipranges/googlebot.json",
		Name:            "GoogleBot",
		Description:     "GoogleBot IP Ranges of the main crawlers",
		Categories:      []Category{Categories["crawler"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "googlebot.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchGoogleBotData,
	},
	"google-bot-special": {
		URL:             "https://developers.google.com/static/search/apis/ipranges/special-crawlers.json",
		Name:            "GoogleBot Special Crawlers",
		Description:     "GoogleBot IP Ranges of the special crawlers",
		Categories:      []Category{Categories["crawler"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "googlebot-special.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchGoogleBotData,
	},
	"google-user-triggered-fetchers-google": {
		URL:             "https://developers.google.com/static/search/apis/ipranges/user-triggered-fetchers-google.json",
		Name:            "GoogleBot Users Triggered (Google)",
		Description:     "GoogleBot IP Ranges of the user triggered crawlers (google IPs)",
		Categories:      []Category{Categories["crawler"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "google-user-triggered-fetchers-google.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchGoogleBotData,
	},
	"bingbot": {
		URL:             "https://www.bing.com/toolbox/bingbot.json",
		Name:            "BingBot",
		Description:     "BingBot IP Ranges",
		Categories:      []Category{Categories["crawler"]},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "bingbot.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchBingBotData,
	},
}
