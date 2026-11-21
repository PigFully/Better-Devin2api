package llm

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
)

func TestWeeklyQuotaExhaustedBecomesGenericBalanceError(t *testing.T) {
	upstream := connect.NewError(connect.CodeFailedPrecondition,
		errors.New("Your weekly usage quota has been exhausted. Visit https://app.devin.ai/settings/usage to purchase more."))
	failure := Classify(upstream)
	if !failure.InsufficientBalance {
		t.Fatal("weekly quota exhaustion not classified as insufficient balance")
	}
	if got := failure.Error(); got != "Insufficient account balance" {
		t.Fatalf("Error() = %q, want generic message without code prefix or upstream text", got)
	}
}

func TestOtherFailedPreconditionUnchanged(t *testing.T) {
	failure := Classify(connect.NewError(connect.CodeFailedPrecondition, errors.New("CASCADE required")))
	if failure.InsufficientBalance {
		t.Fatal("unrelated failed_precondition must not be treated as quota exhaustion")
	}
}

func TestDevinSpecificMessagesRewrittenToStandard(t *testing.T) {
	cases := map[string]string{
		"invalid_argument: This request was flagged by Anthropic's cyber verification policy. You can enroll in the Anthropic Cyber Verification Program at https://windsurf.com/settings to get access.": "Request blocked by cyber safety policy.",
		"invalid_argument: This request was refused by Anthropic's content policy. Please retry with a different model.":                                                                                  "Request blocked by content safety policy.",
		"invalid_argument: There is an issue with this request, please try a different model.":                                                                                                            "Invalid request.",
	}
	for raw, want := range cases {
		failure := ClassifyText(raw)
		if failure.Error() != want {
			t.Fatalf("Error() = %q, want %q", failure.Error(), want)
		}
		if failure.Code != "invalid_argument" {
			t.Fatalf("code = %q, want invalid_argument (status mapping must be kept)", failure.Code)
		}
	}
}
