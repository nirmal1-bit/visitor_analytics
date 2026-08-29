package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"visitor_analytics/internal/data"
	"visitor_analytics/internal/geo"
	"visitor_analytics/internal/parser"
	"visitor_analytics/internal/validator"
)

func generateSlug(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	// Check X-Real-IP
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (app *application) createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TargetURL string `json:"target_url"`
		CustomSlug string `json:"custom_slug"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(strings.TrimSpace(input.TargetURL) != "", "target_url", "must be provided")
	v.Check(validator.IsURL(input.TargetURL), "target_url", "must be a valid URL (including http:// or https://)")

	slug := strings.TrimSpace(input.CustomSlug)
	if slug == "" {
		slug = generateSlug(4) // 8 hex characters
	} else {
		v.Check(len(slug) >= 3, "custom_slug", "must be at least 3 characters long")
		v.Check(len(slug) <= 50, "custom_slug", "must not be more than 50 characters long")
	}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Check if already exists
	if _, err := app.models.Links.GetBySlug(slug); err == nil {
		v.AddError("custom_slug", "slug already in use")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	link := &data.Link{
		ID:        generateSlug(8),
		Slug:      slug,
		TargetURL: input.TargetURL,
		ShortURL:  fmt.Sprintf("%s/r/%s", strings.TrimRight(app.config.baseURL, "/"), slug),
		CreatedAt: time.Now().UTC(),
		Visits:    []data.Visit{},
	}

	err = app.models.Links.Insert(link)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"link": link}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listLinksHandler(w http.ResponseWriter, r *http.Request) {
	links := app.models.Links.GetAll()
	for _, l := range links {
		l.ShortURL = fmt.Sprintf("%s/r/%s", strings.TrimRight(app.config.baseURL, "/"), l.Slug)
	}
	err := app.writeJSON(w, http.StatusOK, envelope{"links": links}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getLinkAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.Links.GetBySlug(slug)
	if err != nil {
		if err == data.ErrLinkNotFound {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	link.ShortURL = fmt.Sprintf("%s/r/%s", strings.TrimRight(app.config.baseURL, "/"), link.Slug)

	err = app.writeJSON(w, http.StatusOK, envelope{"link": link}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) redirectLinkHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		app.notFoundResponse(w, r)
		return
	}

	link, err := app.models.Links.GetBySlug(slug)
	if err != nil {
		if err == data.ErrLinkNotFound {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()
	referer := r.Referer()
	acceptLang := r.Header.Get("Accept-Language")
	if len(acceptLang) > 20 {
		acceptLang = acceptLang[:20]
	}

	// Geolocation & Network metadata lookup
	loc, err := app.geoProvider.Lookup(clientIP)
	if err != nil {
		loc = &geo.Location{
			IP:       clientIP,
			Country:  "Unknown",
			City:     "Unknown",
			Region:   "Unknown",
			Timezone: "Unknown",
			ISP:      "Unknown",
			Org:      "Unknown",
			ASN:      "Unknown",
		}
	}

	// Parse User-Agent for OS, Browser, Device
	clientInfo := parser.ParseUserAgent(userAgent)

	visit := data.Visit{
		ID:         generateSlug(6),
		VisitedAt:  time.Now().UTC(),
		IP:         loc.IP,
		Country:    loc.Country,
		City:       loc.City,
		Region:     loc.Region,
		Timezone:   loc.Timezone,
		ISP:        loc.ISP,
		Org:        loc.Org,
		ASN:        loc.ASN,
		Browser:    clientInfo.Browser,
		OS:         clientInfo.OS,
		DeviceType: clientInfo.DeviceType,
		Referer:    referer,
		Language:   acceptLang,
		UserAgent:  userAgent,
	}

	_, _ = app.models.Links.RecordVisit(slug, visit)

	// Redirect to target URL
	http.Redirect(w, r, link.TargetURL, http.StatusFound)
}

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
