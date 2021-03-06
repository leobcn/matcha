/*
Package application provides access to application resources. Image assets
must be in the app's .xcassets file (iOS) or res/drawable folder (Android).
Disable "Compress PNG Files" and "Remove Text Metadata from PNG Files" in Xcode
if loading image resources is not working. Android does not allow uppercase
image names or folders and this restriction carries over to Matcha as well.

    // Display an image.
    img, err := application.LoadImage("example")
    if err != nil {
        imageview.Image = img
    }

    // or
    imageview.Image = application.MustLoadImage("example")
*/
package application

import (
	"errors"
	"runtime"

	"gomatcha.io/matcha/bridge"
)

// // AssetsDir returns the path to the app's assets directory. `NSBundle.mainBundle.resourcePath`
// func AssetsDir() (string, error) {
// 	return bridge.Bridge("").Call("assetsDir").ToString(), nil
// }

func OpenURL(url string) error {
	success := true
	if runtime.GOOS == "android" {
		success = bridge.Bridge("").Call("openURL", bridge.String(url)).ToBool()
	} else {
		success = bridge.Bridge("").Call("openURL:", bridge.String(url)).ToBool()
	}
	if !success {
		return errors.New("Unable to open URL")
	}
	return nil
}
