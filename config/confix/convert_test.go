package confix

import (
	"reflect"
	"testing"

	"github.com/bluebrown/kobold/config"
	"github.com/bluebrown/kobold/config/confix/old"
)

func TestMakeConfig(t *testing.T) {
	t.Parallel()

	oc := new(old.Config)
	if err := old.ReadFile("testdata/give.toml", oc); err != nil {
		t.Fatal(err)
	}

	give, err := MakeConfig(oc)
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
