package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Companies struct {
	Name    string `gorm:"primary_key" json:"name"`
	Created int    `json:"created"`
	Product string `json:"product"`
}

func main() {
	router := gin.Default() //라우터와 미들웨어를 구성하기위한 Gin의 인스턴스를 생성
	router.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest") //다음에 오는 파라미터값을 기본값으러 설정
		lastname := c.Query("lastname")                   // shortcut for c.Request.URL.Query().Get("lastname")
		///welcome?firstname=go&lastname=gin로 호출하면 Hello go gin이라는 결과를 얻을 수 있다.
		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})
	router.Run() //서버를 시작하고 클라이언트 요청을 수신대기함
	//포트 번호가 지정 가능하다
}
