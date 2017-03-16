package main

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	server := Server{}
	r := gin.Default()
	r.GET("/", server.ServeIndex)
	r.GET("/auth", server.Auth)
	r.Run()
}
func buildInitalQuery() (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://accounts.google.com/o/oauth2/auth", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("client_id", client_id)
	q.Add("redirect_uri", "https://auth.k8s.wdc.sl.g2trk.com/auth")
	q.Add("response_type", "code")
	q.Add("scope", "openid email profile")
	q.Add("access_type", "offline")
	q.Add("prompt", "consent")
	q.Add("hd", "go2mobi.com")
	req.URL.RawQuery = q.Encode()

	return req, nil
}
func generateUser(email, clientId, clientSecret, idToken, refreshToken string) *KubectlUser {
	return &KubectlUser{
		Name: email,
		KubeUserInfo: &KubeUserInfo{
			AuthProvider: &AuthProvider{
				APConfig: &APConfig{
					ClientID:     clientId,
					ClientSecret: clientSecret,
					IdToken:      idToken,
					IdpIssuerUrl: "https://accounts.google.com",
					RefreshToken: refreshToken,
				},
				Name: "oidc",
			},
		},
	}
}
