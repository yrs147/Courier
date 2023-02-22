package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/spf13/viper"
	"github.com/trycourier/courier-go/v2"
	"golang.org/x/net/html"
)

func main() {
	// read the HTML file
	htmlFile, err := ioutil.ReadFile("example.html")
	if err != nil {
		log.Fatal(err)
	}

	// parse the HTML file
	doc, err := html.Parse(bytes.NewReader(htmlFile))
	if err != nil {
		log.Fatal(err)
	}

	// find all text nodes under the body tag
	var f func(*html.Node)
	var text string
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	var bodyNode *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)

	if bodyNode != nil {
		f(bodyNode)
		if text != "" {
			// save the extracted text to a file
			const inputFile = "./input.txt"
			err = os.WriteFile(inputFile, []byte(text), os.ModePerm)
			if err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}

			// read the text from the input file
			fileBytes, err := os.ReadFile(inputFile)
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}
			msgPrefix := "Summarize these terms of service in 10 points, without losing the context, and only show information which may be relevant to the user "
			msg := msgPrefix + string(fileBytes)

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
			err = client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
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

			// save the response to a file
			const outPutFile = "./output.txt"
			err = os.WriteFile(outPutFile, []byte(output), os.ModePerm)
			if err != nil {
				log.Fatalf("Failed to write file: %v", err)
			}

			// send the summarized terms to the user via email or push notification using the Courier API.
			// initialize the Courier client
			courclient := courier.CreateClient("Courier_KEY", nil)
			requestID, err := courclient.SendMessage(context.Background(), courier.SendMessageRequestBody{
				Message: map[string]interface{}{
					"to": map[string]string{
						"email": "example@gmail.com",
					},
					"content": map[string]string{
						"title": "Hey ! Here are your terms and conditions!",
						"body":  output,
					},
				},
			})

			if err != nil {
				log.Fatalln(err)
			}
			log.Println(requestID)

		}
	}
}
