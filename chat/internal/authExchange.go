package chat

import (
    "net/url"

    "github.com/gin-gonic/gin"
)

// AuthExchangeHandler принимает access/refresh токены через query,
// устанавливает HttpOnly куки и редиректит на главную без параметров
func AuthExchangeHandler(c *gin.Context) {
    access := c.Query("access_token")
    refresh := c.Query("refresh_token")

    if access == "" || refresh == "" {
        c.JSON(400, gin.H{
            "error": "tokens are required",
        })
        return
    }

    setAuthCookies(c, access, refresh)

    // Очищаем query из URL при редиректе
    dest := &url.URL{Path: "/"}
    c.Redirect(302, dest.String())
}


