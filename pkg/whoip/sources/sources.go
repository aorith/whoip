package sources

import (
	"encoding/gob"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	utils "github.com/aorith/whoip/internal"
)

// Category represents the type of a category.
type Category struct {
	ID          string
	Description string
}

// Predefined Category instances.
var (
	Crawler     = Category{"crawler", "IP ranges used by web crawlers and bots"}
	Residential = Category{"residential", "IP ranges assigned to residential users by ISPs"}
	Business    = Category{"business", "IP ranges assigned to businesses"}
	Mobile      = Category{"mobile", "IP ranges used by mobile carriers for their data services"}
	DataCenter  = Category{"datacenter", "IP ranges belonging to data centers and hosting providers"}
	Education   = Category{"education", "IP ranges assigned to educational institutions"}
	Government  = Category{"government", "IP ranges used by government agencies"}
	Healthcare  = Category{"healthcare", "IP ranges used by healthcare providers"}
	CDN         = Category{"cdn", "IP ranges used by content delivery networks"}
	ISP         = Category{"isp", "IP ranges owned by internet service providers"}
	VPN         = Category{"vpn", "IP ranges used by VPN and proxy services"}
	Spam        = Category{"spam", "IP ranges identified as sources of spam activity"}
	Malicious   = Category{"malicious", "IP ranges identified as sources of malicious activity"}
	Private     = Category{"private", "Non-routable IP ranges used for private networks"}
	IoT         = Category{"iot", "IP ranges used by Internet of Things devices"}
	Telecom     = Category{"telecom", "IP ranges used by telecommunications companies"}
	RnD         = Category{"rnd", "IP ranges used by research and development networks"}
	SocialMedia = Category{"social", "IP ranges belonging to social media platforms"}
	Gaming      = Category{"gaming", "IP ranges used by online gaming platforms"}
)

// IPSource holds the data for a specific IP ranges source.
type IPSource struct {
	URL             string
	Name            string
	Description     string
	Categories      []Category
	DataFilename    string
	RefreshInterval time.Duration
	MetaData        IPMetaData
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
func (src *IPSource) load() error {
	file, err := os.Open(src.DataFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	var data IPMetaData
	if err := gob.NewDecoder(file).Decode(&data); err != nil {
		if removeErr := os.Remove(src.DataFilename); removeErr != nil {
			log.Printf("Failed to remove corrupt data file '%s': %v", src.DataFilename, removeErr)
		}
		return err
	}

	if time.Since(data.LastUpdate) < src.RefreshInterval {
		src.MetaData = data
	}
	return nil
}

// Predefined IP range sources.
var iprangeSources = map[string]*IPSource{
	"aws": {
		URL:             "https://ip-ranges.amazonaws.com/ip-ranges.json",
		Name:            "Amazon AWS",
		Description:     "Amazon AWS IP Ranges",
		Categories:      []Category{DataCenter},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "aws.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchAWSData,
	},
	"google": {
		URL:             "https://www.gstatic.com/ipranges/cloud.json",
		Name:            "Google Cloud",
		Description:     "Google Cloud IP Ranges",
		Categories:      []Category{DataCenter},
		DataFilename:    filepath.Join(utils.GetDataDirectory(), "google.bin"),
		RefreshInterval: 24 * time.Hour,
		Fetcher:         fetchGoogleData,
	},
}
