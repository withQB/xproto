package router

const (
	capabilityLengthenedRootInterval = 1 << iota
	capabilityCryptographicSetups
	capabilitySetupACKs // nolint:deadcode,varcheck
	capabilityDedupedCoordinateInfo
	capabilitySoftState
	capabilityHybridRouting
)

const ourVersion uint8 = 1
const ourCapabilities uint32 = capabilityLengthenedRootInterval | capabilityCryptographicSetups | capabilityDedupedCoordinateInfo | capabilitySoftState | capabilityHybridRouting
