// Package liuyao 暴露嵌入的 Web 前端静态资源。
//
// 使用 go:embed 将 web/ 目录编译进二进制，使服务可脱离源码目录独立部署。
package liuyao

import "embed"

//go:embed web
var WebFS embed.FS
