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

//App is an exported class whitch can be used from native java code to control the http server
type App struct {
	server *http.Server
}

//NewApp initializes and returns an new App
func NewApp() *App {
	return &App{
		server: &http.Server{
			Addr:           "127.0.0.1:0",
			Handler:        http.DefaultServeMux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

//Start is called by the native portion of the webapp to start the web server.
//It returns the server root URL (without the trailing slash) and any errors.
func (app *App) Start() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		err = app.server.Serve(listener)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return listener.Addr().String(), nil
}

//Stop is called by the native portion of the webapp to stop the web server.
func (app *App) Stop() {
	//waiting for app.server.Close() in 1.8
}`

func buildDotGradle(b []byte) ([]byte, error) {
	b = bytes.Replace(b, []byte("runProguard false"), []byte("minifyEnabled true"), 1)
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
	i := bytes.Index(b, []byte("<application"))
	if i < 0 {
		return nil, fmt.Errorf("Could not find <application tag in AndroidManifest.xml")
	}
	buf := bytes.Buffer{}
	buf.Write(b[:i])
	buf.Write([]byte(androidManifestDotXMLTextSystem))
	buf.Write(b[i:])
	return buf.Bytes(), nil
}

const androidManifestDotXMLTextSystem = `
    <uses-permission android:name="android.permission.INTERNET" />

`

const mainDotJava = `
package gowebview;

import android.app.Activity;
import android.os.Bundle;
import android.view.KeyEvent;
import android.webkit.WebSettings;
import android.webkit.WebViewClient;
import android.widget.Toast;
import android.webkit.WebView;

import gowebview.App;
import gowebview.Gowebview;

public class Main extends Activity {
    private WebView mWebView;
	private App mSrv;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        mWebView = new WebView(this);
		mSrv = Gowebview.newApp();
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
			mWebView.loadUrl(mSrv.start() + "/");
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
		mSrv.stop();
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
