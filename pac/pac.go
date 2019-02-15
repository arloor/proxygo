package pac

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"net/http"
	"runtime"
)

var buf, err = ioutil.ReadFile("pac.txt")

const pacUrl = "http://127.0.0.1:9999/pac"

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
		fmt.Println("pac设置失败")
		fmt.Println(err)
	} else {
		http.HandleFunc("/pac", handler)
		fmt.Println("设置pac地址为：", pacUrl)
		switch runtime.GOOS {
		case "darwin":
		case "windows":
			setWindowsRegistry()
		case "linux":
		}

	}
	if err := http.ListenAndServe("127.0.0.1:9999", nil); err != nil {
		fmt.Println("serve PAC过程中出错")
	}
}

func setWindowsRegistry() {
	//设置注册表
	key, exists, err := registry.CreateKey(registry.CURRENT_USER, "Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", registry.ALL_ACCESS)
	if err != nil {
		fmt.Println(err)
	}
	defer key.Close()

	if exists {
		//fmt.Println("键已存在")
	} else {
		fmt.Println("此注册表项不存在，已自动新建")
	}
	// 写入字符串
	key.SetStringValue("AutoConfigURL", pacUrl)
	key.SetStringValue("ProxyServer", "127.0.0.1:8081")
	key.SetStringValue("ProxyOverride", "localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;172.32.*;192.168.*;127.0.0.1;<local>")
	// 写入32位整形值
	key.SetDWordValue("MigrateProxy", 0x00000001)
	key.SetDWordValue("ProxyEnable", 0x00000000)
	fmt.Println("设置windows代理相关注册表成功")
}
