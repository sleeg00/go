package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default() //라우터와 미들웨어를 구성하기위한 Gin의 인스턴스를 생성
	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name") //go에서 Param을 받을때는 ?를 쓰지말고 입륵만 입력함 ex)/user/sleeg
		c.String(http.StatusOK, "Hello %s", name)
	})

	router.Run() //서버를 시작하고 클라이언트 요청을 수신대기함
	//포트 번호가 지정 가능하다
}
