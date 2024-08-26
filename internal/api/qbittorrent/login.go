package qbittorrent

import (
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

type QbittorrentAuthentication struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type QbittorrentAuthenticationApi struct {
	cache    *cache.Cache
	username string
	password string
}

func NewQbittorrentAuthenticationApi(e *echo.Group, username, password string) *QbittorrentAuthenticationApi {
	loginApi := &QbittorrentAuthenticationApi{
		cache:    cache.New(cache.NoExpiration, cache.NoExpiration),
		username: username,
		password: password,
	}

	g := e.Group("/auth")
	g.POST("/login", loginApi.login)
	g.GET("/login", loginApi.login)

	return loginApi
}

func (q *QbittorrentAuthenticationApi) login(c echo.Context) error {

	if _, exist := q.cache.Get("auth"); exist {
		return Ok(c)
	}

	var auth QbittorrentAuthentication

	if c.Request().Method == "POST" {
		parsedBody, err := parseLoginBody(c.Request().Body)

		if err != nil {
			return Fails(c)
		}

		auth = *parsedBody
	} else {
		auth.Username = c.QueryParam("username")
		auth.Password = c.QueryParam("password")
	}

	if auth.Username != q.username || auth.Password != q.password {
		return Fails(c)
	}

	q.cache.Set("auth", auth, cache.NoExpiration)

	return Ok(c)
}

func parseLoginBody(body io.ReadCloser) (*QbittorrentAuthentication, error) {
	content, err := io.ReadAll(body)

	if err != nil {
		return nil, err
	}

	var auth QbittorrentAuthentication

	for _, v := range strings.Split(string(content), "&") {
		pair := strings.Split(v, "=")
		if len(pair) != 2 {
			return nil, nil
		}

		switch pair[0] {
		case "username":
			auth.Username = pair[1]
		case "password":
			auth.Password = pair[1]
		}

	}

	return &auth, nil

}
