package gomodule

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

func Test_TestBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			go_testedbinary {
			  name: "smthTestResult.exe",
			  srcs: ["smth.go", "smth_test.go"],
			  pkg: ".",
				testPkg: ".",
				vendorFirst: true
			}
		`),
		"smth.go":      nil,
		"smth_test.go": nil,
	})

	ctx.RegisterModuleType("go_testedbinary", TestedBinFactory)

	config := bood.NewConfig()

	_, err := ctx.ParseBlueprintsFiles(".", config)
	if len(err) != 0 {
		t.Fatalf("ParseBlueprintsFiles error : %s", err)
	}

	_, err = ctx.PrepareBuildActions(config)
	if len(err) != 0 {
		t.Errorf("PrepareBuildActions error : %s", err)
	}

	ninjaContentBuffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(ninjaContentBuffer); err != nil {
		t.Errorf("WriteBuildFile error : %s", err)
	} else {
		ninjaContent := ninjaContentBuffer.String()
		if !strings.Contains(ninjaContent, "out/bin/smthTestResult.exe:") {
			t.Errorf("out/bin/smthTestResult.exe does not exist")
		}
		if !strings.Contains(ninjaContent, "smth.go") {
			t.Errorf("smth.go does not exist")
		}
		if !strings.Contains(ninjaContent, "build vendor: g.gomodule.vendor | go.mod") {
			t.Errorf("build vendor: g.gomodule.vendor | go.mod does not exist")
		}
	}
}
