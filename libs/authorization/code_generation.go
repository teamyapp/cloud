package authorization

import (
	_ "embed"
	"github.com/teamyapp/cloud/libs/errs"
	"os"
	"strings"
	"text/template"
)

// //go:embed resource_type.gotmpl
//go:embed resource_type_operation.gotmpl
var resourceTypeTmpl string

func GenerateCode(config *Config, outputDir string) *errs.Error {
	var outputBuffer strings.Builder
	tmpl, err := template.New("test").Parse(resourceTypeTmpl)

	err = tmpl.Execute(&outputBuffer, config)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	internalErr := writeOrOverwriteToDir(outputBuffer.String(), outputDir)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func writeOrOverwriteToDir(content string, outputDir string) *errs.Error {
	file, err := os.OpenFile(outputDir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errs.NewError(errs.IO, err.Error())
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return errs.NewError(errs.IO, err.Error())
	}

	return nil
}
