package main

import (
	"./models"
	"fmt"
	"time"
)

func main() {
	p := models.Project{
		"123",
		"Project Test",
		"ptest",
		"A test project",
		models.PStatusPublished,
		time.Now(),
	}
	fmt.Println("Created test project\n", p)
	r := models.Release{
		"321",
		"ch12",
		1,
		models.RStatusDraft,
		"abc123",
		time.Now(),
	}
	fmt.Println("\nCreated test release\n", r)
}
