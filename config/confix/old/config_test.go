package old

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestConfig(t *testing.T) {
	os.Setenv("ACR_TOKEN", "test-header")
	os.Setenv("GIT_USR", "test-usr")
	os.Setenv("GIT_PAT", "test-pwd")
	os.Setenv("NAMESPACE", "kobold")

	conf, err := ReadPath("testdata/config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	b1, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}

	b2, err := os.ReadFile("testdata/expected.yaml")
	if err != nil {
		t.Fatal(err)
	}

	s1 := string(b1)
	s2 := string(b2)
	if s1 != s2 {
		t.Errorf("unexpected output\n\nExpected:\n---\n---\n%s\n\nGot:\n---\n---\n%s\n\n", s2, s1)
	}
}
