package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drivev3 "google.golang.org/api/drive/v3"
)

type Service struct {
	config *oauth2.Config
	fileId string
}

func NewService(clientId string, clientSecret string, fileId string) *Service {
	return &Service{
		config: &oauth2.Config{
			RedirectURL:  "http://localhost:3000/callback",
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Scopes:       []string{drivev3.DriveScope},
			Endpoint:     google.Endpoint,
		},
		fileId: fileId,
	}
}

const GOOGLE_OAUTH_API_URL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func (handler *Service) Login(w http.ResponseWriter, r *http.Request) {
	oauthState := handler.generateStateOauthCookie(w)
	u := handler.config.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (handler *Service) Callback(w http.ResponseWriter, r *http.Request) {
	oauthState, err := r.Cookie("state")
	if err != nil {
		log.Println("Cookie Not Found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := handler.downloadFile(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	err = saveFile(data)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	htmlData := `<html>
		<body>
			<h1> Downloaded File</h1>
			<img id="profileImage" src="/img.png">
		</body>
	</html>`

	fmt.Fprintf(w, htmlData)
}

func (handler *Service) generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "state", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (handler *Service) downloadFile(code string) (data []byte, err error) {
	token, err := handler.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	client := handler.config.Client(context.Background(), token)
	driveClient, err := drivev3.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to create drive client ", err)
	}

	res, err := driveClient.Files.Get(handler.fileId).Download()
	if err != nil {
		return nil, fmt.Errorf("unable to download the file ", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to download the file ", err)
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read the respnse", err)
	}

	return data, nil
}
func saveFile(data []byte) (err error) {
	img, _, _ := image.Decode(bytes.NewReader(data))
	out, err := os.Create("./templates/img.png")

	if err != nil {
		return
	}
	err = png.Encode(out, img)
	if err != nil {
		return
	}
	return
}
