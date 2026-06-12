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
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const cmdTimeout = 30 * time.Second

// linuxClient implements Client on Linux via system commands and file parsing.
type linuxClient struct{}

// NewClient creates an NFS management client for the current platform.
// On non-Linux platforms it returns ErrNotSupported.
func NewClient() (Client, error) {
	return &linuxClient{}, nil
}

// ListExports reads /etc/exports and returns parsed entries.
func (c *linuxClient) ListExports(ctx context.Context) ([]*ExportEntry, error) {
	return readExportsFile()
}

// WriteExports atomically writes entries to /etc/exports.
func (c *linuxClient) WriteExports(ctx context.Context, entries []*ExportEntry) error {
	return writeExportsFile(entries)
}

// ReloadExports runs exportfs -r to reload exports from /etc/exports.
func (c *linuxClient) ReloadExports(ctx context.Context) error {
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "exportfs", "-r")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exportfs -r failed: %w\n%s", err, string(out))
	}
	return nil
}

// ListMounts parses /proc/mounts for NFS mounts.
func (c *linuxClient) ListMounts(ctx context.Context) ([]*MountEntry, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var mounts []*MountEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		fsType := fields[2]
		if fsType != "nfs" && fsType != "nfs4" {
			continue
		}
		// fields[0] = spec, fields[1] = mount point, fields[2] = type, fields[3] = opts
		spec := fields[0]
		mountPoint := fields[1]
		opts := strings.Split(fields[3], ",")

		var server, remote string
		if srv, rem, ok := strings.Cut(spec, ":"); ok {
			server = srv
			remote = rem
		} else {
			server = spec
		}

		readOnly := false
		for _, opt := range opts {
			if opt == "ro" {
				readOnly = true
			} else if opt == "rw" {
				readOnly = false
			}
		}
		mounts = append(mounts, &MountEntry{
			Server:     server,
			Remote:     remote,
			MountPoint: mountPoint,
			Type:       fsType,
			Options:    opts,
			ReadOnly:   readOnly,
		})
	}
	return mounts, scanner.Err()
}

// Mount mounts an NFS share.
func (c *linuxClient) Mount(ctx context.Context, entry *MountEntry) error {
	args := []string{"-t", entry.Type}
	if len(entry.Options) > 0 {
		args = append(args, "-o", strings.Join(entry.Options, ","))
	}
	spec := fmt.Sprintf("%s:%s", entry.Server, entry.Remote)
	args = append(args, spec, entry.MountPoint)

	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "mount", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %w\n%s", err, string(out))
	}
	return nil
}

// Unmount unmounts an NFS share by mount point.
func (c *linuxClient) Unmount(ctx context.Context, mountPoint string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "umount", mountPoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount failed: %w\n%s", err, string(out))
	}
	return nil
}

// ServerStatus checks NFS server service status via systemctl.
func (c *linuxClient) ServerStatus(ctx context.Context) (*NFSStatus, error) {
	status := &NFSStatus{}

	// Check if nfs-server is active
	cmdCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "systemctl", "is-active", "nfs-server")
	out, _ := cmd.Output()
	status.Active = strings.TrimSpace(string(out))
	status.Running = status.Active == "active"

	// Check if nfs-server is enabled at boot
	cmdCtx2, cancel2 := context.WithTimeout(ctx, cmdTimeout)
	defer cancel2()
	cmd2 := exec.CommandContext(cmdCtx2, "systemctl", "is-enabled", "nfs-server")
	out2, _ := cmd2.Output()
	status.Enabled = strings.TrimSpace(string(out2)) == "enabled"

	// Get sub state
	cmdCtx3, cancel3 := context.WithTimeout(ctx, cmdTimeout)
	defer cancel3()
	cmd3 := exec.CommandContext(cmdCtx3, "systemctl", "show", "nfs-server", "--property=SubState", "--value")
	out3, _ := cmd3.Output()
	status.SubState = strings.TrimSpace(string(out3))

	return status, nil
}
