package main

import (
	"os"
	"strings"

	"./common"
)

func main() {
	input, err := common.GetInput()
	common.HandleFatalError(err, "Error getting concourse input")

	common.HandleFatalError(common.SetupSSHKey(input.Source), "Error setting up ssh key")

	common.HandleFatalError(
		common.RunGitCommand("clone --single-branch %s --branch %s %s", input.Source.RepoUrl, input.Version.ChangedBranch, os.Args[1]),
		"Error cloning git repo",
	)

	common.HandleFatalError(os.Chdir(os.Args[1]), "Error changing to repo directory")

	common.HandleFatalError(
		common.RunGitCommand("checkout -q %s", input.Version.Ref),
		"Error checking out ref",
	)

	common.HandleFatalError(
		common.RunGitCommand("submodule update --init --depth 1 --recursive"),
		"Error updating submodules",
	)

	common.HandleFatalError(
		common.RunGitCommand("config concourse-ci.branch-name %s", input.Version.ChangedBranch),
		"Error adding branch to git config",
	)

	common.HandleFatalError(
		common.RunGitCommand("config concourse-ci.prs-list %s", strings.Join(input.Version.Branches, ",")),
		"Error adding branch to git config",
	)

	common.HandleFatalError(common.OutputVersion(input.Version), "Error marshaling version json")

	common.HandleFatalError(
		common.RunGitCommandSaveOutputToFile("--no-pager log -1 --pretty=format:\"%ae\"", "committer"),
		"Error fetching git committer",
	)

	common.HandleFatalError(
		common.RunGitCommandSaveOutputToFile("log -1 --format=format:%B", "commit_message"),
		"Error fetching git commit message",
	)

}
