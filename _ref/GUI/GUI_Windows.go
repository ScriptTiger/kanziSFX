package main

import (
	"path/filepath"
	"os"
	"strings"
	"syscall"
	"unsafe"

	. "github.com/ScriptTiger/kanziSFX"
	. "github.com/ScriptTiger/cno/win/gui"
)

const (
	// kanziSFX accelerator
	accelerator int64 = 500000

	// Window dimensions

	WINDOW_WIDTH = 400
	WINDOW_HEIGHT = 115

	// Control IDs

	TEXT = 0
	EDIT_FIELD = 1
	BROWSE_BUTTON = 2
	EXTRACT_BUTTON = 3
	CANCEL_BUTTON = 4

	// Control layout

	PAD = 5
	LINE_HEIGHT = 20
	BROWSE_WIDTH = 20
	EDIT_WIDTH = WINDOW_WIDTH-BROWSE_WIDTH-25
	BUTTON_WIDTH =  55

	// Custom window messages

	WM_EXTRACTION_COMPLETE = 0x7fff
)

var (
	// Path buffer
	pathBuffer = make([]uint16, MAX_PATH)

	// Tar boolean denoting if Kanzi bit stream contains tar or not
	tar bool
)

// Window callback function
func proc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) (uintptr) {
	switch msg {
		case WM_CREATE:
			CreateWindowEx(
				0,
				CStr("static"),
				CStr("Extract to:"),
				WS_CHILD | WS_VISIBLE,
				PAD, PAD, WINDOW_WIDTH, LINE_HEIGHT,
				uintptr(hwnd),
				uintptr(TEXT),
				0, 0,
			)
			defaultPath, _ := os.Executable()
			defaultPath, _ = filepath.EvalSymlinks(defaultPath)
			defaultPath = strings.TrimSuffix(defaultPath, filepath.Ext(defaultPath))
			CreateWindowEx(
				WS_EX_CLIENTEDGE,
				CStr("edit"),
				CStr(defaultPath),
				WS_CHILD | WS_VISIBLE | ES_AUTOHSCROLL,
				PAD, LINE_HEIGHT+PAD, EDIT_WIDTH, LINE_HEIGHT,
				uintptr(hwnd),
				uintptr(EDIT_FIELD),
				0, 0,
			)
			CreateWindowEx(
				0,
				CStr("button"),
				CStr("..."),
				WS_CHILD | WS_VISIBLE,
				EDIT_WIDTH+PAD, LINE_HEIGHT+PAD, BROWSE_WIDTH, LINE_HEIGHT,
				uintptr(hwnd),
				uintptr(BROWSE_BUTTON),
				0, 0,
			)
			CreateWindowEx(
				0,
				CStr("button"),
				CStr("Extract"),
				WS_CHILD | WS_VISIBLE,
				WINDOW_WIDTH-BUTTON_WIDTH*2-25, LINE_HEIGHT*2+PAD*2, BUTTON_WIDTH, LINE_HEIGHT,
				uintptr(hwnd),
				uintptr(EXTRACT_BUTTON),
				0, 0,
			)
			CreateWindowEx(
				0,
				CStr("button"),
				CStr("Cancel"),
				WS_CHILD | WS_VISIBLE,
				WINDOW_WIDTH-BUTTON_WIDTH-20, LINE_HEIGHT*2+PAD*2, BUTTON_WIDTH, LINE_HEIGHT,
				uintptr(hwnd),
				uintptr(CANCEL_BUTTON),
				0, 0,
			)
			return 0
		case WM_DESTROY:
			PostQuitMessage(0)
			return 0
		case WM_COMMAND:
			id := int(wparam & 0xffff)
			switch id {
				case BROWSE_BUTTON:
					var pathGet uintptr
					if tar {
						pathGet = SHBrowseForFolder(uintptr(unsafe.Pointer(&BROWSEINFOW{
							HwndOwner:	hwnd,
							LpszTitle:	(*uint16)(unsafe.Pointer(CStr("Select the directory to extract to."))),
							UlFlags:	BIF_NEWDIALOGSTYLE | BIF_RETURNONLYFSDIRS,
						})))
						if pathGet != 0 {
							SHGetPathFromIDList(pathGet, uintptr(unsafe.Pointer(&pathBuffer[0])))
							CoTaskMemFree(pathGet)
						}
					} else {
						pathGet = GetSaveFileName(uintptr(unsafe.Pointer(&OPENFILENAMEW{
							LStructSize:	uint32(unsafe.Sizeof(OPENFILENAMEW{})),
							HwndOwner:	hwnd,
							LpstrFile:	(*uint16)(unsafe.Pointer(&pathBuffer[0])),
							NMaxFile:	MAX_PATH,
						})))
					}
					if pathGet != 0 {
						SetWindowText(
							GetDlgItem(uintptr(hwnd), EDIT_FIELD),
							uintptr(unsafe.Pointer(&pathBuffer[0])),
						)
					}
				case EXTRACT_BUTTON:
					GetWindowText(
						GetDlgItem(uintptr(hwnd), EDIT_FIELD),
						uintptr(unsafe.Pointer(&pathBuffer[0])),
						MAX_PATH,
					)
					SetWindowText(
						GetDlgItem(uintptr(hwnd), TEXT),
						CStr("Extracting..."),
					)
					DestroyWindow(GetDlgItem(uintptr(hwnd), EDIT_FIELD))
					DestroyWindow(GetDlgItem(uintptr(hwnd), BROWSE_BUTTON))
					DestroyWindow(GetDlgItem(uintptr(hwnd), EXTRACT_BUTTON))
					pathBufferStr := syscall.UTF16ToString(pathBuffer)
					go func() {
						err := Extract(&pathBufferStr, accelerator, nil, REWRITE_PATH)
						if err != nil {MessageBox(0, CStr(err.Error()), CStr("Error"), MB_ICONERROR)}
						PostMessage(
							uintptr(hwnd),
							WM_EXTRACTION_COMPLETE,
							0, 0,
						)
					}()
				case CANCEL_BUTTON:
					DestroyWindow(uintptr(hwnd))
			}
			return 0
		case WM_EXTRACTION_COMPLETE:
			SetWindowText(
				GetDlgItem(uintptr(hwnd), TEXT),
				CStr("Extraction complete!"),
			)
			SetWindowText(
				GetDlgItem(uintptr(hwnd), CANCEL_BUTTON),
				CStr("Okay"),
			)
			return 0
	}
	return DefWindowProc(
		uintptr(hwnd),
		uintptr(msg),
		wparam,
		lparam,
	)
}

func main() {

	// Initialize aliases to use aliases before CreateWindow is called
	// CreateWindow will check if aliases have already initialized or not and automatical initialize if not
	Init_aliases()

	// Set up variables for kanziSFX

	outNamePtr := new(string)
	ctx := make(map[string]any)

	// Call kanziSFX
	err := Extract(outNamePtr, accelerator, ctx, INFO)

	// Report any errors and exit
	if err != nil {
		MessageBox(0, CStr(err.Error()), CStr("Error"), MB_ICONERROR)
		os.Exit(1)
	}

	// Check if Kanzi bit stream contains tar or not
	if ctx["tar"].(bool) {tar = true}

	// Get screen dimensions

	screenWidth := GetSystemMetrics(SM_CXSCREEN)
	screenHeight := GetSystemMetrics(SM_CYSCREEN)

	// Call CreateWindow
	CreateWindow(
		proc,
		COLOR_MENU + 1,
		CStr("kanziSFX: Kanzi self-extracting archive"),
		WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX | WS_VISIBLE,
		(screenWidth - uintptr(WINDOW_WIDTH))/2, (screenHeight - uintptr(WINDOW_HEIGHT))/2, uintptr(WINDOW_WIDTH), uintptr(WINDOW_HEIGHT),
	)
}
