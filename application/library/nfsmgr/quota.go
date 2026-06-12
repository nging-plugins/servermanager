//go:build linux

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
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ListQuota runs repquota -a -u and returns parsed quota reports.
func (c *linuxClient) ListQuota(ctx context.Context) ([]*QuotaReport, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "repquota", "-a", "-u")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("repquota: %w", err)
	}
	return parseRepquota(string(out))
}

// parseRepquota parses the output of repquota -a -u.
func parseRepquota(output string) ([]*QuotaReport, error) {
	lines := strings.Split(output, "\n")
	var reports []*QuotaReport
	var current *QuotaReport
	parsingData := false
	const devicePrefix = "*** Report for user quotas on device "

	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")

		// New device section
		if device, ok := strings.CutPrefix(line, devicePrefix); ok {
			current = &QuotaReport{
				Device: device,
			}
			reports = append(reports, current)
			parsingData = false
			continue
		}

		if current == nil {
			continue
		}

		// Separator line → start parsing data rows
		if strings.HasPrefix(line, "---") {
			parsingData = true
			continue
		}

		if !parsingData || line == "" {
			continue
		}

		entry := parseRepquotaRow(line)
		if entry != nil {
			current.Entries = append(current.Entries, entry)
		}
	}
	return reports, nil
}

// parseRepquotaRow parses a single data line from repquota output.
// Format: User  Status  blockUsed blockSoft blockHard [blockGrace] inodeUsed inodeSoft inodeHard [inodeGrace]
func parseRepquotaRow(line string) *QuotaEntry {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	entry := &QuotaEntry{
		User:   fields[0],
		Status: fields[1],
	}

	idx := 2
	if idx < len(fields) {
		entry.BlockUsed = parseUint64(fields[idx])
		idx++
	}
	if idx < len(fields) {
		entry.BlockSoft = parseUint64(fields[idx])
		idx++
	}
	if idx < len(fields) {
		entry.BlockHard = parseUint64(fields[idx])
		idx++
	}
	// Block grace — present only when soft limit exceeded (non-numeric)
	if idx < len(fields) && !isNumeric(fields[idx]) {
		entry.BlockGrace = fields[idx]
		idx++
	}
	if idx < len(fields) {
		entry.InodeUsed = parseUint64(fields[idx])
		idx++
	}
	if idx < len(fields) {
		entry.InodeSoft = parseUint64(fields[idx])
		idx++
	}
	if idx < len(fields) {
		entry.InodeHard = parseUint64(fields[idx])
		idx++
	}
	if idx < len(fields) && !isNumeric(fields[idx]) {
		entry.InodeGrace = fields[idx]
	}

	return entry
}

func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// SetQuota sets disk quota limits for a user via setquota.
func (c *linuxClient) SetQuota(ctx context.Context, limit *QuotaLimit) error {
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "setquota", "-u",
		limit.User,
		strconv.FormatUint(limit.BlockSoft, 10),
		strconv.FormatUint(limit.BlockHard, 10),
		strconv.FormatUint(limit.InodeSoft, 10),
		strconv.FormatUint(limit.InodeHard, 10),
		limit.MountPoint,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("setquota failed: %w\n%s", err, string(out))
	}
	return nil
}
