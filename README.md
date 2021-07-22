# gowebview

Tool to create an android application from an go http server package. The package must just set some endpoints.

## prerequisites
download and install android [sdk](https://developer.android.com/studio/index.html#downloads) and [ndk](https://developer.android.com/ndk/downloads/index.html)

gomobile: `go get -u golang.org/x/mobile/...`

set environment variables:
- `JAVA_HOME` - where the `javac` is
- `ANDROID_HOME` - android sdk path
- `ANDROID_NDK_ROOT` - android ndk path

example http package that sets the '/' path:
```go
package exampleserver

import "net/http"

func init() {
	//just add some endpoints to the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	//or you can also use your favorite router and then paste it to http.DefaultServeMux
}
```
Check the ID of the target android API with `$ $ANDROID_HOME/tools/android list target`. It will be eg. 13.

Then go to the package directory and run:
```bash
$ gomobile init -ndk $ANDROID_NDK_ROOT
$ gowebview -target 13
```

`gowebview` creates an [gradle](https://gradle.org/) project in build directory.

Go to the build directory and create your android apk:
```bash
$ cd build
$ ./gradlew build
```

You can now install your brand new android application:
```
$ adb install -r build/outputs/apk/build-debug.apk
```
Enjoy your awesome app:

![screenshot](https://github.com/microo8/gowebview/raw/master/screenshot.png)
