package schedulingconfigs

import (
	"math"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// LeastOutstandingScheduling prefers the candidate with the fewest in-flight
// requests after normalizing by configured candidate weight.
type LeastOutstandingScheduling struct {
	Scheduling
	count int
}

func (this *LeastOutstandingScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *LeastOutstandingScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	stats := runtimeStats(call)
	var best CandidateInterface
	bestScore := math.MaxFloat64
	for _, candidate := range this.Candidates {
		outstanding := uint64(0)
		if stats != nil {
			if value, ok := stats.CandidateOutstandingRequests(candidate); ok {
				outstanding = value
			}
		}
		score := float64(outstanding+1) / candidateWeight(candidate)
		if score < bestScore {
			best = candidate
			bestScore = score
		}
	}
	return best
}

func (this *LeastOutstandingScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "leastOutstanding",
		"name":        "LeastOutstanding最少未完成请求算法",
		"description": "根据源站当前in-flight请求数和权重选择压力最低的源站",
		"networks":    []string{"http", "tcp", "udp", "unix"},
	}
}
