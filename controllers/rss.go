package controllers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"

	"github.com/4rneee/noah/models"
	"github.com/gin-gonic/gin"
)

const RFC822_4Y = "02 Jan 2006 15:04 MST" // https://validator.w3.org/feed/docs/warning/ProblematicalRFC822Date.html

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Items       []Item   `xml:"item"`
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Guid        string   `xml:"guid"`
	PubDate     string   `xml:"pubDate"`
}

func Feed(c *gin.Context) {
	const MAX_ITEMS = 20

	var posts []models.Post
	err := models.DB.
		Order("created_at desc").
		Limit(MAX_ITEMS).
		Find(&posts).
		Error
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user, ok := get_current_user(c)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	redaction_date := get_redaction_date(&user)

	channel := Channel{
		Title:       "Notes Of A Human",
		Link:        os.Getenv("BASE_URL"),
		Description: "Notes Of A Human",
		Items:       []Item{},
	}

	for i := range posts {
		post := &posts[i]
		if redaction_date.Compare(post.CreatedAt) < 0 {
			post.Redact()
		}

		link := fmt.Sprintf("%s/post/%d", channel.Link, post.ID)

		channel.Items = append(channel.Items,
			Item{
				Title:       post.Title,
				Link:        link,
				Description: fmt.Sprintf("%s by %s", post.Title, post.UserName),
				Guid:        link,
				PubDate:     post.CreatedAt.Format(RFC822_4Y),
			})
	}

	rssData, err := xml.MarshalIndent(RSS{Version: "2.0", Channel: channel}, "", "  ")

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/xml", []byte(xml.Header+string(rssData)))
}
