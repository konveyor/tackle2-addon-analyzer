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
