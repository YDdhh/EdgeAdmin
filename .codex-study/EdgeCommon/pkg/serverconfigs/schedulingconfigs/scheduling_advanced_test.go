package schedulingconfigs

import (
	"testing"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

type advancedStats struct {
	latencies   map[string]float64
	outstanding map[string]uint64
}

func (this *advancedStats) CandidateLatencyMillis(candidate CandidateInterface, scope string) (float64, bool) {
	value, ok := this.latencies[candidateKey(candidate)]
	return value, ok
}

func (this *advancedStats) CandidateOutstandingRequests(candidate CandidateInterface) (uint64, bool) {
	value, ok := this.outstanding[candidateKey(candidate)]
	return value, ok
}

func newAdvancedCall(options maps.Map) *shared.RequestCall {
	call := shared.NewRequestCall()
	call.Options = options
	return call
}

func assertCandidateName(t *testing.T, candidate CandidateInterface, name string) {
	t.Helper()
	if candidate == nil {
		t.Fatalf("expected candidate %q, got nil", name)
	}
	testCandidate, ok := candidate.(*TestCandidate)
	if !ok {
		t.Fatalf("expected *TestCandidate, got %T", candidate)
	}
	if testCandidate.Name != name {
		t.Fatalf("expected candidate %q, got %q", name, testCandidate.Name)
	}
}

func TestDynamicLatencyScheduling_Next(t *testing.T) {
	s := &DynamicLatencyScheduling{}
	s.Add(&TestCandidate{Name: "a", Weight: 1})
	s.Add(&TestCandidate{Name: "b", Weight: 1})
	s.Start()

	call := newAdvancedCall(maps.Map{
		OptionRuntimeStats: &advancedStats{
			latencies: map[string]float64{
				"a": 80,
				"b": 20,
			},
		},
	})

	assertCandidateName(t, s.Next(call), "b")
}

func TestLeastOutstandingScheduling_Next(t *testing.T) {
	s := &LeastOutstandingScheduling{}
	s.Add(&TestCandidate{Name: "a", Weight: 1})
	s.Add(&TestCandidate{Name: "b", Weight: 10})
	s.Start()

	call := newAdvancedCall(maps.Map{
		OptionRuntimeStats: &advancedStats{
			outstanding: map[string]uint64{
				"a": 1,
				"b": 8,
			},
		},
	})

	assertCandidateName(t, s.Next(call), "b")
}

func TestGeoFailoverScheduling_Next(t *testing.T) {
	s := &GeoFailoverScheduling{}
	s.Add(&TestCandidate{Name: "us", Weight: 1})
	s.Add(&TestCandidate{Name: "sg", Weight: 1})
	s.Start()

	call := newAdvancedCall(maps.Map{
		OptionClientCountry: "CN",
		"countryMap": maps.Map{
			"cn": []string{"sg", "us"},
		},
		"default": []string{"us"},
	})

	assertCandidateName(t, s.Next(call), "sg")
}

func TestGeoProximityScheduling_Next(t *testing.T) {
	s := &GeoProximityScheduling{}
	s.Add(&TestCandidate{Name: "iad", Weight: 1})
	s.Add(&TestCandidate{Name: "sin", Weight: 1})
	s.Start()

	call := newAdvancedCall(maps.Map{
		OptionClientLatitude:  22.3193,
		OptionClientLongitude: 114.1694,
		"candidateLocations": maps.Map{
			"iad": "38.9531,-77.4565",
			"sin": "1.3521,103.8198",
		},
	})

	assertCandidateName(t, s.Next(call), "sin")
}

func TestLatencyPolicyScheduling_Next(t *testing.T) {
	s := &LatencyPolicyScheduling{}
	s.Add(&TestCandidate{Name: "us", Weight: 1})
	s.Add(&TestCandidate{Name: "sg", Weight: 1})
	s.Start()

	call := newAdvancedCall(maps.Map{
		OptionClientRegion: "apac",
		"latencyMatrix": maps.Map{
			"apac": maps.Map{
				"us": 180,
				"sg": 30,
			},
		},
	})

	assertCandidateName(t, s.Next(call), "sg")
}

func TestTrafficDialWeightedScheduling_Next(t *testing.T) {
	s := &TrafficDialWeightedScheduling{}
	s.Add(&TestCandidate{Name: "off", Weight: 100})
	s.Add(&TestCandidate{Name: "on", Weight: 1})
	s.Start()

	call := newAdvancedCall(maps.Map{
		"trafficDials": maps.Map{
			"off": 0,
			"on":  100,
		},
	})

	for i := 0; i < 100; i++ {
		assertCandidateName(t, s.Next(call), "on")
	}
}

func TestFindSchedulingType_Advanced(t *testing.T) {
	for _, code := range []string{
		"dynamicLatency",
		"leastOutstanding",
		"geoFailover",
		"geoProximity",
		"latencyPolicy",
		"trafficDialWeighted",
	} {
		if FindSchedulingType(code) == nil {
			t.Fatalf("expected scheduling type %q to be registered", code)
		}
	}
}
