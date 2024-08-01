package main

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/konveyor/analyzer-lsp/parser"
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/repository"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/nas"
	"github.com/rogpeppe/go-internal/semver"
	"gopkg.in/yaml.v3"
)

const (
	// KonveyorIO namespace for labels.
	KonveyorIO = "konveyor.io"
)

type History = map[uint]byte

// LvRegex - Label value regex.
var LvRegex = regexp.MustCompile(`(\D+)(\d(?:[\d\.]*\d)?)([\+-])?$`)

// Rules settings.
type Rules struct {
	Path         string          `json:"path"`
	Repository   *api.Repository `json:"repository"`
	Identity     *api.Ref        `json:"identity"`
	Labels       Labels          `json:"labels"`
	RuleSets     []api.Ref       `json:"ruleSets"`
	repositories []string
	rules        []string
}

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
	err = r.Labels.injectAlways(r.repositories)
	if err != nil {
		return
	}
	return
}

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
	addon.Activity(
		"[RULESET] fetching: %s",
		r.Path)
	bucket := addon.Bucket()
	err = bucket.Get(r.Path, ruleDir)
	if err != nil {
		return
	}
	entries, err := os.ReadDir(ruleDir)
	if err != nil {
		return
	}
	for _, ent := range entries {
		if ent.Name() == parser.RULE_SET_GOLDEN_FILE_NAME {
			r.repositories = append(r.repositories, ruleDir)
			r.append(ruleDir)
			return
		}
	}
	n := 0
	for _, ent := range entries {
		p := path.Join(ruleDir, ent.Name())
		r.append(p)
		n++
	}
	if n > 0 {
		r.repositories = append(r.repositories, ruleDir)
	}
	return
}

// addRuleSets adds rulesets and their dependencies.
func (r *Rules) addRuleSets() (err error) {
	history := make(History)
	ruleSets := make([]api.RuleSet, 0)
	for _, ref := range r.RuleSets {
		var ruleSet *api.RuleSet
		ruleSet, err = addon.RuleSet.Get(ref.ID)
		if err != nil {
			return
		}
		ruleSets = append(
			ruleSets,
			*ruleSet)
	}
	matched, err := r.Labels.RuleSets()
	if err != nil {
		return
	}
	ruleSets = append(ruleSets, matched...)
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
			r.append(path)
		}
	}
	if n > 1 {
		r.append(ruleDir)
	}
	return
}

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
	r.repositories = append(r.repositories, ruleDir)
	r.append(ruleDir)
	return
}

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
	r.repositories = append(r.repositories, ruleDir)
	r.append(ruleDir)
	return
}

// addSelector adds label selector.
func (r *Rules) addSelector(options *command.Options) (err error) {
	ruleSelector := RuleSelector{Included: r.Labels.Included}
	selector := ruleSelector.String()
	if selector != "" {
		options.Add("--label-selector", selector)
	}
	return
}

// convert windup rules.
func (r *Rules) convert() (err error) {
	cmd := command.New("/usr/bin/windup-shim")
	cmd.Options.Add("convert")
	cmd.Options.Add("--outputdir", RuleDir)
	cmd.Options.Add(RuleDir)
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

// append path.
func (r *Rules) append(p string) {
	for i := range r.rules {
		if r.rules[i] == p {
			return
		}
	}
	switch strings.ToUpper(path.Ext(p)) {
	case "",
		".YAML",
		".YML":
		r.rules = append(r.rules, p)
	}
}

// Labels collection.
type Labels struct {
	Included []string `json:"included,omitempty"`
	Excluded []string `json:"excluded,omitempty"`
}

// RuleSets returns a list of ruleSets matching the 'included' labels.
func (r *Labels) RuleSets() (matched []api.RuleSet, err error) {
	mapped, err := r.ruleSetMap()
	if err != nil {
		return
	}
	for _, included := range r.Included {
		for rule, ruleSets := range mapped {
			if Label(rule).Match(Label(included)) {
				matched = append(
					matched,
					ruleSets...)
			}
		}
	}
	return
}

// ruleSetMap returns a populated RuleSetMap.
func (r *Labels) ruleSetMap() (mp RuleSetMap, err error) {
	mp = make(RuleSetMap)
	ruleSets, err := addon.RuleSet.List()
	if err != nil {
		return
	}
	for _, ruleSet := range ruleSets {
		for _, rule := range ruleSet.Rules {
			for i := range rule.Labels {
				mp[rule.Labels[i]] = append(
					mp[rule.Labels[i]],
					ruleSet)
			}
		}
	}
	return
}

// injectAlways - Replaces the labels in every rule file
// with konveyor.io/include=always.
func (r *Labels) injectAlways(paths []string) (err error) {
	read := func(m any, p string) (err error) {
		f, err := os.Open(p)
		if err != nil {
			return
		}
		defer func() {
			_ = f.Close()
		}()
		d := yaml.NewDecoder(f)
		err = d.Decode(m)
		return
	}
	write := func(m any, p string) (err error) {
		f, err := os.Create(p)
		if err != nil {
			return
		}
		defer func() {
			_ = f.Close()
		}()
		en := yaml.NewEncoder(f)
		err = en.Encode(m)
		return
	}
	inspect := func(p string, info fs.FileInfo, wErr error) (_ error) {
		var err error
		if wErr != nil || info.IsDir() {
			addon.Log.Error(wErr, p)
			return
		}
		switch strings.ToUpper(path.Ext(p)) {
		case "",
			".YAML",
			".YML":
		default:
			return
		}
		key := "labels"
		if path.Base(p) == parser.RULE_SET_GOLDEN_FILE_NAME {
			ruleSet := make(map[any]any)
			err = read(&ruleSet, p)
			if err != nil {
				return
			}
			ruleSet[key] = []string{"konveyor.io/include=always"}
			err = write(&ruleSet, p)
			if err != nil {
				return
			}
		} else {
			rules := make([]map[any]any, 0)
			err = read(&rules, p)
			if err != nil {
				return
			}
			for _, rule := range rules {
				rule[key] = []string{"konveyor.io/include=always"}
			}
			err = write(&rules, p)
			if err != nil {
				return
			}
		}
		return
	}
	ruleSelector := RuleSelector{Included: r.Included}
	selector := ruleSelector.String()
	if selector == "" {
		return
	}
	for _, ruleDir := range paths {
		err = filepath.Walk(ruleDir, inspect)
		if err != nil {
			return
		}
	}
	return
}

// RuleSetMap is a map of labels mapped to ruleSets with those labels.
type RuleSetMap map[string][]api.RuleSet

// Label formatted labels.
// Formats:
// - name
// - name=value
// - namespace/name
// - namespace/name=value
type Label string

// Namespace returns the (optional) namespace.
func (r Label) Namespace() (ns string) {
	s := string(r)
	part := strings.Split(s, "/")
	if len(part) > 1 {
		ns = part[0]
	}
	return
}

// Name returns the name.
func (r Label) Name() (n string) {
	s := string(r)
	_, s = path.Split(s)
	n = strings.Split(s, "=")[0]
	return
}

// Value returns the (optional) value.
func (r Label) Value() (v string) {
	s := string(r)
	_, s = path.Split(s)
	part := strings.SplitN(s, "=", 2)
	if len(part) == 2 {
		v = part[1]
	}
	return
}

// Match returns true when matched.
// Values may contain version expressions.
func (r Label) Match(other Label) (matched bool) {
	if r.Namespace() != other.Namespace() ||
		r.Name() != other.Name() {
		return
	}
	selfMatch := LvRegex.FindStringSubmatch(r.Value())
	otherMatch := LvRegex.FindStringSubmatch(other.Value())
	if len(selfMatch) != 4 {
		matched = r.Value() == other.Value()
		return
	}
	if len(otherMatch) != 4 {
		matched = selfMatch[1] == other.Value()
		return
	}
	if selfMatch[1] != otherMatch[1] {
		return
	}
	n := semver.Compare(selfMatch[2], otherMatch[2])
	switch selfMatch[3] {
	case "+":
		matched = n == 0 || n == 1
	case "-":
		matched = n == 0 || n == -1
	default:
		matched = n == 0
	}
	return
}

// Eq returns true when equal.
func (r Label) Eq(other Label) (matched bool) {
	matched = r.Namespace() == other.Namespace() &&
		r.Name() == other.Name() &&
		r.Value() == other.Value()
	return
}

// RuleSelector - Label-based rule selector.
type RuleSelector struct {
	Included []string
	Excluded []string
}

// String returns string representation.
func (r *RuleSelector) String() (selector string) {
	var other, sources, targets []string
	for _, s := range r.unique(r.Included) {
		label := Label(s)
		if label.Namespace() != KonveyorIO {
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
