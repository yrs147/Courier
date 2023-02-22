package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/spf13/viper"
	"github.com/trycourier/courier-go/v2"
)

type ApiResponse struct {
	Output string `json:"output"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "in.html")
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			searchQuery := r.FormValue("search_query")
			msgPrefix := "Give the summary of Privacy Policy also focus on contact information , app permisions , your personal data , and other aspects in 10 points without losing context  of website "
			msg := msgPrefix + searchQuery
			// call the OpenAI API
			viper.SetConfigFile(".env")
			viper.ReadInConfig()
			apiKey := viper.GetString("API_KEY")
			if apiKey == "" {
				panic("API KEY Missing")
			}
			ctx := context.Background()
			client := gpt3.NewClient(apiKey)

			outputBuilder := strings.Builder{}
			err := client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
				Prompt: []string{
					msg,
				},
				MaxTokens:   gpt3.IntPtr(2000),
				Temperature: gpt3.Float32Ptr(0),
			}, func(resp *gpt3.CompletionResponse) {
				outputBuilder.WriteString(resp.Choices[0].Text)
			})
			if err != nil {
				log.Fatal(err)
			}
			output := strings.TrimSpace(outputBuilder.String())

			// send the summarized terms to the user via JSON response
			apiResponse := ApiResponse{Output: output}
			jsonResponse, err := json.Marshal(apiResponse)
			if err != nil {
				log.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResponse)
			fmt.Println(searchQuery)

			// send the summarized terms to the user via email or push notification using the Courier API.
			// initialize the Courier client
			courclient := courier.CreateClient("Courier_KEY", nil)
			requestID, err := courclient.SendMessage(context.Background(), courier.SendMessageRequestBody{
				Message: map[string]interface{}{
					"to": map[string]string{
						"email": "example@gmail.com",
					},
					"content": map[string]string{
						"title": "Hey ! Here are your terms and conditions for " + searchQuery,
						"body":  output,
					},
				},
			})

			if err != nil {
				log.Fatalln(err)
			}
			log.Println(requestID)
		}
	})

	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
