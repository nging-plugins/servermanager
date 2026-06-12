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
	"os"
	"path/filepath"
	"strings"
)

const fstabPath = "/etc/fstab"

type fstabEntry struct {
	Device     string
	MountPoint string
	Type       string
	Options    []string
	Dump       string
	Pass       string
}

// readFstab parses /etc/fstab and returns only NFS entries.
func readFstab() ([]*fstabEntry, error) {
	b, err := os.ReadFile(fstabPath)
	if err != nil {
		return nil, err
	}
	return parseFstab(string(b)), nil
}

func parseFstab(text string) []*fstabEntry {
	var entries []*fstabEntry
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		fstype := fields[2]
		if fstype != "nfs" && fstype != "nfs4" {
			continue
		}
		entries = append(entries, &fstabEntry{
			Device:     fields[0],
			MountPoint: fields[1],
			Type:       fstype,
			Options:    strings.Split(fields[3], ","),
			Dump:       fstabField(fields, 4, "0"),
			Pass:       fstabField(fields, 5, "0"),
		})
	}
	return entries
}

func fstabField(fields []string, idx int, def string) string {
	if idx < len(fields) {
		return fields[idx]
	}
	return def
}

// AddFstabEntry adds an NFS mount entry to /etc/fstab.
// No-op if mount point already exists in fstab.
func AddFstabEntry(entry *MountEntry) error {
	existing, err := readFstab()
	if err != nil {
		return err
	}
	for _, e := range existing {
		if e.MountPoint == entry.MountPoint {
			return nil // Already exists
		}
	}

	device := entry.Server + ":" + entry.Remote
	line := device + "\t" + entry.MountPoint + "\t" + entry.Type + "\t" +
		strings.Join(entry.Options, ",") + "\t0\t0\n"

	f, err := os.OpenFile(fstabPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}

// RemoveFstabEntry removes an NFS entry from /etc/fstab by mount point.
// Uses atomic write (temp file + rename).
func RemoveFstabEntry(mountPoint string) error {
	b, err := os.ReadFile(fstabPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	var newLines []string
	changed := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 && fields[1] == mountPoint {
			fstype := fields[2]
			if fstype == "nfs" || fstype == "nfs4" {
				changed = true
				continue // Skip this line
			}
		}
		newLines = append(newLines, line)
	}
	if !changed {
		return nil
	}

	dir := filepath.Dir(fstabPath)
	tmp, err := os.CreateTemp(dir, ".fstab.tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(strings.Join(newLines, "\n")); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()
	if err := os.Rename(tmpName, fstabPath); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
