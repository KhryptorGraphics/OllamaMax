# Dashboard Enhancement Plan (Phase 1)

## Goals
- <2s initial load with cached data
- <500ms real-time UI update latency
- 100+ concurrent widgets without perf degradation

## Inventory and Migration Matrix
- KPI Cards: High usage, low complexity → migrate first
- Charts (Chart.js): High usage, medium complexity → extract wrappers in @ollamamax/ui
- Tables: Medium usage, medium complexity → add virtualization and column controls
- Filters: High usage, medium complexity → extract FilterBar with debounce + URL sync

## Real-Time Architecture
- WebSocket connect to `/ws`
- Exponential backoff: 0.5s → 30s, jitter
- Connection status + latency indicator
- Container/presenter split; in-memory cache; memoized selectors

## Layout System
- react-grid-layout
- Persist per-user layout (localStorage → backend preferences later)
- Widget catalog (KPI, Chart, Table, Log Stream)

## Filters & Search
- Multi-field: date ranges, select, text
- 300ms debounce
- Query param persistence; saved presets

## Exporting
- CSV/JSON with UTF-8 + escaping
- Progress modal, cancel
- Export history list with download links

## API Integration
- P2P metrics, consensus status, analytics endpoints
- Normalize DTOs in @ollamamax/api-client
- Graceful partial data + retries

## Testing & Benchmarks
- Playwright E2E + visual regression
- Vitest + RTL for components (>95% coverage)
- Performance suites for WS update cost + export throughput

## Timeline
- Week 3: MVP layout + KPI; real-time infra; plan doc (this)
- Week 4: Charts and Filters; export MVP; perf harness

