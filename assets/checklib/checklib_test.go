package checklib

import (
	"testing"
	"time"

	"../common"
)

func getBranchFixture() StashBranch {
	branch := StashBranch{}
	branch.DisplayID = "feature/my-branch"
	branch.LatestCommit = "my-latest-commit-sha"
	branch.Metadata = StashBranchMetadata{}
	branch.Metadata.LatestCommitMD = StashBranchLatestCommitMD{}
	branch.Metadata.LatestCommitMD.Timestamp = time.Now().Add(time.Hour*24*time.Duration(-2)).UnixNano() / 1000000 //Two days ago in ms
	branch.Metadata.LatestCommitMD.Message = "my commit message"
	branch.Metadata.PullRequestMD = StashBranchPullRequestMD{}
	branch.Metadata.PullRequestMD.PullRequest = StashBranchPullRequest{}
	branch.Metadata.PullRequestMD.Open = 1
	branch.Metadata.PullRequestMD.PullRequest.State = "OPEN"

	return branch
}

func getConcourseInputFixture() common.ConcourseInput {
	input := common.ConcourseInput{}
	input.Source = common.ConcourseSource{}
	input.Source.PROnly = true
	input.Source.DaysBack = 0

	return input
}

func TestValidateInput_PathsExists_PrOnlyFalse(t *testing.T) {
	input := getConcourseInputFixture()
	input.Source.Paths = []string{"one", "two"}
	input.Source.PROnly = false

	err := ValidateInput(input)

	if err == nil {
		t.Error("Expected non-nil error, got nil")
	}
}

func TestValidateInput_PathsExists_PrOnlyTrue(t *testing.T) {
	input := getConcourseInputFixture()
	input.Source.Paths = []string{"one", "two"}
	input.Source.PROnly = true

	err := ValidateInput(input)

	if err != nil {
		t.Error("Expected nil error, got ", err)
	}
}

func TestValidateInput_PathsDontExist_PrOnlyFalse(t *testing.T) {
	input := getConcourseInputFixture()
	input.Source.Paths = []string{}
	input.Source.PROnly = false

	err := ValidateInput(input)

	if err != nil {
		t.Error("Expected nil error, got ", err)
	}
}

func TestProcessBranch_NoFilter_Single(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)

	if len(branches) != 1 {
		t.Error("Expected branches to have length 1, got ", len(branches))
	} else {
		if branches[0] != "feature/my-branch::my-latest-commit-sha" {
			t.Error("Expected first value in branches to be 'feature/my-branch::my-latest-commit-sha', got ", branches[0])
		}
	}

	if len(updatedBranches) != 1 {
		t.Error("Expected updatedBranches to have length 1, got ", len(updatedBranches))
	}
}

func TestProcessBranch_NoFilter_Multiple(t *testing.T) {
	branches := []string{"bugfix/my-bug-branch::my-latest-bug-commit-sha"}
	updatedBranches := []*common.ConcourseVersion{&common.ConcourseVersion{}}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)

	if len(branches) != 2 {
		t.Error("Expected branches to have length 1, got ", len(branches))
	} else {
		if branches[0] != "bugfix/my-bug-branch::my-latest-bug-commit-sha" {
			t.Error("Expected first value in branches to be 'bugfix/my-bug-branch::my-latest-bug-commit-sha', got ", branches[0])
		}
		if branches[1] != "feature/my-branch::my-latest-commit-sha" {
			t.Error("Expected first value in branches to be 'feature/my-branch::my-latest-commit-sha', got ", branches[1])
		}
	}

	if len(updatedBranches) != 2 {
		t.Error("Expected updatedBranches to have length 1, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_IgnoreBranches_true(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.IgnoreBranches = "feature/my-branch"
	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 0 {
		t.Error("Expected branches to have length 0, got ", len(branches))
	}
	if len(updatedBranches) != 0 {
		t.Error("Expected updatedBranches to have length 0, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_IgnoreBranches_false(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.IgnoreBranches = "feature/my-branch1"
	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 1 {
		t.Error("Expected branches to have length 1, got ", len(branches))
	}
	if len(updatedBranches) != 1 {
		t.Error("Expected updatedBranches to have length 1, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_IncludeBranches_true(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.Branches = "feature/my-branch"
	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 1 {
		t.Error("Expected branches to have length 1, got ", len(branches))
	}
	if len(updatedBranches) != 1 {
		t.Error("Expected updatedBranches to have length 1, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_IncludeBranches_false(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.Branches = "feature/my-branch1"
	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 0 {
		t.Error("Expected branches to have length 0, got ", len(branches))
	}
	if len(updatedBranches) != 0 {
		t.Error("Expected updatedBranches to have length 0, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_DateTime_true(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.DaysBack = 1

	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 0 {
		t.Error("Expected branches to have length 0, got ", len(branches))
	}
	if len(updatedBranches) != 0 {
		t.Error("Expected updatedBranches to have length 0, got ", len(updatedBranches))
	}
}

func TestProcessBranch_Filter_DateTime_false(t *testing.T) {
	branches := []string{}
	updatedBranches := []*common.ConcourseVersion{}
	branchToCommitMap := map[string]string{}
	branch := getBranchFixture()
	input := getConcourseInputFixture()

	input.Source.DaysBack = 3

	branches, updatedBranches = processBranch(branches, updatedBranches, branchToCommitMap, branch, input)
	if len(branches) != 1 {
		t.Error("Expected branches to have length 1, got ", len(branches))
	}
	if len(updatedBranches) != 1 {
		t.Error("Expected updatedBranches to have length 1, got ", len(updatedBranches))
	}
}
