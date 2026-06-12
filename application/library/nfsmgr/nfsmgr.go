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

// Package nfsmgr provides NFS (Network File System) management capabilities.
package nfsmgr

import (
	"context"
	"errors"
)

var (
	ErrNotSupported = errors.New("NFS management is not supported on this platform")
	ErrNoEntry      = errors.New("NFS entry not found")
)

// ExportClient represents a client/host allowed to access an NFS export.
type ExportClient struct {
	Host    string   `json:"host"`    // Client IP/network/hostname, e.g. "192.168.1.0/24" or "*"
	Options []string `json:"options"` // Per-client options, e.g. ["rw", "no_root_squash"]
}

// ExportEntry represents a single export entry in /etc/exports.
type ExportEntry struct {
	Path    string         `json:"path"`    // Export path, e.g. "/data"
	Clients []ExportClient `json:"clients"` // Clients allowed to access
	Comment string         `json:"comment,omitempty"`
}

// MountEntry represents a mounted NFS share.
type MountEntry struct {
	Server     string   `json:"server"`     // NFS server address
	Remote     string   `json:"remote"`     // Remote export path
	MountPoint string   `json:"mountPoint"` // Local mount point
	Type       string   `json:"type"`       // "nfs" or "nfs4"
	Options    []string `json:"options"`    // Mount options
	ReadOnly   bool     `json:"readOnly"`   // Whether mount is read-only
}

// NFSStatus represents the NFS server service status.
type NFSStatus struct {
	Running  bool   `json:"running"`
	Enabled  bool   `json:"enabled"`
	Active   string `json:"active"`   // systemd active state
	SubState string `json:"subState"` // systemd sub state
}

// Client defines the interface for NFS management operations.
type Client interface {
	// Export management
	ListExports(ctx context.Context) ([]*ExportEntry, error)
	WriteExports(ctx context.Context, entries []*ExportEntry) error
	ReloadExports(ctx context.Context) error

	// Mount management
	ListMounts(ctx context.Context) ([]*MountEntry, error)
	Mount(ctx context.Context, entry *MountEntry) error
	Unmount(ctx context.Context, mountPoint string) error

	// Server status
	ServerStatus(ctx context.Context) (*NFSStatus, error)

	// Quota management
	ListQuota(ctx context.Context) ([]*QuotaReport, error)
	SetQuota(ctx context.Context, limit *QuotaLimit) error
}

// QuotaEntry represents a single user quota entry from repquota.
type QuotaEntry struct {
	User       string `json:"user"`
	BlockUsed  uint64 `json:"blockUsed"`
	BlockSoft  uint64 `json:"blockSoft"`
	BlockHard  uint64 `json:"blockHard"`
	BlockGrace string `json:"blockGrace,omitempty"`
	InodeUsed  uint64 `json:"inodeUsed"`
	InodeSoft  uint64 `json:"inodeSoft"`
	InodeHard  uint64 `json:"inodeHard"`
	InodeGrace string `json:"inodeGrace,omitempty"`
	Status     string `json:"status"`
}

// QuotaReport contains quota info for one device.
type QuotaReport struct {
	Device      string        `json:"device"`
	Entries     []*QuotaEntry `json:"entries"`
}

// BlockPercent returns usage percentage of hard block limit (0-100).
func (e *QuotaEntry) BlockPercent() int {
	if e.BlockHard == 0 {
		return 0
	}
	pct := int(e.BlockUsed * 100 / e.BlockHard)
	if pct > 100 {
		return 100
	}
	return pct
}

// BlockAvail returns available blocks before hitting hard limit.
func (e *QuotaEntry) BlockAvail() uint64 {
	if e.BlockHard == 0 || e.BlockUsed >= e.BlockHard {
		return 0
	}
	return e.BlockHard - e.BlockUsed
}

// InodePercent returns usage percentage of hard inode limit (0-100).
func (e *QuotaEntry) InodePercent() int {
	if e.InodeHard == 0 {
		return 0
	}
	pct := int(e.InodeUsed * 100 / e.InodeHard)
	if pct > 100 {
		return 100
	}
	return pct
}

// InodeAvail returns available inodes before hitting hard limit.
func (e *QuotaEntry) InodeAvail() uint64 {
	if e.InodeHard == 0 || e.InodeUsed >= e.InodeHard {
		return 0
	}
	return e.InodeHard - e.InodeUsed
}

// QuotaLimit defines quota limits to apply via setquota.
type QuotaLimit struct {
	User       string `json:"user"`       // Username
	MountPoint string `json:"mountPoint"` // Mount point or device path
	BlockSoft  uint64 `json:"blockSoft"`  // Soft limit for blocks (1KB units, 0=unlimited)
	BlockHard  uint64 `json:"blockHard"`  // Hard limit for blocks
	InodeSoft  uint64 `json:"inodeSoft"`  // Soft limit for inodes
	InodeHard  uint64 `json:"inodeHard"`  // Hard limit for inodes
}
