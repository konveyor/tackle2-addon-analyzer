package builder

import (
	"testing"

	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/onsi/gomega"
)

func TestNextId(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	b := Insights{}
	b.input = []output.RuleSet{
		{
			Name: "RULESET-A",
			Violations: map[string]output.Violation{
				"rule-000": {},
				"rule-001": {},
				"rule-002": {},
			},
			Insights: map[string]output.Violation{
				"rule-001": {},
				"rule-003": {},
				"rule-004": {},
			},
		},
	}
	b.ensureUnique()
	cleaned := []output.RuleSet{
		{
			Name: "RULESET-A",
			Violations: map[string]output.Violation{
				"rule-000": {},
				"rule-001": {},
				"rule-002": {},
			},
			Insights: map[string]output.Violation{
				"rule-001_": {},
				"rule-003":  {},
				"rule-004":  {},
			},
		},
	}
	g.Expect(cleaned).To(gomega.Equal(b.input))
}
