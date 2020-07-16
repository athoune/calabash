package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/papey/calabash/internal/state"
)

func main() {
	// Session data and state
	var session state.Session

	// Gin default server
	srv := gin.Default()

	// Routes setup
	// Is this server alive ?
	srv.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})

	// Create a pomodoro session
	srv.POST("/pomodoro", func(c *gin.Context) {
		if session.Started {
			c.JSON(http.StatusConflict, "{message: 'A session is already running'}")
			return
		}

		session = state.NewSession()

		// Start a new and fresh session
		session.Start()
		// Run this new session
		go session.Run()

		// Return an OK status plus the current state
		c.JSON(http.StatusOK, &session)
	})

	// Get session state for current timer
	srv.GET("/pomodoro", func(c *gin.Context) {
		// Ensure a Read lock
		session.Lock.RLock()
		defer session.Lock.RUnlock()

		if !session.Started {
			c.JSON(http.StatusNotFound, "{message: 'No session started'}")
			return
		}

		// Return a OK status plus the current state
		c.JSON(http.StatusOK, &session)
	})

	// Delete a running session
	srv.DELETE("/pomodoro", func(c *gin.Context) {

		// Ensure a session is started
		if !session.Started {
			c.JSON(http.StatusNotFound, "{message: 'No session started'}")
			return
		}

		// Cancel the session using cancel chan
		session.Cancel <- true
		// Close this unused channel
		close(session.Cancel)
		// Create a new and fresh session
		session = state.NewSession()

		// Return an OK status
		c.JSON(http.StatusOK, "{message: 'Session deleted'}")
	})

	// Session start/pause current timer
	srv.PUT("/pomodoro", func(c *gin.Context) {
		// Ensure a session is started
		if !session.Started {
			c.JSON(http.StatusNotFound, "{message: 'No session started'}")
			return
		}

		// Toogle running/paused mode for this session
		session.Toogle()
		// Return an OK status plus the current state
		c.JSON(http.StatusOK, &session)
	})

	// Run the gin server
	err := srv.Run()
	if err != nil {
		panic(err)
	}
}
