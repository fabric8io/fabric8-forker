package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"encoding/json"

	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
)

type errorResponse struct {
	Msg string `json:"msg"`
}

type okResponse struct {
	URL string `json:"url"`
}

var (
	// Commit current build commit set by build script
	Commit = "0"
	// BuildTime set by build script in ISO 8601 (UTC) format: YYYY-MM-DDThh:mm:ssTZD (see https://www.w3.org/TR/NOTE-datetime for details)
	BuildTime = "0"
	// StartTime in ISO 8601 (UTC) format
	StartTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
)

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerContentType, "application/json")
	w.WriteHeader(http.StatusOK)

	type status struct {
		Commit    string `json:"commit"`
		BuildTime string `json:"buildTime"`
		StartTime string `json:"startTime"`
	}
	json.NewEncoder(w).Encode(&status{Commit: Commit, BuildTime: BuildTime, StartTime: StartTime})
}

func fork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerContentType, "application/json")

	authorization := r.Header.Get(headerAuthorization)
	if authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(&errorResponse{Msg: "require '" + headerAuthorization + "' header"})
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(&errorResponse{Msg: "require query argument 'url'"})
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.Replace(authorization, "Bearer ", "", -1)},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	user, _, err := client.Users.Get("")
	fmt.Println("Current user: ", *user.Login)

	owner, repository, err := ParseOwnerAndRepo(url)
	if err != nil {
		panic("Could not locate owner/repo: " + err.Error())
	}

	fmt.Println("Owner", owner)
	fmt.Println("Repo ", repository)

	ownedByUser := *user.Login == owner

	fmt.Println("OwnedByUser", ownedByUser)

	var userRepo *github.Repository
	if !ownedByUser {
		var err error
		userRepo, _, err = client.Repositories.Get(*user.Login, repository)
		if err != nil {
			userRepo, _, err = client.Repositories.CreateFork(owner, repository, &github.RepositoryCreateForkOptions{})
			if err != nil {
				fmt.Println(err)
			}
		}
		count := 0
		for userRepo == nil && count < 5 {
			userRepo, _, err = client.Repositories.Get(*user.Login, repository)
			time.Sleep(1 * time.Second)
			count++
		}
	}
	if userRepo != nil {
		fmt.Println("UserRepo", *userRepo.CloneURL)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&okResponse{URL: *userRepo.CloneURL})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&okResponse{URL: url})
}

func main() {

	host := ":8080"

	http.HandleFunc("/fork", fork)
	http.HandleFunc("/status", status)
	log.Println("Started listening on ", host)
	log.Fatal(http.ListenAndServe(host, nil))
}

// ParseOwnerAndRepo tries to match known URLs to extract Owner and Repository name
func ParseOwnerAndRepo(url string) (owner, repo string, err error) {
	exp, err := regexp.Compile(".*github.com.(.+)/(.+).git")
	if err != nil {
		return "", "", fmt.Errorf("Error in regexp. ??")
	}
	matches := exp.FindAllStringSubmatch(url, -1)
	if len(matches) == 1 {
		return matches[0][1], matches[0][2], nil
	}

	return "", "", fmt.Errorf("URL[%v] does not match a known pattern[%v]", url, exp)
}
