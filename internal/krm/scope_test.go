package krm

import "testing"

type wants struct {
	relPath string
	ignore  bool
}

func TestScopes(t *testing.T) {
	tests := []struct {
		name       string
		giveScopes []string
		wants      []wants
	}{
		{
			name:       "simple exact",
			giveScopes: []string{"stage/deployment.yaml"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "top level dir",
			giveScopes: []string{"stage/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "nested level dir",
			giveScopes: []string{"stage/"},
			wants: []wants{
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "from top level only",
			giveScopes: []string{"/stage/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "mutli scope",
			giveScopes: []string{"stage/", "dev/"},
			wants: []wants{
				{
					relPath: "stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/stage/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "dev/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "my-app/dev/deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "prod/deployment.yaml",
					ignore:  true,
				},
				{
					relPath: "my-app/prod/deployment.yaml",
					ignore:  true,
				},
			},
		},
		{
			name:       "file negation",
			giveScopes: []string{"*.yaml", "!compose.yaml"},
			wants: []wants{
				{
					relPath: "deployment.yaml",
					ignore:  false,
				},
				{
					relPath: "compose.yaml",
					ignore:  true,
				},
			},
		},
		// {
		// 	name:       "folder negation",
		// 	giveScopes: []string{"!/prod/"},
		// 	wants: []wants{
		// 		{
		// 			relPath: "prod/deployment.yaml",
		// 			ignore:  true,
		// 		},
		// 		{
		// 			relPath: "dev/deployment.yaml",
		// 			ignore:  false,
		// 		},
		// 		{
		// 			relPath: "stage/deployment.yaml",
		// 			ignore:  false,
		// 		},
		// 	},
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fn := ignore(tt.giveScopes)
			for _, w := range tt.wants {
				m := fn(w.relPath)
				if m != w.ignore {
					t.Errorf("%q did not have correct ignore. Want %v but got %v", w.relPath, w.ignore, m)
				}
			}
		})
	}
}
