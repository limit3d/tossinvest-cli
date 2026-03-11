package main

import (
	"errors"
	"fmt"

	tossclient "github.com/junghoonkye/toss-investment-cli/internal/client"
)

func userFacingCommandError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, tossclient.ErrNoSession) {
		return fmt.Errorf("no active session; run `tossctl auth login`")
	}

	if tossclient.IsAuthError(err) {
		return fmt.Errorf("stored session is no longer valid; run `tossctl auth login`")
	}

	return err
}
