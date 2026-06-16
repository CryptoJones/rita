# Test Data

This directory holds sample log fixtures used by RITA's unit, integration, and
`cmd` test suites. Each subdirectory is a self-contained input set for an import
test that exercises a particular format, edge case, or detection.

> `go.mod` here scopes the fixtures out of the main module's build/vet so the
> sample files are never compiled or linted as Go.

## Fixtures

| Path | What it exercises |
|---|---|
| `valid_tsv/` | Baseline well-formed Zeek **TSV** logs (conn/dns/http/ssl). The canonical "happy path" import. |
| `valid_json/` | Baseline well-formed Zeek **JSON** logs. Verifies the JSON parser path. |
| `json_with_all_fields/` | JSON logs populating the full set of supported fields. |
| `has_unknown_field/` | Logs containing fields RITA does not model, to confirm unknown fields are tolerated/ignored. |
| `dns_only/` | A dataset with only DNS logs (no conn), exercising DNS-only analysis paths. |
| `missing_host/` | HTTP beaconing traffic with a blank `Host` header (drives the missing-host-header modifier). See `missing_host/README` for the synthetic scenario it encodes. Organized into day directories for rolling-import tests. |
| `dnscat2-ja3-strobe-agent/` | DNS-tunneling / C2-style traffic (dnscat2) with JA3 fingerprints and strobe behavior — drives C2-over-DNS and strobe detection. |
| `open_conns/` | Long-lived "open" connections (split into open vs closed) used to test connection seasoning/linking. |
| `open_sni/` | Open connections with SNI data for SSL linking. |
| `proxy/` | HTTP proxy traffic for proxy-aware linking. |
| `proxy_rolling/` | Proxy traffic arranged for rolling-import tests. |
| `truncated/` | Deliberately truncated/partial log files to test error handling. |
| `text_file/` | A non-log text file to verify that invalid inputs are rejected gracefully. |

## Regenerating the open-connection fixtures

`create_open_logs.sh` generates the `open_conns` (open/closed) fixtures from the
baseline `valid_tsv` logs. Run it from this directory when the open-connection
sample set needs to be rebuilt; it is the source of truth for how those fixtures
were produced.
