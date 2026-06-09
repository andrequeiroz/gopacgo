package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// responseAur represents the top-level JSON response from the AUR RPC API.
type responseAur struct {
	ResultCount int      `json:"resultcount"`
	Type        string   `json:"type"`
	Version     int      `json:"version"`
	Results     []pacAur `json:"results"`
}

// pacAur represents a single package entry returned by the AUR RPC API.
type pacAur struct {
	Name    string `json:"Name"`
	Version string `json:"Version"`
}

// checkAur queries the AUR RPC API for the latest version of each package.
// All packages are sent in a single batch request using the arg[] parameter.
func checkAur(pacs []pac) (*responseAur, error) {
	baseURL := "https://aur.archlinux.org/rpc/"
	params := url.Values{}
	params.Add("v", "5")
	params.Add("type", "info")

	// add each package name as a separate arg[] query parameter
	for _, p := range pacs {
		params.Add("arg[]", p.name)
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("AUR unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AUR invalid status: %d", resp.StatusCode)
	}

	var respAur responseAur
	if err := json.NewDecoder(resp.Body).Decode(&respAur); err != nil {
		return nil, fmt.Errorf("cannot decode AUR response: %w", err)
	}

	return &respAur, nil
}
