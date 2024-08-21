package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	//"syscall"
	//"unsafe"

	"github.com/lxn/win"
	"github.com/mitchellh/go-ps"
	//"golang.org/x/sys/windows"
)

func main() {
	//function GetWindowDesktopNumber(hWindow: HWND): Integer; stdcall; external 'VirtualDesktopAccessor.dll';
	//function IsPinnedWindow(hwnd: HWND): Integer; stdcall; external 'VirtualDesktopAccessor.dll';
	//currentDesktopNumber := GetCurrentDesktopNumber()
	winidActivate := flag.Int("a", 0, "activate window by ID")
	winidClose := flag.Int("c", 0, "close window by ID")
	flag.Parse()

	if *winidActivate > 0 {
		hwnd := win.HWND(*winidActivate)
		// TODO Lebeda - nahradit WinActivate
		if win.IsIconic(hwnd) {
			win.ShowWindow(hwnd, win.SW_RESTORE)
		}
		win.SetForegroundWindow(hwnd)
		win.SetActiveWindow(hwnd)
	} else if *winidClose > 0 {
		hwnd := win.HWND(*winidClose)
		win.SendMessage(hwnd, win.WM_CLOSE, 0, 0)
	} else {
		PrintWindows("", "")
	}

}

func PrintWindows(aFilter, aSelfExe string) {
	hDesktop := win.GetDesktopWindow()
	//foregroundWindow := win.GetForegroundWindow() // current foreground window - current terminal - if use tabs,
	winlist := []string{}
	hWindow := win.GetWindow(hDesktop, win.GW_CHILD)
	for hWindow != 0 {

		// skip current foreground window - if use terminal tabs, show as other
		//if hWindow == foregroundWindow {
		//	hWindow = win.GetWindow(hWindow, win.GW_HWNDNEXT)
		//	continue
		//}

		lDesktop := GetWindowDesktopNumber(hWindow)
		lIsPined := IsPinnedWindow(hWindow)
		title := GetWindowText(hWindow)
		visible := win.IsWindowVisible(hWindow)
		isFs := IsFullScreen(hWindow)
		//isIconic := win.IsIconic(hWindow)
		isMainWindow := (win.GetWindow(hWindow, win.GW_OWNER) == 0) && win.IsWindowVisible(hWindow)

		// start for debug
		//processId := GetWindowThreadProcessId(uintptr(hWindow))
		//p, _ := ps.FindProcess(int(processId))
		//if p != nil {
		//	lExeFile := p.Executable()
		//	if strings.Contains(title, "settings.json - Notepad") {
		//		fmt.Println("title: ", title)
		//		fmt.Println("lDesktop: ", lDesktop)
		//		fmt.Println("lIsPined: ", lIsPined)
		//		fmt.Println("isFs: ", isFs)
		//		fmt.Println("isIconic: ", isIconic)
		//		fmt.Println("visible: ", visible)
		//		fmt.Println("isMainWindow: ", isMainWindow)
		//		fmt.Println("lExeFile: ", lExeFile)
		//		fmt.Println()
		//	}
		//}
		// end for debug

		if (len(title) > 0) && ((lDesktop < 4294967295) || (lIsPined == 1)) && (visible || isFs) && isMainWindow {
			//if (len(title) > 0) && visible {

			processId := GetWindowThreadProcessId(uintptr(hWindow))
			p, _ := ps.FindProcess(int(processId))
			lExeFile := p.Executable()
			//rect := GetWindowDimensions(hWindow)

			if !((lExeFile == "fzfMenu.exe") ||
				(lExeFile == "FastWindowSwitcher.exe") ||
				(lExeFile == "SystemSettings.exe") ||
				(lExeFile == "TextInputHost.exe") ||
				(lExeFile == "ApplicationFrameHost.exe") ||
				(title == "Program Manager")) { // TODO Lebeda - parametrizovat exclude
				//((rect.Left < -10) && (rect.Top < -10))
				lMenuTitle := `[` + strconv.Itoa(lDesktop+1) + `] ` + fmt.Sprintf("%-20v", lExeFile) + ` - ` + title
				//lMenuTitle := fmt.Sprintf("%-20v", lExeFile) + ` - ` + title

				// TODO Lebeda - zjistit aktuální okno     https://gist.github.com/obonyojimmy/d6b263212a011ac7682ac738b7fb4c70  GetForegroundWindow
				// TODO Lebeda - zjistit minimalizované okno
				// TODO Lebeda - zjištěné rozdíly ~14pt
				//strRect := "[" + strconv.Itoa(int(rect.Left)) + "," + // TODO Lebeda - strRect jen na debug
				//	strconv.Itoa(int(rect.Top)) + "," +
				//	strconv.Itoa(int(rect.Right)) + "," +
				//	strconv.Itoa(int(rect.Bottom)) + "]"
				winItem := `[ win] ` + lMenuTitle + " " + // strRect +
					"\t" + `WinList -a=` + strconv.Itoa(int(hWindow)) +
					"\t" + `WinList -c=` + strconv.Itoa(int(hWindow))
				winlist = append(winlist, winItem)
				//fmt.Println(`[win] ` + lMenuTitle + "\t" + `WinList -a=` + strconv.Itoa(int(hWindow)))
			}

		}

		hWindow = win.GetWindow(hWindow, win.GW_HWNDNEXT)
	}

	sort.Strings(winlist)
	for _, s := range winlist {
		fmt.Println(s)
	}
}
