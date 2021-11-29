package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stoewer/go-strcase"
)

// Set CORS permissions
func SetupCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
}

type Repository struct {
	Ter string `json:".ter"`
	Try string `json:".try"`
}

type Error struct {
	Row     uint   `json:"row"`
	Column  uint   `json:"column"`
	Message string `json:"message"`
}

var repos = map[string]Repository{}

const HEPTAVINTIMAL = "0123456789ABCDEFGHIJKLMNOPQ"

func byteToHepta(num byte) byte {
	result := 0
	for i := 0; i < 3; i++ {
		trit := (num & (0b11 << (i * 2))) >> (i * 2)
		switch trit {
		case 0b01:
			fallthrough
		case 0b10:
			result += int(trit) * int(math.Pow(3.0, float64(i)))
		}
	}
	return HEPTAVINTIMAL[result]
}

// Setup playground routes
func SetupRoutes(r *gin.Engine) {
	// Get repositories
	r.GET("/", func(c *gin.Context) {
		keys := []string{}
		for k := range repos {
			keys = append(keys, k)
		}
		c.JSON(http.StatusOK, gin.H{"repos": keys})
	})
	// Get repository files
	r.GET("/:repo", func(c *gin.Context) {
		repo, ok := repos[strcase.KebabCase(strings.ToLower(c.Param("repo")))]
		if ok {
			c.JSON(http.StatusOK, repo)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Send repository files
	r.POST("/:repo", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		_, ok := repos[repo]
		if ok {
			type Body struct {
				Try string `json:".try"`
			}
			body := Body{}

			err := c.Bind(&body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}

			newRepo := repos[repo]
			newRepo.Try = body.Try
			repos[repo] = newRepo
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Compile
	r.POST("/:repo/compile", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		r, ok := repos[repo]
		if ok {

			f, _ := os.OpenFile(repo+".try", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			f.Truncate(0)
			f.Seek(0, 0)
			fmt.Fprintf(f, "%s", []byte(r.Try))
			f.Close()

			out, _ := exec.Command("./terry", "eae.try").Output()

			errors := strings.Split(string(out), "\n")
			problems := []Error{}
			for _, e := range errors[:len(errors)-1] {
				data := strings.Split(e, ":")
				row, _ := strconv.Atoi(data[1])
				column, _ := strconv.Atoi(data[2])
				problems = append(problems, Error{Row: uint(row), Column: uint(column), Message: strings.Split(e, "[0m ")[1]})
			}

			c.JSON(http.StatusOK, gin.H{"problems": problems})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Run
	r.POST("/:repo/run", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		_, ok := repos[repo]
		if ok {
			time.Sleep(1 * time.Second)

			c.JSON(http.StatusOK, gin.H{"out": rand.Int()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Update repository info
	r.PATCH("/:repo", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		_, ok := repos[repo]
		if ok {
			if name := c.Query("name"); name != "" {
				_, ok := repos[name]
				if ok {
					c.JSON(http.StatusConflict, gin.H{"error": "repository already exists"})
					return
				} else {
					repos[name] = repos[repo]
					delete(repos, repo)
				}
			}
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Create repository
	r.PUT("/:repo", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		_, ok := repos[repo]
		if !ok {
			if repo == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository name"})
				return
			}
			repos[repo] = Repository{}
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "repository already exists"})
		}
	})
	// Delete repository
	r.DELETE("/:repo", func(c *gin.Context) {
		repo := strcase.KebabCase(strings.ToLower(c.Param("repo")))
		_, ok := repos[repo]
		if ok {
			delete(repos, repo)
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
}
