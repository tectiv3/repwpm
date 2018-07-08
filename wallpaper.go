package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/murlokswarm/app"
)

const (
	feed_url = "https://api.reddit.com/r/earthporn/hot?raw_json=1"
)

var (
	cache_file      = path.Join(os.TempDir(), "cache.json")
	wallpapers_path = path.Join(os.Getenv("HOME"), "Pictures", "EarthWallpapers")
	invalid_path    = path.Join(wallpapers_path, "invalid")
)

type Reddit struct {
	Thread Thread `json:"data"`
	Kind   string
}
type Thread struct {
	Children []Posts
	Dist     int
}
type Posts struct {
	Post Post `json:"data"`
	Kind string
}

type Post struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Permalink string `json:"permalink"`

	// CreatedUTC uint64 `json:"created_utc"`
	Deleted bool `json:"deleted"`

	Ups   int32 `json:"ups"`
	Downs int32 `json:"downs"`
	Likes bool  `json:"likes"`

	Author              string `json:"author"`
	AuthorFlairCSSClass string `json:"author_flair_css_class"`
	AuthorFlairText     string `json:"author_flair_text"`

	Title  string `json:"title"`
	Score  int32  `json:"score"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
	NSFW   bool   `json:"over_18"`

	Subreddit   string `json:"subreddit"`
	SubredditID string `json:"subreddit_id"`

	IsSelf       bool   `json:"is_self"`
	SelfText     string `json:"selftext"`
	SelfTextHTML string `json:"selftext_html"`

	Hidden            bool   `json:"hidden"`
	LinkFlairCSSClass string `json:"link_flair_css_class"`
	LinkFlairText     string `json:"link_flair_text"`

	NumComments int32  `json:"num_comments"`
	Locked      bool   `json:"locked"`
	Thumbnail   string `json:"thumbnail"`

	Gilded        int32  `json:"gilded"`
	Distinguished string `json:"distinguished"`
	Stickied      bool   `json:"stickied"`

	IsRedditMediaDomain bool `json:"is_reddit_media_domain"`
}

func getPosts() []Posts {
	log.Println("get posts")
	if fi, err := os.Stat(cache_file); os.IsNotExist(err) || time.Since(fi.ModTime()).Hours() > 1 {
		if err := downloadFile(feed_url, cache_file); err != nil {
			log.Println(err)
			return []Posts{}
		}
	}
	data, err := ioutil.ReadFile(cache_file)
	if err != nil {
		log.Println(err)
		return []Posts{}
	}
	log.Println("reading cache file")
	reddit := Reddit{}
	if err := json.Unmarshal(data, &reddit); err != nil {
		log.Println(err)
		return []Posts{}
	}
	log.Printf("got %d (%d) posts\n", len(reddit.Thread.Children), reddit.Thread.Dist)
	return reddit.Thread.Children
}

func downloadFile(url, filepath string) error {
	log.Println("downloading from", url)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Golang_Wallpaper_bot/0.0.1")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// data, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	//     return err
	// }
	// log.Printf("downloaded (%d bytes), saving to file\n", len(data))
	// if err := ioutil.WriteFile(cache_file, data, os.ModePerm); err != nil {
	//     return err
	// }
	return nil
}

func exists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

func downloadNewWallpapers() (int, string) {
	posts := getPosts()
	if len(posts) == 0 {
		log.Println("no new posts")
		return 0, "no new posts"
	}
	urls := []string{}
	for _, post := range posts {
		if strings.HasSuffix(post.Post.URL, ".jpg") {
			urls = append(urls, post.Post.URL)
			log.Println(post.Post.URL)
		}
	}

	log.Printf("got %d URLs\n", len(urls))

	os.MkdirAll(wallpapers_path, os.ModePerm)
	os.MkdirAll(invalid_path, os.ModePerm)

	i := 0
	t := 0
	for _, url := range urls {
		pictureName := path.Base(url)
		pictureFileName := path.Join(wallpapers_path, pictureName)
		pictureFileNameInvalid := path.Join(invalid_path, pictureName)
		log.Println(pictureFileName)
		if !exists(pictureFileName) && !exists(pictureFileNameInvalid) {
			if err := downloadFile(url, pictureFileName); err != nil {
				log.Println(err)
				continue
			}
			log.Println("saved")
			t++
			if isInvalid(pictureFileName) {
				os.Rename(pictureFileName, path.Join(invalid_path, pictureName))
				i++
			}
		} else {
			log.Println("file exists")
		}
	}
	result := fmt.Sprintf("downloaded %d new wallpapers (%d invalid)", t, i)
	log.Println(result)
	return t, result
}

func isInvalid(pictureFileName string) bool {
	reader, err := os.Open(pictureFileName)
	if err != nil {
		log.Println(err)
		return true
	}
	defer reader.Close()
	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		log.Printf("%s: %v\n", path.Base(pictureFileName), err)
		return true
	}
	app.Log("%s %d %d\n", path.Base(pictureFileName), im.Width, im.Height)
	return (im.Height > im.Width || im.Width < 1920 || im.Height < 1080)
}

func nextWallpaper() error {
	command := `
    tell application "System Events"
        tell current desktop
            set initInterval to get change interval
            set change interval to initInterval
         end tell
    end tell
    `
	cmd := exec.Command("osascript", "-e", command)
	output, err := cmd.CombinedOutput()
	prettyOutput := strings.Replace(string(output), "\n", "", -1)

	// Ignore errors from the user hitting the cancel button
	if err != nil && strings.Index(string(output), "User canceled.") < 0 {
		return fmt.Errorf(err.Error() + ": " + prettyOutput + " (" + command + ")")
	}
	log.Println(prettyOutput)
	return nil
}
