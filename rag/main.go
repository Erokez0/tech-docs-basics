package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"google.golang.org/genai"
)

const ( 
	GENAI_API_KEY = "key"

	OPENAI_BASE_URL = "https://api.deepseek.com"
	OPENAI_API_KEY = "key"
	OPENAI_MODEL = "deepseek-chat"
) 

var ctx = context.Background()
var HeaderRegexp = regexp.MustCompile("#+ ")

func embed(contents []*genai.Content) [][]float32 {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: GENAI_API_KEY, Backend: genai.BackendGeminiAPI})
	if err != nil {
		log.Fatal(err)
	}


	result, err := client.Models.EmbedContent(ctx, "gemini-embedding-001", contents, nil)
	if err != nil {
		log.Fatal(err)
	}

	vectors := make([][]float32, len(result.Embeddings))
	
	for i, embedding := range result.Embeddings {
		vectors[i] = embedding.Values
	}

	return vectors
}

func divideMarkdownByHeadings(path string) []string {
	binaryData, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	blocks := HeaderRegexp.Split(string(binaryData), -1)[1:]

	return blocks
}

func main() {

	chatClient := openai.NewClient(
		option.WithAPIKey(OPENAI_API_KEY),
		option.WithBaseURL(OPENAI_BASE_URL),
	)

	blocks := divideMarkdownByHeadings("../Game.md")

	fmt.Println("--- VECTORIZING BLOCKS ---")
	
	contents := make([]*genai.Content, len(blocks))
	for ix, block := range blocks {
		content := genai.NewContentFromText(block, genai.RoleUser)
		contents[ix] = content
	}
	vectors := embed(contents)


	err := os.WriteFile("../vectors.npz", fmt.Append(nil, vectors), 0766)
	if err != nil {
		log.Fatal(err)
	}


	completion, err := chatClient.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model: OPENAI_MODEL,
		Messages: make([]openai.ChatCompletionMessageParamUnion, 0),
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Ответ: %v", completion.Choices[0].Message.Content)
}