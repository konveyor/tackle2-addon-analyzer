package builder

import (
	"bytes"
	"os"
	"testing"

	output "github.com/konveyor/analyzer-lsp/output/v1/konveyor"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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

func TestInsightBuilder(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	report := []output.RuleSet{
		{
			Name: "Test",
			Violations: map[string]output.Violation{
				"rule-001": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-001 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-001 matched here.",
						},
					},
				},
				"rule-002": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-002 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-002 matched here.",
						},
					},
				},
				"rule-003": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-002 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-002 matched here.",
						},
					},
				},
			},
			Insights: map[string]output.Violation{
				"rule-004": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-004 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-004 matched here.",
						},
					},
				},
				"rule-005": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-005 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-005 matched here.",
						},
					},
				},
				"rule-006": {
					Incidents: []output.Incident{
						{
							URI:     "file:///path",
							Message: "rule-006 matched here.",
						},
						{
							URI:     "file:///path2",
							Message: "rule-006 matched here.",
						},
					},
				},
			},
			Tags: []string{
				"tag1",
				"tag2",
				"tag3",
			},
		},
	}

	b, err := yaml.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	_, err = f.WriteString(string(b))
	if err != nil {
		t.Fatal(err)
	}

	builder, err := NewInsights(f.Name())
	g.Expect(err).To(gomega.BeNil())

	bfr := bytes.NewBuffer([]byte{})
	err = builder.Write(bfr)
	g.Expect(err).To(gomega.BeNil())

	expected := `BEGIN-INSIGHTS
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: rule-001
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-001 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-001 matched here.
  codeSnip: ""
  facts: {}
labels: []
---
analysis: 0
ruleset: Test
rule: rule-002
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-002 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-002 matched here.
  codeSnip: ""
  facts: {}
labels: []
---
analysis: 0
ruleset: Test
rule: rule-003
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-002 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-002 matched here.
  codeSnip: ""
  facts: {}
labels: []
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: ""
name: ""
labels: []
---
analysis: 0
ruleset: Test
rule: rule-004
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-004 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-004 matched here.
  codeSnip: ""
  facts: {}
labels: []
---
analysis: 0
ruleset: Test
rule: rule-005
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-005 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-005 matched here.
  codeSnip: ""
  facts: {}
labels: []
---
analysis: 0
ruleset: Test
rule: rule-006
name: ""
incidents:
- insight: 0
  file: /path
  line: 0
  message: rule-006 matched here.
  codeSnip: ""
  facts: {}
- insight: 0
  file: /path2
  line: 0
  message: rule-006 matched here.
  codeSnip: ""
  facts: {}
labels: []
END-INSIGHTS
`

	g.Expect(expected).To(gomega.Equal(bfr.String()))

	tags := builder.Tags()

	g.Expect([]string{
		"tag1",
		"tag2",
		"tag3",
	}).To(gomega.Equal(tags))
}

func TestDepsBuilder(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	report := []output.DepsFlatItem{
		{
			FileURI:  "file:///path/to/file",
			Provider: "java",
			Dependencies: []*output.Dep{
				{
					Name:               "dep001",
					Version:            "1.0",
					Indirect:           false,
					ResolvedIdentifier: "0x001",
					Labels: []string{
						"label001",
						"label002",
						"label003",
					},
				},
			},
		},
		{
			FileURI:  "file:///path/to/file",
			Provider: "java",
			Dependencies: []*output.Dep{
				{
					Name:               "dep002",
					Version:            "2.0",
					Indirect:           false,
					ResolvedIdentifier: "0x002",
					Labels: []string{
						"label001",
						"label002",
						"label003",
					},
				},
			},
		},
		{
			FileURI:  "file:///path/to/file",
			Provider: "java",
			Dependencies: []*output.Dep{
				{
					Name:               "dep003",
					Version:            "3.0",
					Indirect:           false,
					ResolvedIdentifier: "0x003",
					Labels: []string{
						"label001",
						"label002",
						"label003",
					},
				},
			},
		},
	}

	b, err := yaml.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	_, err = f.WriteString(string(b))
	if err != nil {
		t.Fatal(err)
	}

	builder, err := NewDeps(f.Name())
	g.Expect(err).To(gomega.BeNil())

	bfr := bytes.NewBuffer([]byte{})
	err = builder.Write(bfr)
	g.Expect(err).To(gomega.BeNil())

	expected := `BEGIN-DEPS
---
analysis: 0
provider: java
name: dep001
version: "1.0"
labels:
- label001
- label002
- label003
sha: "0x001"
---
analysis: 0
provider: java
name: dep002
version: "2.0"
labels:
- label001
- label002
- label003
sha: "0x002"
---
analysis: 0
provider: java
name: dep003
version: "3.0"
labels:
- label001
- label002
- label003
sha: "0x003"
END-DEPS
`

	g.Expect(expected).To(gomega.Equal(bfr.String()))
}
