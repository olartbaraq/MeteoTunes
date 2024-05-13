package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

var wg *sync.WaitGroup

type WeatherParams struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

type Weather struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type Current struct {
	Temperature float64   `json:"temp"`
	Pressure    float64   `json:"pressure"`
	Humidity    float64   `json:"humidity"`
	Weather     []Weather `json:"weather"`
}

type WeatherResponse struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Timezone  string  `json:"timezone"`
	Current   Current `json:"current"`
}

type ImageData struct {
	Data []ImageInfo `json:"data"`
}

type ImageInfo struct {
	AssetID  string `json:"asset_id"`
	AssetURL string `json:"asset_url"`
	Type     string `json:"type"`
	Width    int32  `json:"width"`
	Height   int32  `json:"height"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"message": "welcome to meteotunes"})
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"message": "welcome to meteojjbjtunes"})
}

func (f *Server) fetchMusic(w http.ResponseWriter, r *http.Request) {
	wg = &sync.WaitGroup{}

	// 1) talk to weather api

	// 1) Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"Error":   err.Error(),
			"message": "error trying to read request body",
		})
		return
	}
	defer r.Body.Close()

	// 2) Unmarshal the request body into your struct
	var params WeatherParams
	err = json.Unmarshal(body, &params)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"Error":   err.Error(),
			"message": "error unmarshalling to weather params",
		})
		return
	}

	// 3) Use the params struct
	jsonStr := &params

	validate := validator.New()

	if err := validate.Struct(jsonStr); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return
		}

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.Field())
			fmt.Println(err.Tag())
			fmt.Println()
		}
	}

	weatherUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%v&lon=%v&appid=%v&units=metric", jsonStr.Latitude, jsonStr.Longitude, f.config.OPEN_WEATHER_KEY)

	resp, err := http.Get(weatherUrl)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"Error":   err.Error(),
			"message": "error making request to open weather API",
		})
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "weather API returned non-OK status",
		})
		return
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"Error":   err.Error(),
			"message": "error reading response body",
		})
		return
	}

	var weatherInfo WeatherResponse
	if err := json.Unmarshal(body, &weatherInfo); err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"Error":   err.Error(),
			"message": "error unmarshalling JSON response",
		})
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusPartialContent)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "data loaded successfully",
		"data":    weatherInfo,
	})

	// 2) goroutine to talk to limewire api and claude3 apis

	imageInfoChan := make(chan *ImageData, 1)

	wg.Add(1)
	go func(u chan<- *ImageData) {
		location := strings.Split(weatherInfo.Timezone, "/")
		prompt := fmt.Sprintf("It's a beautiful day with temperature %v , humidity %v and pressure %v, with nice %v and sometimes ; nicely of %v in the city %v in %v", weatherInfo.Current.Temperature, weatherInfo.Current.Humidity, weatherInfo.Current.Pressure, weatherInfo.Current.Weather[0].Main, weatherInfo.Current.Weather[0].Description, location[1], location[0])
		fmt.Println(prompt)

		reqUrl := "https://api.limewire.com/api/image/generation"

		data := []byte(fmt.Sprintf(`{ "prompt": "%s", "aspect_ratio": "1:1" }`, prompt))

		req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(data))
		if err != nil {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{
				"Error":   err.Error(),
				"message": "error making request to limewire API",
			})
			return
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Api-Version", "v1")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", f.config.LIME_WIRE_KEY))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{
				"Error":   err.Error(),
				"message": "error getting response from limewire API",
			})
			return
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)

		if err != nil {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"Error":   err.Error(),
				"message": "error reading response body",
			})
			return
		}

		fmt.Println(res)
		var imageInfo ImageData

		if err := json.Unmarshal(body, &imageInfo); err != nil {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"Error":   err.Error(),
				"message": "error unmarshalling imageJSON response",
			})
			return
		}

		u <- &imageInfo
		wg.Done()

	}(imageInfoChan)

	select {
	case imageInfo := <-imageInfoChan:
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "data loaded successfully from limewire",
			"data":    imageInfo,
		})

	case <-time.After(15 * time.Second):
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "timeout fetching image from limewire",
		})
		return
	}

	wg.Wait()
	close(imageInfoChan)

	// 3) feed LLM's prompt to music API
}
