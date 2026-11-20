package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// fakeSource — статичне джерело даних для тестів CLI без випадковості.
type fakeSource struct {
	heights []float64
	err     error
	seeded  bool // чи викликався GenerateWithSeed
}

func (f *fakeSource) Generate(count int, seaArea string) ([]float64, error) {
	return f.heights, f.err
}

func (f *fakeSource) GenerateWithSeed(count int, seaArea string, seed int64) ([]float64, error) {
	f.seeded = true
	return f.heights, f.err
}

func TestRun(t *testing.T) {
	static := []float64{0.4, 1.2, 3.0, 6.0, 1.2}
	tests := []struct {
		name       string
		args       []string
		src        *fakeSource
		wantCode   int
		wantStdout string // підрядок у stdout ("" — stdout має бути порожнім)
		wantStderr string // підрядок у stderr ("" — stderr має бути порожнім)
		wantSeeded bool
	}{
		{"success", []string{"analyze", "--count", "5", "--subject", "Чорне море"},
			&fakeSource{heights: static}, exitOK, "середнє:   2.36 м", "", false},
		{"success via go run with --", []string{"--", "analyze", "--count", "5", "--subject", "Чорне море"},
			&fakeSource{heights: static}, exitOK, "Акваторія: Чорне море", "", false},
		{"seed", []string{"analyze", "--count", "5", "--subject", "BlackSea", "--seed", "42"},
			&fakeSource{heights: static}, exitOK, "Seed: 42", "", true},
		{"subject with leading dash", []string{"analyze", "--count", "5", "--subject=-Північ"},
			&fakeSource{heights: static}, exitOK, "Акваторія: -Північ", "", false},
		{"limit found", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "2"},
			&fakeSource{heights: static}, exitOK, "Перша хвиля вища за 2.0 м: №3", "", false},
		{"limit not found", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "6"},
			&fakeSource{heights: static}, exitOK, "Хвиль, вищих за 6.0 м, немає", "", false},
		{"limit out of range", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "13"},
			&fakeSource{}, exitUsage, "", "порогова висота поза діапазоном", false},
		{"no arguments", nil, &fakeSource{}, exitUsage, "", "не вказано підкоманду", false},
		{"help", []string{"help"}, &fakeSource{}, exitOK, "Підкоманди:", "", false},
		{"analyze -h", []string{"analyze", "-h"}, &fakeSource{}, exitOK, "--count N", "", false},
		{"version", []string{"version"}, &fakeSource{}, exitOK, version, "", false},
		{"unknown subcommand", []string{"report"}, &fakeSource{}, exitUsage, "", "невідома підкоманда", false},
		{"unknown flag", []string{"analyze", "--count", "5", "--subject", "A", "--x", "1"},
			&fakeSource{}, exitUsage, "", "flag provided but not defined", false},
		{"missing count", []string{"analyze", "--subject", "A"}, &fakeSource{}, exitUsage, "", "--count", false},
		{"missing count and subject", []string{"analyze"}, &fakeSource{}, exitUsage, "", "--count, --subject", false},
		{"count 4", []string{"analyze", "--count", "4", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"count 101", []string{"analyze", "--count", "101", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"negative count", []string{"analyze", "--count", "-5", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"non-numeric count", []string{"analyze", "--count", "abc", "--subject", "A"}, &fakeSource{}, exitUsage, "", "invalid value", false},
		{"blank subject", []string{"analyze", "--count", "5", "--subject", "   "}, &fakeSource{}, exitUsage, "", "порожня", false},
		{"subject too long", []string{"analyze", "--count", "5", "--subject", strings.Repeat("a", 101)},
			&fakeSource{}, exitUsage, "", "задовга", false},
		{"extra argument", []string{"analyze", "--count", "5", "--subject", "A", "extra"},
			&fakeSource{}, exitUsage, "", "зайві аргументи", false},
		{"generator error", []string{"analyze", "--count", "5", "--subject", "A"},
			&fakeSource{err: errors.New("generator failed")}, exitData, "", "помилка генератора", false},
		{"generator returned wrong count", []string{"analyze", "--count", "6", "--subject", "A"},
			&fakeSource{heights: static}, exitData, "", "замість 6", false},
		{"generator returned value out of range", []string{"analyze", "--count", "5", "--subject", "A"},
			&fakeSource{heights: []float64{1, 2, 3, 4, 13}}, exitData, "", "поза діапазоном", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, tt.src)

			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d\nstderr: %s", code, tt.wantCode, stderr.String())
			}
			checkStream(t, "stdout", stdout.String(), tt.wantStdout)
			checkStream(t, "stderr", stderr.String(), tt.wantStderr)
			if tt.src.seeded != tt.wantSeeded {
				t.Errorf("GenerateWithSeed called = %v, want %v", tt.src.seeded, tt.wantSeeded)
			}
		})
	}
}

func checkStream(t *testing.T, name, got, want string) {
	t.Helper()
	if want == "" {
		if got != "" {
			t.Errorf("%s must be empty, got:\n%s", name, got)
		}
		return
	}
	if !strings.Contains(got, want) {
		t.Errorf("%s does not contain %q:\n%s", name, want, got)
	}
}

// TestReportHasNoANSI — у звичайному виведенні немає керівних ANSI-послідовностей.
func TestReportHasNoANSI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	run([]string{"analyze", "--count", "5", "--subject", "A"}, &stdout, &stderr,
		&fakeSource{heights: []float64{0.4, 1.2, 3.0, 6.0, 1.2}})
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Errorf("stdout contains an ANSI sequence:\n%q", stdout.String())
	}
}
