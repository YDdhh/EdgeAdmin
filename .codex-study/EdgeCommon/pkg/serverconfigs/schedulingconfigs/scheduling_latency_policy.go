package schedulingconfigs

import (
	"math"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// LatencyPolicyScheduling uses a static client-region to candidate latency
// matrix, similar to DNS latency routing demos.
type LatencyPolicyScheduling struct {
	Scheduling
	count int
}

func (this *LatencyPolicyScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *LatencyPolicyScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	region := optionString(call, OptionClientRegion)
	if len(region) == 0 {
		region = optionString(call, "region")
	}

	var best CandidateInterface
	bestLatency := math.MaxFloat64
	for _, candidate := range this.Candidates {
		latency, ok := configuredLatencyMillis(call, candidate, region)
		if !ok || latency <= 0 {
			continue
		}
		if latency < bestLatency {
			best = candidate
			bestLatency = latency
		}
	}
	if best != nil {
		return best
	}
	return this.Candidates[0]
}

func (this *LatencyPolicyScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "latencyPolicy",
		"name":        "LatencyPolicy静态延迟路由算法",
		"description": "根据客户端区域到源站的静态延迟矩阵选择延迟最低源站",
		"networks":    []string{"http"},
	}
}
