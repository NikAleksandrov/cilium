// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package helpers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/vladimirvivien/gexe"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type ciliumCLI struct {
	cmd string
	e   *gexe.Echo
}

func newCiliumCLI() *ciliumCLI { return &ciliumCLI{cmd: "cilium", e: gexe.New()} }

func (c *ciliumCLI) findOrInstall(ctx context.Context) error {
	if _, err := exec.LookPath(c.cmd); err != nil {
		// TODO: try to install cilium-cli using `go install` or similar
		return fmt.Errorf("cilium CLI not installed or could not be found: %w", err)
	}

	stdout := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, c.cmd, "version")
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to determine cilium-cli version: %w", err)
	}

	s := strings.Split(stdout.String(), "\n")
	if len(s) > 0 {
		klog.V(4).Info("Found cilium-cli version", s[0])
	}

	// TODO: check against expected cilium-cli version?

	return nil
}

func (c *ciliumCLI) install(ctx context.Context, helmOpts map[string]string) (context.Context, error) {
	if err := c.findOrInstall(ctx); err != nil {
		return ctx, err
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, c.cmd, "install")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return ctx, fmt.Errorf("failed to determine cilium-cli version: %w", err)
	}

	return ctx, nil
}

func (c *ciliumCLI) uninstall(ctx context.Context) (context.Context, error) {
	if err := c.findOrInstall(ctx); err != nil {
		return ctx, err
	}

	return ctx, nil
}

type ciliumContextKey string

// InstallCiliumWithOpts installs Cilium with the provided Helm options.
func InstallCiliumWithOpts(helmOpts map[string]string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return newCiliumCLI().install(ctx, helmOpts)
	}
}

func InstallCilium() env.Func {
	return InstallCiliumWithOpts(nil)
}

func UninstallCilium() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		return newCiliumCLI().uninstall(ctx)
	}
}
