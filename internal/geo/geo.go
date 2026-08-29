package geo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type Location struct {
	IP        string  `json:"ip"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Region    string  `json:"region"`
	Timezone  string  `json:"timezone"`
	ISP       string  `json:"isp"`
	Org       string  `json:"org"`
	ASN       string  `json:"asn"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type Provider interface {
	Lookup(ip string) (*Location, error)
}

// IPAPIProvider queries ip-api.com for enriched geolocation and network information
type IPAPIProvider struct {
	client *http.Client
}

func NewIPAPIProvider(timeout time.Duration) *IPAPIProvider {
	return &IPAPIProvider{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type ipApiResponse struct {
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	Country    string  `json:"country"`
	RegionName string  `json:"regionName"`
	City       string  `json:"city"`
	Timezone   string  `json:"timezone"`
	ISP        string  `json:"isp"`
	Org        string  `json:"org"`
	AS         string  `json:"as"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Query      string  `json:"query"`
}

func (p *IPAPIProvider) Lookup(ip string) (*Location, error) {
	cleanIP := strings.TrimSpace(ip)
	if host, _, err := net.SplitHostPort(cleanIP); err == nil {
		cleanIP = host
	}

	parsedIP := net.ParseIP(cleanIP)
	if parsedIP == nil || parsedIP.IsLoopback() || parsedIP.IsPrivate() || parsedIP.IsUnspecified() {
		return &Location{
			IP:        cleanIP,
			Country:   "Local Network",
			City:      "Localhost",
			Region:    "Internal",
			Timezone:  "UTC",
			ISP:       "Loopback / Private",
			Org:       "Localhost",
			ASN:       "N/A",
			Latitude:  0.0,
			Longitude: 0.0,
		}, nil
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,regionName,city,timezone,isp,org,as,lat,lon,query", cleanIP)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "VisitorAnalytics/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return defaultUnknown(cleanIP), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return defaultUnknown(cleanIP), nil
	}

	var res ipApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return defaultUnknown(cleanIP), nil
	}

	if res.Status != "success" {
		return defaultUnknown(cleanIP), nil
	}

	return &Location{
		IP:        cleanIP,
		Country:   res.Country,
		City:      res.City,
		Region:    res.RegionName,
		Timezone:  res.Timezone,
		ISP:       res.ISP,
		Org:       res.Org,
		ASN:       res.AS,
		Latitude:  res.Lat,
		Longitude: res.Lon,
	}, nil
}

func defaultUnknown(ip string) *Location {
	return &Location{
		IP:        ip,
		Country:   "Unknown",
		City:      "Unknown",
		Region:    "Unknown",
		Timezone:  "Unknown",
		ISP:       "Unknown",
		Org:       "Unknown",
		ASN:       "Unknown",
		Latitude:  0.0,
		Longitude: 0.0,
	}
}
