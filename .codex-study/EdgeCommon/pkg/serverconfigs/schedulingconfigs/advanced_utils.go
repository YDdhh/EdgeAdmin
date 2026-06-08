package schedulingconfigs

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs/shared"
	"github.com/iwind/TeaGo/maps"
)

const earthRadiusKm = 6371.0

type candidateLocation struct {
	lat float64
	lon float64
}

func callOptions(call *shared.RequestCall) maps.Map {
	if call == nil || call.Options == nil {
		return maps.Map{}
	}
	return call.Options
}

func optionString(call *shared.RequestCall, key string) string {
	options := callOptions(call)
	raw, ok := options[key]
	if !ok || raw == nil {
		return ""
	}
	value := strings.TrimSpace(fmt.Sprint(raw))
	if call != nil && call.Formatter != nil {
		value = strings.TrimSpace(call.Formatter(value))
	}
	return value
}

func optionFloat(call *shared.RequestCall, key string) (float64, bool) {
	options := callOptions(call)
	raw, ok := options[key]
	if !ok {
		return 0, false
	}
	if text, ok := raw.(string); ok && call != nil && call.Formatter != nil {
		raw = call.Formatter(text)
	}
	return toFloat64(raw)
}

func optionMap(call *shared.RequestCall, key string) map[string]interface{} {
	options := callOptions(call)
	raw, ok := options[key]
	if !ok {
		return nil
	}
	return toMap(raw)
}

func runtimeStats(call *shared.RequestCall) RuntimeStatsProvider {
	options := callOptions(call)
	stats, _ := options[OptionRuntimeStats].(RuntimeStatsProvider)
	return stats
}

func candidateWeight(candidate CandidateInterface) float64 {
	weight := candidate.CandidateWeight()
	if weight == 0 {
		return 1
	}
	return float64(weight)
}

func candidateKey(candidate CandidateInterface) string {
	codes := candidateCodes(candidate)
	if len(codes) > 0 {
		return codes[0]
	}
	return fmt.Sprintf("%T:%v", candidate, candidate)
}

func candidateCodes(candidate CandidateInterface) []string {
	if candidate == nil {
		return nil
	}
	codes := []string{}
	for _, code := range candidate.CandidateCodes() {
		code = strings.TrimSpace(code)
		if len(code) > 0 {
			codes = append(codes, code)
		}
	}
	return codes
}

func candidateMatches(candidate CandidateInterface, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) == 0 {
		return false
	}
	for _, candidateCode := range candidateCodes(candidate) {
		if candidateCode == code || strings.EqualFold(candidateCode, code) {
			return true
		}
	}
	return false
}

func findCandidateByCode(candidates []CandidateInterface, code string) CandidateInterface {
	for _, candidate := range candidates {
		if candidateMatches(candidate, code) {
			return candidate
		}
	}
	return nil
}

func firstCandidateByCodes(candidates []CandidateInterface, codes []string) CandidateInterface {
	for _, code := range codes {
		if candidate := findCandidateByCode(candidates, code); candidate != nil {
			return candidate
		}
	}
	return nil
}

func stringList(raw interface{}) []string {
	switch value := raw.(type) {
	case nil:
		return nil
	case []string:
		return value
	case []interface{}:
		result := []string{}
		for _, item := range value {
			text := strings.TrimSpace(fmt.Sprint(item))
			if len(text) > 0 {
				result = append(result, text)
			}
		}
		return result
	case string:
		result := []string{}
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if len(item) > 0 {
				result = append(result, item)
			}
		}
		return result
	default:
		text := strings.TrimSpace(fmt.Sprint(value))
		if len(text) == 0 {
			return nil
		}
		return []string{text}
	}
}

func toMap(raw interface{}) map[string]interface{} {
	switch value := raw.(type) {
	case nil:
		return nil
	case maps.Map:
		return map[string]interface{}(value)
	case map[string]interface{}:
		return value
	case map[string]string:
		result := map[string]interface{}{}
		for k, v := range value {
			result[k] = v
		}
		return result
	case map[string]float64:
		result := map[string]interface{}{}
		for k, v := range value {
			result[k] = v
		}
		return result
	case map[string]int:
		result := map[string]interface{}{}
		for k, v := range value {
			result[k] = v
		}
		return result
	case map[string][]string:
		result := map[string]interface{}{}
		for k, v := range value {
			result[k] = v
		}
		return result
	default:
		return nil
	}
}

func toFloat64(raw interface{}) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int8:
		return float64(value), true
	case int16:
		return float64(value), true
	case int32:
		return float64(value), true
	case int64:
		return float64(value), true
	case uint:
		return float64(value), true
	case uint8:
		return float64(value), true
	case uint16:
		return float64(value), true
	case uint32:
		return float64(value), true
	case uint64:
		return float64(value), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func configuredLatencyMillis(call *shared.RequestCall, candidate CandidateInterface, scope string) (float64, bool) {
	for _, key := range []string{"latencies", "latencyMatrix"} {
		root := optionMap(call, key)
		if len(root) == 0 {
			continue
		}
		for _, scopeKey := range []string{scope, "global", "default"} {
			if len(scopeKey) == 0 {
				continue
			}
			if latencyMap := toMap(root[scopeKey]); len(latencyMap) > 0 {
				if latency, ok := lookupCandidateFloat(latencyMap, candidate); ok {
					return latency, true
				}
			}
		}
		if latency, ok := lookupCandidateFloat(root, candidate); ok {
			return latency, true
		}
	}
	return 0, false
}

func lookupCandidateFloat(values map[string]interface{}, candidate CandidateInterface) (float64, bool) {
	for _, code := range candidateCodes(candidate) {
		if raw, ok := values[code]; ok {
			return toFloat64(raw)
		}
		for key, raw := range values {
			if strings.EqualFold(key, code) {
				return toFloat64(raw)
			}
		}
	}
	return 0, false
}

func candidateTrafficDial(call *shared.RequestCall, candidate CandidateInterface) float64 {
	dial := 100.0
	if v, ok := optionFloat(call, "defaultDial"); ok {
		dial = v
	}
	for _, key := range []string{"trafficDials", "candidateTrafficDials"} {
		if valueMap := optionMap(call, key); len(valueMap) > 0 {
			if candidateDial, ok := lookupCandidateFloat(valueMap, candidate); ok {
				dial = candidateDial
			}
		}
	}
	if dial < 0 {
		return 0
	}
	if dial > 100 {
		return 100
	}
	return dial
}

func lookupCandidateLocation(call *shared.RequestCall, candidate CandidateInterface) (candidateLocation, bool) {
	if located, ok := candidate.(LocationCandidateInterface); ok {
		return candidateLocation{lat: located.CandidateLatitude(), lon: located.CandidateLongitude()}, true
	}

	locations := optionMap(call, "candidateLocations")
	for _, code := range candidateCodes(candidate) {
		for key, raw := range locations {
			if key == code || strings.EqualFold(key, code) {
				return parseLocation(raw)
			}
		}
	}
	return candidateLocation{}, false
}

func parseLocation(raw interface{}) (candidateLocation, bool) {
	switch value := raw.(type) {
	case []float64:
		if len(value) < 2 {
			return candidateLocation{}, false
		}
		return candidateLocation{lat: value[0], lon: value[1]}, true
	case []interface{}:
		if len(value) < 2 {
			return candidateLocation{}, false
		}
		lat, latOK := toFloat64(value[0])
		lon, lonOK := toFloat64(value[1])
		return candidateLocation{lat: lat, lon: lon}, latOK && lonOK
	case string:
		pieces := strings.Split(value, ",")
		if len(pieces) != 2 {
			return candidateLocation{}, false
		}
		lat, latOK := toFloat64(pieces[0])
		lon, lonOK := toFloat64(pieces[1])
		return candidateLocation{lat: lat, lon: lon}, latOK && lonOK
	default:
		valueMap := toMap(value)
		if len(valueMap) == 0 {
			return candidateLocation{}, false
		}
		lat, latOK := toFloat64(firstMapValue(valueMap, "lat", "latitude"))
		lon, lonOK := toFloat64(firstMapValue(valueMap, "lon", "lng", "longitude"))
		return candidateLocation{lat: lat, lon: lon}, latOK && lonOK
	}
}

func firstMapValue(valueMap map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := valueMap[key]; ok {
			return value
		}
	}
	return nil
}

func haversineKm(a candidateLocation, b candidateLocation) float64 {
	lat1 := a.lat * math.Pi / 180
	lat2 := b.lat * math.Pi / 180
	dLat := (b.lat - a.lat) * math.Pi / 180
	dLon := (b.lon - a.lon) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLon*sinLon
	return 2 * earthRadiusKm * math.Asin(math.Sqrt(h))
}
