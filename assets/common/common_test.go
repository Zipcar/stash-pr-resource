package common

import (
	"testing"
)

const (
	agentOutput = `
SSH_AUTH_SOCK=/tmp/ssh-5kZzFAN0j5Ry/agent.21; export SSH_AUTH_SOCK;
SSH_AGENT_PID=22; export SSH_AGENT_PID;
echo Agent pid 22;
`
)

func TestTimeConsuming(t *testing.T) {
	envVars := retrieveEnvVarsFromAgent(agentOutput)

	if len(envVars) != 2 {
		t.Error("Expected envVars to have length 2", len(envVars))
	}

	if envVars["SSH_AUTH_SOCK"] != "/tmp/ssh-5kZzFAN0j5Ry/agent.21" {
		t.Error("Expected SSH_AUTH_SOCK to equal /tmp/ssh-5kZzFAN0j5Ry/agent.21:", envVars["SSH_AUTH_SOCK"])
	}

	if envVars["SSH_AGENT_PID"] != "22" {
		t.Error("Expected SSH_AGENT_PID to equal 22:", envVars["SSH_AGENT_PID"])
	}
}
