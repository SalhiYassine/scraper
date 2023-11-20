package cleaner

import (
	"strings"
	"testing"
)

type htmlTest struct {
	name     string
	input    string
	expected string
}

var htmlTests = []htmlTest{
	{
		name:     "Remove Script Tags",
		input:    `<html><body><script>alert("test");</script><p>Content</p></body></html>`,
		expected: `<html><body><p>Content</p></body></html>`,
	},
	{
		name:     "Remove Style Tags",
		input:    `<html><head><style>body {background-color: #fff;}</style></head><body><p>Content</p></body></html>`,
		expected: `<html><head></head><body><p>Content</p></body></html>`,
	},
	{
		name:     "Remove Comments",
		input:    `<!-- Comment --><div>Text</div>`,
		expected: `<div>Text</div>`,
	},
	{
		name:     "Remove Empty Tags",
		input:    `<div><span></span>Content</div>`,
		expected: `<div>Content</div>`,
	},
	{
		name:     "Remove Attributes",
		input:    `<div class="class" id="id">Content</div>`,
		expected: `<div>Content</div>`,
	},
	{
		name:     "Nested Structure",
		input:    `<div><div><p>Text</p></div></div>`,
		expected: `<div><p>Text</p></div>`,
	},
	{
		name:     "Multiple Script Tags",
		input:    `<script>Script1</script><p>Content</p><script>Script2</script>`,
		expected: `<p>Content</p>`,
	},
	{
		name:     "Script Tags with Attributes",
		input:    `<script type="text/javascript">alert("test");</script>`,
		expected: ``,
	},
	{
		name:     "Mixed Content",
		input:    `<div><!-- Comment --><p style="color:red;">Text</p><script>Script</script></div>`,
		expected: `<div><p>Text</p></div>`,
	},
}

func TestCleanHTML(t *testing.T) {
	for _, tc := range htmlTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CleanHTML(tc.input)
			if err != nil {
				t.Fatalf("Test %v failed with error: %v", tc.name, err)
			}
			if strings.TrimSpace(result) != strings.TrimSpace(tc.expected) {
				t.Errorf("Test %v failed. Expected: %v, got: %v", tc.name, tc.expected, result)
			}
		})
	}
}
