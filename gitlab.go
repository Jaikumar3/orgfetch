package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GitLabRepo struct {
	Name  string `json:"name"`
	Fork  bool   `json:"fork"`
	Size  int    `json:"repository_size"` // size in bytes
	Owner struct {
		Username string `json:"username"`
	} `json:"owner"`
}

type GitLabMember struct {
	Username string `json:"username"`
}

// Fetch group members (users)
func fetchGitLabMembers(token, group string) ([]GitLabMember, error) {
	var members []GitLabMember
	url := fmt.Sprintf("https://gitlab.com/api/v4/groups/%s/members", group)
	resp, err := gitlabApiGet(token, url)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// Fetch group projects (repos)
func fetchGitLabRepos(token, group string) ([]GitLabRepo, error) {
	var repos []GitLabRepo
	url := fmt.Sprintf("https://gitlab.com/api/v4/groups/%s/projects?per_page=100", group)
	resp, err := gitlabApiGet(token, url)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

// Fetch user projects (repos)
func fetchGitLabUserRepos(token, username string) ([]GitLabRepo, error) {
	var repos []GitLabRepo
	url := fmt.Sprintf("https://gitlab.com/api/v4/users/%s/projects?per_page=100", username)
	resp, err := gitlabApiGet(token, url)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

// Helper for GitLab API requests
func gitlabApiGet(token, url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
