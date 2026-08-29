package data

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrLinkNotFound = errors.New("link not found")
)

type Visit struct {
	ID         string    `json:"id"`
	VisitedAt  time.Time `json:"visited_at"`
	IP         string    `json:"ip"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	Region     string    `json:"region"`
	Timezone   string    `json:"timezone"`
	ISP        string    `json:"isp"`
	Org        string    `json:"org"`
	ASN        string    `json:"asn"`
	Browser    string    `json:"browser"`
	OS         string    `json:"os"`
	DeviceType string    `json:"device_type"`
	Referer    string    `json:"referer"`
	Language   string    `json:"language"`
	UserAgent  string    `json:"user_agent"`
}

type Link struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	TargetURL   string    `json:"target_url"`
	ShortURL    string    `json:"short_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	TotalVisits int       `json:"total_visits"`
	Visits      []Visit   `json:"visits"`
}

type LinkModel struct {
	mu    sync.RWMutex
	links map[string]*Link // keyed by slug
}

func NewLinkModel() *LinkModel {
	return &LinkModel{
		links: make(map[string]*Link),
	}
}

func (m *LinkModel) Insert(link *Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if link.Visits == nil {
		link.Visits = make([]Visit, 0)
	}
	m.links[link.Slug] = link
	return nil
}

func (m *LinkModel) GetBySlug(slug string) (*Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	link, exists := m.links[slug]
	if !exists {
		return nil, ErrLinkNotFound
	}

	linkCopy := *link
	linkCopy.Visits = make([]Visit, len(link.Visits))
	copy(linkCopy.Visits, link.Visits)

	return &linkCopy, nil
}

func (m *LinkModel) RecordVisit(slug string, visit Visit) (*Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	link, exists := m.links[slug]
	if !exists {
		return nil, ErrLinkNotFound
	}

	link.TotalVisits++
	link.Visits = append([]Visit{visit}, link.Visits...)

	linkCopy := *link
	linkCopy.Visits = make([]Visit, len(link.Visits))
	copy(linkCopy.Visits, link.Visits)

	return &linkCopy, nil
}

func (m *LinkModel) GetAll() []*Link {
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := make([]*Link, 0, len(m.links))
	for _, link := range m.links {
		linkCopy := *link
		linkCopy.Visits = make([]Visit, len(link.Visits))
		copy(linkCopy.Visits, link.Visits)
		all = append(all, &linkCopy)
	}
	return all
}

type Models struct {
	Links *LinkModel
}

func NewModels() Models {
	return Models{
		Links: NewLinkModel(),
	}
}
