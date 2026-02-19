package migration

import (
	"hash/crc32"
	"strings"
)

// ComputeChecksum returns a CRC32 checksum of the SQL content,
// normalized by trimming whitespace and using Unix line endings.
func ComputeChecksum(sql string) uint32 {
	normalized := strings.ReplaceAll(sql, "\r\n", "\n")
	normalized = strings.TrimSpace(normalized)
	return crc32.ChecksumIEEE([]byte(normalized))
}
