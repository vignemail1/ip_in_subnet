package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"
)

func main() {
	delimiter := flag.String("delimiter", ",", "CSV field delimiter (single character, or \\t for tab)")
	skipHeader := flag.Bool("skip-header", false, "skip the first line (header)")
	column := flag.Int("column", 1, "1-based column number containing the subnet CIDR")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <ip>\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Reads CSV from stdin and prints rows whose subnet column contains the given IP.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if *column < 1 {
		fmt.Fprintln(os.Stderr, "error: -column must be >= 1")
		os.Exit(2)
	}

	delim, err := parseDelimiter(*delimiter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid -delimiter: %v\n", err)
		os.Exit(2)
	}

	ip, err := netip.ParseAddr(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid IP %q: %v\n", flag.Arg(0), err)
		os.Exit(2)
	}

	exitCode, err := scanCSV(os.Stdin, ip, delim, *skipHeader, *column)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func parseDelimiter(s string) (rune, error) {
	if s == `\t` {
		return '\t', nil
	}

	r := []rune(s)
	if len(r) != 1 {
		return 0, fmt.Errorf("must be exactly one character, or \\t for tab")
	}

	return r[0], nil
}

// scanCSV returns exit code 0 if at least one row matched, 1 if none.
func scanCSV(r io.Reader, ip netip.Addr, delimiter rune, skipHeader bool, column int) (int, error) {
	cr := csv.NewReader(bufio.NewReader(r))
	cr.Comma = delimiter
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	columnIndex := column - 1
	lineNum := 0
	matched := 0

	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 1, fmt.Errorf("line %d: %w", lineNum+1, err)
		}

		lineNum++

		if skipHeader && lineNum == 1 {
			continue
		}

		if columnIndex >= len(record) {
			continue
		}

		subnet := strings.TrimSpace(record[columnIndex])
		if subnet == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(subnet)
		if err != nil {
			continue
		}

		if prefix.Contains(ip) {
			fmt.Println(joinRecord(record, delimiter))
			matched++
		}
	}

	if matched == 0 {
		return 1, nil
	}

	return 0, nil
}

func joinRecord(record []string, delimiter rune) string {
	var b strings.Builder

	for i, field := range record {
		if i > 0 {
			b.WriteRune(delimiter)
		}

		needsQuote := strings.ContainsRune(field, delimiter) ||
			strings.ContainsRune(field, '"') ||
			strings.ContainsRune(field, '\n') ||
			strings.ContainsRune(field, '\r')

		if needsQuote {
			b.WriteByte('"')
			b.WriteString(strings.ReplaceAll(field, `"`, `""`))
			b.WriteByte('"')
		} else {
			b.WriteString(field)
		}
	}

	return b.String()
}
