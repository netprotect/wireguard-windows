/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/windows/version"
)

var easterEggIndex = -1

func onAbout(owner walk.Form) {
	showError(runAboutDialog(owner), owner)
}

func runAboutDialog(owner walk.Form) error {
	vbl := walk.NewVBoxLayout()
	vbl.SetMargins(walk.Margins{80, 20, 80, 20})
	vbl.SetSpacing(10)

	var disposables walk.Disposables
	defer disposables.Treat()

	dlg, err := walk.NewDialogWithFixedSize(owner)
	if err != nil {
		return err
	}
	disposables.Add(dlg)
	dlg.SetTitle("About WireGuard")
	dlg.SetLayout(vbl)
	if icon, err := loadLogoIcon(32); err == nil {
		dlg.SetIcon(icon)
	}

	font, _ := walk.NewFont("Segoe UI", 9, 0)
	dlg.SetFont(font)

	iv, err := walk.NewImageView(dlg)
	if err != nil {
		return err
	}
	iv.SetCursor(walk.CursorHand())
	iv.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			win.ShellExecute(dlg.Handle(), nil, windows.StringToUTF16Ptr("https://www.wireguard.com/"), nil, nil, win.SW_SHOWNORMAL)
		} else if easterEggIndex >= 0 && button == walk.RightButton {
			if icon, err := loadSystemIcon("moricons", int32(easterEggIndex), 128); err == nil {
				iv.SetImage(icon)
				easterEggIndex++
			} else {
				easterEggIndex = -1
				if logo, err := loadLogoIcon(128); err == nil {
					iv.SetImage(logo)
				}
			}
		}
	})
	if logo, err := loadLogoIcon(128); err == nil {
		iv.SetImage(logo)
	}

	wgLbl, err := walk.NewTextLabel(dlg)
	if err != nil {
		return err
	}
	wgFont, _ := walk.NewFont("Segoe UI", 16, walk.FontBold)
	wgLbl.SetFont(wgFont)
	wgLbl.SetTextAlignment(walk.AlignHCenterVNear)
	wgLbl.SetText("WireGuard")

	detailsLbl, err := walk.NewTextLabel(dlg)
	if err != nil {
		return err
	}
	detailsLbl.SetTextAlignment(walk.AlignHCenterVNear)
	_, appVersion := version.RunningNameVersion()
	detailsLbl.SetText(fmt.Sprintf("App version: %s\nGo backend version: %s\nGo version: %s\nOperating system: %s\nArchitecture: %s", appVersion, device.WireGuardGoVersion, strings.TrimPrefix(runtime.Version(), "go"), version.OsName(), runtime.GOARCH))

	copyrightLbl, err := walk.NewTextLabel(dlg)
	if err != nil {
		return err
	}
	copyrightFont, _ := walk.NewFont("Segoe UI", 7, 0)
	copyrightLbl.SetFont(copyrightFont)
	copyrightLbl.SetTextAlignment(walk.AlignHCenterVNear)
	copyrightLbl.SetText("Copyright © 2015-2019 Jason A. Donenfeld. All Rights Reserved.")

	buttonCP, err := walk.NewComposite(dlg)
	if err != nil {
		return err
	}
	hbl := walk.NewHBoxLayout()
	hbl.SetMargins(walk.Margins{VNear: 10})
	buttonCP.SetLayout(hbl)
	walk.NewHSpacer(buttonCP)
	closePB, err := walk.NewPushButton(buttonCP)
	if err != nil {
		return err
	}
	closePB.SetAlignment(walk.AlignHCenterVNear)
	closePB.SetText("Close")
	closePB.Clicked().Attach(func() {
		dlg.Accept()
	})
	donatePB, err := walk.NewPushButton(buttonCP)
	if err != nil {
		return err
	}
	donatePB.SetAlignment(walk.AlignHCenterVNear)
	donatePB.SetText("♥ Donate!")
	donatePB.Clicked().Attach(func() {
		if easterEggIndex == -1 {
			easterEggIndex = 0
		}
		win.ShellExecute(dlg.Handle(), nil, windows.StringToUTF16Ptr("https://www.wireguard.com/donations/"), nil, nil, win.SW_SHOWNORMAL)
		dlg.Accept()
	})
	walk.NewHSpacer(buttonCP)

	dlg.SetDefaultButton(donatePB)
	dlg.SetCancelButton(closePB)

	disposables.Spare()

	dlg.Run()

	return nil
}
