// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"demo/internal/middleware"
	"flag"
	"fmt"

	"demo/internal/config"
	"demo/internal/handler"
	"demo/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/demo.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	// 全局注册日志中间件
	server.Use(middleware.RequestLog)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
