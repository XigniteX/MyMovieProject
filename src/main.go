package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
	"golang.org/x/oauth2"
	"net/http"
	"fmt"
	"context"
	"io/ioutil"
)


var conf_trakt *oauth2.Config
var trakt_version = "2"
var trakt_api_key = "dc5e44dfa4e2ed9123c5a8b446802acad6ce20bfd7ccf07ffe9b80b85327aa09"
var tmdb_api_key = "d6e5abc928615e9b29a3f78afd635add"

func init(){
	var endpoint_trakt = oauth2.Endpoint{}
	endpoint_trakt.AuthURL = "https://trakt.tv/oauth/authorize?response_type=code&client_id=dc5e44dfa4e2ed9123c5a8b446802acad6ce20bfd7ccf07ffe9b80b85327aa09&redirect_uri=localhost:8000"
	endpoint_trakt.TokenURL = "https://api.trakt.tv/oauth/token"

	conf_trakt = &oauth2.Config{
		ClientID:     "dc5e44dfa4e2ed9123c5a8b446802acad6ce20bfd7ccf07ffe9b80b85327aa09",
		ClientSecret: "bbb080674bbe76acca667bc51723017e0e6561584455a99254eaafb29d7c393e",
		RedirectURL:  "http://localhost:8000/traktCallBack",
		Scopes:       []string{ },
		Endpoint:     endpoint_trakt,
	}
}

func main(){
	router := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))

	router.Use(sessions.Sessions("mysession", store))

	router.GET("/login", loginHandler)
	router.GET("/traktCallBack", traktAuthHandler)
	router.GET("/api/profile", traktProfile)
	router.GET("/api/movie/search", traktMovieSearch)
	router.GET("/api/movies", traktMyMovies)
	router.GET("/api/series", traktMySeries)
	router.GET("/api/settings", traktSettings)

	router.Run("localhost:8000")

}

func getTraktLogin() string{
	return conf_trakt.AuthCodeURL("state")
}

func loginHandler(c *gin.Context)  {
	c.Writer.Write([]byte("<html><head><title>Trakt Login</title></head> <body><a href='" + getTraktLogin() + "'><button>Login with Trakt!</button></body></html>"))
}

func traktAuthHandler(c *gin.Context){
	session := sessions.Default(c)

	code := c.Query("code")

	token,err := conf_trakt.Exchange(context.TODO(), code)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		c.Writer.WriteString(err.Error())
		return
	}

	fmt.Println(token.AccessToken)
	session.Set("token", token.AccessToken)
	session.Save()

	c.Writer.WriteString(string("Connected to Trakt.tv"))

}

func traktProfile (c *gin.Context){

	endpoint := "https://api.trakt.tv/users/XigniteX"
	data := doCall(c, endpoint)

	c.JSON(http.StatusOK, data)


}

func traktMovieSearch (c *gin.Context){

	query := c.Query("query")
	endpoint := "https://api.trakt.tv/search/movie?query=" + query

	data := doCall(c, endpoint)

	c.JSON(http.StatusOK, data)

}


func traktMyMovies (c *gin.Context){

	endpoint := "https://api.trakt.tv/users/XigniteX/watched/movies"
	data := doCall(c, endpoint)
	c.JSON(http.StatusOK, string(data))

}

func traktMySeries (c *gin.Context){
	endpoint := "https://api.trakt.tv/users/XigniteX/watched/shows"
	data := doCall(c, endpoint)
	c.JSON(http.StatusOK, data)

}

func traktSettings (c *gin.Context){
	endpoint := "https://api.trakt.tv/users/settings"
	data := doCall(c, endpoint)
	c.JSON(http.StatusOK, data)
}

func doCall(c *gin.Context, endPoint string) string{
	fmt.Println(endPoint)
	session := sessions.Default(c)
	token := session.Get("token")

	if token == nil {
		panic("No session found")
	}

	var newToken oauth2.Token
	newToken.AccessToken = token.(string)


	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{})


	client := conf_trakt.Client(ctx, &newToken)


	req, _ := http.NewRequest("GET", endPoint, nil)
	req.Header.Set("trakt-api-version", trakt_version)
	req.Header.Set("trakt-api-key", trakt_api_key)
	resp, err := client.Do(req)


	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "error.html", nil)
	}

	defer resp.Body.Close()
	data, _:= ioutil.ReadAll(resp.Body)


	return string(data)
}