package checklib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"../common"
)

// StashBranchPullRequest the structure of the pull request response from Stash
type StashBranchPullRequest struct {
	State string `json:"state"`
	ID    int    `json:"id"`
}

// StashBranchPullRequestMD the structure of the pull request metadata response from Stash
type StashBranchPullRequestMD struct {
	PullRequest StashBranchPullRequest `json:"pullRequest"`
	Open        int                    `json:"open"`
}

// StashBranchLatestCommitMD the structure of the commit metadata response from Stash
type StashBranchLatestCommitMD struct {
	Timestamp int64  `json:"authorTimestamp"`
	Message   string `json:"message"`
}

// StashBranchMetadata the structure of the branch metadata response from Stash
type StashBranchMetadata struct {
	LatestCommitMD StashBranchLatestCommitMD `json:"com.atlassian.bitbucket.server.bitbucket-branch:latest-commit-metadata"`
	PullRequestMD  StashBranchPullRequestMD  `json:"com.atlassian.bitbucket.server.bitbucket-ref-metadata:outgoing-pull-request-metadata"`
}

// StashBranch the structure of the branch response from Stash
type StashBranch struct {
	DisplayID    string              `json:"displayId"`
	LatestCommit string              `json:"latestCommit"`
	Metadata     StashBranchMetadata `json:"metadata"`
}

// StashBranches the structure of the branches response from Stash
type StashBranches struct {
	Branches []StashBranch `json:"values"`
}

// StashPullRequestChangePage the structure of the pull request change page response from Stash
type StashPullRequestChangePage struct {
	Changes    []StashPullRequestChange `json:"values"`
	IsLastPage bool                     `json:"isLastPage"`
}

// StashPullRequestChange the structure of the pull request change response from Stash
type StashPullRequestChange struct {
	Path StashPullRequestPath `json:"path"`
}

// StashPullRequestPath the structure of the pull request path response from Stash
type StashPullRequestPath struct {
	Parent string `json:"parent"`
	Name   string `json:"name"`
}

// ValidateInput returns an errors object if validation doesn't pass, nil otherwise
func ValidateInput(input common.ConcourseInput) error {
	if !input.Source.PROnly && len(input.Source.Paths) > 0 {
		return errors.New("Cannot pass paths when pronly is false")
	}
	return nil
}

func filterOutByDateAndTime(latestCommitTime int64, input common.ConcourseInput) bool {
	if input.Source.DaysBack > 0 {
		cutOffTime := time.Now().Add(time.Hour*24*time.Duration(-input.Source.DaysBack)).UnixNano() / 1000000

		if latestCommitTime < cutOffTime {
			return true
		}
	}

	return false
}

func filterOutByBranchName(branch StashBranch, input common.ConcourseInput) bool {
	if input.Source.Branches != "" {
		rBranches := regexp.MustCompile(input.Source.Branches)
		if !rBranches.MatchString(branch.DisplayID) {
			return true
		}
	}

	if input.Source.IgnoreBranches != "" {
		rIgnoreBranches := regexp.MustCompile(input.Source.IgnoreBranches)
		if rIgnoreBranches.MatchString(branch.DisplayID) {
			return true
		}
	}

	return false
}

func commitMarkedAsSkip(branch StashBranch) bool {
	message := branch.Metadata.LatestCommitMD.Message
	if strings.Contains(message, "[ci skip]") || strings.Contains(message, "[skip ci]") {
		return true
	}

	return false
}

func notANewCommit(branch StashBranch, branchToCommitMap map[string]string) bool {
	if previousLatestCommit, ok := branchToCommitMap[branch.DisplayID]; ok {
		if previousLatestCommit == branch.LatestCommit {
			return true
		}
	}

	return false
}

func noOpenPR(branch StashBranch) bool {
	if branch.Metadata.PullRequestMD.PullRequest.State == "OPEN" || branch.Metadata.PullRequestMD.Open > 0 {
		return false
	}

	return true
}

func pathNotInPrs(branch StashBranch, input common.ConcourseInput) bool {
	if len(input.Source.Paths) > 0 {
		pullRequestChangedPaths := getStashBranchPullRequestChangePaths(input, branch.Metadata.PullRequestMD.PullRequest.ID)
		for _, changedPath := range pullRequestChangedPaths {
			for _, desiredPath := range input.Source.Paths {

				//this is just string matching instead of actual path matching, but that is ok as a first pass
				if strings.HasPrefix(changedPath, desiredPath) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func processBranch(branches []string, updatedBranches []*common.ConcourseVersion,
	branchToCommitMap map[string]string, branch StashBranch, input common.ConcourseInput) ([]string, []*common.ConcourseVersion) {

	branchDateAndTime := branch.Metadata.LatestCommitMD.Timestamp

	if input.Source.PROnly {
		if noOpenPR(branch) {
			return branches, updatedBranches
		}
	}

	if filterOutByDateAndTime(branchDateAndTime, input) {
		return branches, updatedBranches
	}

	if filterOutByBranchName(branch, input) {
		return branches, updatedBranches
	}

	if commitMarkedAsSkip(branch) {
		return branches, updatedBranches
	}

	if pathNotInPrs(branch, input) {
		return branches, updatedBranches
	}

	branches = append(branches, fmt.Sprintf("%s%s%s", branch.DisplayID, common.BranchesSeperator, branch.LatestCommit))

	if notANewCommit(branch, branchToCommitMap) {
		return branches, updatedBranches
	}

	updatedBranches = append(updatedBranches, &common.ConcourseVersion{
		ChangedBranch: branch.DisplayID,
		Ref:           branch.LatestCommit,
	})

	return branches, updatedBranches
}

// GetBranchToCommitMap returns a map from the flattened list of branches and sha's
func GetBranchToCommitMap(input common.ConcourseInput) map[string]string {
	branchToCommitMap := map[string]string{}
	for _, branch := range input.Version.Branches {
		branchCommit := strings.Split(branch, common.BranchesSeperator)
		branchToCommitMap[branchCommit[0]] = branchCommit[1]
	}

	return branchToCommitMap
}

// SetBranchesNode returns a list of updated branches only
func SetBranchesNode(updatedBranches []*common.ConcourseVersion, branches []string) []*common.ConcourseVersion {
	for _, updatedBranch := range updatedBranches {
		if len(updatedBranch.Branches) < 1 {
			updatedBranch.Branches = branches
		}
	}

	return updatedBranches
}

// ParseStashBranches returns a populated list of all branches and updatedBranches after a post-filtering process
func ParseStashBranches(branches []string, updatedBranches []*common.ConcourseVersion,
	branchToCommitMap map[string]string, respBody []byte, input common.ConcourseInput) ([]string, []*common.ConcourseVersion) {

	stashBranches := StashBranches{}
	err := json.Unmarshal(respBody, &stashBranches)
	common.HandleFatalError(err, "Error parsing stash branches response json")

	for _, branch := range stashBranches.Branches {
		branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	}

	return branches, updatedBranches
}

// GetStashBranches returns the response of a call to the Stash service returning all branches associated to a given project and repo
func GetStashBranches(input common.ConcourseInput) []byte {
	url := fmt.Sprintf("https://%s:%s@%s/rest/api/1.0/projects/%s/repos/%s/branches?limit=1000&details=true",
		input.Source.Username,
		input.Source.Password,
		input.Source.StashUrl,
		input.Source.ProjectName,
		input.Source.RepoName)

	resp, err := http.DefaultClient.Get(url)
	common.HandleFatalError(err, "Error getting stash response")

	respBody, err := ioutil.ReadAll(resp.Body)
	common.HandleFatalError(err, "Error reading stash response")

	if resp.StatusCode != 200 {
		fmt.Printf(string(respBody))
		common.HandleFatalError(
			fmt.Errorf("Expected 200 response code but got %v from %s", resp.StatusCode, url),
			"Error reading stash branches response",
		)
	}

	return respBody
}

func getStashBranchPullRequestChangePage(url string) StashPullRequestChangePage {
	resp, err := http.DefaultClient.Get(url)
	common.HandleFatalError(err, "Error getting stash response")

	respBody, err := ioutil.ReadAll(resp.Body)
	common.HandleFatalError(err, "Error reading stash response")

	if resp.StatusCode != 200 {
		fmt.Printf(string(respBody))
		common.HandleFatalError(
			fmt.Errorf("Expected 200 response code but got %v from %s", resp.StatusCode, url),
			"Error reading stash pull request response",
		)
	}

	pullRequestChangePage := StashPullRequestChangePage{}
	err = json.Unmarshal(respBody, &pullRequestChangePage)
	common.HandleFatalError(err, "Error parsing stash pull request response json")

	return pullRequestChangePage
}

func getStashBranchPullRequestChangePaths(input common.ConcourseInput, pullRequestID int) []string {
	url := fmt.Sprintf("https://%s:%s@%s/rest/api/1.0/projects/%s/repos/%s/pull-requests/%v/changes",
		input.Source.Username,
		input.Source.Password,
		input.Source.StashUrl,
		input.Source.ProjectName,
		input.Source.RepoName,
		pullRequestID)
	pullRequestChangePage := getStashBranchPullRequestChangePage(url)

	if !pullRequestChangePage.IsLastPage {
		common.HandleFatalError(
			errors.New("Pull requests with paged change list is not yet implemented.  This PR is too large for this resource to handle"),
			"Error parsing stash pull request response json",
		)
	}

	pullRequestChangePages := []StashPullRequestChangePage{pullRequestChangePage}

	return buildPullRequestChangesPathArray(pullRequestChangePages)
}

func buildPullRequestChangesPathArray(pullRequestChangePages []StashPullRequestChangePage) []string {
	changePaths := []string{}
	for _, page := range pullRequestChangePages {
		for _, change := range page.Changes {
			path := fmt.Sprintf("%s/%s", change.Path.Parent, change.Path.Name)
			if change.Path.Parent == "" {
				path = change.Path.Name
			}
			changePaths = append(changePaths, path)
		}
	}
	return changePaths
}

// InitUpdatedBranches appends version information to each updated branch returned from Stash
func InitUpdatedBranches(input common.ConcourseInput) []*common.ConcourseVersion {
	updatedBranches := []*common.ConcourseVersion{}
	if len(input.Version.Branches) > 0 {
		updatedBranches = append(updatedBranches, &input.Version)
	}

	return updatedBranches
}
