package ccpanel

import (
	"testing"

	"github.com/WncFht/devin2api/internal/store"
)

func TestPlanFloatNilIsUnreported(t *testing.T) {
	if planFloat(map[string]any{"weekly_quota_remaining": nil}, "weekly_quota_remaining") != nil {
		t.Fatal("nil value must stay unreported, not become 0")
	}
	if planFloat(map[string]any{}, "weekly_quota_remaining") != nil {
		t.Fatal("absent key must stay unreported")
	}
}

func TestQuotaLimitsOverageDebtAndUnreportedWeekly(t *testing.T) {
	daily := 100.0
	point := &store.QuotaSample{DailyRemaining: &daily, OverageBalanceMicros: -6065858, PromptCredits: -1}
	limits := quotaLimits(point)
	if limits["state"] != "overage" {
		t.Fatalf("state = %v, want overage", limits["state"])
	}
	if limits["weekly_percent"] != nil {
		t.Fatalf("weekly must be null when unreported, got %v", limits["weekly_percent"])
	}
	if got := limits["overage_usd"].(float64); got > -6.0 || got < -6.1 {
		t.Fatalf("overage_usd = %v, want ~-6.07", got)
	}
	if _, ok := limits["prompt_credits"]; ok {
		t.Fatal("prompt_credits -1 must be treated as not applicable")
	}
}

func TestQuotaLimitsStates(t *testing.T) {
	zero, low, ok := 0.0, 10.0, 80.0
	cases := []struct {
		name  string
		point *store.QuotaSample
		want  string
	}{
		{"weekly exhausted", &store.QuotaSample{WeeklyRemaining: &zero}, "weekly_exhausted"},
		{"daily low", &store.QuotaSample{DailyRemaining: &low}, "daily_low"},
		{"healthy", &store.QuotaSample{DailyRemaining: &ok, WeeklyRemaining: &ok}, "ok"},
		{"nothing reported", &store.QuotaSample{}, "unknown"},
	}
	for _, c := range cases {
		if got := quotaLimits(c.point)["state"]; got != c.want {
			t.Errorf("%s: state = %v, want %s", c.name, got, c.want)
		}
	}
}
