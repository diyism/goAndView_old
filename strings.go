package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func buildDotGradle(b []byte) ([]byte, error) {
	b = bytes.Replace(b, []byte("runProguard false"), []byte("minifyEnabled true"), 1)
	return append(b, []byte(buildDotGradleTextSystem)...), nil
}

const buildDotGradleTextSystem = `
android {
    defaultConfig {
        minSdkVersion 19
    }
}

repositories {
    flatDir{
        dirs 'libs'
    }
}

dependencies {
    compile(name:'backend', ext:'aar')
}

task genGoMobileAAR(type:Exec) {
  workingDir '..'
  commandLine 'gomobile', 'bind', '-o', 'androidapp/libs/backend.aar', '.'
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

func mainDotJava([]byte) ([]byte, error) {
	tmpl := template.Must(template.New("mainDotJavaText").Parse(mainDotJavaTextSystem))
	params := struct {
		PkgPath string
		PkgName string
		ClsName string
	}{
		PkgPath: pkgPath,
		PkgName: pkgName,
		ClsName: strings.ToTitle(pkgName),
	}
	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, fmt.Errorf("error writing main.java :%s", err)
	}
	return buf.Bytes(), nil
}

const mainDotJavaTextSystem = `
package {{.PkgPath}};

import android.app.Activity;
import android.os.Bundle;
import android.view.KeyEvent;
import android.webkit.WebSettings;
import android.webkit.WebViewClient;
import android.widget.Toast;

import android.webkit.WebView;
import go.{{.PkgName}}.{{.ClsName}};

public class Main extends Activity {
    private WebView mWebView;
	private {{.ClsName}}.App mSrv;

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
		mSrv = {{.ClsName}}.NewApp();
        try {
			mWebView.loadUrl(mSrv.Start() + "/");
        } catch (Exception e) {
            Toast.makeText(this,"Error:"+e.toString(),Toast.LENGTH_LONG).show();
            e.printStackTrace();
            this.finish();
        }
    }

    // Send a graceful shut down signal to the server. onPause is guaranteed
	// to be called by Android while onStop or onDestroy may not be called.
    @Override
    protected void onPause() {
        super.onPause();
		mSrv.Stop();
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
