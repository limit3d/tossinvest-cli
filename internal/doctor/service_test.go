package doctor

import (
	"strings"
	"testing"

	"github.com/junghoonkye/tossinvest-cli/internal/config"
)

func TestCheckLiveOrderActionsDisabled(t *testing.T) {
	check := checkLiveOrderActions(config.Status{})
	if check.Status != CheckInfo {
		t.Fatalf("expected info status, got %s", check.Status)
	}
	if check.Name != "live_order_actions" {
		t.Fatalf("unexpected check name: %s", check.Name)
	}
}

func TestCheckDangerousAutomationEnabled(t *testing.T) {
	check := checkDangerousAutomation(config.Status{
		Trading: config.Trading{
			DangerousAutomation: config.DangerousAutomation{
				CompleteTradeAuth: true,
				AcceptFXConsent:   true,
			},
		},
	})
	if check.Status != CheckWarn {
		t.Fatalf("expected warn status, got %s", check.Status)
	}
	if !strings.Contains(check.Detail, "complete_trade_auth") {
		t.Fatalf("unexpected dangerous automation detail: %s", check.Detail)
	}
	if !strings.Contains(check.Detail, "accept_fx_consent") {
		t.Fatalf("unexpected dangerous automation detail: %s", check.Detail)
	}
}

func TestCheckLegacyConfig(t *testing.T) {
	check := checkLegacyConfig(config.Status{
		Exists:       true,
		LegacyFields: []string{"trading.allow_dangerous_execute"},
	})
	if check.Status != CheckWarn {
		t.Fatalf("expected warn status, got %s", check.Status)
	}
	if check.Name != "legacy_config" {
		t.Fatalf("unexpected check name: %s", check.Name)
	}
}
