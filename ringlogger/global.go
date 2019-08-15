/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ringlogger

import (
	"log"
	"path/filepath"
	"os"
	"golang.zx2c4.com/wireguard/windows/conf"
)

var Global *Ringlogger

func InitGlobalLogger(tag string) error {
	if Global != nil {
		return nil
	}
	root, err := conf.RootDirectory()
	if err != nil {
		return err
	}

	mapped_log_filename := "log.bin"

	env_log_filename := os.Getenv("WG_LOG_NAME")
	if(env_log_filename != "") {
		mapped_log_filename = env_log_filename
	}

	Global, err = NewRinglogger(filepath.Join(root, mapped_log_filename), tag)
	if err != nil {
		return err
	}
	log.SetOutput(Global)
	log.SetFlags(0)
	return nil
}
