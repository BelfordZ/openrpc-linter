package reporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/open-rpc/openrpc-linter/types"
)

func TestFormatSeverityColPlainPadsToFixedWidth(t *testing.T) {
	cases := []struct {
		severity types.Severity
		want     string
	}{
		{types.SeverityError, "error  "},
		{types.SeverityWarn, "warning"},
		{types.SeverityInfo, "info   "},
		{"", "error  "},
	}
	for _, tc := range cases {
		got := formatSeverityCol(tc.severity, false)
		if got != tc.want {
			t.Errorf("severity %q: got %q (len %d), want %q (len %d)", tc.severity, got, len(got), tc.want, len(tc.want))
		}
	}
}

func TestFormatSeverityColColorWrapsLabel(t *testing.T) {
	cases := []struct {
		severity types.Severity
		code     string
	}{
		{types.SeverityError, "\x1b[91m"},
		{types.SeverityWarn, "\x1b[93m"},
		{types.SeverityInfo, "\x1b[94m"},
	}
	for _, tc := range cases {
		got := formatSeverityCol(tc.severity, true)
		if !strings.HasPrefix(got, tc.code) {
			t.Errorf("severity %q: expected prefix %q, got %q", tc.severity, tc.code, got)
		}
		if !strings.Contains(got, ansiReset) {
			t.Errorf("severity %q: expected reset code in %q", tc.severity, got)
		}
	}
}

func TestSupportsColorRespectsNoColorAndBuffer(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if supportsColor(&bytes.Buffer{}) {
		t.Fatal("NO_COLOR should always disable color")
	}
}

func TestSupportsColorBufferIsNotTTY(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("CLICOLOR_FORCE", "")
	if supportsColor(&bytes.Buffer{}) {
		t.Fatal("bytes.Buffer is not a TTY; color should be disabled")
	}
}

func TestSupportsColorForceColorWins(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "1")
	if !supportsColor(&bytes.Buffer{}) {
		t.Fatal("FORCE_COLOR=1 should enable color even on non-TTY")
	}
}

func TestStyledCellPlainPadsToWidth(t *testing.T) {
	got := styledCell("path", 8, ansiDarkGrey, false)
	if got != "path    " {
		t.Fatalf("plain styledCell = %q, want padded to 8", got)
	}
}

func TestStyledCellColorWrapsTextNotPad(t *testing.T) {
	got := styledCell("path", 8, ansiDarkGrey, true)
	want := ansiDarkGrey + "path" + ansiReset + "    "
	if got != want {
		t.Fatalf("colored styledCell = %q, want %q", got, want)
	}
}

func TestStyledCellOverflowKeepsNoPad(t *testing.T) {
	got := styledCell("longer-than-width", 4, ansiDarkGrey, true)
	want := ansiDarkGrey + "longer-than-width" + ansiReset
	if got != want {
		t.Fatalf("overflow styledCell = %q, want %q", got, want)
	}
}

func TestColorizePassthroughWhenDisabled(t *testing.T) {
	if got := colorize("x", ansiLightGrey, false); got != "x" {
		t.Fatalf("disabled colorize should pass through, got %q", got)
	}
}

func TestColorizeWrapsWhenEnabled(t *testing.T) {
	got := colorize("x", ansiLightGrey, true)
	want := ansiLightGrey + "x" + ansiReset
	if got != want {
		t.Fatalf("colorize = %q, want %q", got, want)
	}
}
