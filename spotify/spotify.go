package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/m0thm4n/Dusty/util"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const (
	DEFAULTCOVERURL string = "https://github.com/golang/go/blob/master/doc/gopher/fiveyears.jpg"
)

type SpotifyAPI struct {
	ClientID       string
	ClientSecretID string
	BearerToken    string
}

type SpotifyPlaylistTracks struct {
	Items []struct {
		Track struct {
			Album struct {
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Name   string `json:"name"`
				Images []struct {
					Url string `json:"url"`
				} `json:"images"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Name string `json:"name"`
		} `json:"track"`
	} `json:"items"`
	Limit    int         `json:"limit"`
	Next     string      `json:"next"`
	Offset   int         `json:"offset"`
	Previous interface{} `json:"previous"`
	Total    int         `json:"total"`
}

type SpotifyAlbumTracks struct {
	Name   string `json:"name"` //album name
	Images []struct {
		Url string `json:"url"` //album cover url
	} `json:"images"`
	Tracks struct {
		Items []struct {
			Name    string `json:"name"` //track name
			Artists []struct {
				Name string `json:"name"` //track artist name
			} `json:"Artists"`
		} `json:"items"`
	} `json:"tracks"`
}

type SpotifySingleTrack struct {
	Name   string `json:"name"` //track name
	Images []struct {
		Url string `json:"url"` //track cover url
	} `json:"images"`
	Artists []struct {
		Name string `json:"name"` //track artist name
	}
}

type SpotifyPlaylistInfo struct {
	Name  string `json:"name"`
	Owner struct {
		DisplayName string `json:"display_name"`
		ID          string `json:"id"`
	} `json:"owner"`
}

type SpotifyPlaylist struct {
	TrackName   string
	CoverUrl    string
	ArtistNames string
}

//!!!! Spotify api endpoints requires bearer token, clientID and
//clientSecretID needed only for acquiring the bearer token.
//endpoint functions could be call by a different struct that
//contains only acquired bearer token.
func NewSpotifyAPI(clientID, clientSecretID string) *SpotifyAPI {
	spoAPI := &SpotifyAPI{
		ClientID:       clientID,
		ClientSecretID: clientSecretID,
	}

	bearerToken, err := spoAPI.getAPIToken()
	if err != nil {
		log.Printf("Error while getting Spotify OAUTH Token: %v", err)
	}

	return &SpotifyAPI{
		ClientID:       clientID,
		ClientSecretID: clientSecretID,
		BearerToken:    bearerToken.AccessToken,
	}
}

//getAPIToken returns Spotify oauth token
func (s *SpotifyAPI) getAPIToken() (*oauth2.Token, error) {
	ctx := context.Background()
	conf := &clientcredentials.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecretID,
		TokenURL:     "https://accounts.spotify.com/api/token",
	}

	tok, err := conf.Token(ctx)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func (s *SpotifyAPI) do(method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", "Bearer "+s.BearerToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//GetPlaylistInfo wrapper function of getPlaylistInfo function.
func (s *SpotifyAPI) GetPlaylistInfo(id string) (*SpotifyPlaylistInfo, error) {
	spotifyPlaylistInfo, err := s.getPlaylistInfo(id)
	if err != nil {
		return nil, err
	}
	return spotifyPlaylistInfo, nil
}

//GetPlaylistInfo sends request to get information about playlist to
//Spotify API and decodes API response to SpotifyPlaylistInfo struct.
func (s *SpotifyAPI) getPlaylistInfo(id string) (*SpotifyPlaylistInfo, error) {
	url := "https://api.spotify.com/v1/playlists/" + id

	resp, err := s.do("GET", url)
	if err != nil {
		return nil, fmt.Errorf("Error while getting playlist info: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var spotifyPl SpotifyPlaylistInfo
	err = decoder.Decode(&spotifyPl)
	if err != nil {
		return nil, err
	}

	return &spotifyPl, nil
}

func (s *SpotifyAPI) GetSpotifyPlaylist(id string, urlType int) ([]SpotifyPlaylist, error) {
	switch urlType {
	case util.SPOTIFYPLAYLISTURL:
		playlist, err := s.HandlePlaylist(id)
		if err != nil {
			return nil, err
		}
		return playlist, nil
	case util.SPOTIFYALBUMURL:
		playlist, err := s.HandleAlbum(id)
		if err != nil {
			return nil, err
		}
		return playlist, nil
	case util.SPOTIFYTRACKURL:
		playlist, err := s.HandleTrack(id)
		if err != nil {
			return nil, err
		}
		return playlist, nil
	}
	return nil, nil
}

//HandlePlaylist eliminates required information to create a download queue in
//the bot package from the getPlaylistTracks function response  and creates a slice
//of SpotifyPlaylist struct.
func (s *SpotifyAPI) HandlePlaylist(id string) ([]SpotifyPlaylist, error) {
	plTracks, err := s.getPlaylistTracks(id)
	if err != nil {
		return nil, err
	}

	playlist := []SpotifyPlaylist{}

	items := plTracks.Items

	for i, value := range items {
		if i > 20 {
			break
		}

		trackName := value.Track.Name

		coverUrl := ""
		album := value.Track.Album
		if album.Images[1].Url != "" {
			coverUrl = album.Images[1].Url
		} else {
			coverUrl = DEFAULTCOVERURL
		}

		artistNames := ""
		artists := value.Track.Artists
		for artistIndex := range artists {
			artistNames += artists[artistIndex].Name + " "
		}

		spotifyPlaylist := SpotifyPlaylist{
			TrackName:   trackName,
			CoverUrl:    coverUrl,
			ArtistNames: artistNames,
		}

		playlist = append(playlist, spotifyPlaylist)
	}
	return playlist, nil
}

//getPlaylistTracks sends request to get information about playlist's tracks
//to Spotify API and decodes API response to SpotifyPlaylistTracks struct.
func (s *SpotifyAPI) getPlaylistTracks(id string) (*SpotifyPlaylistTracks, error) {
	url := "https://api.spotify.com/v1/playlists/" + id + "/tracks"
	resp, err := s.do("GET", url)
	if err != nil {
		return nil, fmt.Errorf("Error while getting response from playlist tracks endpoint: %v", err)
	}

	// spotify playlist
	decoder := json.NewDecoder(resp.Body)

	var spotifyPlTracks SpotifyPlaylistTracks
	err = decoder.Decode(&spotifyPlTracks)
	if err != nil {
		return nil, fmt.Errorf("Error while getting playlist track info: %v", err)
	}

	return &spotifyPlTracks, nil
}

//HandleAlbum eliminates required information to create a download queue in
//the bot package from the getAlbumTracks function response  and creates a slice
//of SpotifyPlaylist struct.
func (s *SpotifyAPI) HandleAlbum(id string) ([]SpotifyPlaylist, error) {
	albumTracks, err := s.getAlbumTracks(id)
	if err != nil {
		return nil, err
	}

	playlist := []SpotifyPlaylist{}

	//this is album playlist so track covers will be
	//same for all the tracks.
	coverUrl := albumTracks.Images[1].Url

	items := albumTracks.Tracks.Items
	for _, value := range items {
		//get track name
		trackName := value.Name

		//get artist name
		artistNames := ""
		artists := value.Artists
		for artistIndex := range artists {
			artistNames += artists[artistIndex].Name + " "
		}

		spotifyPlaylist := SpotifyPlaylist{
			TrackName:   trackName,
			CoverUrl:    coverUrl,
			ArtistNames: artistNames,
		}

		playlist = append(playlist, spotifyPlaylist)
	}
	return playlist, nil
}

//getAlbumTracks sends request to get information about album's tracks to
//Spotify API and decodes API Response to SpotifyAlbumTracks struct.
func (s *SpotifyAPI) getAlbumTracks(id string) (*SpotifyAlbumTracks, error) {
	url := "https://api.spotify.com/v1/albums/" + id

	resp, err := s.do("GET", url)
	if err != nil {
		return nil, fmt.Errorf("Error while getting album tracks info: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var spotifyAlbumTracks SpotifyAlbumTracks
	err = decoder.Decode(&spotifyAlbumTracks)
	if err != nil {
		return nil, err
	}

	return &spotifyAlbumTracks, nil
}

//HandleTrack eliminates required information to create a download queue in
//the bot package from the getTrack function response  and creates a slice
//of SpotifyPlaylist struct.
func (s *SpotifyAPI) HandleTrack(id string) ([]SpotifyPlaylist, error) {
	track, err := s.getTrack(id)
	if err != nil {
		return nil, err
	}

	playlist := []SpotifyPlaylist{}

	artistNames := ""
	for _, value := range track.Artists {
		artistNames += value.Name + " "
	}

	spotifyPlaylist := SpotifyPlaylist{
		TrackName:   track.Name,
		CoverUrl:    track.Images[1].Url,
		ArtistNames: artistNames,
	}
	playlist = append(playlist, spotifyPlaylist)
	return playlist, nil
}

//getTrack sends request to get information about track to
//Spotify API and decodes API Response to SpotifySingleTrack struct.
func (s *SpotifyAPI) getTrack(id string) (*SpotifySingleTrack, error) {
	url := "https://api.spotify.com/v1/tracks/" + id

	resp, err := s.do("GET", url)
	if err != nil {
		return nil, fmt.Errorf("Error while getting track info: %v", err)
	}

	decoder := json.NewDecoder(resp.Body)
	var spotifySingleTrack SpotifySingleTrack
	err = decoder.Decode(&spotifySingleTrack)
	if err != nil {
		return nil, err
	}

	return &spotifySingleTrack, nil
}

//func completeAuth(w http.ResponseWriter, r *http.Request) {
//	token, err := auth.Token(state, r)
//	if err != nil {
//		http.Error(w, "Couldn't get token", http.StatusForbidden)
//		log.Fatalln(err)
//	}
//
//	if st := r.FormValue("state"); st != state {
//		http.NotFound(w, r)
//		log.Fatalf("State mismatch: %s != %s\n", st, state)
//	}
//
//	// Use the token to get an authenticated client
//	client := auth.NewClient(token)
//	fmt.Fprintf(w, "Login Completed!")
//	ch <- &client
//}
//
//// Auth authenticates with Spotify and refreshes the token
//func spotifyAuth(s *discordgo.Session, m *discordgo.Message) *spotify.Client {
//	fmt.Println(s.ClientID, s.ClientSecretID)
//
//	if s.ClientID == "" || s.ClientSecretID == "" {
//		fmt.Println("Please configure your Spotify client ID and secret in the config file at C:\\workspace\\go\\src\\spotify-bot\\")
//		os.Exit(1)
//	}
//
//	// shouldRefresh, err := cmd.Flags().GetBool("refresh")
//	// if err != nil {
//	// 	log.Fatalln(err)
//	// }
//
//	fmt.Println("Getting token...")
//	auth.SetAuthInfo(s.ClientID, s.ClientSecretID)
//	http.HandleFunc("/callback", completeAuth)
//	go http.ListenAndServe(":8888", nil)
//	url := auth.AuthURL(state)
//	fmt.Println("Please log in to Spotify by clicking the following link:", url)
//	_, _ = s.ChannelMessageSend(m.ChannelID, "Please log in to Spotify by clicking the following link: "+url)
//	//wait for auth to finish
//	client := <-ch
//	user, err := client.CurrentUser()
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	// conf.Token = *token
//	// marshalToken, err := json.Marshal(conf.Token)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	// viper.Set("auth", string(marshalToken))
//	// if err := viper.WriteConfigAs(cfgFile); err != nil {
//	// 	glog.Fatal("Error writing config:", err)
//	// }
//	fmt.Println("Login successful as", user.ID)
//	_, _ = s.ChannelMessageSend(m.ChannelID, "Login successful as "+user.ID)
//
//	return client
//}



func GetUsersPlaylists(userID string, sess *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()

	//if userID == "" {
	//	fmt.Fprintf(os.Stderr, "Error: missing user ID\n")
	//	flag.Usage()
	//	return nil, err
	//}

	config := &clientcredentials.Config {
		ClientID: SpotifyAPI{}.ClientID,
		ClientSecret: SpotifyAPI{}.ClientID,
		TokenURL: spotifyauth.TokenURL,
	}

	token, err := config.Token(context.Background())
	if err != nil {
		log.Fatalf("Couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)

	client := spotify.New(httpClient)

	user, err := client.GetUsersPublicProfile(ctx, spotify.ID(userID))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("User ID:", user.ID)
	fmt.Println("Display name:", user.DisplayName)
	fmt.Println("Spotify URI:", string(user.URI))
	fmt.Println("Endpoint:", user.Endpoint)
	fmt.Println("Followers:", user.Followers.Count)
}