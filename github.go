package main

import (

	"encoding/json"

	"fmt"

	"io/ioutil"

	"net/http"

	"strings"

)

type Repo struct {

	Name     string `json:"name"`

	Fork     bool   `json:"fork"`

	Size     int    `json:"size"` // size in KB

	Owner    struct {

		Login string `json:"login"`

	} `json:"owner"`

}



type Member struct {

	Login string `json:"login"`

}

func fetchMembers(token, org string) ([]Member, error) {

	var members []Member

	url := fmt.Sprintf("https://api.github.com/orgs/%s/members", org)

	for url != "" {

		resp, err := apiGet(token, url)

		if err != nil {

			return nil, err

		}

		var page []Member

		if err := json.Unmarshal(resp.Body, &page); err != nil {

			return nil, err

		}

		members = append(members, page...)

		url = resp.Next

	}

	return members, nil

}



func fetchRepos(token, org string) ([]Repo, error) {

	var repos []Repo

	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100", org)

	for url != "" {

		resp, err := apiGet(token, url)

		if err != nil {

			return nil, err

		}

		var page []Repo

		if err := json.Unmarshal(resp.Body, &page); err != nil {

			return nil, err

		}

		repos = append(repos, page...)

		url = resp.Next

	}

	return repos, nil

}

func fetchUserRepos(token, username string) ([]Repo, error) {

	var repos []Repo

	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100", username)

	for url != "" {

		resp, err := apiGet(token, url)

		if err != nil {

			return nil, err

		}

		var page []Repo

		if err := json.Unmarshal(resp.Body, &page); err != nil {

			return nil, err

		}

		repos = append(repos, page...)

		url = resp.Next

	}

	return repos, nil

}

// Helper for GitHub API requests with pagination

type apiResponse struct {

	Body []byte

	Next string

}



func apiGet(token, url string) (apiResponse, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {

		return apiResponse{}, err

	}

	req.Header.Set("Authorization", "token "+token)

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)

	if err != nil {

		return apiResponse{}, err

	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {

		return apiResponse{}, err

	}

	// Parse Link header for pagination

	next := ""

	if link := resp.Header.Get("Link"); link != "" {

		for _, part := range splitLinks(link) {

			if part.Rel == "next" {

				next = part.URL

			}

		}

	}

	return apiResponse{Body: body, Next: next}, nil

}



type linkPart struct {

	URL string

	Rel string

}



func splitLinks(header string) []linkPart {

	var parts []linkPart

	for _, s := range splitAndTrim(header, ",") {

		var url, rel string

		if i := strings.Index(s, "<"); i != -1 {

			j := strings.Index(s, ">")

			if j > i {

				url = s[i+1 : j]

			}

		}

		if i := strings.Index(s, "rel="); i != -1 {

			rel = s[i+5 : len(s)-1]

		}

		if url != "" && rel != "" {

			parts = append(parts, linkPart{URL: url, Rel: rel})

		}

	}

	return parts

}



func splitAndTrim(s, sep string) []string {

	var out []string

	for _, part := range split(s, sep) {

		out = append(out, trim(part))

	}

	return out

}

func split(s, sep string) []string { return strings.Split(s, sep) }

func trim(s string) string         { return strings.TrimSpace(s) }

