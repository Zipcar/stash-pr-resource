package main

import (
	"encoding/json"
	"fmt"

	"./checklib"
	"./common"
)

func main() {
	input, err := common.GetInput()
	common.HandleFatalError(err, "Error getting concourse input")

	checklib.ValidateInput(input)
	common.HandleFatalError(err, "Error while validating input")

	branches := []string{}
	updatedBranches := checklib.InitUpdatedBranches(input)
	branchToCommitMap := checklib.GetBranchToCommitMap(input)

	branches, updatedBranches = checklib.ParseStashBranches(branches, updatedBranches, branchToCommitMap, checklib.GetStashBranches(input), input)
	updatedBranches = checklib.SetBranchesNode(updatedBranches, branches)

	output, err := json.Marshal(updatedBranches)
	common.HandleFatalError(err, "Error marshaling concourse output")

	fmt.Println(string(output))
}
