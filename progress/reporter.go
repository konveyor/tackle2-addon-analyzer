package progress

import (
	"sync/atomic"
	"time"

	"github.com/konveyor/analyzer-lsp/progress"
	"github.com/konveyor/tackle2-hub/shared/addon/adapter"
)

type AddonReporter struct {
	events                chan progress.Event
	closed                bool
	droppedEvents         atomic.Uint64
	addon                 *adapter.Adapter
	lastReportedExecution *time.Time
}

func (a *AddonReporter) Report(event progress.Event) {

	switch event.Stage {
	case progress.StageProviderInit:
		if event.Message == "" {
			return
		}
		a.addon.Activity("[ANALYZER] %s", event.Message)
	case progress.StageRuleParsing:
		if event.Total > 0 {
			a.addon.Activity("[ANALYZER] Loaded %d rules", event.Total)
		}
	case progress.StageProviderPrepare:
		if event.Current != 0 || event.Total != event.Current {
			return
		}
		a.addon.Activity("[ANALYZER] %s", event.Message)
	case progress.StageRuleExecution:
		if event.Total == 0 {
			return
		}
		if event.Current == event.Total {
			a.addon.Activity("[ANALYZER] processed %d rules out of %d", event.Current, event.Total)
			// Just in case there is ever more than one analysis run per addon
			a.lastReportedExecution = nil
			return
		}
		if event.Current == 0 {
			a.addon.Activity("[ANALYZER] starting to process %v rules", event.Total)
			a.addon.Total(event.Total)
			t := time.Now()
			a.lastReportedExecution = &t
		} else if a.lastReportedExecution == nil || time.Since(*a.lastReportedExecution) >= 5*time.Second {
			a.addon.Activity("[ANALYZER] processed %d rules out of %d", event.Current, event.Total)
			a.addon.Completed(event.Current)
			t := time.Now()
			a.lastReportedExecution = &t
		}
	case progress.StageComplete:
		if event.Message == "" {
			return
		}
		a.addon.Activity("[ANALYZER] %s", event.Message)
	}

}

func NewAddonReporter(addon *adapter.Adapter) *AddonReporter {
	return &AddonReporter{
		events:        make(chan progress.Event),
		closed:        false,
		droppedEvents: atomic.Uint64{},
		addon:         addon,
	}
}
