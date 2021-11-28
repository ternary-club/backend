package main

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	Row     uint `json:"row"`
	Column  uint `json:"column"`
	Message string
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
		repo, ok := repos[c.Param("repo")]
		if ok {
			c.JSON(http.StatusOK, repo)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Send repository files
	r.POST("/:repo", func(c *gin.Context) {
		repo := c.Param("repo")
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
		repo := c.Param("repo")
		_, ok := repos[repo]
		if ok {
			time.Sleep(1 * time.Second)
			if rand.Float32() > .9 {
				c.JSON(http.StatusBadRequest, gin.H{
					"problems": []Error{
						{
							Row:     1,
							Column:  2,
							Message: "ugly assertion detected",
						},
						{
							Row:     4,
							Column:  3,
							Message: "cannot use kek that way",
						},
						{
							Row:     5,
							Column:  8,
							Message: "variable is not poggers",
						},
					}})
				return
			}

			if c.Query("compile") == "false" {
				c.Status(http.StatusOK)
				return
			}

			tryteCount := rand.Intn(10) + 1
			trits := make([]byte, tryteCount*3)
			crand.Read(trits)

			trytes := make([]string, tryteCount)
			for i := range trytes {
				trytes[i] += string(byteToHepta(trits[i*3+0]))
				trytes[i] += string(byteToHepta(trits[i*3+1]))
				trytes[i] += string(byteToHepta(trits[i*3+2]))
			}

			ternary := strings.Join(trytes, "\n")
			fmt.Println(ternary)

			newRepo := repos[repo]
			newRepo.Ter = ternary
			repos[repo] = newRepo

			c.JSON(http.StatusOK, gin.H{".ter": ternary})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Run
	r.POST("/:repo/run", func(c *gin.Context) {
		repo := c.Param("repo")
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
		repo := c.Param("repo")
		_, ok := repos[repo]
		if ok {
			if name := c.Query("name"); name != "" {
				repos[name] = repos[repo]
				delete(repos, repo)
			}
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
	// Create repository
	r.PUT("/:repo", func(c *gin.Context) {
		repo := c.Param("repo")
		_, ok := repos[repo]
		if !ok {
			if strings.Trim(repo, " ") == "" {
				c.JSON(http.StatusConflict, gin.H{"error": "invalid repository name"})
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
		repo := c.Param("repo")
		_, ok := repos[repo]
		if ok {
			delete(repos, repo)
			c.Status(http.StatusOK)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "repository does not exist"})
		}
	})
}
