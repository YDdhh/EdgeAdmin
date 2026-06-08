package schedulingconfigs

// CandidateInterface 候选对象接口
type CandidateInterface interface {
	// CandidateWeight 权重
	CandidateWeight() uint

	// CandidateCodes 代号
	CandidateCodes() []string
}

const (
	OptionRuntimeStats    = "_runtimeStats"
	OptionClientCountry   = "clientCountry"
	OptionClientRegion    = "clientRegion"
	OptionClientLatitude  = "clientLatitude"
	OptionClientLongitude = "clientLongitude"
	OptionLatencyScope    = "latencyScope"
)

// RuntimeStatsProvider provides request-time data that is not part of static
// candidate config, such as EWMA latency and current in-flight request count.
type RuntimeStatsProvider interface {
	CandidateLatencyMillis(candidate CandidateInterface, scope string) (float64, bool)
	CandidateOutstandingRequests(candidate CandidateInterface) (uint64, bool)
}

// LocationCandidateInterface can be implemented by candidates that have a
// stable geographic location. Algorithms also support locations from options.
type LocationCandidateInterface interface {
	CandidateLatitude() float64
	CandidateLongitude() float64
}
