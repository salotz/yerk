// Package api defines stable, versioned product resource types for yerk.
//
// These types are the shared in-process model for collectors, printers, and
// future serializers (json/yaml, schema, yerk get). They are distinct from:
//
//   - internal/config TOML load structs (file schema + validation)
//   - CLI tabwriter / display helpers (presentation only)
//
// Adapters map config → api; status collectors fill observed api types;
// commands print api values. See ADR 011.
package api
