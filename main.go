package main

import (
	"./models"
	"fmt"
)

func main() {
	p := models.NewProject("Project Test", "ptest", "A test project")
	fmt.Println("Created test project\n", p)
	r := models.NewRelease(1, "ch12")
	fmt.Println("\nCreated test release\n", r)
	page := models.NewPage(1, "manga/chapters/rnrl/001.jpg")
	fmt.Println("\nCreated test page\n", page)
}
