package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type LoginResult struct {
	Status           string `json:"status"`
	Message          string `json:"message,omitempty"`
	StorageStatePath string `json:"storage_state_path"`
	CookieCount      int    `json:"cookie_count,omitempty"`
	OriginCount      int    `json:"origin_count,omitempty"`
}

type LoginRunner interface {
	Login(context.Context, LoginConfig) (*LoginResult, error)
}

type PythonLoginRunner struct{}

func (PythonLoginRunner) Login(ctx context.Context, cfg LoginConfig) (*LoginResult, error) {
	if cfg.HelperDir == "" {
		return nil, fmt.Errorf("auth helper directory is not configured")
	}
	if cfg.StorageStatePath == "" {
		return nil, fmt.Errorf("auth helper storage-state path is not configured")
	}

	cmd := exec.CommandContext(
		ctx,
		cfg.PythonBin,
		"-m",
		"tossctl_auth_helper",
		"login",
		"--storage-state",
		cfg.StorageStatePath,
	)
	cmd.Dir = cfg.HelperDir
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("auth helper failed: %w", err)
	}

	var result LoginResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("decode auth helper output: %w", err)
	}
	if result.Status != "ok" {
		if result.Message == "" {
			result.Message = "auth helper returned a non-ok status"
		}
		return nil, fmt.Errorf("%s", result.Message)
	}
	if result.StorageStatePath == "" {
		result.StorageStatePath = cfg.StorageStatePath
	}

	return &result, nil
}
