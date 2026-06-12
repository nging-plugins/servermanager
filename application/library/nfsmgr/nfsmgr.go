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
}
