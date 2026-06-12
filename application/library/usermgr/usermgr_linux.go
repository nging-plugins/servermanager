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

package usermgr

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const userCmdTimeout = 30 * time.Second

type linuxClient struct{}

// NewClient creates a user management client for Linux.
func NewClient() (Client, error) {
	return &linuxClient{}, nil
}

func (c *linuxClient) List(ctx context.Context) ([]*User, error) {
	// Read /etc/passwd
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer passwd.Close()

	// Pre-read shadow and group for status and groups
	shadowMap := readShadowMap()
	groupMap := readGroupMap()

	var users []*User
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(parts[2])
		gid, _ := strconv.Atoi(parts[3])
		username := parts[0]

		u := &User{
			Username: username,
			UID:      uid,
			GID:      gid,
			Comment:  parts[4],
			HomeDir:  parts[5],
			Shell:    parts[6],
			Locked:   shadowMap[username],
			System:   uid < 1000 && uid > 0, // root (0) is not "system"
			Groups:   groupMap[username],
		}
		users = append(users, u)
	}
	return users, scanner.Err()
}

func readShadowMap() map[string]bool {
	locked := map[string]bool{}
	f, err := os.Open("/etc/shadow")
	if err != nil {
		return locked
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 || parts[0] == "" {
			continue
		}
		// Password field starts with "!" or "*" or is empty → locked
		pw := parts[1]
		locked[parts[0]] = pw == "" || strings.HasPrefix(pw, "!") || strings.HasPrefix(pw, "*")
	}
	scanner.Err() // ignore read errors, return best-effort map
	return locked
}

func readGroupMap() map[string][]string {
	groups := map[string][]string{}
	f, err := os.Open("/etc/group")
	if err != nil {
		return groups
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		groupName := parts[0]
		members := strings.Split(parts[3], ",")
		for _, m := range members {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			groups[m] = append(groups[m], groupName)
		}
	}
	scanner.Err() // ignore read errors, return best-effort map
	return groups
}

func (c *linuxClient) Get(ctx context.Context, username string) (*User, error) {
	// Use getent to find the user
	cmd := exec.CommandContext(ctx, "getent", "passwd", username)
	out, err := cmd.Output()
	if err != nil {
		return nil, ErrUserNotFound
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return nil, ErrUserNotFound
	}
	parts := strings.Split(line, ":")
	if len(parts) < 7 {
		return nil, ErrUserNotFound
	}
	uid, _ := strconv.Atoi(parts[2])
	gid, _ := strconv.Atoi(parts[3])

	shadowMap := readShadowMap()
	groupMap := readGroupMap()

	return &User{
		Username: parts[0],
		UID:      uid,
		GID:      gid,
		Comment:  parts[4],
		HomeDir:  parts[5],
		Shell:    parts[6],
		Locked:   shadowMap[parts[0]],
		System:   uid < 1000 && uid > 0,
		Groups:   groupMap[parts[0]],
	}, nil
}

func (c *linuxClient) Add(ctx context.Context, u *User, password string) error {
	if err := validateUsername(u.Username); err != nil {
		return err
	}
	if password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	args := buildUserAddArgs(u)
	cmdCtx, cancel := context.WithTimeout(ctx, userCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "useradd", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("useradd failed: %w\n%s", err, string(out))
	}
	if err := setPassword(ctx, u.Username, password); err != nil {
		return err
	}
	return nil
}

func buildUserAddArgs(u *User) []string {
	args := []string{"-m"}
	if u.HomeDir != "" {
		args = append(args, "-d", u.HomeDir)
	}
	if u.Shell != "" {
		args = append(args, "-s", u.Shell)
	}
	if u.Comment != "" {
		args = append(args, "-c", u.Comment)
	}
	if len(u.Groups) > 0 {
		args = append(args, "-G", strings.Join(u.Groups, ","))
	}
	args = append(args, u.Username)
	return args
}

func (c *linuxClient) Edit(ctx context.Context, username string, u *User, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	args := buildUserModArgs(u)
	if len(args) == 0 && password == "" {
		return nil // Nothing to change
	}
	if len(args) > 0 {
		args = append(args, username)
		cmdCtx, cancel := context.WithTimeout(ctx, userCmdTimeout)
		defer cancel()
		cmd := exec.CommandContext(cmdCtx, "usermod", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("usermod failed: %w\n%s", err, string(out))
		}
	}
	if password != "" {
		if err := setPassword(ctx, username, password); err != nil {
			return err
		}
	}
	return nil
}

func buildUserModArgs(u *User) []string {
	var args []string
	if u.HomeDir != "" {
		args = append(args, "-d", u.HomeDir)
	}
	if u.Shell != "" {
		args = append(args, "-s", u.Shell)
	}
	if u.Comment != "" {
		args = append(args, "-c", u.Comment)
	}
	if len(u.Groups) > 0 {
		args = append(args, "-aG", strings.Join(u.Groups, ","))
	}
	return args
}

func (c *linuxClient) Delete(ctx context.Context, username string, removeHome bool) error {
	if username == "root" {
		return ErrRootDeletion
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	args := []string{}
	if removeHome {
		args = append(args, "-r")
	}
	args = append(args, username)
	cmdCtx, cancel := context.WithTimeout(ctx, userCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "userdel", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("userdel failed: %w\n%s", err, string(out))
	}
	return nil
}

func (c *linuxClient) Lock(ctx context.Context, username string) error {
	return passwdCmd(ctx, "-l", username)
}

func (c *linuxClient) Unlock(ctx context.Context, username string) error {
	return passwdCmd(ctx, "-u", username)
}

func passwdCmd(ctx context.Context, flag string, username string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, userCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "passwd", flag, username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("passwd %s failed: %w\n%s", flag, err, string(out))
	}
	return nil
}

func setPassword(ctx context.Context, username, password string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, userCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "chpasswd")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		stdin.Write([]byte(username + ":" + password + "\n"))
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chpasswd failed: %w\n%s", err, string(out))
	}
	return nil
}

func (c *linuxClient) AvailableShells(ctx context.Context) ([]string, error) {
	f, err := os.Open("/etc/shells")
	if err != nil {
		// Fallback to common shells
		return []string{"/bin/bash", "/bin/sh", "/bin/dash", "/bin/zsh", "/usr/sbin/nologin"}, nil
	}
	defer f.Close()
	var shells []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		shells = append(shells, line)
	}
	if len(shells) == 0 {
		shells = []string{"/bin/bash", "/bin/sh"}
	}
	return shells, scanner.Err()
}

func validateUsername(name string) error {
	if name == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrInvalidInput)
	}
	if len(name) > 32 {
		return fmt.Errorf("%w: username too long (max 32)", ErrInvalidInput)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("%w: username cannot start with '-'", ErrInvalidInput)
	}
	for _, c := range name {
		if c == ':' || c == '\n' || c == '\r' || c == '!' || c == ',' {
			return fmt.Errorf("%w: invalid character in username", ErrInvalidInput)
		}
	}
	return nil
}
