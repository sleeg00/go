package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()                    //라우터와 미들웨어를 구성하기위한 Gin의 인스턴스를 생성
	r.GET("/ping", func(c *gin.Context) { //Get요청(URL)이 "/ping"이면 실행되도록 라우터를 등록
		c.JSON(200, gin.H{ //응답은 이렇게 하겠다
			"message": "pong",
		})
	})
	r.Run(":9000") //서버를 시작하고 클라이언트 요청을 수신대기함
	//포트 번호가 지정 가능하다
}
