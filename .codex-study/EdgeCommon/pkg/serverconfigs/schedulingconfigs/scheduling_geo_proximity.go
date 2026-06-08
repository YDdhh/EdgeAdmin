package schedulingconfigs

import (
	"math"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// GeoProximityScheduling chooses the geographically closest candidate.
type GeoProximityScheduling struct {
	Scheduling
	count int
}

func (this *GeoProximityScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *GeoProximityScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	lat, latOK := optionFloat(call, OptionClientLatitude)
	lon, lonOK := optionFloat(call, OptionClientLongitude)
	if !latOK || !lonOK {
		return this.Candidates[0]
	}
	clientLocation := candidateLocation{lat: lat, lon: lon}
	biases := optionMap(call, "biases")

	var best CandidateInterface
	bestScore := math.MaxFloat64
	for _, candidate := range this.Candidates {
		location, ok := lookupCandidateLocation(call, candidate)
		if !ok {
			continue
		}
		score := haversineKm(clientLocation, location)
		if bias, ok := lookupCandidateFloat(biases, candidate); ok {
			score -= bias
		}
		if score < bestScore {
			best = candidate
			bestScore = score
		}
	}
	if best != nil {
		return best
	}
	return this.Candidates[0]
}

func (this *GeoProximityScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "geoProximity",
		"name":        "GeoProximity地理邻近算法",
		"description": "根据客户端与源站经纬度距离选择最近源站，支持bias偏移演示",
		"networks":    []string{"http"},
	}
}
