package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func router(app *gin.Engine) {
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	app.GET("/albums", func(c *gin.Context) {
		db := CustomDb{
			path: "./data/albums",
		}
		data, err := db.GetAlbums()
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		c.JSON(http.StatusOK, data)
	})

	app.POST("/albums", func(c *gin.Context) {
		var newAlbums []Album

		if err := c.BindJSON(&newAlbums); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err)
			return
		}

		db := CustomDb{
			path: "./data/albums",
		}

		data, err := db.AddAlbums(newAlbums)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		c.JSON(http.StatusCreated, data)
	})

	app.DELETE("/albums", func(c *gin.Context) {
		var keys []string

		if err := c.BindJSON(&keys); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err)
			return
		}

		db := CustomDb{path: "./data/albums"}

		data, err := db.DeleteAlbums(keys)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"keysDeleted": data,
		})
	})

	app.PATCH("/albums", func(c *gin.Context) {
		var albumsData []AlbumWithId

		if err := c.BindJSON(&albumsData); err != nil {
			log.Println(err, "err")
			c.JSON(http.StatusBadRequest, err)
			return
		}

		db := CustomDb{path: "./data/albums"}

		data, err := db.UpdateAlbums(albumsData)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		c.JSON(http.StatusOK, data)
	})
}
