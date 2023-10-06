package main

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestRuleSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// all clauses
	selector := RuleSelector{}
	selector.Included = []string{
		"p1",
		"p2",
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	expected :=
		"(p1||p2)||((konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2))"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// other
	selector = RuleSelector{}
	selector.Included = []string{
		"p1",
		"p2",
	}
	expected = "(p1||p2)"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// sources and targets
	selector = RuleSelector{}
	selector.Included = []string{
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	expected =
		"(konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// sources
	selector = RuleSelector{}
	selector.Included = []string{
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
	}
	expected = "(konveyor.io/source=s1||konveyor.io/source=s2)"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// targets
	selector = RuleSelector{}
	selector.Included = []string{
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	expected = "(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// other and targets
	selector = RuleSelector{}
	selector.Included = []string{
		"p1",
		"p2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	expected = "(p1||p2)||(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(selector.String()).To(gomega.Equal(expected))
}

func TestLabelMatch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// match name
	included := Label("konveyor.io/target=thing")
	rule := Label("konveyor.io/target=thing")
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	// name not matched.
	included = "konveyor.io/target=dog"
	rule = "konveyor.io/target=cat+"
	g.Expect(rule.Match(included)).To(gomega.BeFalse())
	// match versioned
	included = "konveyor.io/target=thing4"
	rule = "konveyor.io/target=thing4"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	// match versioned plus
	included = "konveyor.io/target=thing4"
	rule = "konveyor.io/target=thing4+"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	// match versioned ALL
	included = "konveyor.io/target=thing"
	rule = "konveyor.io/target=thing4+"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	// match version greater-than
	included = "konveyor.io/target=thing5"
	rule = "konveyor.io/target=thing4+"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	included = "konveyor.io/target=thing4.1"
	rule = "konveyor.io/target=thing4.0+"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
	// match version less-than
	included = "konveyor.io/target=thing3"
	rule = "konveyor.io/target=thing4-"
	g.Expect(rule.Match(included)).To(gomega.BeTrue())
}
