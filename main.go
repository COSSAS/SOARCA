package main

import (

	// docs "github.com/go-project-name/docs"
	routes "soarca/routes"
)

//	@title			Swagger Example API
// //	@version		1.0

// type album struct {
// 	ID     string  `json:"id"`
// 	Title  string  `json:"title"`
// 	Artist string  `json:"artist"`
// 	Price  float64 `json:"price"`
// }

// var albums = []album{
// 	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
// 	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
// 	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
// }

// func getAlbums(c *gin.Context) {
// 	c.IndentedJSON(http.StatusOK, albums)
// }

// func postAlbums(c *gin.Context) {
// 	var newAlbum album

// 	// Call BindJSON to bind the received JSON to
// 	// newAlbum.
// 	if err := c.BindJSON(&newAlbum); err != nil {
// 		return
// 	}

// 	// Add the new album to the slice.
// 	albums = append(albums, newAlbum)
// 	c.IndentedJSON(http.StatusCreated, newAlbum)
// }

// // @BasePath /api/v1

// // PingExample godoc
// // @Summary ping example
// // @Schemes
// // @Description do ping
// // @Tags example
// // @Accept json
// // @Produce json
// // @Success 200 {string} Helloworld
// // @Router /example/helloworld [get]
// func Helloworld(g *gin.Context) {
// 	g.JSON(http.StatusOK, "helloworld")
// }

// // func main() {
// // 	router := gin.Default()
// // 	router.GET("/albums", getAlbums)
// // 	router.POST("/albums", postAlbums)
// // 	router.Run("localhost:8080")
// // }
// r := gin.Default()
// docs.SwaggerInfo.BasePath = "/api/v1"
// v1 := r.Group("/api/v1")
// {
// 	eg := v1.Group("/example")
// 	{
// 		eg.GET("/helloworld", Helloworld)
// 	}
// }

func main() {
	api := routes.Setup()
	api.Run(":8080")

}
