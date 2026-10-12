package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type gifResult struct {
	Title    string
	ImageURL string
	PageURL  string
	Provider string
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func search(cfg config, query string) ([]gifResult, error) {
	if cfg.provider == "tenor" {
		return searchTenor(cfg, query)
	}
	return searchGiphy(cfg, query)
}

func searchGiphy(cfg config, query string) ([]gifResult, error) {
	if cfg.apiKey == "" {
		return nil, fmt.Errorf("GIF_API_KEY is required for Giphy")
	}
	endpoint := fmt.Sprintf(
		"https://api.giphy.com/v1/gifs/search?api_key=%s&q=%s&limit=%d&rating=%s",
		url.QueryEscape(cfg.apiKey), url.QueryEscape(query), cfg.limit, url.QueryEscape(cfg.rating),
	)
	var resp giphyResp
	if err := getJSON(endpoint, &resp); err != nil {
		return nil, err
	}
	out := make([]gifResult, 0, len(resp.Data))
	for _, d := range resp.Data {
		img := firstNonEmpty(
			d.Images.FixedHeight.URL,
			d.Images.DownsizedMedium.URL,
			d.Images.Original.URL,
		)
		if img == "" {
			continue
		}
		out = append(out, gifResult{Title: d.Title, ImageURL: img, PageURL: d.URL, Provider: "GIPHY"})
	}
	return out, nil
}

func searchTenor(cfg config, query string) ([]gifResult, error) {
	if cfg.apiKey == "" {
		return nil, fmt.Errorf("GIF_API_KEY is required for Tenor")
	}
	endpoint := fmt.Sprintf(
		"https://tenor.googleapis.com/v2/search?key=%s&q=%s&limit=%d&media_filter=gif,tinygif",
		url.QueryEscape(cfg.apiKey), url.QueryEscape(query), cfg.limit,
	)
	var resp tenorResp
	if err := getJSON(endpoint, &resp); err != nil {
		return nil, err
	}
	out := make([]gifResult, 0, len(resp.Results))
	for _, r := range resp.Results {
		img := ""
		if m, ok := r.MediaFormats["gif"]; ok {
			img = m.URL
		}
		if img == "" {
			if m, ok := r.MediaFormats["tinygif"]; ok {
				img = m.URL
			}
		}
		if img == "" {
			continue
		}
		out = append(out, gifResult{Title: r.Title, ImageURL: img, PageURL: r.ItemURL, Provider: "Tenor"})
	}
	return out, nil
}

type giphyResp struct {
	Data []struct {
		Title  string `json:"title"`
		URL    string `json:"url"`
		Images struct {
			FixedHeight struct {
				URL string `json:"url"`
			} `json:"fixed_height"`
			DownsizedMedium struct {
				URL string `json:"url"`
			} `json:"downsized_medium"`
			Original struct {
				URL string `json:"url"`
			} `json:"original"`
		} `json:"images"`
	} `json:"data"`
}

type tenorResp struct {
	Results []struct {
		Title        string                          `json:"title"`
		ItemURL      string                          `json:"itemurl"`
		MediaFormats map[string]struct{ URL string } `json:"media_formats"`
	} `json:"results"`
}

func getJSON(endpoint string, out any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "devzat-gif")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("provider API error %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("could not parse provider response: %w", err)
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
