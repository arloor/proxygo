// +build windows

package pac

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

func init() {
	setWindowsRegistry = doSetWindowsRegistry
}

func doSetWindowsRegistry() {
	//设置注册表
	key, exists, err := registry.CreateKey(registry.CURRENT_USER, "Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings", registry.ALL_ACCESS)
	if err != nil {
		log.Println(err)
	}
	defer key.Close()

	if exists {
		//log.Println("键已存在")
	} else {
		log.Println("此注册表项不存在，已自动新建")
	}
	// 写入字符串
	key.SetStringValue("AutoConfigURL", pacUrl)
	key.SetStringValue("ProxyServer", "127.0.0.1:8081")
	key.SetStringValue("ProxyOverride", "localhost;127.*;10.*;172.16.*;172.17.*;172.18.*;172.19.*;172.20.*;172.21.*;172.22.*;172.23.*;172.24.*;172.25.*;172.26.*;172.27.*;172.28.*;172.29.*;172.30.*;172.31.*;172.32.*;192.168.*;127.0.0.1;<local>")
	// 写入32位整形值
	key.SetDWordValue("MigrateProxy", 0x00000001)
	key.SetDWordValue("ProxyEnable", 0x00000000)
	log.Println("自动设置windows代理相关注册表成功")
}
