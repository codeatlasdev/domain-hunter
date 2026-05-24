package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/codeatlasdev/domain-hunter/internal/scanner"
)

type Format string

const (
	FormatTXT  Format = "txt"
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

type Exporter struct {
	formats        []Format
	files          map[Format]*os.File
	csvWriter      *csv.Writer
	jsonItems      []jsonEntry
	mu             sync.Mutex
	timestamp      string
	showRegistered bool
	regFile        *os.File
}

type jsonEntry struct {
	Domain    string `json:"domain"`
	TLD       string `json:"tld"`
	CheckedAt string `json:"checked_at"`
}

func New(formats []Format) (*Exporter, error) {
	return NewWithOptions(formats, false)
}

func NewWithOptions(formats []Format, showRegistered bool) (*Exporter, error) {
	ts := time.Now().Format("20060102-150405")
	e := &Exporter{
		formats:        formats,
		files:          make(map[Format]*os.File),
		timestamp:      ts,
		showRegistered: showRegistered,
	}

	for _, f := range formats {
		filename := fmt.Sprintf("results-%s.%s", ts, f)
		file, err := os.Create(filename)
		if err != nil {
			e.Close()
			return nil, fmt.Errorf("create %s: %w", filename, err)
		}
		e.files[f] = file

		if f == FormatCSV {
			e.csvWriter = csv.NewWriter(file)
			e.csvWriter.Write([]string{"domain", "tld", "checked_at"})
		}
	}

	if showRegistered {
		filename := fmt.Sprintf("registered-%s.txt", ts)
		file, err := os.Create(filename)
		if err != nil {
			e.Close()
			return nil, fmt.Errorf("create %s: %w", filename, err)
		}
		e.regFile = file
	}

	return e, nil
}

func (e *Exporter) Append(r scanner.Result) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !r.Available && !r.Error && e.showRegistered && e.regFile != nil {
		fmt.Fprintln(e.regFile, r.Domain)
	}

	if !r.Available {
		return
	}

	ts := r.Timestamp.Format(time.RFC3339)

	for _, f := range e.formats {
		switch f {
		case FormatTXT:
			fmt.Fprintln(e.files[f], r.Domain)
		case FormatCSV:
			e.csvWriter.Write([]string{r.Domain, r.TLD, ts})
			e.csvWriter.Flush()
		case FormatJSON:
			e.jsonItems = append(e.jsonItems, jsonEntry{
				Domain:    r.Domain,
				TLD:       r.TLD,
				CheckedAt: ts,
			})
		}
	}
}

func (e *Exporter) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Finalize JSON
	if f, ok := e.files[FormatJSON]; ok && f != nil {
		data, _ := json.MarshalIndent(e.jsonItems, "", "  ")
		f.Truncate(0)
		f.Seek(0, 0)
		f.Write(data)
	}

	for _, f := range e.files {
		if f != nil {
			f.Close()
		}
	}

	if e.regFile != nil {
		e.regFile.Close()
	}
}

func (e *Exporter) Filenames() []string {
	var names []string
	for _, f := range e.formats {
		names = append(names, fmt.Sprintf("results-%s.%s", e.timestamp, f))
	}
	if e.showRegistered {
		names = append(names, fmt.Sprintf("registered-%s.txt", e.timestamp))
	}
	return names
}
