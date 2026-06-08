package schedulingconfigs

import (
	"strings"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

// GeoFailoverScheduling maps client geography to an ordered failover list.
type GeoFailoverScheduling struct {
	Scheduling
	count int
}

func (this *GeoFailoverScheduling) Start() {
	this.count = len(this.Candidates)
}

func (this *GeoFailoverScheduling) Next(call *shared.RequestCall) CandidateInterface {
	if this.count == 0 {
		return nil
	}

	country := strings.ToLower(optionString(call, OptionClientCountry))
	region := strings.ToLower(optionString(call, OptionClientRegion))

	for _, lookup := range []struct {
		key   string
		value string
	}{
		{key: "countryMap", value: country},
		{key: "regionMap", value: region},
	} {
		if len(lookup.value) == 0 {
			continue
		}
		rules := optionMap(call, lookup.key)
		if len(rules) == 0 {
			continue
		}
		for key, raw := range rules {
			if strings.EqualFold(key, lookup.value) {
				if candidate := firstCandidateByCodes(this.Candidates, stringList(raw)); candidate != nil {
					return candidate
				}
			}
		}
	}

	if candidate := firstCandidateByCodes(this.Candidates, stringList(callOptions(call)["default"])); candidate != nil {
		return candidate
	}
	return this.Candidates[0]
}

func (this *GeoFailoverScheduling) Summary() maps.Map {
	return maps.Map{
		"code":        "geoFailover",
		"name":        "GeoFailover地域故障转移算法",
		"description": "按国家、区域、默认规则选择有序源站列表中的第一个可用源站",
		"networks":    []string{"http"},
	}
}
