package main

import (
	"math/rand"
	"regexp"

	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-hub/api"
)

var TagExp = regexp.MustCompile("(.+)(=)(.+)")

// Tagger tags an application.
type Tagger struct {
	Enabled bool `json:"enabled"`
}

// AddOptions adds analyzer options.
func (r *Tagger) AddOptions(options *command.Options) (err error) {
	return
}

// Update updates application tags.
//   - Ensures categories exist.
//   - Endures tags exist.
//   - Replaces associated tags (by source).
func (r *Tagger) Update(appID uint, tags []string) (err error) {
	addon.Activity("[TAG] Tagging Application %d.", appID)
	catMap, err := r.ensureCategories(tags)
	if err != nil {
		return
	}
	wanted, err := r.ensureTags(catMap, tags)
	if err != nil {
		return
	}
	err = r.ensureAssociated(appID, wanted)
	if err != nil {
		return
	}
	return
}

// ensureCategories ensures categories exist.
// Returns the map of category names to IDs.
func (r *Tagger) ensureCategories(tags []string) (catMap map[string]uint, err error) {
	catMap = map[string]uint{}
	wanted := []api.TagCategory{}
	for _, s := range tags {
		m := TagExp.FindStringSubmatch(s)
		if len(m) == 4 {
			catMap[m[1]] = 0
		}
	}
	for name := range catMap {
		wanted = append(
			wanted,
			api.TagCategory{
				Name: name,
				Rank: uint(rand.Intn(10)),
			})
	}
	for _, cat := range wanted {
		err = addon.TagCategory.Ensure(&cat)
		if err != nil {
			return
		}
		catMap[cat.Name] = cat.ID
	}
	return
}

// ensureTags ensures tags exist.
// Returns the wanted tag IDs.
func (r *Tagger) ensureTags(catMap map[string]uint, tags []string) (tagIds []uint, err error) {
	mp := map[TagRef]int{}
	wanted := []api.Tag{}
	for _, s := range tags {
		m := TagExp.FindStringSubmatch(s)
		if len(m) == 4 {
			ref := TagRef{
				Category: m[1],
				Name:     m[3],
			}
			mp[ref] = 0
		}
	}
	for ref := range mp {
		catRef := api.Ref{
			ID: catMap[ref.Category],
		}
		wanted = append(
			wanted,
			api.Tag{
				Name:     ref.Name,
				Category: catRef,
			})
	}
	for _, tag := range wanted {
		err = addon.Tag.Ensure(&tag)
		if err != nil {
			return
		}
		tagIds = append(tagIds, tag.ID)
	}
	return
}

// ensureAssociated ensure wanted tags are associated.
func (r *Tagger) ensureAssociated(appID uint, wanted []uint) (err error) {
	tags := addon.Application.Tags(appID)
	tags.Source(Source)
	err = tags.Replace(wanted)
	return
}

// TagRef -
type TagRef struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}
