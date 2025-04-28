package main

import (
	"errors"
	"testing"

	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/onsi/gomega"
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
		"((p1||p2)||((konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2)))"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// all clauses plus excluded
	selector = RuleSelector{}
	selector.Included = []string{
		"p1",
		"p2",
		"konveyor.io/source=s1",
		"konveyor.io/source=s2",
		"konveyor.io/target=t1",
		"konveyor.io/target=t2",
	}
	selector.Excluded = []string{
		"x1",
		"x2",
	}
	expected =
		"(((p1||p2)||((konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2)))&&!(x1||x2))"
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
		"((konveyor.io/source=s1||konveyor.io/source=s2)&&(konveyor.io/target=t1||konveyor.io/target=t2))"
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
	expected = "((p1||p2)||(konveyor.io/target=t1||konveyor.io/target=t2))"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// excluded (one)
	selector = RuleSelector{}
	selector.Excluded = []string{
		"x1",
	}
	expected = "!x1"
	g.Expect(selector.String()).To(gomega.Equal(expected))
	// excluded (many)
	selector = RuleSelector{}
	selector.Excluded = []string{
		"x1",
		"x2",
	}
	expected = "!(x1||x2)"
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

func TestIncidentSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	// Empty.
	scope := Scope{}
	selector := scope.incidentSelector()
	g.Expect("").To(gomega.Equal(selector))
	// Included.
	scope = Scope{}
	scope.Packages.Included = []string{"a", "b"}
	selector = scope.incidentSelector()
	g.Expect("(!package||package=a||package=b)").To(gomega.Equal(selector))
	// Excluded.
	scope = Scope{}
	scope.Packages.Excluded = []string{"C", "D"}
	selector = scope.incidentSelector()
	g.Expect("!(package||package=C||package=D)").To(gomega.Equal(selector))
	// Included and Excluded.
	scope = Scope{}
	scope.Packages.Included = []string{"a", "b"}
	scope.Packages.Excluded = []string{"C", "D"}
	selector = scope.incidentSelector()
	g.Expect("(!package||package=a||package=b) && !(package=C||package=D)").To(gomega.Equal(selector))
}

func TestInjectorDefaults(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	inj := ResourceInjector{}
	inj.dict = make(map[string]any)
	r := &Resource{
		Fields: []Field{
			{
				Name:    "Name",
				Key:     "person.name",
				Default: "Elmer",
			},
			{
				Name: "Age",
				Key:  "person.age",
			},
		},
	}
	err := inj.addDefaults(r)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(inj.dict[r.Fields[0].Key]).To(gomega.Equal(r.Fields[0].Default))
	g.Expect(inj.dict[r.Fields[1].Key]).To(gomega.BeNil())
}

func TestInjectorTypeCast(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	inj := ResourceInjector{}
	inj.dict = make(map[string]any)
	r := &Resource{
		Fields: []Field{
			{
				Name:    "Name",
				Key:     "person.name",
				Default: "Elmer",
			},
			{
				Name:    "Age",
				Key:     "person.age",
				Type:    "integer",
				Default: "18",
			},
			{
				Name:    "Resident",
				Key:     "person.resident",
				Type:    "boolean",
				Default: "true",
			},
			{
				Name:    "Member",
				Key:     "person.member",
				Type:    "boolean",
				Default: 1,
			},
			{
				Name:    "One",
				Key:     "person.one",
				Type:    "integer",
				Default: true,
			},
		},
	}
	err := inj.addDefaults(r)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(inj.dict[r.Fields[0].Key]).To(gomega.Equal(r.Fields[0].Default))
	g.Expect(inj.dict[r.Fields[1].Key]).To(gomega.Equal(18))
	g.Expect(inj.dict[r.Fields[2].Key]).To(gomega.BeTrue())
	g.Expect(inj.dict[r.Fields[3].Key]).To(gomega.BeTrue())
	g.Expect(inj.dict[r.Fields[4].Key]).To(gomega.Equal(1))

	// cast error.
	inj.dict = make(map[string]any)
	r.Fields = append(
		r.Fields,
		Field{
			Name:    "Resident",
			Key:     "person.parent",
			Type:    "integer",
			Default: "true",
		})
	err = inj.addDefaults(r)
	g.Expect(errors.Is(err, &TypeError{})).To(gomega.BeTrue())
}

func TestInject(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	key := "location"
	path := "/tmp/x"
	inj := Injector{}
	inj.Use(make(map[string]any))
	inj.dict[key] = path
	md := &Metadata{}
	md.Provider.InitConfig = []provider.InitConfig{
		{Location: "$(" + key + ")"},
	}
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(md.Provider.InitConfig[0].Location).To(gomega.Equal(path))
}

func TestRawInject(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	key := "location"
	path := "/tmp/x"
	inj := Injector{}
	inj.Use(make(map[string]any))
	inj.dict[key] = path
	md := map[string]any{
		"Location": "$(" + key + ")",
	}
	md2 := inj.inject(md).(map[string]any)
	g.Expect(md2["Location"]).To(gomega.Equal(path))
}
