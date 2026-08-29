package parser

import (
	"strings"
)

type ClientInfo struct {
	Browser    string `json:"browser"`
	OS         string `json:"os"`
	DeviceType string `json:"device_type"` // Desktop, Mobile, Tablet, Bot
}

func ParseUserAgent(ua string) ClientInfo {
	if ua == "" {
		return ClientInfo{
			Browser:    "Unknown",
			OS:         "Unknown",
			DeviceType: "Unknown",
		}
	}

	lowerUA := strings.ToLower(ua)

	info := ClientInfo{
		Browser:    "Other",
		OS:         "Other",
		DeviceType: "Desktop",
	}

	// 1. Device Type Detection
	switch {
	case strings.Contains(lowerUA, "bot") || strings.Contains(lowerUA, "crawler") ||
		strings.Contains(lowerUA, "spider") || strings.Contains(lowerUA, "curl") ||
		strings.Contains(lowerUA, "wget") || strings.Contains(lowerUA, "postman"):
		info.DeviceType = "Bot / CLI"
	case strings.Contains(lowerUA, "ipad") || strings.Contains(lowerUA, "tablet"):
		info.DeviceType = "Tablet"
	case strings.Contains(lowerUA, "mobi") || strings.Contains(lowerUA, "iphone") || strings.Contains(lowerUA, "android"):
		info.DeviceType = "Mobile"
	default:
		info.DeviceType = "Desktop"
	}

	// 2. OS Detection
	switch {
	case strings.Contains(ua, "Windows NT 10.0"):
		info.OS = "Windows 10/11"
	case strings.Contains(ua, "Windows NT 6.3"):
		info.OS = "Windows 8.1"
	case strings.Contains(ua, "Windows NT 6.1"):
		info.OS = "Windows 7"
	case strings.Contains(ua, "Windows"):
		info.OS = "Windows"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		info.OS = "iOS"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		info.OS = "macOS"
	case strings.Contains(ua, "Android"):
		info.OS = "Android"
	case strings.Contains(ua, "Linux"):
		info.OS = "Linux"
	case strings.Contains(ua, "CrOS"):
		info.OS = "ChromeOS"
	}

	// 3. Browser Detection (Order matters: Chrome UA contains Safari, Edge contains Chrome)
	switch {
	case strings.Contains(ua, "Edg/") || strings.Contains(ua, "Edge/"):
		info.Browser = "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera/"):
		info.Browser = "Opera"
	case strings.Contains(ua, "Brave"):
		info.Browser = "Brave"
	case strings.Contains(ua, "Vivaldi/"):
		info.Browser = "Vivaldi"
	case strings.Contains(ua, "Chrome/") || strings.Contains(ua, "CriOS/"):
		info.Browser = "Chrome"
	case strings.Contains(ua, "Firefox/") || strings.Contains(ua, "FxiOS/"):
		info.Browser = "Firefox"
	case strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome/"):
		info.Browser = "Safari"
	case strings.Contains(ua, "curl/"):
		info.Browser = "curl"
	}

	return info
}
