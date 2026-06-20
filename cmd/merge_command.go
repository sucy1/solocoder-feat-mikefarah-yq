package cmd

import (
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/cobra"
)

func evaluateMerge(cmd *cobra.Command, args []string, expression string) error {
	out := cmd.OutOrStdout()

	var err error

	if writeInplace {
		colorsEnabled = forceColor
		writeInPlaceHandler := yqlib.NewWriteInPlaceHandler(args[0], noBackup)
		out, err = writeInPlaceHandler.CreateTempFile()
		if err != nil {
			return err
		}
		defer func() {
			if err == nil {
				err = writeInPlaceHandler.FinishWriteInPlace(true)
			}
		}()
	}

	format, err := yqlib.FormatFromString(outputFormat)
	if err != nil {
		return err
	}

	printerWriter, err := configurePrinterWriter(format, out)
	if err != nil {
		return err
	}

	encoder, err := configureEncoder()
	if err != nil {
		return err
	}

	printer := yqlib.NewPrinter(encoder, printerWriter)

	decoder, err := configureDecoder(true)
	if err != nil {
		return err
	}

	mergeExpression := buildMergeExpression(expression)
	yqlib.GetLogger().Debugf("merge expression: %v", mergeExpression)

	allAtOnceEvaluator := yqlib.NewAllAtOnceEvaluator()
	err = allAtOnceEvaluator.EvaluateFiles(mergeExpression, args, printer, decoder)

	if err == nil && exitStatus && !printer.PrintedAnything() {
		return fmt.Errorf("no matches found")
	}

	return err
}

func buildMergeExpression(expression string) string {
	var mergeExp string
	if mergeStrategy == "append" {
		mergeExp = ". as $item ireduce ([]; . + $item)"
	} else {
		mergeExp = ". as $item ireduce ({}; . * $item)"
	}

	if expression != "" {
		return fmt.Sprintf("%v | %v", mergeExp, expression)
	}
	return mergeExp
}
