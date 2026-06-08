package nodes

import (
	"testing"
	"time"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/schedulingconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
)

func TestOriginSchedulingStats_BeginEnd(t *testing.T) {
	stats := NewOriginSchedulingStats()
	origin := &serverconfigs.OriginConfig{Id: 1}

	finish := stats.Begin(origin)
	outstanding, ok := stats.CandidateOutstandingRequests(origin)
	if !ok || outstanding != 1 {
		t.Fatalf("expected outstanding=1, got %d, ok=%v", outstanding, ok)
	}

	time.Sleep(time.Millisecond)
	finish(true)
	finish(true)

	outstanding, ok = stats.CandidateOutstandingRequests(origin)
	if !ok || outstanding != 0 {
		t.Fatalf("expected outstanding=0, got %d, ok=%v", outstanding, ok)
	}
	latency, ok := stats.CandidateLatencyMillis(origin, "global")
	if !ok || latency <= 0 {
		t.Fatalf("expected positive latency, got %.3f, ok=%v", latency, ok)
	}
}

func TestHTTPRequest_PrepareSchedulingRequestCall(t *testing.T) {
	req := &HTTPRequest{}
	call := shared.NewRequestCall()
	req.prepareSchedulingRequestCall(call)

	if _, ok := call.Options[schedulingconfigs.OptionRuntimeStats].(*OriginSchedulingStats); !ok {
		t.Fatalf("expected runtime stats provider")
	}
	if call.Options[schedulingconfigs.OptionClientCountry] != "${geo.country.id}" {
		t.Fatalf("expected default client country formatter")
	}
	if call.Options[schedulingconfigs.OptionClientRegion] != "${geo.province.id}" {
		t.Fatalf("expected default client region formatter")
	}
}
