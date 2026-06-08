package schedulingconfigs

import (
	"math/rand"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// TrafficDialWeightedScheduling combines candidate weight and traffic dial.
type TrafficDialWeightedScheduling struct {
	Scheduling
	count int
}

func (this *TrafficDialWeightedScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *TrafficDialWeightedScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	weights := make([]float64, 0, this.count)
	sum := 0.0
	for _, candidate := range this.Candidates {
		effectiveWeight := candidateWeight(candidate) * candidateTrafficDial(call, candidate) / 100
		if effectiveWeight < 0 {
			effectiveWeight = 0
		}
		weights = append(weights, effectiveWeight)
		sum += effectiveWeight
	}
	if sum <= 0 {
		return nil
	}

	point := rand.Float64() * sum
	for index, weight := range weights {
		point -= weight
		if point <= 0 {
			return this.Candidates[index]
		}
	}
	return this.Candidates[this.count-1]
}

func (this *TrafficDialWeightedScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "trafficDialWeighted",
		"name":        "TrafficDialWeighted流量拨盘权重算法",
		"description": "组合源站权重和traffic dial百分比进行加权调度",
		"networks":    []string{"http", "tcp", "udp", "unix"},
	}
}
