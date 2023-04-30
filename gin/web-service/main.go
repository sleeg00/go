package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrance", Price: 56.99},
	{ID: "2", Title: "Houny", Artist: "Andanta", Price: 54},
	{ID: "3", Title: "KK", Artist: "gojila", Price: 100},
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums) //구조체를 JSON으로 직렬화하고 응답에 추가하도록 호출
}

func postAlbums(c *gin.Context) {
	var newAlbum album

	if err := c.BindJSON(&newAlbum); err != nil { //받은 JSON 데이터를 &newAlbum 변수에 역직렬화(바인딩)한다.
		return
	}

	albums = append(albums, newAlbum)            //albums에 newAlbum을 추가
	c.IndentedJSON(http.StatusCreated, newAlbum) //201코드 반환 -> 추가했다는 코드
}
func main() {
	router := gin.Default()          //라우터 초기화
	router.GET("/albums", getAlbums) //함수이름을 전달하는 것!! 메소드에 반환값을 전달하는 것이 아니다!!
	router.POST("/post/albums", postAlbums)
	router.Run("localhost:8080")
}
