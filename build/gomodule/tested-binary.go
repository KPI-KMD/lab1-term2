package gomodule

import (
	"fmt"
	"path"
	"regexp"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

var (
	pctx = blueprint.NewPackageContext("github.com/KPI-KMD/lab1-term2/build/gomodule")

	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cmd /c cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cmd /c cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")

	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command:     "cmd /c cd $workDir && go test -v $testPkg > $testLogPath",
		Description: "test go pkg $testPkg",
	}, "workDir", "testLogPath", "testPkg")
)

type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Pkg         string
		Srcs        []string
		SrcsExclude []string
		TestPkg     string
		VendorFirst bool
	}
}

func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	// TODO: Імплементууйте генерацію правил збірки для ninja-файла.
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	testLogPath := path.Join(config.BaseOutputDir, "testLog.txt")

	var inputs []string
	var testInputs []string
	inputErors := false
	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.SrcsExclude); err == nil {
			for _, pathName := range matches {
				if isTest, _ := regexp.Match("^.*_test.go$", []byte(pathName)); isTest {
					testInputs = append(testInputs, pathName)
				} else {
					inputs = append(inputs, pathName)
				}
			}
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	if len(tb.properties.TestPkg) > 0 {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Test module %s", tb.properties.TestPkg),
			Rule:        goTest,
			Outputs:     []string{testLogPath},
			Implicits:   append(testInputs, inputs...),
			Args: map[string]string{
				"testLogPath": testLogPath,
				"workDir":     ctx.ModuleDir(),
				"testPkg":     tb.properties.TestPkg,
			},
		})
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   inputs,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.Pkg,
		},
	})
}

func TestedBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
