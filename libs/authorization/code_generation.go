package authorization

import (
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/teamyapp/cloud/libs/telemetry"

	"github.com/teamyapp/cloud/libs/errs"
)

//go:embed resource_type.gotmpl
var resourceTypeCodeTemplate string

//go:embed resource_operation.gotmpl
var resourceOperationCodeTemplate string

//go:embed resource_type_operation.gotmpl
var resourceTypeOperationCodeTemplate string

//go:embed query.gotmpl
var queryCodeTemplate string

type CodeTemplate struct {
	Name    string
	Content string
}

var codeTemplates []CodeTemplate = []CodeTemplate{
	{
		Name:    "resource_type",
		Content: resourceTypeCodeTemplate,
	},
	{
		Name:    "resource_operation",
		Content: resourceOperationCodeTemplate,
	},
	{
		Name:    "resource_type_operation",
		Content: resourceTypeOperationCodeTemplate,
	},
	{
		Name:    "query",
		Content: queryCodeTemplate,
	},
}

var templateOperations = template.FuncMap{
	"pascalToCamel": pascalToCamel,
}

func GenerateCode(config *Config, logger telemetry.Logger, outputDir string) *errs.Error {
	for _, codeTemplate := range codeTemplates {
		err := generateCodeForTemplate(config, logger, codeTemplate, outputDir)
		if err != nil {
			return errs.NewError(errs.Unknown, err.Message)
		}

		logger.Info(fmt.Sprintf("Generated code for template %v", codeTemplate.Name))
	}

	return nil
}

func generateCodeForTemplate(config *Config, logger telemetry.Logger, codeTemplate CodeTemplate, outputDir string) *errs.Error {
	tmplName := codeTemplate.Name
	tmplContent := codeTemplate.Content
	logger.Info(fmt.Sprintf("Generating code for template: %v", tmplName))
	outputFileName := fmt.Sprintf("%s.go", tmplName)
	outputFilePath := filepath.Join(outputDir, outputFileName)

	var outputBuffer strings.Builder
	tmpl, err := template.New("test").Funcs(templateOperations).Parse(tmplContent)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	err = tmpl.Execute(&outputBuffer, config)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	formattedCode, err := format.Source([]byte(outputBuffer.String()))
	if err != nil {
		logger.Info(fmt.Sprintf("Error formatting code: %v", err.Error()))
		return errs.NewError(errs.Unknown, err.Error())
	}

	internalErr := overwriteFileContent(string(formattedCode), outputFilePath)
	return internalErr
}

func overwriteFileContent(content string, outputFilePath string) *errs.Error {
	file, err := os.OpenFile(outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

func pascalToCamel(pascalCase string) string {
	if len(pascalCase) == 0 {
		return ""
	}

	firstLetter := strings.ToLower(string(pascalCase[0]))
	restLetters := pascalCase[1:]
	return firstLetter + restLetters
}
