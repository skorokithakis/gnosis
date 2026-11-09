package index

import "testing"

func TestSegmentCJK(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"multi-word run gains spaces", "状态流转", "状态 流转"},
		{"single-word run unchanged", "失败", "失败"},
		{"latin text untouched", "doing ready tab", "doing ready tab"},
		{"mixed latin and CJK splits both", "doing状态流转", "doing 状态 流转"},
		{"no CJK at all", "just ascii!", "just ascii!"},
		{"punctuation bounds the run", "长工失败回报撞：worker", "长工 失败 回报 撞：worker"},
		{"CJK glued to a digit splits", "重试5次", "重试 5 次"},
	}
	for _, c := range cases {
		if got := SegmentCJK(c.in); got != c.want {
			t.Errorf("SegmentCJK(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSegmentCJKRestoreRoundTrip(t *testing.T) {
	// RestoreSnippetSpaces must undo exactly the spacing SegmentCJK injects,
	// so segmented text round-trips to the original entry text.
	for _, in := range []string{
		"长工失败回报撞非法状态流转",
		"macbridge 解析缺陷：提醒解析",
		"audit finding: task-service.update_result:doing-ready-illegal-transition，尚未修复",
	} {
		if got := RestoreSnippetSpaces(SegmentCJK(in)); got != in {
			t.Errorf("round trip of %q = %q, want the original", in, got)
		}
	}
}

func TestRestoreSnippetSpaces(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"collapses segmentation spacing", "状态 流转", "状态流转"},
		{"keeps latin spacing", "doing ready", "doing ready"},
		{"mixed run", "doing 状态 流转 end", "doing 状态流转 end"},
	}
	for _, c := range cases {
		if got := RestoreSnippetSpaces(c.in); got != c.want {
			t.Errorf("RestoreSnippetSpaces(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// A space the author deliberately wrote between two CJK characters is
// indistinguishable from injected segmentation spacing and is collapsed in
// snippets. This is a display-only trade-off; stored entry text is untouched.
func TestRestoreSnippetSpaces_collapsesAuthoredCJKSpaces(t *testing.T) {
	if got := RestoreSnippetSpaces("你好 世界"); got != "你好世界" {
		t.Fatalf("got %q, want %q", got, "你好世界")
	}
}

func TestHasCJK(t *testing.T) {
	if !HasCJK("状态流转") || !HasCJK("mixed 中文 text") {
		t.Error("HasCJK should detect CJK characters")
	}
	if HasCJK("plain ascii") || HasCJK("") {
		t.Error("HasCJK should not flag latin-only text")
	}
}
