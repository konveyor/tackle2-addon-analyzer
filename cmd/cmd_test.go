package main

import (
	"errors"
	"os"
	path2 "path"
	"path/filepath"
	"strconv"
	"strings"
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

	var fetched struct {
		targets []uint
		files   []uint
	}

	// Mock the restClient.
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, _ ...binding.Param) (err error) {
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
					fetched.targets = append(fetched.targets, id)
				case 2:
					r.ID = id
					r.Name = "TargetB"
					r.Labels = []api.TargetLabel{
						{Label: "konveyor.io/target=D"},
						{Label: "konveyor.io/target=E"},
						{Label: "konveyor.io/target=F"},
					}
					fetched.targets = append(fetched.targets, id)
				default:
					err = &binding.NotFound{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
		DoFileGet: func(path, destination string) (err error) {
			_, idStr := path2.Split(path)
			nStr, _ := strconv.Atoi(idStr)
			id := uint(nStr)
			switch id {
			case 10, 20:
				fetched.files = append(fetched.files, id)
			default:
				err = &binding.NotFound{}
			}
			return
		},
		DoPost: func(path string, object any) (err error) {
			switch object.(type) {
			case *api.TaskReport:
			default:
				err = &binding.NotFound{}
			}
			return
		},
		DoPut: func(path string, object any, _ ...binding.Param) (err error) {
			switch object.(type) {
			case *api.TaskReport:
			default:
				err = &binding.NotFound{}
			}
			return
		},
		DoIsDir: func(path string, must bool) (isDir bool, err error) {
			return
		},
	})
	addon.Use(richClient)
	addon.Load()

	// Test profile applied.
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
	g.Expect(fetched.targets).To(gomega.Equal([]uint{1, 2}))

	// Test files fetched.
	err = d2.Rules.addFiles()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(fetched.files).To(gomega.Equal([]uint{10, 20}))
}

func TestResourceInjectorWithIdentity(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Setup temp directory for file writes
	tmpDir := t.TempDir()

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Mock identity search for maven kind
				if strings.Contains(path, "identities") {
					// Check if looking for direct identity (application-specific)
					if strings.Contains(path, "applications/1/identities") {
						// Return empty list for direct search
						*r = []api.Identity{}
					} else {
						// Return default maven identity for indirect search
						*r = []api.Identity{
							{
								Resource: api.Resource{ID: 10},
								Kind:     "maven",
								Name:     "maven-creds",
								User:     "testuser",
								Settings: `<?xml version="1.0"?>
<settings>
  <servers>
    <server>
      <id>test</id>
    </server>
  </servers>
</settings>`,
							},
						}
					}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with identity resource selector
	md := &Metadata{
		Provider: provider.Config{
			Address: "localhost:8080",
			InitConfig: []provider.InitConfig{
				{
					ProviderSpecificConfig: map[string]any{
						"mavenSettingsFile": "$(maven.settings.path)",
					},
				},
			},
		},
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "settings",
						Key:  "maven.settings.path",
						Path: filepath.Join(tmpDir, "maven-settings.xml"),
					},
					{
						Name: "user",
						Key:  "maven.user",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify settings file was written
	settingsPath := filepath.Join(tmpDir, "maven-settings.xml")
	content, err := os.ReadFile(settingsPath)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(content)).To(gomega.ContainSubstring("<settings>"))

	// Verify user was injected
	g.Expect(inj.dict["maven.user"]).To(gomega.Equal("testuser"))

	// Verify path was injected
	g.Expect(inj.dict["maven.settings.path"]).To(gomega.Equal(settingsPath))

	// Verify variable substitution in provider config
	g.Expect(md.Provider.InitConfig[0].ProviderSpecificConfig["mavenSettingsFile"]).To(gomega.Equal(settingsPath))
}

func TestResourceInjectorWithSetting(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *any:
				// Mock setting value retrieval
				if strings.Contains(path, "settings/mvn.insecure.enabled") {
					*r = "true"
				} else {
					err = &binding.NotFound{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with setting resource selector
	md := &Metadata{
		Provider: provider.Config{
			Address: "localhost:8080",
			InitConfig: []provider.InitConfig{
				{
					ProviderSpecificConfig: map[string]any{
						"mavenInsecure": "$(maven.insecure)",
					},
				},
			},
		},
		Resources: []Resource{
			{
				Selector: "setting:key=mvn.insecure.enabled",
				Fields: []Field{
					{
						Name: "value",
						Key:  "maven.insecure",
						Type: "boolean",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify boolean value was cast and injected
	g.Expect(inj.dict["maven.insecure"]).To(gomega.BeTrue())

	// Verify variable substitution in provider config
	g.Expect(md.Provider.InitConfig[0].ProviderSpecificConfig["mavenInsecure"]).To(gomega.BeTrue())
}

func TestResourceInjectorWithDefaults(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Return empty list - no identities found
				*r = []api.Identity{}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with default values
	md := &Metadata{
		Provider: provider.Config{
			Address: "localhost:$(port)",
			InitConfig: []provider.InitConfig{
				{
					ProviderSpecificConfig: map[string]any{
						"timeout": "$(timeout)",
					},
				},
			},
		},
		Resources: []Resource{
			{
				Selector: "identity:kind=nonexistent",
				Fields: []Field{
					{
						Name:    "port",
						Key:     "port",
						Default: 8080,
					},
					{
						Name:    "timeout",
						Key:     "timeout",
						Type:    "integer",
						Default: "30",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify defaults were used (defaults keep their original type, not converted to float64)
	g.Expect(inj.dict["port"]).To(gomega.Equal(8080))
	g.Expect(inj.dict["timeout"]).To(gomega.Equal(30))

	// Verify variable substitution with defaults
	g.Expect(md.Provider.Address).To(gomega.Equal("localhost:8080"))
	g.Expect(md.Provider.InitConfig[0].ProviderSpecificConfig["timeout"]).To(gomega.Equal(float64(30)))
}

func TestResourceInjectorFieldNotMatched(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Mock identity without the requested field
				if strings.Contains(path, "identities") && !strings.Contains(path, "applications/") {
					*r = []api.Identity{
						{
							Resource: api.Resource{ID: 10},
							Kind:     "maven",
							Name:     "maven-creds",
							// Note: no "nonexistent" field
						},
					}
				} else {
					*r = []api.Identity{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata requesting a field that doesn't exist
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "nonexistent",
						Key:  "test.key",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &FieldNotMatched{})).To(gomega.BeTrue())
}

func TestResourceInjectorSelectorNotSupported(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, _ ...binding.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with unsupported selector
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "unsupported:key=value",
				Fields: []Field{
					{
						Name: "field",
						Key:  "test.key",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &SelectorNotSupported{})).To(gomega.BeTrue())
}

func TestResourceInjectorKeyConflict(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Return maven identity
				if strings.Contains(path, "identities") && !strings.Contains(path, "applications/") {
					*r = []api.Identity{
						{
							Resource: api.Resource{ID: 10},
							Kind:     "maven",
							User:     "testuser",
						},
					}
				} else {
					*r = []api.Identity{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with duplicate key
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "duplicate.key",
					},
				},
			},
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "duplicate.key", // Same key as above
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &KeyConflictError{})).To(gomega.BeTrue())
}

func TestResourceInjectorMultipleResources(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Mock identity search - check path and params to determine which identity to return
				if strings.Contains(path, "applications/1/identities") {
					// Direct search - return empty
					*r = []api.Identity{}
				} else if strings.Contains(path, "identities") {
					// Indirect search - check filter params for kind
					var kind string
					for _, p := range params {
						if p.Key == "filter" {
							if strings.Contains(p.Value, "kind='maven'") {
								kind = "maven"
							} else if strings.Contains(p.Value, "kind='source'") {
								kind = "source"
							}
						}
					}

					switch kind {
					case "maven":
						*r = []api.Identity{
							{
								Resource: api.Resource{ID: 10},
								Kind:     "maven",
								User:     "maven-user",
								Default:  true,
							},
						}
					case "source":
						*r = []api.Identity{
							{
								Resource: api.Resource{ID: 20},
								Kind:     "source",
								User:     "git-user",
								Default:  true,
							},
						}
					default:
						*r = []api.Identity{}
					}
				} else {
					*r = []api.Identity{}
				}
			case *any:
				// Mock setting value retrieval
				if strings.Contains(path, "settings/app.timeout") {
					*r = "60" // Return as string for type casting
				} else {
					err = &binding.NotFound{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with multiple resources
	// This metadata references both seeded builtins and resource-injected variables
	// Also includes undefined variable references to test they're ignored
	md := &Metadata{
		Provider: provider.Config{
			Address: "$(builtin.protocol)://$(builtin.host):$(builtin.port)",
			InitConfig: []provider.InitConfig{
				{
					Location: "$(undefined.path)",
					ProviderSpecificConfig: map[string]any{
						"mavenUser":   "$(maven.user)",
						"gitUser":     "$(git.user)",
						"timeout":     "$(app.timeout)",
						"defaultPort": "$(default.port)",
						"environment": "$(builtin.env)",
					},
				},
			},
		},
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "maven.user",
					},
				},
			},
			{
				Selector: "identity:kind=source",
				Fields: []Field{
					{
						Name: "user",
						Key:  "git.user",
					},
				},
			},
			{
				Selector: "setting:key=app.timeout",
				Fields: []Field{
					{
						Name: "value",
						Key:  "app.timeout",
						Type: "integer",
					},
				},
			},
			{
				Selector: "identity:kind=nonexistent",
				Fields: []Field{
					{
						Name:    "port",
						Key:     "default.port",
						Default: 9090,
					},
				},
			},
		},
	}

	// Execute injection with seeded builtins
	// This tests the complete flow: Use() seeds dict, then Inject() augments with build()
	inj := ResourceInjector{}
	builtins := map[string]any{
		"builtin.protocol": "https",
		"builtin.host":     "example.com",
		"builtin.port":     8443,
		"builtin.env":      "production",
	}
	inj.Use(builtins)
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify seeded builtins remain in dict
	g.Expect(inj.dict["builtin.protocol"]).To(gomega.Equal("https"))
	g.Expect(inj.dict["builtin.host"]).To(gomega.Equal("example.com"))
	g.Expect(inj.dict["builtin.port"]).To(gomega.Equal(8443))
	g.Expect(inj.dict["builtin.env"]).To(gomega.Equal("production"))

	// Verify all resources were injected (augmented by build())
	g.Expect(inj.dict["maven.user"]).To(gomega.Equal("maven-user"))
	g.Expect(inj.dict["git.user"]).To(gomega.Equal("git-user"))
	g.Expect(inj.dict["app.timeout"]).To(gomega.Equal(60))
	g.Expect(inj.dict["default.port"]).To(gomega.Equal(9090))

	// Verify variable substitution for seeded builtins
	g.Expect(md.Provider.Address).To(gomega.Equal("https://example.com:8443"))

	// Verify variable substitution for resource-injected values (may be float64 after JSON round-trip)
	config := md.Provider.InitConfig[0].ProviderSpecificConfig
	g.Expect(config["mavenUser"]).To(gomega.Equal("maven-user"))
	g.Expect(config["gitUser"]).To(gomega.Equal("git-user"))
	g.Expect(config["timeout"]).To(gomega.Equal(float64(60)))
	g.Expect(config["defaultPort"]).To(gomega.Equal(float64(9090)))
	g.Expect(config["environment"]).To(gomega.Equal("production"))

	// Verify undefined variable references are left unchanged (not substituted)
	// This protects against cases where $(var) pattern occurs naturally in metadata
	g.Expect(md.Provider.InitConfig[0].Location).To(gomega.Equal("$(undefined.path)"))
}

func TestResourceInjectorComplexVariableSubstitution(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Test complex variable substitution scenarios
	inj := Injector{}
	inj.Use(map[string]any{
		"host":     "example.com",
		"port":     8080,
		"protocol": "https",
		"enabled":  true,
	})

	// Test substitution in nested structures
	md := &Metadata{
		Provider: provider.Config{
			Address: "$(protocol)://$(host):$(port)",
			InitConfig: []provider.InitConfig{
				{
					Location: "/path/to/config",
					ProviderSpecificConfig: map[string]any{
						"url":     "$(protocol)://$(host):$(port)/api",
						"enabled": "$(enabled)",
						"nested": map[string]any{
							"server": "$(host)",
						},
						"list": []any{
							"$(host)",
							"$(port)",
						},
					},
				},
			},
		},
	}

	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify complex substitutions
	g.Expect(md.Provider.Address).To(gomega.Equal("https://example.com:8080"))
	config := md.Provider.InitConfig[0].ProviderSpecificConfig
	g.Expect(config["url"]).To(gomega.Equal("https://example.com:8080/api"))
	g.Expect(config["enabled"]).To(gomega.BeTrue())

	// Verify nested map
	nested, ok := config["nested"].(map[string]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(nested["server"]).To(gomega.Equal("example.com"))

	// Verify list (numbers become float64 after JSON round-trip)
	list, ok := config["list"].([]any)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(list[0]).To(gomega.Equal("example.com"))
	g.Expect(list[1]).To(gomega.Equal(float64(8080)))
}

func TestParsedSelector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Test basic selector
	p := ParsedSelector{}
	p.With("identity:kind=maven")
	g.Expect(p.kind).To(gomega.Equal("identity"))
	g.Expect(p.name).To(gomega.Equal("kind"))
	g.Expect(p.value).To(gomega.Equal("maven"))
	g.Expect(p.ns).To(gomega.Equal(""))

	// Test selector with namespace
	p = ParsedSelector{}
	p.With("default/identity:kind=maven")
	g.Expect(p.ns).To(gomega.Equal("default"))
	g.Expect(p.kind).To(gomega.Equal("identity"))
	g.Expect(p.name).To(gomega.Equal("kind"))
	g.Expect(p.value).To(gomega.Equal("maven"))

	// Test selector without value
	p = ParsedSelector{}
	p.With("setting:key")
	g.Expect(p.kind).To(gomega.Equal("setting"))
	g.Expect(p.name).To(gomega.Equal("key"))
	g.Expect(p.value).To(gomega.Equal(""))

	// Test minimal selector
	p = ParsedSelector{}
	p.With("name")
	g.Expect(p.kind).To(gomega.Equal(""))
	g.Expect(p.name).To(gomega.Equal("name"))
	g.Expect(p.value).To(gomega.Equal(""))
}

func TestResourceInjectorApplicationError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client to return error when getting application
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				// Task loads successfully with application reference
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				// But getting the full application fails
				err = errors.New("failed to retrieve application")
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with identity resource
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "maven.user",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("failed to retrieve application"))
}

func TestResourceInjectorIdentitySearchError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Return error when searching for identities
				if strings.Contains(path, "identities") {
					err = errors.New("identity search failed")
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with identity resource
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "maven.user",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("identity search failed"))
}

func TestResourceInjectorDirectIdentityMatch(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Return identity for direct (application-specific) search
				if strings.Contains(path, "applications/1/identities") {
					// Check if looking for maven role
					for _, p := range params {
						if p.Key == "filter" && strings.Contains(p.Value, "role='maven'") {
							*r = []api.Identity{
								{
									Resource: api.Resource{ID: 15},
									Kind:     "maven",
									Name:     "app-specific-maven",
									User:     "direct-user",
								},
							}
							return
						}
					}
				}
				// Return empty for indirect search (should not be reached)
				*r = []api.Identity{}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata with identity selector using role (direct match)
	md := &Metadata{
		Provider: provider.Config{
			Address: "localhost:8080",
			InitConfig: []provider.InitConfig{
				{
					ProviderSpecificConfig: map[string]any{
						"mavenUser": "$(maven.user)",
					},
				},
			},
		},
		Resources: []Resource{
			{
				Selector: "identity:role=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "maven.user",
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).To(gomega.BeNil())

	// Verify direct identity was used
	g.Expect(inj.dict["maven.user"]).To(gomega.Equal("direct-user"))

	// Verify variable substitution
	g.Expect(md.Provider.InitConfig[0].ProviderSpecificConfig["mavenUser"]).To(gomega.Equal("direct-user"))
}

func TestResourceInjectorTypeCastError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// Mock the hub client
	richClient := binding.New("")
	richClient.Use(&client.Stub{
		DoGet: func(path string, object any, params ...client.Param) (err error) {
			switch r := object.(type) {
			case *api.Task:
				r.Application = &api.Ref{ID: 1}
			case *api.Application:
				r.ID = 1
				r.Name = "TestApp"
			case *[]api.Identity:
				// Return identity with non-numeric value
				if strings.Contains(path, "identities") && !strings.Contains(path, "applications/") {
					*r = []api.Identity{
						{
							Resource: api.Resource{ID: 10},
							Kind:     "maven",
							Name:     "maven-creds",
							User:     "not-a-number", // This will fail integer cast
						},
					}
				} else {
					*r = []api.Identity{}
				}
			default:
				err = &binding.NotFound{}
			}
			return
		},
	})

	addon.Use(richClient)
	addon.Load()

	// Create metadata requesting integer type for user field
	md := &Metadata{
		Resources: []Resource{
			{
				Selector: "identity:kind=maven",
				Fields: []Field{
					{
						Name: "user",
						Key:  "maven.user",
						Type: "integer", // Request integer but value is string
					},
				},
			},
		},
	}

	// Execute injection
	inj := ResourceInjector{}
	err := inj.Inject(md)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &TypeError{})).To(gomega.BeTrue())
	g.Expect(err.Error()).To(gomega.ContainSubstring("cast failed"))
}
