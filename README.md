<!-- Output copied to clipboard! -->

<!-----

Yay, no errors, warnings, or alerts!

Conversion time: 0.401 seconds.


Using this Markdown file:

1. Paste this output into your source file.
2. See the notes and action items below regarding this conversion run.
3. Check the rendered output (headings, lists, code blocks, tables) for proper
   formatting and use a linkchecker before you publish this page.

Conversion notes:

* Docs to Markdown version 1.0β34
* Sat Apr 08 2023 03:21:35 GMT-0700 (PDT)
* Source doc: Untitled document
----->



# **AI-Converter**

AI-Converter is a command-line utility that allows users to interact with OpenAI's GPT-3.5 Turbo model, generate text based on custom prompts, and save the generated results to a file and clipboard. Users can define and execute multiple custom commands based on pre-defined prompts.


## **Table of Contents**



* [Installation](./README.md#installation)
* [Usage](./README.md#usage)
* [Features](./README.md#Features)
* [Configuration](./README.md#configuration)
* [Contributing](./README.md#contributing)
* [License](./README.md#license)


## **Features**

- Interactive command-line interface
- Customizable commands with pre-defined prompts
- Input text using a Vim-like editor
- Automatically saves results to a file
- Copies results to the system clipboard

## **Installation**


1. Install[ Go](https://golang.org/doc/install) if you haven't already.

2. Run `go install`

```bash
go install github.com/YanniHu1996/ai-converter
```

## **Usage**

To use AI-Converter, simply run the following command in your terminal:

bash


```
./ai-converter [command]
```


Replace `[command]` with the name of the custom command you want to execute. The utility will prompt you to enter text, which it will send to the OpenAI API. The generated response will be saved to a file and copied to your clipboard.


## **Configuration**

To create custom commands, edit the `commands.json` file in the `~/.ai-converter/` directory. Here's an example configuration:

json


```
[
  {
    "id": 1,
    "name": "improver",
    "prompt": "Correct and improve the following text"
  },
  {
    "id": 2,
    "name": "summarizer",
    "prompt": "Summarize the following text"
  }
]
```


Each command object should have an `id`, `name`, and `prompt`. The `name` is used to execute the command, while the `prompt` is sent to the OpenAI API as part of the input.


## **Contributing**

Contributions are welcome! Please follow these steps to contribute:



1. Fork the repository.
2. Create a new branch with a descriptive name.
3. Make your changes and commit them with clear and concise commit messages.
4. Push your changes to your fork.
5. Open a pull request, describing the changes you made and their purpose.

Please ensure your code follows the project's style guidelines and passes all tests.


## **License**

AI-Converter is licensed under the[ MIT License](https://choosealicense.com/licenses/mit/). See `LICENSE` for more information.
