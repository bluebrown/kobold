package confix

import (
	"reflect"
	"testing"

	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/config/confix/old"
)

func TestMakeConfig(t *testing.T) {
	t.Parallel()
	v1, err := old.ReadPath("testdata/give.yaml")
	if err != nil {
		t.Fatal(err)
	}
	v1.Defaults()

	give, err := MakeConfig(v1)
	if err != nil {
		t.Fatal(err)
	}

	var want config.Config
	if err := config.ReadFile("testdata/want.toml", &want); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(give, &want) {
		t.Fatalf("want=%#v\n\ngive=%#v", &want, give)
	}
}
