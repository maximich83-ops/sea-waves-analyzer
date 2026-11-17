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
		{"успіх", []string{"analyze", "--count", "5", "--subject", "Чорне море"},
			&fakeSource{heights: static}, exitOK, "середнє:   2.36 м", "", false},
		{"успіх через go run з --", []string{"--", "analyze", "--count", "5", "--subject", "Чорне море"},
			&fakeSource{heights: static}, exitOK, "Акваторія: Чорне море", "", false},
		{"seed", []string{"analyze", "--count", "5", "--subject", "BlackSea", "--seed", "42"},
			&fakeSource{heights: static}, exitOK, "Seed: 42", "", true},
		{"subject з дефісом", []string{"analyze", "--count", "5", "--subject=-Північ"},
			&fakeSource{heights: static}, exitOK, "Акваторія: -Північ", "", false},
		{"limit знайдено", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "2"},
			&fakeSource{heights: static}, exitOK, "Перша хвиля вища за 2.0 м: №3", "", false},
		{"limit не знайдено", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "6"},
			&fakeSource{heights: static}, exitOK, "Хвиль, вищих за 6.0 м, немає", "", false},
		{"limit поза діапазоном", []string{"analyze", "--count", "5", "--subject", "A", "--limit", "13"},
			&fakeSource{}, exitUsage, "", "порогова висота поза діапазоном", false},
		{"без аргументів", nil, &fakeSource{}, exitUsage, "", "не вказано підкоманду", false},
		{"help", []string{"help"}, &fakeSource{}, exitOK, "Підкоманди:", "", false},
		{"analyze -h", []string{"analyze", "-h"}, &fakeSource{}, exitOK, "--count N", "", false},
		{"version", []string{"version"}, &fakeSource{}, exitOK, version, "", false},
		{"невідома підкоманда", []string{"report"}, &fakeSource{}, exitUsage, "", "невідома підкоманда", false},
		{"невідомий прапорець", []string{"analyze", "--count", "5", "--subject", "A", "--x", "1"},
			&fakeSource{}, exitUsage, "", "flag provided but not defined", false},
		{"відсутній count", []string{"analyze", "--subject", "A"}, &fakeSource{}, exitUsage, "", "--count", false},
		{"відсутні обидва", []string{"analyze"}, &fakeSource{}, exitUsage, "", "--count, --subject", false},
		{"count=4", []string{"analyze", "--count", "4", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"count=101", []string{"analyze", "--count", "101", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"count від'ємний", []string{"analyze", "--count", "-5", "--subject", "A"}, &fakeSource{}, exitUsage, "", "поза межами", false},
		{"count нечисловий", []string{"analyze", "--count", "abc", "--subject", "A"}, &fakeSource{}, exitUsage, "", "invalid value", false},
		{"порожній subject", []string{"analyze", "--count", "5", "--subject", "   "}, &fakeSource{}, exitUsage, "", "порожня", false},
		{"задовгий subject", []string{"analyze", "--count", "5", "--subject", strings.Repeat("a", 101)},
			&fakeSource{}, exitUsage, "", "задовга", false},
		{"зайвий аргумент", []string{"analyze", "--count", "5", "--subject", "A", "extra"},
			&fakeSource{}, exitUsage, "", "зайві аргументи", false},
		{"помилка генератора", []string{"analyze", "--count", "5", "--subject", "A"},
			&fakeSource{err: errors.New("generator failed")}, exitData, "", "помилка генератора", false},
		{"генератор повернув не стільки значень", []string{"analyze", "--count", "6", "--subject", "A"},
			&fakeSource{heights: static}, exitData, "", "замість 6", false},
		{"генератор повернув значення поза діапазоном", []string{"analyze", "--count", "5", "--subject", "A"},
			&fakeSource{heights: []float64{1, 2, 3, 4, 13}}, exitData, "", "поза діапазоном", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr, tt.src)

			if code != tt.wantCode {
				t.Errorf("код = %d, очікувалось %d\nstderr: %s", code, tt.wantCode, stderr.String())
			}
			checkStream(t, "stdout", stdout.String(), tt.wantStdout)
			checkStream(t, "stderr", stderr.String(), tt.wantStderr)
			if tt.src.seeded != tt.wantSeeded {
				t.Errorf("GenerateWithSeed викликано = %v, очікувалось %v", tt.src.seeded, tt.wantSeeded)
			}
		})
	}
}

func checkStream(t *testing.T, name, got, want string) {
	t.Helper()
	if want == "" {
		if got != "" {
			t.Errorf("%s має бути порожнім, отримано:\n%s", name, got)
		}
		return
	}
	if !strings.Contains(got, want) {
		t.Errorf("%s не містить %q:\n%s", name, want, got)
	}
}

// TestReportHasNoANSI — у звичайному виведенні немає керівних ANSI-послідовностей.
func TestReportHasNoANSI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	run([]string{"analyze", "--count", "5", "--subject", "A"}, &stdout, &stderr,
		&fakeSource{heights: []float64{0.4, 1.2, 3.0, 6.0, 1.2}})
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Errorf("stdout містить ANSI-послідовність:\n%q", stdout.String())
	}
}
