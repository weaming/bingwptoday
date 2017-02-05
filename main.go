package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	fp "path/filepath"
	"strings"
	"time"
)

// HTTP GET timeout
const TIMEOUT = 20

var CA, _ = user.Current()
var outdir = flag.String("d", CA.HomeDir+"/Downloads/BingWallpapers", "Directory to store wallpapers.")
var number = flag.String("n", "1", "Number of wallpapers to download.")

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}

func main() {
	flag.Parse()

	// create directory if not exists
	var _, err = os.Stat(*outdir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(*outdir, 0755)
		if err != nil {
			panic(err)
		}
	}

	api := get_api_json(*number)
	const wp_url_base = "http://www.bing.com"
	for _, image := range api.Images {
		url := wp_url_base + image.Url
		name := strings.SplitN(image.CopyRight, " (©", 2)[0] + fp.Ext(image.Url)
		downloadImage(url, fp.Join(*outdir, name))
	}
}

type API struct {
	Images []struct {
		CopyRight     string `json:"copyright"`
		CopyRightLink string `json:"copyrightlink"`
		Url           string `json:"url"`
		UrlBase       string `json:"urlbase"`
	} `json:"images"`
}

func get_api_json(number string) (api *API) {
	var ajaxAPI = fmt.Sprintf("http://www.bing.com/HPImageArchive.aspx?format=js&n=%v", number)

	c := &http.Client{
		Timeout: TIMEOUT * time.Second,
	}
	resp, err := c.Get(ajaxAPI)

	if err != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		log.Println("Trouble making GET request!")
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Trouble reading reesponse body!")
		return
	}

	err = json.Unmarshal(contents, &api)
	check(err)
	return
}

func downloadImage(url, out string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in downloadImage(): ", r)
		}
	}()

	if _, err := os.Stat(out); err == nil {
		log.Printf("Ignore existed: %v \n\t\t=> %v\n\n", url, out)
		return
	} else {
		log.Printf("%v \n\t\t=> %v\n\n", url, out)
	}

	c := &http.Client{
		Timeout: TIMEOUT * time.Second,
	}
	resp, err := c.Get(url)

	if err != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		log.Println("Trouble making GET photo request!")
		return
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Trouble reading reesponse body!")
		return
	}

	err = ioutil.WriteFile(out, contents, 0644)
	if err != nil {
		log.Println("Trouble creating file!")
		return
	}
}
