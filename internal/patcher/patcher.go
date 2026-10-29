package patcher

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func SelfPatchAndRestart(exePath string, hashSum uint16, markerOffset int64) error {
	input, err := os.ReadFile(exePath)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint16(input[markerOffset+16:markerOffset+18], hashSum)

	tempExe := filepath.Join(os.TempDir(), "integrity_patcher.exe")
	err = os.WriteFile(tempExe, input, os.ModePerm)
	if err != nil {
		return err
	}

	cmd := exec.Command(tempExe, "--update-mode", exePath)
	err = cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func ExecuteUpdate(originalExePath string) {
	time.Sleep(1 * time.Second)

	currentExe, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}

	patchedData, err := os.ReadFile(currentExe)
	if err != nil {
		os.Exit(1)
	}

	err = os.WriteFile(originalExePath, patchedData, os.ModePerm)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
