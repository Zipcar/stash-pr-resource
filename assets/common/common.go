package common

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// GetInput retrieves input parameters from stdin in the expected format
func GetInput() (ConcourseInput, error) {
	input := ConcourseInput{}

	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &input)
		if err != nil {
			return input, err
		}

		return input, nil
	}

	return input, errors.New("No input received")
}

// HandleFatalError exits status 1 and outputs an error message to stderr if a non-nil error is passed
func HandleFatalError(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%s: %s", msg, err.Error()))
	os.Exit(1)
}

// SetupSSHKey sets up an SSH key on the file system to access Stash based on the private key value specified in the input
func SetupSSHKey(source ConcourseSource) error {
	homeDir := os.Getenv("HOME")

	err := ioutil.WriteFile("/tmp/git-private-key", []byte(source.PrivateKey), os.FileMode(0600))
	if err != nil {
		return err
	}

	sshAgent, _ := exec.Command("ssh-agent").Output()

	for key, value := range retrieveEnvVarsFromAgent(string(sshAgent)) {
		os.Setenv(key, value)
	}

	err = exec.Command("ssh-add", "/tmp/git-private-key").Run()
	if err != nil {
		return err
	}

	err = os.MkdirAll(homeDir+"/.ssh", os.FileMode(0600))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(homeDir+"/.ssh/config", []byte("StrictHostKeyChecking no"), os.FileMode(0600))
	if err != nil {
		return err
	}

	return nil
}

// RunGitCommand generically runs a Git command
func RunGitCommand(command string, formating ...interface{}) error {
	args := strings.Split(fmt.Sprintf(command, formating...), " ")
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OutputVersion prints a version string to standard out based on the given ConcourseVersion object
func OutputVersion(version ConcourseVersion) error {
	output, err := json.Marshal(struct {
		Version *ConcourseVersion `json:"version"`
	}{
		Version: &version,
	})

	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func retrieveEnvVarsFromAgent(agentOutput string) map[string]string {
	m := map[string]string{}

	pattern := regexp.MustCompile("(.*)=(.*); export.*")
	scanner := bufio.NewScanner(strings.NewReader(pattern.ReplaceAllString(agentOutput, "$1=$2")))
	for scanner.Scan() {
		line := scanner.Text()
		envVar := strings.Split(line, "=")
		if len(envVar) != 2 {
			continue
		}
		m[envVar[0]] = envVar[1]
	}

	return m
}
