package krm

import (
	"context"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"

	"github.com/bluebrown/kobold/internal/events"
	"github.com/bluebrown/kobold/kobold/config"
	"github.com/google/go-containerregistry/pkg/name"
)

type testPipeOptions struct {
	associations []config.FileTypeSpec
	resolvers    []config.ResolverSpec
}

func testPipe(caseDir string, opts testPipeOptions, events ...events.PushData) (filesys.FileSystem, error) {
	outFs := filesys.MakeFsInMemory()
	w := kio.LocalPackageWriter{
		PackagePath: "/",
		FileSystem: filesys.FileSystemOrOnDisk{
			FileSystem: outFs,
		},
	}

	rend := NewRenderer(
		WithWriter(w),
		WithSelector(NewSelector(opts.resolvers, opts.associations)),
	)

	if _, err := rend.Render(context.Background(), "testdata/"+caseDir, events); err != nil {
		return nil, err
	}

	return outFs, nil
}

func Test_renderer_Render(t *testing.T) {
	type wantFieldValue struct {
		rnodeIndex int
		field      string
		value      string
	}
	tests := []struct {
		name                 string
		giveDir              string
		giveOpts             testPipeOptions
		giveEvents           []events.PushData
		wantSourceFieldValue map[string][]wantFieldValue
	}{
		{
			name:    "kube simple",
			giveDir: "kube",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3"},
				{Image: "test.azurecr.io/nginx", Tag: "v1", Digest: "sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"deployment.yaml": {
					{
						rnodeIndex: 0,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
					},
					{
						rnodeIndex: 1,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:v1@sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73",
					},
				},
			},
		},
		{
			name:    "kube nested",
			giveDir: "nested",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3"},
				{Image: "test.azurecr.io/nginx", Tag: "v1", Digest: "sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"child/deployment.yaml": {
					{
						rnodeIndex: 0,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
					},
					{
						rnodeIndex: 1,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:v1@sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73",
					},
				},
			},
		},
		{
			name:    "kube kpt",
			giveDir: "kpt",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3"},
				{Image: "test.azurecr.io/nginx", Tag: "v1", Digest: "sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"deployment.yaml": {
					{
						rnodeIndex: 0,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:latest@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3",
					},
					{
						rnodeIndex: 1,
						field:      "spec.template.spec.containers.0.image",
						value:      "test.azurecr.io/nginx:v1@sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73",
					},
				},
			},
		},
		{
			name:    "helm krm ignore",
			giveDir: "krm-ignore",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{},
		},
		// FIXME: if there is a yaml parse error, the whole render process will fail
		// needs to use .krm ignore to ignore invalid yaml portions
		// {
		// 	name:    "helm skip errors",
		// 	giveDir: "helm-skip-errors",
		// 	giveEvents: []events.PushData{
		// 		{Image: "index.docker.io/bluebrown/busybox", Tag: "latest", Digest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869"},
		// 	},
		// 	wantSourceFieldValue: map[string][]wantFieldValue{
		// 		"pod.yaml": {
		// 			{
		// 				rnodeIndex: 0,
		// 				field:      "spec.template.image",
		// 				value:      "index.docker.io/bluebrown/busybox@sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
		// 			},
		// 		},
		// 	},
		// },
		{
			name:    "cr",
			giveDir: "custom-resource",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{},
		},
		{
			name:    "compose",
			giveDir: "compose",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
				{Image: "index.docker.io/bluebrown/busybox", Tag: "latest", Digest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"compose.yaml": {
					{
						rnodeIndex: 0,
						field:      "services.foo.image",
						value:      "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
					},
					{
						rnodeIndex: 0,
						field:      "services.bar.image",
						value:      "index.docker.io/bluebrown/busybox:latest@sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
					},
				},
			},
		},
		{
			name:    "ko",
			giveDir: "ko",
			giveEvents: []events.PushData{
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
				{Image: "test.azurecr.io/stuff", Tag: "v1", Digest: "sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				".ko.yaml": {
					{
						rnodeIndex: 0,
						field:      "defaultBaseImage",
						value:      "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
					},
					{
						rnodeIndex: 0,
						field:      "baseImageOverrides.github\\.com/my-user/my-repo/cmd/app",
						value:      "test.azurecr.io/stuff:v1@sha256:993518ca49ede3c4e751fe799837ede16e60bc410452e3922602ebceda9b4c73",
					},
					{
						rnodeIndex: 0,
						field:      "baseImageOverrides.github\\.com/my-user/my-repo/cmd/foo",
						value:      "test.azurecr.io/ubuntu",
					},
				},
			},
		},
		{
			name:    "default registry",
			giveDir: "default-registry",
			giveEvents: []events.PushData{
				{Image: "index.docker.io/bluebrown/busybox", Tag: "latest", Digest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"compose.yaml": {
					{
						rnodeIndex: 0,
						field:      "services.foo.image",
						value:      "index.docker.io/bluebrown/busybox:latest@sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
					},
				},
			},
		},
		{
			name:    "custom-resolver-helm",
			giveDir: "custom-resolver-helm",
			giveOpts: testPipeOptions{
				resolvers: []config.ResolverSpec{
					{Name: "my-helm", Paths: []string{"path.to.image", "another.path"}},
				},
				associations: []config.FileTypeSpec{{Kind: "my-helm", Pattern: "values.yaml"}},
			},
			giveEvents: []events.PushData{
				{Image: "index.docker.io/bluebrown/echoserver", Tag: "latest", Digest: "sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869"},
				{Image: "test.azurecr.io/nginx", Tag: "latest", Digest: "sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248"},
			},
			wantSourceFieldValue: map[string][]wantFieldValue{
				"values.yaml": {
					{
						rnodeIndex: 0,
						field:      "path.to.image",
						value:      "index.docker.io/bluebrown/echoserver:latest@sha256:3b3128d9df6bbbcc92e2358e596c9fbd722a437a62bafbc51607970e9e3b8869",
					},
					{
						rnodeIndex: 0,
						field:      "another.path",
						value:      "test.azurecr.io/nginx:latest@sha256:220611111e8c9bbe242e9dc1367c0fa89eef83f26203ee3f7c3764046e02b248",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs, err := testPipe(tt.giveDir, tt.giveOpts, tt.giveEvents...)
			if err != nil {
				t.Fatal(err)
			}
			for source, fieldValues := range tt.wantSourceFieldValue {
				if !fs.Exists(source) {
					t.Errorf("%s: file does not exist in outFs", source)
					continue
				}
				f, err := fs.Open(source)
				if err != nil {
					t.Errorf("%s: failed to open file %v", source, err)
					continue
				}
				defer f.Close()
				rnodes, err := (&kio.ByteReader{Reader: f}).Read()
				if err != nil {
					t.Errorf("%s: failed to parse buffer into rnode: %v", source, err)
					continue
				}
				for _, fieldValue := range fieldValues {
					if len(rnodes) < fieldValue.rnodeIndex+1 {
						t.Errorf("%s: %d: rnodes len is less than desired index %d", source, fieldValue.rnodeIndex, fieldValue.rnodeIndex)
						continue
					}
					imgValI, err := rnodes[fieldValue.rnodeIndex].GetFieldValue(fieldValue.field)
					if err != nil {
						t.Errorf("%s: %d: %s: could not get image node", source, fieldValue.rnodeIndex, fieldValue.field)
						continue
					}
					imgRefRaw, ok := imgValI.(string)
					if !ok {
						t.Errorf("%s: %d: %s: could not convert image node to map node", source, fieldValue.rnodeIndex, fieldValue.field)
						continue
					}
					if imgRefRaw != fieldValue.value {
						t.Errorf("%s: %d: %s: image ref does not match value:\ngot:  %q\nwant: %q\n", source, fieldValue.rnodeIndex, fieldValue.field, imgRefRaw, fieldValue.value)
						continue
					}
					_, err = name.ParseReference(imgRefRaw)
					if err != nil {
						t.Errorf("%s: %d: image not does not contain a valid image ref %v", source, fieldValue.rnodeIndex, err)
						continue
					}
				}
			}
		})
	}
}
