package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	// 获取原始请求的路径
	path := req.URL.Path

	if path == "/favicon.ico" {
		return
	}

	parts := strings.SplitN(path, "/", 3)

	domain := parts[1]

	path = "/" + parts[2]

	// 去掉前缀的 "/localhost:8080"
	targetURL := "https://" + domain

	// 解析目标 URL
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)
	// 修改请求头
	req.URL.Host = target.Host
	req.URL.Scheme = target.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = target.Host
	req.URL.Path = path
	// 转发请求至目标 URL
	proxy.ServeHTTP(res, req)
}

func main() {
	// 创建处理函数
	http.HandleFunc("/", handleRequestAndRedirect)

	// 启动服务并监听端口
	fmt.Println("Starting server on port 10086...")
	log.Fatal(http.ListenAndServe("localhost:10086", nil))
}

//运行 golang 代码，然后将订阅链接 https:// 替换成 http://localhost:10086/ 然后更新，以后每次更新前都需要先运行 golang 代码。
