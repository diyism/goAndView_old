package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

const webapp = `
package gowebview

import (
	"log"
	"net"
	"net/http"
	"time"

	//importing the users package that will attach the handlers to the DefaultServeMux
	_ "{import}"
)

var server = &http.Server{
	Addr:           "127.0.0.1:0",
	Handler:        http.DefaultServeMux,
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20,
}

//Start is called by the native portion of the webapp to start the web server.
//It returns the server root URL (without the trailing slash) and any errors.
func Start() string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		err = server.Serve(listener)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return listener.Addr().String()
}

//Stop is called by the native portion of the webapp to stop the web server.
func Stop() {
	server.Close()
}`

func buildDotGradle(b []byte) ([]byte, error) {
	b = bytes.Replace(b, []byte("runProguard false"), []byte("minifyEnabled true"), 1)
	b = bytes.Replace(b, []byte("mavenCentral"), []byte("jcenter"), 1)
	return append(b, []byte(buildDotGradleTextSystem)...), nil
}

const buildDotGradleTextSystem = `
android {
    defaultConfig {
        minSdkVersion 15
    }
}

repositories {
    flatDir {
        dirs 'libs'
    }
}

dependencies {
    compile(name:'gowebview', ext:'aar')
}

task genGoMobileAAR(type:Exec) {
  workingDir '.'
  commandLine 'gomobile', 'bind', '-o', 'libs/gowebview.aar', '.'
}

preBuild.dependsOn(genGoMobileAAR)
`

func gradlewrapperDotProperties(b []byte) ([]byte, error) {
	s := bufio.NewScanner(bytes.NewReader(b))
	buf := bytes.Buffer{}
	for s.Scan() {
		txt := s.Text()
		if strings.Index(txt, "distributionUrl") == 0 {
			txt = fmt.Sprintf("distributionUrl=https\\://services.gradle.org/distributions/gradle-%s-all.zip", *gradle)
		}
		buf.WriteString(txt + "\n")
	}
	return buf.Bytes(), s.Err()
}

func androidManifestDotXML(b []byte) ([]byte, error) {
	return bytes.Replace(b, []byte("<application"), []byte(androidManifestDotXMLTextSystem), 1), nil
}

const androidManifestDotXMLTextSystem = `
    <uses-permission android:name="android.permission.INTERNET" />

	<application android:theme="@android:style/Theme.Holo.NoActionBar.Fullscreen"
`

const mainDotJava = `
package {package};

import android.app.Activity;
import android.os.Bundle;
import android.view.KeyEvent;
import android.webkit.WebSettings;
import android.webkit.WebViewClient;
import android.widget.Toast;
import android.webkit.WebView;

import gowebview.Gowebview;

public class Main extends Activity {
    private WebView mWebView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        mWebView = new WebView(this);
		WebSettings webSettings = mWebView.getSettings();
		webSettings.setJavaScriptEnabled(true);
		mWebView.setWebViewClient(new WebViewClient());
        setContentView(mWebView);
    }

	// We start the server on onResume
    @Override
    protected void onResume() {
        super.onResume();
        try {
			String address = Gowebview.start();
			mWebView.loadUrl("http://" + address + "/");
        } catch (Exception e) {
            Toast.makeText(this,"Error:" + e.toString(),Toast.LENGTH_LONG).show();
            e.printStackTrace();
            this.finish();
        }
    }

    // Send a graceful shut down signal to the server. onPause is guaranteed
	// to be called by Android while onStop or onDestroy may not be called.
    @Override
    protected void onPause() {
        super.onPause();
		Gowebview.stop();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
    }

    // We override back key press to close the app rather than pass it to the WebView
    @Override
    public boolean dispatchKeyEvent(KeyEvent event) {
        if(event.getKeyCode() == KeyEvent.KEYCODE_BACK){
            this.finish();
            return true;
        }
        return super.dispatchKeyEvent(event);
    }
}
`
const gitignore = `androidapp.iml
.gradle/
.idea/
local.properties
build/
`
