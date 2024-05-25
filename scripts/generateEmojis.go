package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Emoji struct {
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
}

func main() {
	emojis := must(loadEmojiJson())

	if err := os.MkdirAll("../../internal/emoji", 0755); err != nil {
		panic(err)
	}

	out := must(os.Create("../../internal/emoji/emoji.go"))
	must(out.Write([]byte("package emoji\n\ntype Emoji struct {\n\tEmoji string\n\tDescription string\n}\n\nvar Emojis []Emoji = []Emoji{\n")))
	for _, emoji := range emojis {
		must(out.Write([]byte(fmt.Sprintf("\t{Emoji: \"%s\", Description: \"%s\"},\n", emoji.Emoji, emoji.Description))))
	}
	must(out.Write([]byte("}\n")))
}

func loadEmojiJson() ([]Emoji, error) {
	emojisRaw, err := os.ReadFile("../../emoji.json")
	if err != nil {
		return nil, err
	}

	var emojis []Emoji
	err = json.Unmarshal(emojisRaw, &emojis)
	if err != nil {
		return nil, err
	}

	return emojis, nil
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
