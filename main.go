package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Result represents the input and result text in JSON format
type Result struct {
	Input  string `json:"input"`
	Result string `json:"result"`
}

func callOpenAI(input string) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()
	// prompt := "I want you to act as an English translator, spelling corrector and improver. I will speak to you in any language and you will detect the language, translate it and answer in the corrected and improved version of my text, in English. I want you to replace my simplified A0-level words and sentences with more appropriate, beautiful and elegant. Keep the meaning same. I want you to only reply the correction, the improvements and nothing else, do not write any explanations."
	prompt := "Correct and improve the following text"
	content := fmt.Sprintf("%s: ```%s```", prompt, input)
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 500,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {

		return "", err
	}
	defer stream.Close()

	result := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return "", err
		}

		fmt.Printf(response.Choices[0].Delta.Content)
		result += response.Choices[0].Delta.Content
	}
	fmt.Println()

	return strings.TrimSpace(strings.Trim(result, "\"")), err

}
func readInput() (string, error) {
	// msg := ""
	// survey.AskOne(&survey.Editor{Message: "Please press ENTER to start typing your text"}, &msg)
	tmpFile, err := ioutil.TempFile("", "text-*.txt")
	if err != nil {
		return "", err

	}
	defer os.Remove(tmpFile.Name())

	vimCmd := exec.Command("vim", tmpFile.Name())
	vimCmd.Stdin = os.Stdin
	vimCmd.Stdout = os.Stdout
	vimCmd.Stderr = os.Stderr

	if err := vimCmd.Run(); err != nil {
		return "", err

	}

	content, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), err
}

func writeResult(input, result string) error {
	home, err := GetUserHomeDir()
	if err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(home, ".ai-converter"), 0755); err != nil && !os.IsExist(err) {
		return err
	}

	outputPath := filepath.Join(home, ".ai-converter", "convert-result.json")
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	var results []Result
	if err := json.NewDecoder(file).Decode(&results); err != nil && err != io.EOF {
		return err
	}

	results = append(results, Result{Input: input, Result: result})
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputPath, resultJSON, 0644); err != nil {
		return err
	}
	return nil
}

func copyToClipboard(text string) error {
	// Copy text to the clipboard
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func GetUserHomeDir() (string, error) {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Get the home directory path
	homeDir := currentUser.HomeDir

	return homeDir, nil
}

func main() {

	input, err := readInput()
	if err != nil {
		log.Fatalln(err)
	}

	if input == "" {
		return
	}

	improvedText, err := callOpenAI(input)
	if err != nil {
		log.Fatalln(err)
	}

	if err := copyToClipboard(improvedText); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Copied to clipboard!")

	if err := writeResult(input, improvedText); err != nil {
		log.Fatalln(err)
	}
}
