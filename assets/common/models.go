package common

import (
	"encoding/json"
	"strings"
)

// BranchesSeperator indicates the separation within a single string of the branch name and commit sha in a Stash reponse
const (
	BranchesSeperator = "::"
)

// ConcourseSource the structure defining the expected source input parameter format, supports both check and in
type ConcourseSource struct {
	StashUrl       string   `json:"stash_url"`
	ProjectName    string   `json:"project_name"`
	RepoName       string   `json:"repo_name"`
	PROnly         bool     `json:"pronly"`
	DaysBack       int      `json:"days_back"`
	Branches       string   `json:"branches"`
	IgnoreBranches string   `json:"ignore_branches"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	RepoUrl        string   `json:"repo"`
	PrivateKey     string   `json:"private_key"`
	Paths          []string `json:"paths"`
}

// ConcourseInput the structure defining the expected input parameter format of the script
type ConcourseInput struct {
	Source  ConcourseSource  `json:"source"`
	Version ConcourseVersion `json:"version"`
}

// ConcourseVersion the structure defining the expected version input parameter format
type ConcourseVersion struct {
	Branches      []string `json:"the_branches"`
	ChangedBranch string   `json:"changed_branch"`
	Ref           string   `json:"ref"`
}

// MarshalJSON converts the ConcourseVersion struct into a marshalled JSON object
func (v *ConcourseVersion) MarshalJSON() ([]byte, error) {
	m := map[string]string{}
	m["changed_branch"] = v.ChangedBranch
	m["ref"] = v.Ref
	m["the_branches"] = strings.Join(v.Branches, ",")
	return json.Marshal(m)
}

// UnmarshalJSON populates the given ConcourseVersion struct with the given byte stream
func (v *ConcourseVersion) UnmarshalJSON(b []byte) error {
	m := map[string]string{}

	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	v.ChangedBranch = m["changed_branch"]
	v.Ref = m["ref"]
	if len(m["the_branches"]) > 0 {
		v.Branches = strings.Split(m["the_branches"], ",")
	}
	return nil
}
