package util

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	//Spotify url types
	SPOTIFYPLAYLISTURL = 1
	SPOTIFYALBUMURL    = 2
	SPOTIFYTRACKURL    = 3
	SPOTIFYUNKNOWNURL  = -1
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	ytUrlRegex = `(?m)^(http(s)??\:\/\/)?(www\.)?((youtube\.com\/watch\?v=)|(youtu.be\/))([a-zA-Z0-9\-_])+`

	//var ytPlaylistUrlRegex = `/^.*(youtu.be\/|list=)([^#\&\?]*).
	ytPlaylistUrlRegex = `^.*(youtube.com\/playlist\?list=)([^#\&\?]*)`

	//var ytPlaylistUrlRegex = `^https?:\/\/(www.youtube.com|youtube.com)\/playlist(.*)$`
	durationRegex = `P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`

	//http spotify urls regex for track, album, playlist ID's (matching group number is 3)
	spotifyHttpUrlRegex      = `(?m)^(https:\/\/open.spotify.com\/(playlist\/|album\/|track\/)([a-zA-Z0-9]+))(.*)$`
	spotifyHttpPlaylistRegex = `^(https:\/\/open.spotify.com\/playlist\/[[a-zA-Z0-9]{22}\?.*)$`
	spotifyHttpAlbumRegex    = `^(https:\/\/open.spotify.com\/album\/[[a-zA-Z0-9]{22}\?.*)$`
	spotifyHttpTrackRegex    = `^(https:\/\/open.spotify.com\/track\/[[a-zA-Z0-9]{22}\?.*)$`

	spotifyIDRegex = `[a-zA-Z0-9]{22}`

	//youtube url regex
	ytPlaylistUrlRegex = `^(https?:\/\/www.youtube.com\/watch\?v=[a-zA-Z0-9]{11}\&list=[a-zA-Z0-9_-]{34}.*)$`

	ytTrackUrlRegex = `^(https?:\/\/www.youtube.com\/watch\?v=[a-zA-Z0-9_-]{11})$`

	ytPlaylistIDRegex = `[a-zA-Z0-9_-]{34}`
	ytTrackIDRegex    = `[a-zA-Z0-9_-]{11}`
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func FormatVideoTitle(videoTitle string) string {
	newTitle := strings.Trim(videoTitle, " ")
	newTitle = strings.ReplaceAll(newTitle, " ", "_")

	return newTitle
}

//GetWorkingDirPath returns working path
func GetWorkingDirPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Error while getting working dir: %v", err)
	}
	return dir, nil
}

func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("Error while deleting %s file: %v", path, err)
	}
	return nil
}

//DeleteSoundAndCoverFile deletes given sound file and cover file.
func DeleteSoundAndCoverFile(soundFilePath, coverFilePath string) {
	err := DeleteFile(soundFilePath)
	if err != nil {
		log.Printf("Error while deleting sound file %s: %v\n", soundFilePath, err)
	} else {
		log.Printf("%s is deleted.", soundFilePath)
	}

	errCover := DeleteFile(coverFilePath)
	if errCover != nil {
		log.Printf("Error while deleting cover file %s: %v\n", coverFilePath, err)
	} else {
		log.Printf("%s is deleted.", coverFilePath)
	}
}

//IsSpotifyUrl checks given URL is a valid Spotify URL or not.
//If given URL is valid then returns true, otherwise false.
func IsSpotifyUrl(url string) bool {
	re := regexp.MustCompile(spotifyHttpUrlRegex)
	if re.MatchString(url) {
		return true
	}
	return false
}

/*
//IsYoutubeUrl check given is a valid Youtube URL or not.
//If given URL is valid then returns true, otherwise false.
func IsYoutubeUrl(url string) bool {
	re := regexp.MustCompile()
}
*/

//GetSpotifyID returns ID of playlist, album or track from the given URL.
//("https://open.spotify.com/track/<ID>?si=NoAgqqb6Sp2vV-1IBzzM-g")
func GetSpotifyID(url string) string {
	if IsSpotifyUrl(url) {
		reId := regexp.MustCompile(spotifyIDRegex)
		return reId.FindString(url)
	}
	return ""
}

//GetSpotifyUrlType returns the type of the given url in integer.
func GetSpotifyUrlType(url string) int {
	if strings.Contains(url, "playlist") {
		return SPOTIFYPLAYLISTURL
	} else if strings.Contains(url, "album") {
		return SPOTIFYALBUMURL
	} else if strings.Contains(url, "track") {
		return SPOTIFYTRACKURL
	} else {
		return SPOTIFYUNKNOWNURL
	}
}

//GetSpotifyPlaylistID returns playlist ID
//from the given url
//TODO: Error checking
func GetSpotifyPlaylistID(url string) string {
	//format 1: spotify:playlist:76tzi26o8O920CYAvVbeYO
	if strings.HasPrefix(url, "spotify") {
		return strings.Trim(strings.Split(url, ":")[2], " ")
	}

	//format 2: https://open.spotify.com/playlist/76tzi26o8O920CYAvVbeYO?si=WKrHWhGVQTSmF7GbeqI5sw
	if strings.HasPrefix(url, "https") {
		trimmed := strings.TrimPrefix(url, "https://open.spotify.com/")
		return strings.Trim(strings.Split(strings.Split(trimmed, "/")[1], "?")[0], " ")
	}

	return ""

	/*
		seperated := strings.Split(url, "?")[1]
		playlistID := strings.Trim(seperated, "si=")
		return playlistID
	*/
}

//GetYtVideoID trims video id from the given url.
func GetYtVideoID(url string) string {
	//format: https://www.youtube.com/watch?v=qT6XCvDUUsU
	return strings.TrimPrefix(url, "https://www.youtube.com/watch?v=")
}

//GetYoutubePlaylistID trims youtube playlist id from the url.
//Checks given url if it is not playlist url return empty string ("")
//otherwise trims playlist id and returns.
func GetYoutubePlaylistID(url string) string {
	if !ValidateYoutubePlaylistUrl(url) {
		return ""
	}
	return strings.TrimPrefix(url, "https://www.youtube.com/playlist?list=")
}

//GetCoverImage downloads album cover image from the
//given url and returns its path.
func GetCoverImage(coverUrl string) (string, error) {
	resp, err := http.Get(coverUrl)
	if err != nil {
		return "", fmt.Errorf("Error while getting cover image: %v", err)
	}
	defer resp.Body.Close()

	imgFileName := RandStringRunes(15) + ".jpg"

	imgFile, err := os.Create(imgFileName)
	if err != nil {
		return "", fmt.Errorf("Error while creating cover image file: %v", err)
	}
	defer imgFile.Close()

	_, err = io.Copy(imgFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error while getting cover image file: %v", err)
	}
	return imgFileName, nil
}

//ValiadateYoutubeUrl validates whether given url is a youtube url or not.
func ValidateYoutubeUrl(url string) bool {
	r, err := regexp.Compile(ytUrlRegex)
	if err != nil {
		log.Println(err)
		return false
	}
	return r.MatchString(url)
}

/*
//ValidateSpotifyTrackUrl validates whether given url is a Spotify track url or not.
func ValidateSpotifyTrackUrl(url string) bool {

}
*/

//ValidateYoutubePlaylistUrl validates whether given url is a youtube playlist url or not.
func ValidateYoutubePlaylistUrl(url string) bool {
	r, err := regexp.Compile(ytPlaylistUrlRegex)
	if err != nil {
		log.Println(err)
		return false
	}

	return r.MatchString(url)
}

//ParseISO8601 takes a duration in format ISO8601 and parses to
//MM:SS format.
func ParseISO8601(duration string) string {
	r, err := regexp.Compile(durationRegex)
	if err != nil {
		log.Println(err)
		return ""
	}

	matches := r.FindStringSubmatch(duration)

	years := parseInt64(matches[1])
	months := parseInt64(matches[2])
	days := parseInt64(matches[3])
	hours := parseInt64(matches[4])
	minutes := parseInt64(matches[5])
	seconds := parseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)

	return time.Duration(years*24*365*hour +
		months*30*24*hour + days*24*hour +
		hours*hour + minutes*minute + seconds*second).String()
}

func parseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}

	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}

	return int64(parsed)
}
