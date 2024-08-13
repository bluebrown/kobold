package git

import "testing"

func TestPackageURI_UnmarshalText(t *testing.T) {
	type fields struct {
		Repo string
		Ref  string
		Pkg  string
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Valid url with package",
			fields: fields{
				Repo: "git@github.com:some-org/some-repo",
				Ref:  "main",
				Pkg:  "some-pkg",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git?ref=main&pkg=some-pkg"),
			},
			wantErr: false,
		},
		{
			name: "Valid url with package specified first",
			fields: fields{
				Repo: "git@github.com:some-org/some-repo",
				Ref:  "main",
				Pkg:  "some-pkg",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git?pkg=some-pkg&ref=main"),
			},
			wantErr: false,
		},
		{
			name: "Valid url only uses first declaration of query params",
			fields: fields{
				Repo: "git@github.com:some-org/some-repo",
				Ref:  "main",
				Pkg:  "some-pkg",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git?ref=main&pkg=some-pkg&ref=fail&pkg=fail"),
			},
			wantErr: false,
		},
		{
			name: "Valid url without package",
			fields: fields{
				Repo: "git@github.com:some-org/some-repo",
				Ref:  "main",
				Pkg:  "",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git?ref=main"),
			},
			wantErr: false,
		},
		{
			name: "Missing query params",
			fields: fields{
				Repo: "",
				Ref:  "",
				Pkg:  "",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git"),
			},
			wantErr: true,
		},
		{
			name: "Missing ref",
			fields: fields{
				Repo: "",
				Ref:  "",
				Pkg:  "",
			},
			args: args{
				b: []byte("git@github.com:some-org/some-repo.git?pkg=some-pkg"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := &PackageURI{
				Repo: tt.fields.Repo,
				Ref:  tt.fields.Ref,
				Pkg:  tt.fields.Pkg,
			}
			if err := uri.UnmarshalText(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
