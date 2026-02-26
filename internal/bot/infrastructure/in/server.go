package in

import "github.com/gin-gonic/gin"

func BuildServer() error {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	err := router.Run()
	if err != nil {
		return err
	} // listens on 0.0.0.0:8080 by default
	return nil
}
