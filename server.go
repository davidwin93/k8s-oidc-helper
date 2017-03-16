package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"strings"

	"gopkg.in/gin-gonic/gin.v1"
)

type Server struct{}

func (s *Server) ServeIndex(context *gin.Context) {
	req, err := buildInitalQuery()
	if err != nil {
		context.AbortWithError(500, err)
		return
	}
	context.Redirect(307, req.URL.String())
}

const client_id = "957307268185-i0rjb14s8r9g8g61okcj422ee8vagv95.apps.googleusercontent.com"
const client_secret = "RrYOyJiCwFkiQKd7--bmdeJJ"

func (s *Server) Auth(c *gin.Context) {
	oauthErr := c.DefaultQuery("error", "-1")
	if oauthErr != "-1" {
		c.String(500, "Could not create K8s user. Error: %s", oauthErr)
		return
	}
	oauthCode := c.DefaultQuery("code", "-1")
	if oauthCode == "-1" {
		c.String(500, "Invalid OAuth Code.")
		return
	}
	form := url.Values{}
	form.Add("code", oauthCode)
	form.Add("client_id", client_id)
	form.Add("client_secret", client_secret)
	form.Add("redirect_uri", "https://auth.k8s.wdc.sl.g2trk.com/auth")
	form.Add("grant_type", "authorization_code")
	resp, err := http.PostForm("https://www.googleapis.com/oauth2/v4/token", form)

	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Println(resp.Status)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	defer resp.Body.Close()
	respT := GoogleTokenResponse{}

	err = json.NewDecoder(resp.Body).Decode(&respT)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	email, err := s.getUserEmail(respT)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	if !strings.Contains(email, "@go2mobi.com") {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	log.Printf("%+v", respT)
	user := generateUser(email, client_id, client_secret, respT.IDToken, respT.RefreshToken)
	c.YAML(200, user)
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
}
type UserInfo struct {
	Email string `json:"email"`
}
type KubectlUser struct {
	Name         string        `yaml:"name"`
	KubeUserInfo *KubeUserInfo `yaml:"user"`
}

type KubeUserInfo struct {
	AuthProvider *AuthProvider `yaml:"auth-provider"`
}

type AuthProvider struct {
	APConfig *APConfig `yaml:"config"`
	Name     string    `yaml:"name"`
}

type APConfig struct {
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	IdToken      string `yaml:"id-token"`
	IdpIssuerUrl string `yaml:"idp-issuer-url"`
	RefreshToken string `yaml:"refresh-token"`
}

func (s *Server) getUserEmail(gtr GoogleTokenResponse) (string, error) {
	uri, _ := url.Parse("https://www.googleapis.com/oauth2/v3/userinfo")
	q := uri.Query()
	q.Set("alt", "json")
	q.Set("access_token", gtr.AccessToken)
	uri.RawQuery = q.Encode()
	resp, err := http.Get(uri.String())
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	ui := &UserInfo{}
	err = json.NewDecoder(resp.Body).Decode(ui)
	if err != nil {
		return "", err
	}
	return ui.Email, nil
}
