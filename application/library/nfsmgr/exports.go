/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package nfsmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const exportsPath = "/etc/exports"

// readExportsFile reads and parses /etc/exports, returns entries.
func readExportsFile() ([]*ExportEntry, error) {
	b, err := os.ReadFile(exportsPath)
	if err != nil {
		return nil, err
	}
	return parseExports(string(b))
}

// writeExportsFile atomically writes entries to /etc/exports.
// Uses temp file + rename to prevent corruption.
func writeExportsFile(entries []*ExportEntry) error {
	content := generateExports(entries)
	dir := filepath.Dir(exportsPath)
	tmp, err := os.CreateTemp(dir, ".exports.tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()
	return os.Rename(tmpName, exportsPath)
}

// parseExports parses /etc/exports format text into ExportEntry slice.
func parseExports(text string) ([]*ExportEntry, error) {
	var entries []*ExportEntry
	lines := strings.Split(text, "\n")
	var buf strings.Builder

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		// Strip comments
		if commentIdx := strings.IndexByte(line, '#'); commentIdx >= 0 {
			line = strings.TrimSpace(line[:commentIdx])
		}
		if line == "" {
			continue
		}
		// Continuation line: ends with backslash
		if rest, ok := strings.CutSuffix(line, `\`); ok {
			buf.WriteString(rest)
			buf.WriteByte(' ')
			continue
		}
		buf.WriteString(line)
		full := buf.String()
		buf.Reset()

		entry := parseExportLine(full)
		if entry != nil {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// parseExportLine parses a single (possibly continuation-joined) export line.
// Format: /path    client1(opts) client2(opts)
func parseExportLine(line string) *ExportEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Split path and clients. The path is the first whitespace-separated token.
	parts := splitExportLine(line)
	if len(parts) < 2 {
		return nil
	}

	entry := &ExportEntry{
		Path: parts[0],
	}
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		host, opts := parseExportClient(part)
		client := ExportClient{Host: host, Options: opts}
		entry.Clients = append(entry.Clients, client)
	}
	return entry
}

// splitExportLine splits an exports line respecting parentheses.
func splitExportLine(line string) []string {
	var parts []string
	var buf strings.Builder
	depth := 0
	for _, ch := range line {
		switch ch {
		case '(':
			depth++
			buf.WriteRune(ch)
		case ')':
			depth--
			buf.WriteRune(ch)
		case ' ', '\t':
			if depth > 0 {
				buf.WriteRune(ch)
			} else {
				if buf.Len() > 0 {
					parts = append(parts, buf.String())
					buf.Reset()
				}
			}
		default:
			buf.WriteRune(ch)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

// parseExportClient parses a client spec like "192.168.1.0/24(rw,sync)".
func parseExportClient(s string) (host string, options []string) {
	host, rest, ok := strings.Cut(s, "(")
	if !ok {
		return s, nil
	}
	host = strings.TrimSpace(host)
	inner := strings.TrimSuffix(rest, ")")
	for opt := range strings.SplitSeq(inner, ",") {
		opt = strings.TrimSpace(opt)
		if opt != "" {
			options = append(options, opt)
		}
	}
	return host, options
}

// generateExports generates /etc/exports format text from entries.
func generateExports(entries []*ExportEntry) string {
	var b strings.Builder
	for _, e := range entries {
		if e.Comment != "" {
			b.WriteString("# ")
			b.WriteString(e.Comment)
			b.WriteByte('\n')
		}
		b.WriteString(e.Path)
		// Pad to tab position
		if len(e.Path) < 16 {
			b.WriteByte('\t')
		}
		b.WriteByte('\t')

		for i, c := range e.Clients {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(c.Host)
			if len(c.Options) > 0 {
				b.WriteByte('(')
				b.WriteString(strings.Join(c.Options, ","))
				b.WriteByte(')')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// FormatExportClient formats a client spec for display.
func FormatExportClient(c *ExportClient) string {
	if len(c.Options) == 0 {
		return c.Host
	}
	return fmt.Sprintf("%s(%s)", c.Host, strings.Join(c.Options, ","))
}
