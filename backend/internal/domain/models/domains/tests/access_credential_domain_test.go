package domains_test

import (
	"testing"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/stretchr/testify/assert"
)

func TestAccessCredentialDomainAdmitsOnlyTheConfiguredOwner(t *testing.T) {
	cases := []struct {
		name       string
		credential vo.AccessCredentialVo
		username   string
		password   string
		expected   bool
	}{
		{name: "correct credentials", credential: vo.AccessCredentialVo{OwnerUsername: "james", OwnerPassword: "secret"}, username: "james", password: "secret", expected: true},
		{name: "wrong password", credential: vo.AccessCredentialVo{OwnerUsername: "james", OwnerPassword: "secret"}, username: "james", password: "guess", expected: false},
		{name: "wrong username", credential: vo.AccessCredentialVo{OwnerUsername: "james", OwnerPassword: "secret"}, username: "eve", password: "secret", expected: false},
		{name: "unset credentials admit nobody", credential: vo.AccessCredentialVo{}, username: "", password: "", expected: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			admitted := domains.NewAccessCredentialDomain(testCase.credential).AdmitsOwner(testCase.username, testCase.password)

			assert.Equal(t, testCase.expected, admitted)
		})
	}
}

func TestAccessCredentialDomainAdmitsOnlyTheConfiguredAgentKey(t *testing.T) {
	cases := []struct {
		name     string
		apiKey   string
		given    string
		expected bool
	}{
		{name: "matching key", apiKey: "agentkey", given: "agentkey", expected: true},
		{name: "other key", apiKey: "agentkey", given: "other", expected: false},
		{name: "unset key admits nobody", apiKey: "", given: "", expected: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			admitted := domains.NewAccessCredentialDomain(vo.AccessCredentialVo{AgentApiKey: testCase.apiKey}).AdmitsAgent(testCase.given)

			assert.Equal(t, testCase.expected, admitted)
		})
	}
}
