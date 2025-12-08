package main

import (
	"path/filepath"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	. "github.com/ScriptTiger/kanziSFX"
	. "github.com/ScriptTiger/cno/win/gui"
)

const (
	// kanziSFX accelerator
	ACCELERATOR int64 = 500000

	// Window dimensions

	WINDOW_WIDTH = 400
	WINDOW_HEIGHT = 115

	// Control IDs

	TEXT = 0
	EDIT_FIELD = 1
	BROWSE_BUTTON = 2
	EXTRACT_BUTTON = 3
	CANCEL_BUTTON = 4
	PROGRESS_BAR = 5

	// Control layout

	PAD = 5
	LINE_HEIGHT = 20
	BROWSE_WIDTH = 20
	EDIT_WIDTH = WINDOW_WIDTH-BROWSE_WIDTH-25
	BUTTON_WIDTH =  55

	// Custom window messages

	WM_EXTRACTION_COMPLETE = 0x7fff
	WM_EXTRACTION_FAILED = 0x7ffe
)

var (
	// Default path
	defaultPath string

	// Path buffer
	pathBuffer = make([]uint16, MAX_PATH)

	// Tar boolean denoting if Kanzi bit stream contains tar or not
	tar bool

	// Progress tracker provided to kanziSFX
	progress [2]int64

	// Errors
	err error
)

// Function to report error and exit
func errExit(err error, code int) {
	MessageBox(0, CStr(err.Error()), CStr("Error"), MB_ICONERROR)
	os.Exit(code)
}

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
					var pbhwnd uintptr
					if progress[1] != 0 {
						pbhwnd = CreateWindowEx(
							0,
							CStr(PROGRESS_CLASS),
							0,
							WS_CHILD | WS_VISIBLE,
							PAD, LINE_HEIGHT+PAD, WINDOW_WIDTH-25, LINE_HEIGHT,
							uintptr(hwnd),
							uintptr(PROGRESS_BAR),
							0, 0,
						)
						SendMessage(
							pbhwnd,
							PBM_SETRANGE,
							0,
							uintptr(100<<16),
						)
					}
					pathBufferStr := syscall.UTF16ToString(pathBuffer)
					running := true
					go func() {
						err = Extract(&pathBufferStr, ACCELERATOR, nil, &progress, REWRITE_PATH)
						running = false
						if err != nil {
							PostMessage(
								uintptr(hwnd),
								WM_EXTRACTION_FAILED,
								0, 0,
							)
						} else {
							PostMessage(
								uintptr(hwnd),
								WM_EXTRACTION_COMPLETE,
								0, 0,
							)
						}
					}()
					if progress[1] != 0 {
						go func() {
							for ; running; {
								SendMessage(
									uintptr(pbhwnd),
									PBM_SETPOS,
									uintptr(int((float64(progress[0])/float64(progress[1]))*100)),
									0,
								)
								time.Sleep(100*time.Millisecond)
							}
							if err != nil {
								SendMessage(
									uintptr(pbhwnd),
									PBM_SETPOS,
									0,
									0,
								)
							} else {
								SendMessage(
									uintptr(pbhwnd),
									PBM_SETPOS,
									100,
									0,
								)
							}
						}()
					}
				case CANCEL_BUTTON:
					DestroyWindow(uintptr(hwnd))
			}
			return 0
		case WM_EXTRACTION_FAILED:
			SetWindowText(
				GetDlgItem(uintptr(hwnd), TEXT),
				CStr("A problem occurred during extraction!"),
			)
			SetWindowText(
				GetDlgItem(uintptr(hwnd), CANCEL_BUTTON),
				CStr("Okay"),
			)
			MessageBox(uintptr(hwnd), CStr(err.Error()), CStr("Error"), MB_ICONERROR)
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
	err = Extract(outNamePtr, ACCELERATOR, ctx, nil, INFO)

	// If there was an error, report it and exit
	if err != nil {errExit(err, 1)}

	// Check if Kanzi bit stream contains tar or not
	if ctx["tar"].(bool) {tar = true}

	// Check if output size is present to use with progress tracking
	if value, hasKey := ctx["outputSize"]; hasKey {progress[1] = value.(int64)}

	// Generate default path
	defaultPath, err = os.Executable()
	if err != nil {errExit(err, 2)
	} else {
		defaultPath, err = filepath.EvalSymlinks(defaultPath)
		if err != nil {errExit(err, 3)}
	}
	defaultPath = strings.TrimSuffix(filepath.Base(defaultPath), filepath.Ext(defaultPath))

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
