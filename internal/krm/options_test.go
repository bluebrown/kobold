package krm

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		GiveExpr  string
		WantType  string
		WantTag   string
		WantError bool
	}{
		{
			name:     "Latest Exact",
			GiveExpr: "tag: latest; type: exact",
			WantType: TypeExact,
			WantTag:  "latest",
		},
		{
			name:     "Short Semver",
			GiveExpr: "tag: ^1; type: semver",
			WantType: TypeSemver,
			WantTag:  "^1",
		},
		{
			name:     "Comma Semver",
			GiveExpr: "tag: >= 2.3, < 3; type: semver",
			WantType: TypeSemver,
			WantTag:  ">= 2.3, < 3",
		},
		{
			name:     "Regex",
			GiveExpr: "tag: foo-.*; type: regex",
			WantType: TypeRegex,
			WantTag:  "foo-.*",
		},
		{
			name:      "Unknown Key",
			GiveExpr:  "nope: true",
			WantError: true,
		},
		{
			name:      "Not Key Value Pair",
			GiveExpr:  "nope",
			WantError: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts, err := ParseOpts(tt.GiveExpr)

			if tt.WantError {
				if err == nil {
					t.Errorf("want error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("%v", err)
				return
			}

			if opts.Type != tt.WantType {
				t.Errorf("type %q does not match expected type %q", opts.Type, tt.WantType)
			}

			if opts.Tag != tt.WantTag {
				t.Errorf("tag %q does not match expected tag %q", opts.Tag, tt.WantTag)
			}
		})
	}
}
