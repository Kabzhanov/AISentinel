package iox

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// TestLockedLineWriterConcurrentNoInterleave hammers WriteLine from many
// goroutines concurrently and asserts every line comes out on the reader
// side exactly as it went in — never merged with another goroutine's bytes.
// Run with -race to also catch any data race on the shared buffer.
func TestLockedLineWriterConcurrentNoInterleave(t *testing.T) {
	var buf bytes.Buffer
	lw := NewLockedLineWriter(&buf)

	const goroutines = 50
	const linesPerGoroutine = 200

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < linesPerGoroutine; i++ {
				line := fmt.Sprintf("g%03d-line%04d-%s", id, i, strings.Repeat("x", 40))
				if err := lw.WriteLine([]byte(line)); err != nil {
					t.Errorf("WriteLine: %v", err)
				}
			}
		}(g)
	}
	wg.Wait()

	seen := make(map[string]int, goroutines*linesPerGoroutine)
	scanner := bufio.NewScanner(&buf)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !isWellFormedLine(line) {
			t.Fatalf("corrupted/interleaved line: %q", line)
		}
		seen[line]++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}

	if want := goroutines * linesPerGoroutine; len(seen) != want {
		t.Fatalf("got %d distinct lines, want %d", len(seen), want)
	}
	for line, count := range seen {
		if count != 1 {
			t.Fatalf("line %q seen %d times, want 1", line, count)
		}
	}
}

// isWellFormedLine checks the "gNNN-lineNNNN-xxxx...x" shape produced above;
// an interleaved write would break this shape (e.g. two prefixes glued
// together) and get caught here.
func isWellFormedLine(line string) bool {
	var g, i int
	var suffix string
	n, err := fmt.Sscanf(line, "g%03d-line%04d-%s", &g, &i, &suffix)
	return err == nil && n == 3 && strings.Trim(suffix, "x") == ""
}
