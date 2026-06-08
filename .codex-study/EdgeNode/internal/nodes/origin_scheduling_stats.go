package nodes

import (
	"strconv"
	"sync"
	"time"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/schedulingconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

const originSchedulingLatencyAlpha = 0.2

var SharedOriginSchedulingStats = NewOriginSchedulingStats()

type OriginSchedulingStats struct {
	locker sync.RWMutex
	states map[int64]*OriginSchedulingStat
}

type OriginSchedulingStat struct {
	Outstanding   uint64
	LatencyMillis float64
	UpdatedAt     int64
}

func NewOriginSchedulingStats() *OriginSchedulingStats {
	return &OriginSchedulingStats{
		states: map[int64]*OriginSchedulingStat{},
	}
}

func (this *OriginSchedulingStats) Begin(origin *serverconfigs.OriginConfig) func(success bool) {
	if origin == nil || origin.Id <= 0 {
		return func(success bool) {}
	}

	startedAt := time.Now()
	this.locker.Lock()
	state := this.state(origin.Id)
	state.Outstanding++
	state.UpdatedAt = startedAt.Unix()
	this.locker.Unlock()

	var once sync.Once
	return func(success bool) {
		once.Do(func() {
			this.End(origin, time.Since(startedAt), success)
		})
	}
}

func (this *OriginSchedulingStats) End(origin *serverconfigs.OriginConfig, latency time.Duration, success bool) {
	if origin == nil || origin.Id <= 0 {
		return
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	state := this.state(origin.Id)
	if state.Outstanding > 0 {
		state.Outstanding--
	}
	state.UpdatedAt = time.Now().Unix()
	if success && latency > 0 {
		latencyMillis := float64(latency.Microseconds()) / 1000
		if state.LatencyMillis <= 0 {
			state.LatencyMillis = latencyMillis
		} else {
			state.LatencyMillis = originSchedulingLatencyAlpha*latencyMillis + (1-originSchedulingLatencyAlpha)*state.LatencyMillis
		}
	}
}

func (this *OriginSchedulingStats) CandidateLatencyMillis(candidate schedulingconfigs.CandidateInterface, scope string) (float64, bool) {
	originId, ok := originIdFromCandidate(candidate)
	if !ok {
		return 0, false
	}

	this.locker.RLock()
	state := this.states[originId]
	this.locker.RUnlock()
	if state == nil || state.LatencyMillis <= 0 {
		return 0, false
	}
	return state.LatencyMillis, true
}

func (this *OriginSchedulingStats) CandidateOutstandingRequests(candidate schedulingconfigs.CandidateInterface) (uint64, bool) {
	originId, ok := originIdFromCandidate(candidate)
	if !ok {
		return 0, false
	}

	this.locker.RLock()
	state := this.states[originId]
	this.locker.RUnlock()
	if state == nil {
		return 0, true
	}
	return state.Outstanding, true
}

func (this *OriginSchedulingStats) state(originId int64) *OriginSchedulingStat {
	state := this.states[originId]
	if state == nil {
		state = &OriginSchedulingStat{}
		this.states[originId] = state
	}
	return state
}

func originIdFromCandidate(candidate schedulingconfigs.CandidateInterface) (int64, bool) {
	if candidate == nil {
		return 0, false
	}
	if origin, ok := candidate.(*serverconfigs.OriginConfig); ok && origin.Id > 0 {
		return origin.Id, true
	}
	for _, code := range candidate.CandidateCodes() {
		originId, err := strconv.ParseInt(code, 10, 64)
		if err == nil && originId > 0 {
			return originId, true
		}
	}
	return 0, false
}

func (this *HTTPRequest) prepareSchedulingRequestCall(call *shared.RequestCall) {
	if call == nil {
		return
	}
	if call.Options == nil {
		call.Options = maps.Map{}
	}
	call.Options[schedulingconfigs.OptionRuntimeStats] = SharedOriginSchedulingStats
	if _, ok := call.Options[schedulingconfigs.OptionClientCountry]; !ok {
		call.Options[schedulingconfigs.OptionClientCountry] = "${geo.country.id}"
	}
	if _, ok := call.Options[schedulingconfigs.OptionClientRegion]; !ok {
		call.Options[schedulingconfigs.OptionClientRegion] = "${geo.province.id}"
	}
	if _, ok := call.Options[schedulingconfigs.OptionLatencyScope]; !ok {
		call.Options[schedulingconfigs.OptionLatencyScope] = "global"
	}
}
