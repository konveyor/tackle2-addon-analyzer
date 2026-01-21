package main

import (
	"errors"
	path2 "path"
	"strconv"
	"testing"

	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
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

func TestRulesetLabelMatch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	allRuleSets := []api.RuleSet{
		{
			Name: "target a",
			Rules: []api.Rule{
				{
					Name:   "target a",
					Labels: []string{"konveyor.io/target=a"},
				},
			},
		},
		{
			Name: "target b",
			Rules: []api.Rule{
				{
					Name:   "target b",
					Labels: []string{"konveyor.io/target=b"},
				},
			},
		},
		{
			Name: "no target",
			Rules: []api.Rule{
				{
					Name:   "no target",
					Labels: []string{"konveyor.io/source=b"},
				},
			},
		},
		{
			Name: "thing 4+",
			Rules: []api.Rule{
				{
					Name:   "4.0+",
					Labels: []string{"konveyor.io/thing=thing4.0+"},
				},
			},
		},
		{
			Name: "thing 4-",
			Rules: []api.Rule{
				{
					Name:   "4.0-",
					Labels: []string{"konveyor.io/thing=thing4.0-"},
				},
			},
		},
	}

	l := Labels{
		Included: []string{"konveyor.io/target=a"},
		Excluded: []string{},
	}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "target a",
			Rules: []api.Rule{
				{
					Name:   "target a",
					Labels: []string{"konveyor.io/target=a"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/target=a", "konveyor.io/target=b"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "target a",
			Rules: []api.Rule{
				{
					Name:   "target a",
					Labels: []string{"konveyor.io/target=a"},
				},
			},
		},
		{
			Name: "target b",
			Rules: []api.Rule{
				{
					Name:   "target b",
					Labels: []string{"konveyor.io/target=b"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/target"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "target a",
			Rules: []api.Rule{
				{
					Name:   "target a",
					Labels: []string{"konveyor.io/target=a"},
				},
			},
		},
		{
			Name: "target b",
			Rules: []api.Rule{
				{
					Name:   "target b",
					Labels: []string{"konveyor.io/target=b"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/thing=thing4"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "thing 4+",
			Rules: []api.Rule{
				{
					Name:   "4.0+",
					Labels: []string{"konveyor.io/thing=thing4.0+"},
				},
			},
		},
		{
			Name: "thing 4-",
			Rules: []api.Rule{
				{
					Name:   "4.0-",
					Labels: []string{"konveyor.io/thing=thing4.0-"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/thing=thing5"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "thing 4+",
			Rules: []api.Rule{
				{
					Name:   "4.0+",
					Labels: []string{"konveyor.io/thing=thing4.0+"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/thing=thing4.1"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "thing 4+",
			Rules: []api.Rule{
				{
					Name:   "4.0+",
					Labels: []string{"konveyor.io/thing=thing4.0+"},
				},
			},
		},
	}))
	l.Included = []string{"konveyor.io/thing=thing3"}
	g.Expect(l.selectedRuleSets(allRuleSets)).To(gomega.Equal([]api.RuleSet{
		{
			Name: "thing 4-",
			Rules: []api.Rule{
				{
					Name:   "4.0-",
					Labels: []string{"konveyor.io/thing=thing4.0-"},
				},
			},
		},
	}))

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

func TestProfile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the restClient.
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		GetFn: func(path string, object any, _ ...binding.Param) (err error) {
			_, idStr := path2.Split(path)
			nStr, _ := strconv.Atoi(idStr)
			id := uint(nStr)
			switch r := object.(type) {
			case *api.Task:
			case *api.AnalysisProfile:
				r.ID = id
				r.Name = "Test"
				r.Mode = api.ApMode{
					WithDeps: true,
				}
				r.Scope = api.ApScope{
					WithKnownLibs: true,
					Packages: api.InExList{
						Included: []string{
							"pA",
							"pB",
						},
						Excluded: []string{
							"pC",
							"pD",
						},
					},
				}
				r.Rules = api.ApRules{
					Targets: []api.ApTargetRef{
						{ID: 1},
						{ID: 2},
					},
					Labels: api.InExList{
						Included: []string{
							"tA",
							"tB",
						},
						Excluded: []string{
							"tC",
							"tD",
						},
					},
					Files: []api.Ref{
						{ID: 10},
						{ID: 20},
					},
					Repository: &api.Repository{
						URL:    "http://rules.com/pub",
						Branch: "master",
						Path:   "test",
					},
					Identity: &api.Ref{
						ID: 30,
					},
				}
			case *api.Target:
				switch id {
				case 1:
					r.ID = id
					r.Name = "TargetA"
					r.Labels = []api.TargetLabel{
						{Label: "konveyor.io/target=A"},
						{Label: "konveyor.io/target=B"},
						{Label: "konveyor.io/target=C"},
					}
				case 2:
					r.ID = id
					r.Name = "TargetB"
					r.Labels = []api.TargetLabel{
						{Label: "konveyor.io/target=D"},
						{Label: "konveyor.io/target=E"},
						{Label: "konveyor.io/target=F"},
					}
				default:
					err = &binding.NotFound{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
		PostFn: func(path string, object any) (err error) {
			switch object.(type) {
			case *api.TaskReport:
			default:
				err = &binding.NotFound{}
			}
			return
		},
		PutFn: func(path string, object any, _ ...binding.Param) (err error) {
			switch object.(type) {
			case *api.TaskReport:
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})
	addon.Use(richClient)
	addon.Load()

	// Test
	d := Data{}
	d.Profile = api.Ref{ID: 1}
	err := applyProfile(&d)
	g.Expect(err).To(gomega.BeNil())

	// Validate the profile has been applied to the Data.
	d2 := Data{}
	d2.Profile = api.Ref{ID: 1}
	d2.Mode.WithDeps = true
	d2.Scope.WithKnownLibs = true
	d2.Scope.Packages.Included = []string{
		"pA",
		"pB",
	}
	d2.Scope.Packages.Excluded = []string{
		"pC",
		"pD",
	}
	d2.Rules.Repository = &api.Repository{
		URL:    "http://rules.com/pub",
		Branch: "master",
		Path:   "test",
	}
	d2.Rules.Labels.Included = []string{
		"konveyor.io/target=A",
		"konveyor.io/target=B",
		"konveyor.io/target=C",
		"konveyor.io/target=D",
		"konveyor.io/target=E",
		"konveyor.io/target=F",
		"tA",
		"tB",
	}
	d2.Rules.Labels.Excluded = []string{
		"tC",
		"tD",
	}
	d2.Rules.Identity = &api.Ref{
		ID: 30,
	}
	d2.Rules.ruleFiles = []api.Ref{
		{ID: 10},
		{ID: 20},
	}
	g.Expect(d2).To(gomega.Equal(d))
}
