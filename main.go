/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/elevate"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/ringlogger"
	"golang.zx2c4.com/wireguard/windows/tunnel"
	"golang.zx2c4.com/wireguard/windows/ui"
)

var flags = [...]string{
	"(no argument): elevate and install manager service for current user",
	"/installmanagerservice",
	"/installtunnelservice CONFIG_PATH",
	"/uninstallmanagerservice",
	"/uninstalltunnelservice TUNNEL_NAME",
	"/managerservice",
	"/tunnelservice CONFIG_PATH",
	"/ui CMD_READ_HANDLE CMD_WRITE_HANDLE CMD_EVENT_HANDLE LOG_MAPPING_HANDLE",
	"/dumplog OUTPUT_PATH",
}

var cli = ""
var isCli = false

func fatal(v ...interface{}) {
	if isCli {
		fmt.Println(fmt.Sprint(v...))
	} else {
		windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprint(v...)), windows.StringToUTF16Ptr("Error"), windows.MB_ICONERROR)
	}

	os.Exit(1)
}

func fatalf(format string, v ...interface{}) {
	fatal(fmt.Sprintf(format, v...))
}

func info(title string, format string, v ...interface{}) {
	if isCli {
		fmt.Println(fmt.Sprintf(format, v...))
	} else {
		windows.MessageBox(0, windows.StringToUTF16Ptr(fmt.Sprintf(format, v...)), windows.StringToUTF16Ptr(title), windows.MB_ICONINFORMATION)
	}
}

func usage() {
	builder := strings.Builder{}
	for _, flag := range flags {
		builder.WriteString(fmt.Sprintf("    %s\n", flag))
	}
	info("Command Line Options", "Usage: %s [\n%s]", os.Args[0], builder.String())
	os.Exit(1)
}

func checkForWow64() {
	var b bool
	p, err := windows.GetCurrentProcess()
	if err != nil {
		fatal(err)
	}
	err = windows.IsWow64Process(p, &b)
	if err != nil {
		fatalf("Unable to determine whether the process is running under WOW64: %v", err)
	}
	if b {
		fatal("You must use the 64-bit version of WireGuard on this computer.")
	}
}

func checkForAdminGroup() {
	// This is not a security check, but rather a user-confusion one.
	processToken, err := elevate.OpenCurrentProcessToken() //TODO: Change to windows.OpenCurrentProcessToken once https://go-review.googlesource.com/c/sys/+/192337 lands
	if err != nil {
		fatalf("Unable to open current process token: %v", err)
	}
	defer processToken.Close()
	if !elevate.TokenIsElevatedOrElevatable(processToken) {
		fatalf("WireGuard may only be used by users who are a member of the Builtin %s group.", elevate.AdminGroupName())
	}
}

func checkForAdminDesktop() {
	adminDesktop, err := elevate.IsAdminDesktop()
	if !adminDesktop && err == nil {
		fatalf("WireGuard is running, but the UI is only accessible from desktops of the Builtin %s group.", elevate.AdminGroupName())
	}
}

func execElevatedManagerServiceInstaller() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	err = elevate.ShellExecute(path, "/installmanagerservice", "", windows.SW_SHOW)
	if err != nil {
		return err
	}
	os.Exit(0)
	return windows.ERROR_ACCESS_DENIED // Not reached
}

func pipeFromHandleArgument(handleStr string) (*os.File, error) {
	handleInt, err := strconv.ParseUint(handleStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handleInt), "pipe"), nil
}

func main() {
	isCli = cli == "true"
	checkForWow64()

	if len(os.Args) <= 1 {
		checkForAdminGroup()
		if !isCli && ui.RaiseUI() {
			return
		}
		err := execElevatedManagerServiceInstaller()
		if err != nil {
			fatal(err)
		}
		return
	}
	switch os.Args[1] {
	case "/installmanagerservice":
		if len(os.Args) != 2 {
			usage()
		}
		go ui.WaitForRaiseUIThenQuit()
		err := manager.InstallManager()
		if err != nil {
			if err == manager.ErrManagerAlreadyRunning {
				checkForAdminDesktop()
			}
			fatal(err)
		}
		checkForAdminDesktop()
		time.Sleep(30 * time.Second)
		fatal("WireGuard system tray icon did not appear after 30 seconds.")
		return
	case "/uninstallmanagerservice":
		if len(os.Args) != 2 {
			usage()
		}
		err := manager.UninstallManager()
		if err != nil {
			fatal(err)
		}
		return
	case "/managerservice":
		if len(os.Args) != 2 {
			usage()
		}
		err := manager.Run()
		if err != nil {
			fatal(err)
		}
		return
	case "/installtunnelservice":
		if len(os.Args) != 3 {
			usage()
		}
		err := manager.InstallTunnel(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	case "/uninstalltunnelservice":
		if len(os.Args) != 3 {
			usage()
		}
		err := manager.UninstallTunnel(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	case "/tunnelservice":
		if len(os.Args) != 3 {
			usage()
		}
		err := tunnel.Run(os.Args[2])
		if err != nil {
			fatal(err)
		}
		return
	case "/ui":
		if len(os.Args) != 6 {
			usage()
		}
		err := elevate.DropAllPrivileges(false)
		if err != nil {
			fatal(err)
		}
		readPipe, err := pipeFromHandleArgument(os.Args[2])
		if err != nil {
			fatal(err)
		}
		writePipe, err := pipeFromHandleArgument(os.Args[3])
		if err != nil {
			fatal(err)
		}
		eventPipe, err := pipeFromHandleArgument(os.Args[4])
		if err != nil {
			fatal(err)
		}
		ringlogger.Global, err = ringlogger.NewRingloggerFromInheritedMappingHandle(os.Args[5], "GUI")
		if err != nil {
			fatal(err)
		}
		manager.InitializeIPCClient(readPipe, writePipe, eventPipe)
		ui.RunUI()
		return
	case "/dumplog":
		if len(os.Args) != 3 {
			usage()
		}
		file, err := os.Create(os.Args[2])
		if err != nil {
			fatal(err)
		}
		defer file.Close()
		err = ringlogger.DumpTo(file, true)
		if err != nil {
			fatal(err)
		}
		return
	case "/runtunnelservice":
		if len(os.Args) != 3 {
			usage()
		}

		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		if !filepath.IsAbs(os.Args[2]) {
			cleanPath, absErr := filepath.Abs(os.Args[2])
			if absErr != nil {
				fmt.Println("Converting config path to absolute failed.")
			}
			fmt.Println(fmt.Printf("%s is now %s", os.Args[2], cleanPath))
			os.Args[2] = cleanPath
		}

		// Secretly remove the tunnel if it's already connected.
		manager.UninstallTunnel(strings.ReplaceAll(filepath.Base(os.Args[2]), ".conf", ""))

		err := ringlogger.InitGlobalLogger("CLI")
		go func() {
			// Follow log.
			ticker := time.NewTicker(time.Millisecond / 2)
			cursor := ringlogger.Global.GetLatestCursor()

			for {
				select {
				case <-ticker.C:
					var items []ringlogger.FollowLine
					items, cursor = ringlogger.Global.FollowFromCursor(cursor)

					if len(items) == 0 {
						continue
					}

					for _, line := range items {
						fmt.Println(line.Line)
					}

				case <-done:
					ticker.Stop()
					break
				}
			}

		}()

		fmt.Println("[CLI] Initializing and installing tunnel.")
		err = manager.InstallTunnel(os.Args[2])
		if err != nil {
			fatal(err)
		}
		fmt.Println("[CLI] Tunnel installed.")

		go func() {
			sig := <-sigs
			fmt.Println(fmt.Sprintf("[CLI] Received a signal of type %s, ending tunnel.", sig))
			manager.UninstallTunnel(strings.ReplaceAll(filepath.Base(os.Args[2]), ".conf", ""))
			done <- true
		}()
		<-done
		manager.UninstallTunnel(strings.ReplaceAll(filepath.Base(os.Args[2]), ".conf", ""))
		fmt.Println("[CLI] Tunnel ended.")
		return
	}
	usage()
}
