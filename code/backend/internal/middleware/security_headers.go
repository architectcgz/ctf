package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有 API 响应设置安全 HTTP 头。
// CSP 在 JSON API 响应上不会被执行（浏览器只在文档载入时强制执行 CSP），
// 但作为纵深防御：若出现 content-type 混淆，可阻止浏览器将 API 响应当作 HTML 解析。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

