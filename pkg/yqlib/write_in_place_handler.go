package yqlib

import (
	"os"
)

type writeInPlaceHandler interface {
	CreateTempFile() (*os.File, error)
	FinishWriteInPlace(evaluatedSuccessfully bool) error
}

type writeInPlaceHandlerImpl struct {
	inputFilename string
	tempFile      *os.File
	noBackup      bool
	backupFile    string
}

func NewWriteInPlaceHandler(inputFile string, noBackup bool) writeInPlaceHandler {

	return &writeInPlaceHandlerImpl{inputFile, nil, noBackup, ""}
}

func (w *writeInPlaceHandlerImpl) CreateTempFile() (*os.File, error) {
	file, err := createTempFile()

	if err != nil {
		return nil, err
	}
	info, err := os.Stat(w.inputFilename)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(file.Name(), info.Mode())

	if err != nil {
		return nil, err
	}

	if err = changeOwner(info, file); err != nil {
		return nil, err
	}
	log.Debugf("WriteInPlaceHandler: writing to tempfile: %v", file.Name())
	w.tempFile = file
	return file, err
}

func (w *writeInPlaceHandlerImpl) createBackup() error {
	if w.noBackup {
		return nil
	}
	backupPath := w.inputFilename + ".bak"
	data, err := os.ReadFile(w.inputFilename)
	if err != nil {
		return err
	}
	err = os.WriteFile(backupPath, data, 0600)
	if err != nil {
		return err
	}
	w.backupFile = backupPath
	log.Debugf("WriteInPlaceHandler: created backup at %v", backupPath)
	return nil
}

func (w *writeInPlaceHandlerImpl) FinishWriteInPlace(evaluatedSuccessfully bool) error {
	log.Debugf("Going to write in place, evaluatedSuccessfully=%v, target=%v", evaluatedSuccessfully, w.inputFilename)
	safelyCloseFile(w.tempFile)
	if evaluatedSuccessfully {
		if err := w.createBackup(); err != nil {
			log.Debugf("WriteInPlaceHandler: warning - could not create backup: %v", err)
		}
		log.Debug("Moving temp file to target")
		err := tryRenameFile(w.tempFile.Name(), w.inputFilename)
		if err != nil && w.backupFile != "" {
			log.Debugf("WriteInPlaceHandler: rename failed, attempting to restore backup from %v", w.backupFile)
			restoreErr := w.restoreBackup()
			if restoreErr != nil {
				log.Debugf("WriteInPlaceHandler: failed to restore backup: %v", restoreErr)
			}
			return err
		}
		if err == nil && w.backupFile != "" {
			tryRemoveTempFile(w.backupFile)
			w.backupFile = ""
		}
		return err
	}
	tryRemoveTempFile(w.tempFile.Name())

	return nil
}

func (w *writeInPlaceHandlerImpl) restoreBackup() error {
	if w.backupFile == "" {
		return nil
	}
	data, err := os.ReadFile(w.backupFile)
	if err != nil {
		return err
	}
	return os.WriteFile(w.inputFilename, data, 0600)
}
