# 3. Indicator-based threat scoring with bucketed severities

- Status: Accepted
- Date: 2026-06-15

## Context

RITA's job is to surface likely C2 / malicious activity from benign bulk traffic.
A single score is not enough: different behaviors (beaconing, long connections,
strobes, C2-over-DNS, threat-intel hits) are evidence of different things, and
some are binary (a threat-intel match either happened or not) while others are
graduated (a 9-hour connection is "more" long than a 2-hour one).

## Decision

Score each internal/external host pair by combining a set of **indicators** and
**modifiers** into a `ThreatMixtape` result row (`analysis/`):

- **Beaconing** is a weighted blend of subscores (timestamp regularity, data-size
  consistency, duration, histogram), with the weights configured in
  `scoring.beacon` and required to sum to 1.
- **Graduated indicators** (long connection, C2) use `calculateBucketedScore`,
  which linearly interpolates a score between configured
  Base/Low/Med/High thresholds, mapping into the None/Low/Med/High impact
  categories.
- **Binary indicators** (strobe, threat-intel) contribute a fixed configured
  impact score.
- **Modifiers** (prevalence, first-seen, missing-host-header, rare signature,
  MIME-type mismatch) nudge the score up or down based on context.

A connection pair with `Count >= strobeConnThreshold` (86400 ≈ one connection per
second over a day) is recorded as a **strobe** and skips beacon analysis, since
such volume is not beacon-like.

## Consequences

- **Positive:** Operators can tune each indicator's weight/thresholds in
  `config.hjson` without code changes.
- **Positive:** The mixtape preserves *why* a pair scored highly (which
  indicators fired), which is what analysts need to triage.
- **Negative:** Many tunable knobs mean defaults matter a lot and must be chosen
  carefully; misconfiguration can hide or over-report threats.
- **Negative:** Interpolation/threshold ordering must be validated
  (`score_thresholds_range` custom validator enforces strictly increasing
  buckets).
