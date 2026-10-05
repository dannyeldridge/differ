package gui

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

// differ runs as a bare binary, not an .app bundle, so macOS has no icon file to
// show in the Dock and Cmd+Tab. Set one on the running application instead.
static void setAppIcon(const void *bytes, int length) {
	NSData *data = [NSData dataWithBytes:bytes length:length];
	NSImage *image = [[NSImage alloc] initWithData:data];
	if (image == nil) {
		return;
	}
	dispatch_async(dispatch_get_main_queue(), ^{
		[NSApp setApplicationIconImage:image];
	});
}
*/
import "C"

import (
	_ "embed"
	"unsafe"
)

//go:embed icon.png
var appIcon []byte

func setAppIcon() {
	C.setAppIcon(unsafe.Pointer(&appIcon[0]), C.int(len(appIcon)))
}
