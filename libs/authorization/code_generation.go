package authorization

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/teamyapp/cloud/libs/errs"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

//go:embed *.gotmpl
var fs embed.FS

var funcMap = template.FuncMap{
	"ToUpper":       strings.ToUpper,
	"ToLower":       strings.ToLower,
	"PascalToCamel": PascalToCamel,
}

func PascalToCamel(pascalCase string) string {
	if len(pascalCase) == 0 {
		return ""
	}

	firstLetter := strings.ToLower(string(pascalCase[0]))
	restLetters := pascalCase[1:]
	return firstLetter + restLetters
}

func GenerateCode(config *Config, outputDir string) *errs.Error {
	templateFiles, _ := fs.ReadDir(".")
	for _, templateFile := range templateFiles {
		templateFileName := templateFile.Name()
		fmt.Println("Generating code for template: ", templateFileName)
		tmplString, err := fs.Open(templateFile.Name())
		if err != nil {
			return errs.NewError(errs.IO, err.Error())
		}

		tmplContent, err := ioutil.ReadAll(tmplString)
		if err != nil {
			return errs.NewError(errs.IO, err.Error())
		}

		outputFileName := strings.Split(templateFileName, ".")[0] + ".go"
		outputFilePath := outputDir + "/" + outputFileName
		generateCodeErr := generateCodeForSingleTmpl(config, string(tmplContent), outputFilePath)
		if generateCodeErr != nil {
			return errs.NewError(errs.Unknown, err.Error())
		}

		fmt.Println("Generated code for template", templateFileName)
	}

	return nil
}

func generateCodeForSingleTmpl(config *Config, tmplContent string, outputFilePath string) *errs.Error {
	var outputBuffer strings.Builder
	tmpl, err := template.New("test").Funcs(funcMap).Parse(tmplContent)

	err = tmpl.Execute(&outputBuffer, config)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	formattedCode, err := format.Source([]byte(outputBuffer.String()))
	if err != nil {
		fmt.Println("Error formatting code: ", err)
		return errs.NewError(errs.Unknown, err.Error())
	}

	internalErr := writeOrOverwriteToDir(string(formattedCode), outputFilePath)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func writeOrOverwriteToDir(content string, outputFilePath string) *errs.Error {
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
