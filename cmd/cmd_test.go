package main

import (
	"github.com/konveyor/tackle2-addon/command"
	"github.com/onsi/gomega"
	"testing"
)

func TestRuleSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// all clauses
	rules := Rules{}
	rules.Labels.Included = []string{
		"p1",
		"p2",
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	options := command.Options{}
	err := rules.addSelector(&options)
	selector :=
		"(p1||p2)||((konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2))"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
	// other
	rules = Rules{}
	rules.Labels.Included = []string{
		"p1",
		"p2",
	}
	options = command.Options{}
	err = rules.addSelector(&options)
	selector = "(p1||p2)"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
	// sources and targets
	rules = Rules{}
	rules.Labels.Included = []string{
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	options = command.Options{}
	err = rules.addSelector(&options)
	selector =
		"(konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
	// sources
	rules = Rules{}
	rules.Labels.Included = []string{
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
	}
	options = command.Options{}
	err = rules.addSelector(&options)
	selector = "(konveyor.io/source=s1||konveyor.io/source=s2)"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
	// targets
	rules = Rules{}
	rules.Labels.Included = []string{
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	options = command.Options{}
	err = rules.addSelector(&options)
	selector = "(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
	// other and targets
	rules = Rules{}
	rules.Labels.Included = []string{
		"p1",
		"p2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	options = command.Options{}
	err = rules.addSelector(&options)
	selector = "(p1||p2)||(konveyor.io/target=t1||konveyor.io/target=t2)"
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(options)).To(gomega.Equal(2))
	g.Expect(options[1]).To(gomega.Equal(selector))
}
