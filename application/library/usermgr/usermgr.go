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

// Package usermgr provides Linux system user account management.
package usermgr

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotSupported  = errors.New("user management is not supported on this platform")
	ErrRootDeletion  = errors.New("cannot delete root user")
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidInput  = errors.New("invalid input")
)

// User represents a Linux system user account.
type User struct {
	Username string   `json:"username"`
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
	Comment  string   `json:"comment"`  // GECOS field (full name)
	HomeDir  string   `json:"homeDir"`
	Shell    string   `json:"shell"`
	Locked   bool     `json:"locked"`   // shadow entry starts with "!"
	System   bool     `json:"system"`   // UID < 1000
	Groups   []string `json:"groups"`   // supplementary groups
}

// IsRoot returns true if this is the root user.
func (u *User) IsRoot() bool {
	return u.Username == "root" || u.UID == 0
}

// ValidatePassword validates a password before it is handed to system tools
// such as chpasswd. chpasswd parses its stdin line by line as
// "username:password" records, so a password containing a newline or carriage
// return would let a caller inject additional records (e.g. an arbitrary
// "root:..." line), and a colon would corrupt the record format. Rejecting
// these characters mirrors the denylist used by validateUsername().
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("%w: password cannot be empty", ErrInvalidInput)
	}
	if len(password) > 4096 {
		return fmt.Errorf("%w: password too long (max 4096)", ErrInvalidInput)
	}
	for _, c := range password {
		if c == ':' || c == '\n' || c == '\r' {
			return fmt.Errorf("%w: invalid character in password", ErrInvalidInput)
		}
	}
	return nil
}

// Client defines the interface for system user management.
type Client interface {
	List(ctx context.Context) ([]*User, error)
	Get(ctx context.Context, username string) (*User, error)
	Add(ctx context.Context, u *User, password string) error
	Edit(ctx context.Context, username string, u *User, password string) error
	Delete(ctx context.Context, username string, removeHome bool) error
	Lock(ctx context.Context, username string) error
	Unlock(ctx context.Context, username string) error
	AvailableShells(ctx context.Context) ([]string, error)
}
