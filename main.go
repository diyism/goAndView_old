package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const buildPath = "build"

var (
	apitarget = flag.String("apitarget", "", "Required. Android build target. To list possible targets run $ANDROID_HOME/tools/android list targets")
	gradle    = flag.String("gradle", "2.2", "Gradle version")
	plugin    = flag.String("plugin", "1.3.0", "Android gradle plugin version")
)

type modBytes func([]byte) ([]byte, error)

func modFile(fpath string, modfn modBytes) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatalf("unable to read %s: %s", fpath, err)
	}
	w, err := modfn(b)
	if err != nil {
		log.Fatalf("error processing %s: %s", fpath, err)
	}
	if err := ioutil.WriteFile(fpath, w, 0644); err != nil {
		log.Fatalf("umable to write changes to build.gradle: %s", err)
	}
}

func javaImportPath(goImportPath string) string {
	parts := strings.Split(goImportPath, "/")
	domainparts := strings.Split(parts[0], ".")
	ret := ""
	for _, domainpart := range domainparts {
		ret = domainpart + "." + ret
	}
	ret = ret[:len(ret)-1]

	for _, part := range parts[1:] {
		ret = ret + "." + part
	}
	return ret
}

func main() {
	flag.Parse()

	if *apitarget == "" {
		log.Fatalf("-target must be specified")
		return
	}

	// Check for existance of the android tool in the sdk path
	toolPath := filepath.Join(os.Getenv("ANDROID_HOME"), "tools", "android")
	if runtime.GOOS == "windows" {
		toolPath = toolPath + ".bat"
	}
	if _, err := os.Stat(toolPath); err != nil {
		log.Fatalf("couldn't find %s. Environment variable ANDROID_HOME must be defined and point to a valid Android SDK folder", toolPath)
		return
	}

	// figure out the package name from the current working directory
	// and set the import path for the java app
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("couldn't fetch working dir path: %s", err)
		return
	}

	pkg, err := build.ImportDir(wd, build.IgnoreVendor)
	if err != nil || pkg.ImportPath == "" {
		log.Fatalf("could not derive package path from import path")
		return
	}
	pkgPath := javaImportPath(pkg.ImportPath) + "." + buildPath

	outPath := "./" + buildPath

	// Project Generation: Run the android tool and then generate the go files
	var out []byte
	fmt.Println(toolPath, "create", "project", "--name", pkg.Name,
		"--package", pkgPath, "--activity", "Main", "--target", *apitarget,
		"--gradle", "--gradle-version", *plugin, "--path", outPath)
	if out, err = exec.Command(toolPath, "create", "project", "--name", pkg.Name,
		"--package", pkgPath, "--activity", "Main", "--target", *apitarget,
		"--gradle", "--gradle-version", *plugin, "--path", outPath,
	).CombinedOutput(); err != nil {
		log.Fatalf("error in android tool: %s\n%s", err, out)
		return
	}
	fmt.Println(string(out))

	// build.gradle: add dependencies for gomobile bind; fix minifyEnabled
	modFile(filepath.Join(outPath, "build.gradle"), buildDotGradle)

	// libs: create the folder
	libsPath := filepath.Join(outPath, "libs")
	if _, err = os.Stat(libsPath); os.IsNotExist(err) {
		if err = os.Mkdir(libsPath, os.ModeDir|0775); err != nil {
			log.Fatalf("unable to create libs folder: %v", err)
			return
		}
	}

	// gradle-wrapper.properties: update to a more modern gradle version
	modFile(filepath.Join(outPath, "gradle", "wrapper", "gradle-wrapper.properties"), gradlewrapperDotProperties)

	// AndroidManifest.xml: add permissions for INTERNET and NETWORK_STATE
	modFile(filepath.Join(outPath, "src", "main", "AndroidManifest.xml"), androidManifestDotXML)

	// strings.xml: modify the value of app_name to title
	modFile(filepath.Join(outPath, "src", "main", "res", "values", "strings.xml"), func(b []byte) ([]byte, error) {
		return bytes.Replace(b, []byte("Main"), []byte(pkg.Name), 1), nil
	})

	// main.xml: remove the current layout file
	if err = os.Remove(filepath.Join(outPath, "src", "main", "res", "layout", "main.xml")); err != nil {
		log.Fatalf("unable to remove main.xml: %s", err)
		return
	}

	// finally write the new contents of Main.java
	fpath := filepath.Join(outPath, "src", "main", "java")
	for _, s := range strings.Split(pkgPath, ".") {
		fpath = filepath.Join(fpath, s)
	}
	fpath = filepath.Join(fpath, "Main.java")
	if err = ioutil.WriteFile(fpath, []byte(mainDotJava), 0644); err != nil {
		log.Fatalf("unable to write Main.java: %s", err)
		return
	}

	//writing the server.go file, the server has Start and Stop exported functions to control the http server
	//it also imports the users package that will setup the DefaultServeMux
	srccode, err := format.Source(bytes.Replace([]byte(webapp), []byte("{import}"), []byte(pkg.ImportPath), 1))
	if err != nil {
		log.Fatalf("unable to write server.go: %s", err)
		return
	}
	if err = ioutil.WriteFile("./"+buildPath+"/webapp.go", srccode, 0644); err != nil {
		log.Fatalf("unable to write server.go: %s", err)
		return
	}

	// and generate gowebview.aar
	if out, err = exec.Command("gomobile", "bind", "-o", filepath.Join(outPath, "libs", "gowebview.aar"), "./"+buildPath).CombinedOutput(); err != nil {
		log.Fatalf("could not generate libs/gowebview.aar: %s: %s", err, out)
		return
	}
	fmt.Println(string(out))

	// write a gitignore so we don't checkin local properties or build files
	if err = ioutil.WriteFile(filepath.Join(outPath, ".gitignore"), []byte(gitignore), 0644); err != nil {
		log.Fatalf("error writing .gitignore:%s", err)
		return
	}
}
