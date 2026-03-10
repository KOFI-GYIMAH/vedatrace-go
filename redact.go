package vedatrace

import (
	"maps"
	"strings"
)

// redact applies the RedactFields rules to a copy of meta, replacing matching
// leaf values with "[REDACTED]". The original map is never mutated.
func redact(meta LogMetadata, fields []string) LogMetadata {
	if len(meta) == 0 || len(fields) == 0 {
		return meta
	}
	out := cloneMetadata(meta)
	for _, field := range fields {
		parts := strings.SplitN(field, ".", 2)
		if len(parts) == 1 {
			if _, ok := out[parts[0]]; ok {
				out[parts[0]] = "[REDACTED]"
			}
		} else {
			// nested path — recurse into the sub-map if present
			key, rest := parts[0], parts[1]
			if sub, ok := out[key]; ok {
				if subMap, ok := sub.(map[string]any); ok {
					out[key] = redact(LogMetadata(subMap), []string{rest})
				} else if subMeta, ok := sub.(LogMetadata); ok {
					out[key] = redact(subMeta, []string{rest})
				}
			}
		}
	}
	return out
}

// cloneMetadata performs a shallow copy of a LogMetadata map.
func cloneMetadata(m LogMetadata) LogMetadata {
	out := make(LogMetadata, len(m))
	maps.Copy(out, m)
	return out
}
