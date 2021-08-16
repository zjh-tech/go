// +build !dev

package enet

import "github.com/gin-gonic/gin"

func setGinMode() {
	gin.SetMode(gin.ReleaseMode)
}
