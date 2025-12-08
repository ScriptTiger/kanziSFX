package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/ScriptTiger/kanziSFX"
	. "github.com/ScriptTiger/cno"
	. "github.com/ScriptTiger/cno/lnx/gui"
	"github.com/ebitengine/purego"
)

// kanziSFX accelerator
const ACCELERATOR int64 = 500000

var (
	// Widget handles
	hwnd, labelHwnd,
	pathBoxHwnd, entryHwnd, browseButtonHwnd,
	buttonBoxHwnd, cancelButtonHwnd, extractButtonHwnd,
	pbHwnd uintptr

	// Path string
	path string

	// Tar boolean denoting if Kanzi bit stream contains tar or not
	tar bool

	// Progress tracker provided to kanziSFX
	progress [2]int64

	// Status of extraction
	running bool

	// Instructions
	instructions string

	// Errors
	err error
)

// Function to report error and exit
func errExit(err error, code int) {
	errorDialog := Gtk_message_dialog_new(
		0,
		0,
		GTK_MESSAGE_ERROR,
		GTK_BUTTONS_OK,
		CStr(err.Error()),
	)
	Gtk_window_set_title(errorDialog, CStr("Error"))
	Gtk_dialog_run(errorDialog)
	os.Exit(code)
}

// Callback for browse button "clicked" signal
func browseButtonClick() {

	var action uintptr
	if tar {action = GTK_FILE_CHOOSER_ACTION_CREATE_FOLDER
	} else {action = GTK_FILE_CHOOSER_ACTION_SAVE}

	browseDialog := Gtk_file_chooser_dialog_new(
		CStr(instructions+"."),
		hwnd,
		action,
		CStr("Cancel"), GTK_RESPONSE_CANCEL,
		CStr("Okay"), GTK_RESPONSE_ACCEPT,
		0, 0,
	)

	if Gtk_dialog_run(browseDialog) == GTK_RESPONSE_ACCEPT {
		pathPtr := Gtk_file_chooser_get_filename(browseDialog)
		if pathPtr != 0 {
			path = GStr(pathPtr)
			G_free(pathPtr)
			Gtk_entry_set_text(entryHwnd, CStr(path))
		}
	}

	Gtk_widget_destroy(browseDialog)
	return
}

// Callback for extract button "clicked" signal
func extractButtonClick() {
	pathPtr := Gtk_entry_get_text(entryHwnd)
	if pathPtr != 0 {path = GStr(pathPtr)}
	Gtk_label_set_text(labelHwnd, CStr("Extracting..."))
	Gtk_widget_destroy(entryHwnd)
	Gtk_widget_destroy(browseButtonHwnd)
	Gtk_widget_destroy(extractButtonHwnd)
	if progress[1] != 0 {
		pbHwnd = Gtk_progress_bar_new()
		Gtk_box_pack_start(pathBoxHwnd, pbHwnd, 1, 1, 5)
		Gtk_widget_show(pbHwnd)
	}
	running = true
	G_idle_add(purego.NewCallback(progressCallback), 0)
	go func() {
		err = Extract(&path, ACCELERATOR, nil, &progress, REWRITE_PATH)
		running = false
	}()
}

// Callback for tracking progress of extraction and updating window accordingly
func progressCallback() bool {
	if running {
		if progress[1] != 0 {Gtk_progress_bar_set_fraction(pbHwnd, float64(progress[0])/float64(progress[1]))}
		return true
	}
	if err != nil {
		if progress[1] != 0 {Gtk_progress_bar_set_fraction(pbHwnd, 0)}
		Gtk_label_set_text(labelHwnd, CStr("A problem occurred during extraction!"))
		Gtk_button_set_label(cancelButtonHwnd, CStr("Okay"))
		errorDialog := Gtk_message_dialog_new(
			hwnd,
			GTK_DIALOG_MODAL,
			GTK_MESSAGE_ERROR,
			GTK_BUTTONS_OK,
			CStr(err.Error()),
		)
		Gtk_window_set_title(errorDialog, CStr("Error"))
		Gtk_dialog_run(errorDialog)
		Gtk_widget_destroy(errorDialog)
	} else {
		if progress[1] != 0 {Gtk_progress_bar_set_fraction(pbHwnd, 1)}
		Gtk_label_set_text(labelHwnd, CStr("Extraction complete!"))
		Gtk_button_set_label(cancelButtonHwnd, CStr("Okay"))
	}
	return false
}


func main() {

	// Lock this main go routine to the current OS thread
	runtime.LockOSThread()

	// Initialize GTK
	Gtk_init(0, 0)

	// Set up variables for kanziSFX
	outNamePtr := new(string)
	ctx := make(map[string]any)

	// Call kanziSFX
	err = Extract(outNamePtr, ACCELERATOR, ctx, nil, INFO)

	// If there was an error, report it and exit
	if err != nil {errExit(err, 1)}

	// Check if Kanzi bit stream contains tar or not
	if ctx["tar"].(bool) {
		tar = true
		instructions = "Specify the directory to extract to"
	} else {instructions = "Specify the file to extract to"}

	// Check if output size is present to use with progress tracking
	if value, hasKey := ctx["outputSize"]; hasKey {progress[1] = value.(int64)}

	path, err := os.Executable()
	if err != nil {errExit(err, 2)
	} else {
		path, err = filepath.EvalSymlinks(path)
		if err != nil {errExit(err, 3)}
	}
	path = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	// Create and set up main window
	hwnd = Gtk_window_new(0)
	Gtk_window_set_title(hwnd, CStr("kanziSFX: Kanzi self-extracting archive"))
	Gtk_window_set_resizable(hwnd, 0)

	// Create main box with vertical orientation and add it to the window
	mainBoxHwnd := Gtk_box_new(1, 5)
	Gtk_container_add(hwnd, mainBoxHwnd)

	// Main box widgets
	labelHwnd = Gtk_label_new(CStr(instructions+":"))
	pathBoxHwnd = Gtk_box_new(0, 5)
	buttonBoxHwnd = Gtk_box_new(0, 5)
	Gtk_box_pack_start(mainBoxHwnd, labelHwnd, 1, 1, 5)
	Gtk_box_pack_start(mainBoxHwnd, pathBoxHwnd, 0, 0, 5)
	Gtk_box_pack_start(mainBoxHwnd, buttonBoxHwnd, 0, 0, 5)

	// Path box widgets
	entryHwnd = Gtk_entry_new()
	Gtk_entry_set_text(entryHwnd, CStr(path))
	browseButtonHwnd = Gtk_button_new_with_label(CStr("..."))
	Gtk_box_pack_start(pathBoxHwnd, entryHwnd, 1, 1, 5)
	Gtk_box_pack_start(pathBoxHwnd, browseButtonHwnd, 0, 0, 5)

	// Button box widgets
	cancelButtonHwnd = Gtk_button_new_with_label(CStr("Cancel"))
	extractButtonHwnd = Gtk_button_new_with_label(CStr("Extract"))
	Gtk_box_pack_end(buttonBoxHwnd, cancelButtonHwnd, 0, 0, 5)
	Gtk_box_pack_end(buttonBoxHwnd, extractButtonHwnd, 0, 0, 5)

	// Connect event signals to callbacks
	G_signal_connect_data(browseButtonHwnd, CStr("clicked"), purego.NewCallback(browseButtonClick), 0, 0, 0)
	G_signal_connect_data(extractButtonHwnd, CStr("clicked"), purego.NewCallback(extractButtonClick), 0, 0, 0)
	G_signal_connect_data(cancelButtonHwnd, CStr("clicked"), purego.NewCallback(Gtk_main_quit), 0, 0, 0)
	G_signal_connect_data(hwnd, CStr("destroy"), purego.NewCallback(Gtk_main_quit), 0, 0, 0)

	// Show everything and start the main loop
	Gtk_widget_show_all(hwnd)
	Gtk_main()
}