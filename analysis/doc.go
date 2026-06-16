// Package analysis computes RITA's threat indicators from imported connection
// data.
//
// For each internal/external host pair (a "uconn") the analyzer evaluates a set
// of indicators -- beaconing (a weighted blend of timestamp, data-size,
// duration, and histogram subscores), long connections, strobes, C2-over-DNS,
// and threat-intel hits -- and combines them, together with prevalence and
// first-seen modifiers, into a ThreatMixtape result row. Prorated indicators are
// bucketed into Base/Low/Med/High severities via calculateBucketedScore.
//
// Work is parallelized with a pool of analysis workers feeding a
// database.BulkWriter; results stream to ClickHouse as they are produced.
package analysis
