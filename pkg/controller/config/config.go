package config

// TestMode enables test mode in the operator and applies
// the following changes:
// - Disables BookKeeper minimum number of replicas
// - Disables Pravega Controller minimum number of replicas
// - Disables Segment Store minimum number of replicas
var TestMode bool
