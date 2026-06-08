package schedulingconfigs

import (
	"math"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// DynamicLatencyScheduling chooses the lowest observed EWMA latency candidate.
type DynamicLatencyScheduling struct {
	Scheduling
	count int
}

func (this *DynamicLatencyScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *DynamicLatencyScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	scope := optionString(call, OptionLatencyScope)
	if len(scope) == 0 {
		scope = "global"
	}

	stats := runtimeStats(call)
	var best CandidateInterface
	bestScore := math.MaxFloat64
	hasLatency := false
	for _, candidate := range this.Candidates {
		latency, ok := configuredLatencyMillis(call, candidate, scope)
		if !ok && stats != nil {
			latency, ok = stats.CandidateLatencyMillis(candidate, scope)
		}
		if !ok || latency <= 0 {
			continue
		}

		score := latency / math.Sqrt(candidateWeight(candidate))
		if score < bestScore {
			best = candidate
			bestScore = score
			hasLatency = true
		}
	}
	if hasLatency {
		return best
	}

	if fallback := firstCandidateByCodes(this.Candidates, stringList(callOptions(call)["fallback"])); fallback != nil {
		return fallback
	}
	return this.Candidates[0]
}

func (this *DynamicLatencyScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "dynamicLatency",
		"name":        "DynamicLatency动态延迟算法",
		"description": "根据运行时EWMA延迟选择源站，缺少延迟数据时按故障转移顺序回退",
		"networks":    []string{"http", "tcp", "udp", "unix"},
	}
}
