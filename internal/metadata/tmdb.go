package metadata

import (
"context"
"encoding/json"
"fmt"
"net/http"
"net/url"
"time"
)

const tmdbBaseURL = "https://api.themoviedb.org/3"

type Client struct {
apiKey string
http   *http.Client
}

func NewClient(apiKey string) *Client {
return &Client{apiKey: apiKey, http: &http.Client{Timeout: 15 * time.Second}}
}

type TMDBMovie struct {
ID               int     `json:"id"`
Title            string  `json:"title"`
OriginalTitle    string  `json:"original_title"`
Overview         string  `json:"overview"`
ReleaseDate      string  `json:"release_date"`
PosterPath       string  `json:"poster_path"`
BackdropPath     string  `json:"backdrop_path"`
VoteAverage      float64 `json:"vote_average"`
Popularity       float64 `json:"popularity"`
GenreIDs         []int   `json:"genre_ids"`
OriginalLanguage string  `json:"original_language"`
}

type TMDBMovieDetail struct {
TMDBMovie
ImdbID              string          `json:"imdb_id"`
Runtime             int             `json:"runtime"`
Status              string          `json:"status"`
Tagline             string          `json:"tagline"`
Budget              int64           `json:"budget"`
Revenue             int64           `json:"revenue"`
Genres              []TMDBGenre     `json:"genres"`
ProductionCompanies []TMDBCompany   `json:"production_companies"`
BelongsToCollection *TMDBCollection `json:"belongs_to_collection"`
Images              *TMDBImages     `json:"images,omitempty"`
Videos              *TMDBVideos     `json:"videos,omitempty"`
}

type TMDBGenre      struct{ ID int `json:"id"`; Name string `json:"name"` }
type TMDBCompany    struct{ ID int `json:"id"`; Name string `json:"name"` }
type TMDBCollection struct{ ID int `json:"id"`; Name string `json:"name"`; PosterPath string `json:"poster_path"` }

type TMDBImages struct {
Backdrops []TMDBImage `json:"backdrops"`
Posters   []TMDBImage `json:"posters"`
}

type TMDBImage struct {
FilePath    string  `json:"file_path"`
Width       int     `json:"width"`
Height      int     `json:"height"`
VoteAverage float64 `json:"vote_average"`
}

type TMDBVideos struct {
Results []TMDBVideo `json:"results"`
}

type TMDBVideo struct {
Key      string `json:"key"`
Site     string `json:"site"`
Type     string `json:"type"`
Official bool   `json:"official"`
}

func (c *Client) SearchMovies(ctx context.Context, query string, year int) ([]TMDBMovie, error) {
params := url.Values{
"api_key":       {c.apiKey},
"query":         {query},
"include_adult": {"false"},
"language":      {"en-US"},
"page":          {"1"},
}
if year > 0 {
params.Set("year", fmt.Sprintf("%d", year))
}
var result struct {
Results []TMDBMovie `json:"results"`
}
return result.Results, c.get(ctx, "/search/movie", params, &result)
}

func (c *Client) GetMovie(ctx context.Context, tmdbID int) (*TMDBMovieDetail, error) {
params := url.Values{
"api_key":            {c.apiKey},
"language":           {"en-US"},
"append_to_response": {"images,videos"},
"include_image_language": {"en,null"},
}
var detail TMDBMovieDetail
return &detail, c.get(ctx, fmt.Sprintf("/movie/%d", tmdbID), params, &detail)
}

func (c *Client) FindByIMDB(ctx context.Context, imdbID string) (*TMDBMovieDetail, error) {
params := url.Values{
"api_key":         {c.apiKey},
"external_source": {"imdb_id"},
}
var raw struct {
MovieResults []TMDBMovie `json:"movie_results"`
}
if err := c.get(ctx, fmt.Sprintf("/find/%s", imdbID), params, &raw); err != nil {
return nil, err
}
if len(raw.MovieResults) == 0 {
return nil, fmt.Errorf("no results for imdb_id %s", imdbID)
}
return c.GetMovie(ctx, raw.MovieResults[0].ID)
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
u := fmt.Sprintf("%s%s?%s", tmdbBaseURL, path, params.Encode())
req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
if err != nil {
return err
}
req.Header.Set("Accept", "application/json")
resp, err := c.http.Do(req)
if err != nil {
return fmt.Errorf("tmdb: %w", err)
}
defer resp.Body.Close()
if resp.StatusCode == http.StatusNotFound {
return fmt.Errorf("tmdb: not found")
}
if resp.StatusCode != http.StatusOK {
return fmt.Errorf("tmdb: status %d", resp.StatusCode)
}
return json.NewDecoder(resp.Body).Decode(out)
}
