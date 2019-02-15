package pac

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

var buf, err = ioutil.ReadFile("pac.txt")

const pacUrl = "http://127.0.0.1:9999/pac"

//在windows平台才会有真实的操作
var setWindowsRegistry func() = func() {}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, string(buf))
	switch runtime.GOOS {
	case "darwin":
	case "windows":
		setWindowsRegistry()
	case "linux":
	}
}

func ServePAC() {
	if err != nil {
		log.Println("pac设置失败", err)
	} else {
		http.HandleFunc("/pac", handler)
		log.Println("pac地址为：", pacUrl)
		switch runtime.GOOS {
		case "darwin":
		case "windows":
			setWindowsRegistry()
		case "linux":
		}
		if err := http.ListenAndServe("127.0.0.1:9999", nil); err != nil {
			log.Println("serve PAC过程中出错", err)
		}
	}

}
