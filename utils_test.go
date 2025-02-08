package main

import (
	"testing"
)

func TestHumanSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Bytes", 500, "500 B"},
		{"Exactly 1KB", 1 << 10, "1.00 KB"},
		{"1.5KB", 1536, "1.50 KB"},
		{"Exactly 1MB", 1 << 20, "1.00 MB"},
		{"2.5MB", 2<<20 + 512<<10, "2.50 MB"},
		{"Exactly 1GB", 1 << 30, "1.00 GB"},
		{"3.75GB", 3<<30 + 768<<20, "3.75 GB"},
		{"Exactly 1TB", 1 << 40, "1.00 TB"},
		{"5.25TB", 5<<40 + 256<<30, "5.25 TB"},
		{"Exactly 1PB", 1 << 50, "1.00 PB"},
		{"100PB", 100 << 50, "100.00 PB"},
		{"Edge case EB", 1 << 60, "1.00 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanSize(tt.size); got != tt.want {
				t.Errorf("HumanSize(%d) = %v, want %v", tt.size, got, tt.want)
			} else {
				t.Logf("HumanSize(%d) = %v", tt.size, got)
			}
		})
	}

}
