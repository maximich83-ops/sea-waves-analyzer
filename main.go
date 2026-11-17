// sea-waves-analyzer — CLI-утиліта аналізу висот морських хвиль
// (варіант 5 «Морські хвилі»).
//
// main.go відповідає лише за: вибір підкоманди, читання прапорців,
// виклик генератора й предметного пакета, виведення та код завершення.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/waves"
	"github.com/maximich83-ops/sea-waves-analyzer/analyzer"
)

const (
	programName = "sea-waves-analyzer"
	version     = "0.3.0"
	variantInfo = "Варіант 5 «Морські хвилі»"
)

// Коди завершення утиліти.
const (
	exitOK    = 0 // успіх
	exitData  = 1 // помилка отримання або аналізу даних
	exitUsage = 2 // помилка виклику або аргументів
)

// waveSource — мале джерело даних. У програмі це пакет waves,
// у тестах — статична реалізація без випадковості.
type waveSource interface {
	Generate(count int, seaArea string) ([]float64, error)
	GenerateWithSeed(count int, seaArea string, seed int64) ([]float64, error)
}

// wavesPackage адаптує функції пакета waves до інтерфейсу waveSource.
type wavesPackage struct{}

func (wavesPackage) Generate(count int, seaArea string) ([]float64, error) {
	return waves.Generate(count, seaArea)
}

func (wavesPackage) GenerateWithSeed(count int, seaArea string, seed int64) ([]float64, error) {
	return waves.GenerateWithSeed(count, seaArea, seed)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, wavesPackage{}))
}

// run виконує програму й повертає код завершення.
// Потоки та джерело даних передаються ззовні, тому run можна тестувати.
func run(args []string, stdout, stderr io.Writer, src waveSource) int {
	// go run . -- analyze ... передає "--" програмі; зібраний файл — ні.
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "помилка: не вказано підкоманду")
		printUsage(stderr)
		return exitUsage
	}

	switch args[0] {
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr, src)
	case "help", "-h", "-help", "--help":
		if len(args) > 1 {
			fmt.Fprintf(stderr, "помилка: help не приймає аргументів: %q\n", args[1:])
			return exitUsage
		}
		printUsage(stdout)
		return exitOK
	case "version":
		if len(args) > 1 {
			fmt.Fprintf(stderr, "помилка: version не приймає аргументів: %q\n", args[1:])
			return exitUsage
		}
		fmt.Fprintf(stdout, "%s %s (%s)\n", programName, version, variantInfo)
		return exitOK
	default:
		fmt.Fprintf(stderr, "помилка: невідома підкоманда %q\n", args[0])
		printUsage(stderr)
		return exitUsage
	}
}

// analyzeOptions — перевірені параметри підкоманди analyze.
type analyzeOptions struct {
	count    int
	seaArea  string
	seed     int64
	useSeed  bool
	limit    float64
	useLimit bool
}

// parseAnalyzeArgs читає й перевіряє прапорці analyze.
// Повертає flag.ErrHelp, якщо користувач запросив довідку.
func parseAnalyzeArgs(args []string) (analyzeOptions, error) {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // повідомлення формуємо самі
	count := fs.Int("count", 0, "кількість значень (5–100), обов'язковий")
	subject := fs.String("subject", "", "назва морської акваторії, обов'язковий")
	seed := fs.Int64("seed", 0, "фіксований seed для відтворюваної генерації")
	limit := fs.Float64("limit", 0, "порогова висота для пошуку першої вищої хвилі, м")

	if err := fs.Parse(args); err != nil {
		return analyzeOptions{}, err
	}
	if fs.NArg() > 0 {
		return analyzeOptions{}, fmt.Errorf("зайві аргументи: %q", fs.Args())
	}

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })
	var missing []string
	for _, name := range []string{"count", "subject"} {
		if !set[name] {
			missing = append(missing, "--"+name)
		}
	}
	if len(missing) > 0 {
		return analyzeOptions{}, fmt.Errorf("відсутні обов'язкові прапорці: %s", strings.Join(missing, ", "))
	}

	if err := analyzer.ValidateCount(*count); err != nil {
		return analyzeOptions{}, err
	}
	area, err := analyzer.NormalizeSeaArea(*subject)
	if err != nil {
		return analyzeOptions{}, err
	}
	if set["limit"] {
		if err := analyzer.ValidateLimit(*limit); err != nil {
			return analyzeOptions{}, err
		}
	}
	return analyzeOptions{
		count: *count, seaArea: area,
		seed: *seed, useSeed: set["seed"],
		limit: *limit, useLimit: set["limit"],
	}, nil
}

func runAnalyze(args []string, stdout, stderr io.Writer, src waveSource) int {
	opts, err := parseAnalyzeArgs(args)
	if errors.Is(err, flag.ErrHelp) {
		printAnalyzeUsage(stdout)
		return exitOK
	}
	if err != nil {
		fmt.Fprintf(stderr, "помилка: %v\n", err)
		fmt.Fprintf(stderr, "Довідка: %s analyze -h\n", programName)
		return exitUsage
	}

	var heights []float64
	if opts.useSeed {
		heights, err = src.GenerateWithSeed(opts.count, opts.seaArea, opts.seed)
	} else {
		heights, err = src.Generate(opts.count, opts.seaArea)
	}
	if err != nil {
		fmt.Fprintf(stderr, "помилка генератора: %v\n", err)
		return exitData
	}
	if len(heights) != opts.count {
		fmt.Fprintf(stderr, "помилка даних: генератор повернув %d значень замість %d\n", len(heights), opts.count)
		return exitData
	}

	series, err := analyzer.NewWaveSeries(opts.seaArea, heights)
	if err != nil {
		fmt.Fprintf(stderr, "помилка даних: %v\n", err)
		return exitData
	}
	printReport(stdout, series.Analyze(), opts)
	if opts.useLimit {
		if pos, ok := series.FirstAbove(opts.limit); ok {
			fmt.Fprintf(stdout, "\nПерша хвиля вища за %.1f м: №%d\n", opts.limit, pos)
		} else {
			fmt.Fprintf(stdout, "\nХвиль, вищих за %.1f м, немає\n", opts.limit)
		}
	}
	return exitOK
}

// printReport виводить звіт; округлення виконується лише тут, для відображення.
func printReport(w io.Writer, r analyzer.Report, opts analyzeOptions) {
	fmt.Fprintf(w, "Акваторія: %s\n", r.SeaArea)
	if opts.useSeed {
		fmt.Fprintf(w, "Seed: %d\n", opts.seed)
	}
	parts := make([]string, len(r.Heights))
	for i, h := range r.Heights {
		parts[i] = fmt.Sprintf("%.1f", h)
	}
	fmt.Fprintf(w, "Висоти хвиль, м (%d): %s\n\n", len(r.Heights), strings.Join(parts, " "))

	fmt.Fprintln(w, "Статистика:")
	fmt.Fprintf(w, "  кількість: %d\n", r.Stats.Count)
	fmt.Fprintf(w, "  сума:      %.2f м\n", r.Stats.Sum)
	fmt.Fprintf(w, "  мінімум:   %.2f м\n", r.Stats.Min)
	fmt.Fprintf(w, "  максимум:  %.2f м\n", r.Stats.Max)
	fmt.Fprintf(w, "  середнє:   %.2f м\n\n", r.Stats.Mean)

	fmt.Fprintln(w, "Категорії:")
	for _, cc := range r.Categories {
		fmt.Fprintf(w, "  %-15s %d\n", cc.Category.String()+":", cc.Count)
	}

	fmt.Fprintln(w, "\nПовтори:")
	if len(r.Repeats) == 0 {
		fmt.Fprintln(w, "  повторів немає")
	}
	for _, rep := range r.Repeats {
		fmt.Fprintf(w, "  %.1f м — %d раз(и), пар: %d\n", rep.Height, rep.Occurrences, rep.Pairs())
	}
	fmt.Fprintf(w, "  усього однакових пар: %d\n", r.TotalPairs())

	if r.HasDangerous() {
		fmt.Fprintln(w, "\n!!! ПОПЕРЕДЖЕННЯ: виявлено небезпечні хвилі понад 5 м !!!")
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `%s %s — аналізатор висот морських хвиль
%s

Використання:
  %s <підкоманда> [прапорці]

Підкоманди:
  analyze   згенерувати й проаналізувати висоти хвиль
  help      показати цю довідку
  version   показати версію

Приклади:
  %s analyze --count 20 --subject "Чорне море"
  %s analyze --count 20 --subject "Чорне море" --seed 42
  %s analyze -h
`, programName, version, variantInfo, programName, programName, programName, programName)
}

func printAnalyzeUsage(w io.Writer) {
	fmt.Fprintf(w, `Використання:
  %s analyze --count N --subject НАЗВА [--seed S]

Прапорці:
  --count N        кількість значень, від %d до %d (обов'язковий)
  --subject НАЗВА  назва морської акваторії, до %d символів (обов'язковий)
  --seed S         ціле число для відтворюваної генерації (необов'язковий)
  --limit H        порогова висота 0–12 м: номер першої вищої хвилі (необов'язковий)

Значення, що починається з "-", передавайте через "=":
  %s analyze --count 10 --subject=-Північ
`, programName, analyzer.MinCount, analyzer.MaxCount, analyzer.MaxSeaAreaRunes, programName)
}
