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
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func main() {
	// init
	var (
		inputReader    InputReader     = InputReadFunc(vimReader)
		chatter        Chatter         = ChatFunc(callOpenAI)
		resultHandlers []ResultHandler = []ResultHandler{
			copyToClipboard,
			writeResult,
		}

		root = cobra.Command{
			Use: "ai-converter",
		}
	)

	cmds, err := loadCommands()
	if err != nil {
		log.Fatalln(err)
	}

	for _, cmd := range cmds {
		root.AddCommand(&cobra.Command{
			Use:   cmd.Name,
			Short: fmt.Sprintf("the prompt is %q", cmd.Prompt),
			RunE: func(_ *cobra.Command, args []string) error {
				return loopUntilSignalReceived(func() error {
					input, err := inputReader.Read()
					if err != nil {
						return err
					}

					if input == "" {
						return nil
					}

					content := fmt.Sprintf("%s: ```%s```", cmd.Prompt, input)
					response, err := chatter.Chat(content)
					if err != nil {
						return err
					}

					res := Result{Input: input, Result: response, CommandID: cmd.ID}
					for _, h := range resultHandlers {
						if err := h(res); err != nil {
							log.Println(err)
						}
					}

					return nil
				})
			},
		})
	}

	if err := root.Execute(); err != nil {
		log.Fatalln(err)
	}
}

// Result represents the input and result text in JSON format
type Result struct {
	Input     string `json:"input"`
	Result    string `json:"result"`
	CommandID int
}

type ResultHandler func(Result) error

type Command struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

func loadCommands() ([]Command, error) {
	// Get the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Construct the path to the commands.json file
	path := filepath.Join(home, ".ai-converter", "commands.json")

	// Create the file if it does not exist
	err = createFileIfNotExist(path)
	if err != nil {
		return nil, err
	}

	// Read the contents of the JSON file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON content into a slice of Command structs
	var commands []Command
	err = json.Unmarshal(content, &commands)
	if err != nil {
		return nil, err
	}

	// append built-in command
	if len(commands) == 0 {
		commands = append(commands, Command{
			ID:     1,
			Name:   "improver",
			Prompt: "Correct and improve the following text",
		})
	}

	return commands, nil
}

func createFileIfNotExist(filePath string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create the directory for the file if it does not exist
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return err
		}

		// Create the file with default content
		defaultContent := []byte("[]")
		err = ioutil.WriteFile(filePath, defaultContent, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

type Chatter interface {
	Chat(content string) (string, error)
}

type ChatFunc func(string) (string, error)

func (f ChatFunc) Chat(content string) (string, error) {
	return f(content)
}

func callOpenAI(content string) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()
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

type InputReader interface {
	Read() (string, error)
}

type InputReadFunc func() (string, error)

func (f InputReadFunc) Read() (string, error) {
	return f()
}

func vimReader() (string, error) {
	msg := ""
	if err := survey.AskOne(&survey.Editor{Message: "Please press ENTER to start typing your text"}, &msg); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(msg)), nil
}

func writeResult(r Result) error {
	home, err := GetUserHomeDir()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(home, ".ai-converter", fmt.Sprintf("%v-convert-results.json", r.CommandID))
	if err := createFileIfNotExist(outputPath); err != nil {

	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return err
	}

	var results []Result
	if err := json.Unmarshal(data, &results); err != nil {
		return err
	}
	for _, res := range results {
		res.CommandID = r.CommandID
	}

	results = append(results, r)
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputPath, resultJSON, 0644); err != nil {
		return err
	}
	return nil
}

func copyToClipboard(r Result) error {
	// Copy text to the clipboard
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(r.Result)
	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Copied to clipboard!")
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

func loopUntilSignalReceived(f func() error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigCh:
			return nil
		default:
			if err := f(); err != nil {
				return err
			}
		}
	}

}
