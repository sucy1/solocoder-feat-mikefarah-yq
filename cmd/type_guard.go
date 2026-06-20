package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/cobra"
)

var typeGuardExitCode = 3

var yamlTagToTypeGuard = map[string]string{
	"!!str":   "string",
	"!!int":   "int",
	"!!float": "float",
	"!!bool":  "bool",
	"!!seq":   "array",
	"!!map":   "object",
	"!!null":  "null",
}

var typeGuardToYamlTag = map[string]string{
	"string": "!!str",
	"int":    "!!int",
	"float":  "!!float",
	"bool":   "!!bool",
	"array":  "!!seq",
	"object": "!!map",
	"null":   "!!null",
}

func evaluateTypeGuard(cmd *cobra.Command, args []string, expression string) error {
	path, expectedType, err := parseTypeGuard(typeGuard)
	if err != nil {
		return err
	}

	_, typeExists := typeGuardToYamlTag[expectedType]
	if !typeExists {
		return fmt.Errorf("unsupported type '%v' in --type-guard. Supported types: string, int, float, bool, array, object, null", expectedType)
	}

	decoder, err := configureDecoder(false)
	if err != nil {
		return err
	}

	encoder, err := configureEncoder()
	if err != nil {
		return err
	}

	tagExp := fmt.Sprintf("%v | tag", path)
	if expression != "" {
		tagExp = fmt.Sprintf("%v | %v | tag", expression, path)
	}

	stringEvaluator := yqlib.NewStringEvaluator()

	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		pipingStdin := stat != nil && (stat.Mode()&os.ModeCharDevice) == 0
		if nullInput || !pipingStdin {
			cmd.Println(cmd.UsageString())
			return nil
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(os.Stdin)
		if err != nil {
			return fmt.Errorf("--type-guard: error reading stdin: %w", err)
		}
		tagResult, err := stringEvaluator.Evaluate(tagExp, buf.String(), encoder, decoder)
		if err != nil {
			return fmt.Errorf("--type-guard: error evaluating path: %w", err)
		}
		return checkTypeGuardResult(path, expectedType, tagResult)
	}

	for _, filename := range args {
		if filename == "-" {
			var buf bytes.Buffer
			_, err = buf.ReadFrom(os.Stdin)
			if err != nil {
				return fmt.Errorf("--type-guard: error reading stdin: %w", err)
			}
			tagResult, err := stringEvaluator.Evaluate(tagExp, buf.String(), encoder, decoder)
			if err != nil {
				return fmt.Errorf("--type-guard: error evaluating path from stdin: %w", err)
			}
			if err := checkTypeGuardResult(path, expectedType, tagResult); err != nil {
				return err
			}
			continue
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("--type-guard: error reading file '%v': %w", filename, err)
		}

		tagResult, err := stringEvaluator.Evaluate(tagExp, string(data), encoder, decoder)
		if err != nil {
			return fmt.Errorf("--type-guard: error evaluating path in '%v': %w", filename, err)
		}

		if err := checkTypeGuardResult(path, expectedType, tagResult); err != nil {
			return err
		}
	}

	return nil
}

func checkTypeGuardResult(path, expectedType, tagResult string) error {
	tagLines := strings.Split(strings.TrimSpace(tagResult), "\n")
	for _, line := range tagLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		actualType, typeExists := yamlTagToTypeGuard[line]
		if !typeExists {
			return fmt.Errorf("--type-guard: unknown tag '%v'", line)
		}
		if actualType != expectedType {
			return &typeGuardError{
				Path:         path,
				ExpectedType: expectedType,
				ActualType:   actualType,
			}
		}
	}
	return nil
}

func parseTypeGuard(guard string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(guard), " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid --type-guard format '%v', expected '<path> <type>' (e.g. '.name string')", guard)
	}
	path := parts[0]
	expectedType := strings.ToLower(strings.TrimSpace(parts[1]))
	return path, expectedType, nil
}

type typeGuardError struct {
	Path         string
	ExpectedType string
	ActualType   string
}

func (e *typeGuardError) Error() string {
	return fmt.Sprintf("type guard failed: path '%v' expected type '%v' but got '%v'", e.Path, e.ExpectedType, e.ActualType)
}

func (e *typeGuardError) ExitCode() int {
	return typeGuardExitCode
}

func isTypeGuardError(err error) bool {
	var tgErr *typeGuardError
	return errors.As(err, &tgErr)
}
