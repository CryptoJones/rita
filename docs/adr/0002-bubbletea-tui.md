# 2. Use Bubble Tea for the results viewer

- Status: Accepted
- Date: 2026-06-15

## Context

Analysts need to browse scored threat results interactively — sort by score,
filter, and drill into a host pair's indicators — directly in a terminal, often
over SSH on a sensor with no GUI. RITA needed a results viewer that works well in
a plain terminal and is maintainable as the result schema evolves.

## Decision

Build the viewer (`viewer/`) on **Bubble Tea** (with **Lip Gloss** for styling
and **Bubbles** for table/viewport components).

## Consequences

- **Positive:** The Elm-style model/update/view architecture gives a clear,
  testable structure for interactive state (selection, sorting, paging).
- **Positive:** Runs anywhere a terminal does; no browser or GUI stack required.
- **Positive:** The Charm component ecosystem provides batteries-included tables,
  viewports, and key handling.
- **Negative:** Adds the Charm dependency set to the binary.
- **Negative:** A TUI is harder to test than a pure function; viewer tests focus
  on model state transitions rather than rendered output.
- **Note:** The viewer is a strictly read-only consumer of analysis output and
  could, if desired, be split into an optional component later.
