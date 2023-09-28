package main

import (
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"os"
	"path"
	"strconv"
	"strings"
)

type History = map[uint]byte

//
// Rules settings.
type Rules struct {
	Path       string          `json:"path"`
	Repository *api.Repository `json:"repository"`
	Identity   *api.Ref        `json:"identity"`
	Labels     Labels          `json:"labels"`
	rules      []string
}

//
// Build assets.
func (r *Rules) Build() (err error) {
	err = r.addFiles()
	if err != nil {
		return
	}
	err = r.addRepository()
	if err != nil {
		return
	}
	err = r.addRuleSets()
	if err != nil {
		return
	}
	err = r.convert()
	if err != nil {
		return
	}
	return
}

//
// AddOptions adds analyzer options.
func (r *Rules) AddOptions(options *command.Options) (err error) {
	for _, path := range r.rules {
		options.Add("--rules", path)
	}
	err = r.addSelector(options)
	if err != nil {
		return
	}
	return
}

//
// addFiles add uploaded rules files.
func (r *Rules) addFiles() (err error) {
	if r.Path == "" {
		return
	}
	ruleDir := path.Join(RuleDir, "/files")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	r.rules = append(r.rules, ruleDir)
	bucket := addon.Bucket()
	err = bucket.Get(r.Path, ruleDir)
	if err != nil {
		return
	}
	return
}

//
// addRuleSets adds rulesets and their dependencies.
func (r *Rules) addRuleSets() (err error) {
	history := make(History)
	ruleSets, err := r.Labels.ruleSets()
	if err != nil {
		return
	}
	for _, ruleSet := range ruleSets {
		if _, found := history[ruleSet.ID]; found {
			continue
		}
		addon.Activity(
			"[RULESET] fetching: id=%d (%s)",
			ruleSet.ID,
			ruleSet.Name)
		history[ruleSet.ID] = 0
		err = r.addRules(&ruleSet)
		if err != nil {
			return
		}
		err = r.addRuleSetRepository(&ruleSet)
		if err != nil {
			return
		}
		err = r.addDeps(&ruleSet, history)
		if err != nil {
			return
		}
	}
	return
}

//
// addDeps adds ruleSet dependencies.
func (r *Rules) addDeps(ruleSet *api.RuleSet, history History) (err error) {
	for _, ref := range ruleSet.DependsOn {
		if _, found := history[ref.ID]; found {
			continue
		}
		history[ref.ID] = 0
		var ruleSet *api.RuleSet
		ruleSet, err = addon.RuleSet.Get(ref.ID)
		if err != nil {
			return
		}
		addon.Activity(
			"[RULESET] fetching (dep): id=%d (%s)",
			ruleSet.ID,
			ruleSet.Name)
		err = r.addRules(ruleSet)
		if err != nil {
			return
		}
		err = r.addRuleSetRepository(ruleSet)
		if err != nil {
			return
		}
		err = r.addDeps(ruleSet, history)
		if err != nil {
			return
		}
	}
	return
}

//
// addRules adds rules
func (r *Rules) addRules(ruleset *api.RuleSet) (err error) {
	ruleDir := path.Join(
		RuleDir,
		"/rulesets",
		strconv.Itoa(int(ruleset.ID)),
		"rules")
	err = nas.MkDir(ruleDir, 0755)
	if err != nil {
		return
	}
	n := len(ruleset.Rules)
	for _, ruleset := range ruleset.Rules {
		fileRef := ruleset.File
		if fileRef == nil {
			continue
		}
		path := path.Join(ruleDir, fileRef.Name)
		err = addon.File.Get(ruleset.File.ID, path)
		if err != nil {
			break
		}
		if n == 1 {
			r.rules = append(r.rules, path)
		}
	}
	if n > 1 {
		r.rules = append(r.rules, ruleDir)
	}
	return
}

//
// addRuleSetRepository adds ruleset repository.
func (r *Rules) addRuleSetRepository(ruleset *api.RuleSet) (err error) {
	if ruleset.Repository == nil {
		return
	}
	rootDir := path.Join(
		RuleDir,
		"/rulesets",
		strconv.Itoa(int(ruleset.ID)),
		"repository")
	err = nas.MkDir(rootDir, 0755)
	if err != nil {
		return
	}
	var ids []api.Ref
	if ruleset.Identity != nil {
		ids = []api.Ref{*ruleset.Identity}
	}
	rp, err := repository.New(
		rootDir,
		ruleset.Repository,
		ids)
	if err != nil {
		return
	}
	err = rp.Fetch()
	if err != nil {
		return
	}
	ruleDir := path.Join(rootDir, ruleset.Repository.Path)
	r.rules = append(r.rules, ruleDir)
	return
}

//
// addRepository adds custom repository.
func (r *Rules) addRepository() (err error) {
	if r.Repository == nil {
		return
	}
	rootDir := path.Join(
		RuleDir,
		"repository")
	err = nas.MkDir(rootDir, 0755)
	if err != nil {
		return
	}
	var ids []api.Ref
	if r.Identity != nil {
		ids = []api.Ref{*r.Identity}
	}
	rp, err := repository.New(
		rootDir,
		r.Repository,
		ids)
	if err != nil {
		return
	}
	err = rp.Fetch()
	if err != nil {
		return
	}
	ruleDir := path.Join(rootDir, r.Repository.Path)
	r.rules = append(r.rules, ruleDir)
	return
}

//
// addSelector adds label selector.
func (r *Rules) addSelector(options *command.Options) (err error) {
	ruleSelector := RuleSelector{Included: r.Labels.Included}
	selector := ruleSelector.String()
	if selector != "" {
		options.Add("--label-selector", selector)
	}
	return
}

//
// convert windup rules.
func (r *Rules) convert() (err error) {
	output := path.Join(RuleDir, "converted")
	cmd := command.Command{Path: "/usr/bin/windup-shim"}
	cmd.Options.Add("convert")
	cmd.Options.Add("--outputdir", output)
	cmd.Options.Add(RuleDir)
	err = cmd.Run()
	if err != nil {
		return
	}
	converted, err := os.ReadDir(output)
	if err != nil {
		return
	}
	if len(converted) > 0 {
		r.rules = append(r.rules, output)
	}
	return
}

//
// Labels collection.
type Labels struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

//
// ruleSets returns list of ruleSets with these labels.
func (r *Labels) ruleSets() (matched []api.RuleSet, err error) {
	var found []api.RuleSet
	for _, name := range r.Included {
		f := hub.Filter{}
		f.And("Labels").Eq(name)
		found, err = addon.RuleSet.Find(f)
		if err == nil {
			matched = append(matched, found...)
		} else {
			return
		}
	}
	return
}

//
// Label formatted labels.
// Formats:
// - name
// - name=value
// - namespace/name
// - namespace/name=value
type Label string

//
// Namespace returns the (optional) namespace.
func (r *Label) Namespace() (ns string) {
	s := string(*r)
	part := strings.Split(s, "/")
	if len(part) > 1 {
		ns = part[0]
	}
	return
}

//
// Name returns the name.
func (r *Label) Name() (n string) {
	s := string(*r)
	_, s = path.Split(s)
	n = strings.Split(s, "=")[0]
	return
}

//
// Value returns the (optional) value.
func (r *Label) Value() (v string) {
	s := string(*r)
	_, s = path.Split(s)
	part := strings.SplitN(s, "=", 2)
	if len(part) == 2 {
		v = part[1]
	}
	return
}

//
// RuleSelector - Label-based rule selector.
type RuleSelector struct {
	Included []string
	Excluded []string
}

//
// String returns string representation.
func (r *RuleSelector) String() (selector string) {
	var other, sources, targets []string
	for _, s := range r.unique(r.Included) {
		label := Label(s)
		if label.Namespace() != "konveyor.io" {
			other = append(other, s)
			continue
		}
		switch label.Name() {
		case "source":
			sources = append(sources, s)
		case "target":
			targets = append(targets, s)
		default:
			other = append(other, s)
		}
	}
	var ands []string
	ands = append(ands, r.join("||", sources...))
	ands = append(ands, r.join("||", targets...))
	selector = r.join("||", other...)
	selector = r.join("||", selector, r.join("&&", ands...))
	if strings.HasPrefix(selector, "((") {
		selector = selector[1 : len(selector)-1]
	}
	return
}

//
// join clauses.
func (r *RuleSelector) join(operator string, operands ...string) (joined string) {
	var packed []string
	for _, s := range operands {
		if len(s) > 0 {
			packed = append(packed, s)
		}
	}
	switch len(packed) {
	case 0:
	case 1:
		joined = strings.Join(packed, operator)
	default:
		joined = "(" + strings.Join(packed, operator) + ")"
	}
	return
}

//
// unique returns unique strings.
func (r *RuleSelector) unique(in []string) (out []string) {
	mp := make(map[string]int)
	for _, s := range in {
		if _, found := mp[s]; !found {
			out = append(out, s)
			mp[s] = 0
		}
	}
	return
}
