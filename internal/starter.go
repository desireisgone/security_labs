package internal

import (
	"fmt"
	"os"

	"selfcheck/internal/patcher"
	"selfcheck/internal/scanner"
)

var Marker = [18]byte{0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0xAA, 0xBB, 0, 0}

func StartSelfCheck() {
	if len(os.Args) == 3 && os.Args[1] == "--update-mode" {
		originalPath := os.Args[2]
		patcher.ExecuteUpdate(originalPath)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Unable determine path: %v\n", err)
		os.Exit(1)
	}

	hashSumBuffer := Marker[16:]
	hashSumValue := uint16(hashSumBuffer[0])<<8 | uint16(hashSumBuffer[1])

	scanner := scanner.NewScanner(Marker[:16], 512)
	res, err := scanner.ScanFile(exePath)
	if err != nil {
		fmt.Printf("Hash calculation error: %v\n", err)
		os.Exit(1)
	}

	if hashSumValue == 0 {
		fmt.Println("Hash initialization")
		fmt.Printf("Hash sum: 0x%04X\n", res.CalculatedHash)

		err = patcher.SelfPatchAndRestart(exePath, res.CalculatedHash, res.MarkerOffset)
		if err != nil {
			fmt.Printf("Patch file error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Hash check")
		fmt.Printf("Current hash sum: 0x%04X\n", hashSumValue)

		if res.CalculatedHash == hashSumValue {
			fmt.Println("Hash check success")
		} else {
			fmt.Println("Hashes do not match")
			os.Exit(2)
		}
	}
}
