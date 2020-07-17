// Package akarifavo is a generate Akari favorite statement from the text.
package akarifavo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kechako/go-yahoo/da"
)

// A Akari represents a generator of Akari favorite statement.
type Akari struct {
	client *da.Client
}

// New returns a new *Akari.
func New(appID string) *Akari {
	return &Akari{
		client: da.New(appID),
	}
}

// Say returns a Akari favorite statement that generated from text.
// Returns a empty string if the text does not have a favorite.
func (a *Akari) Say(ctx context.Context, text string) (string, error) {
	favorite, err := a.findFavorite(ctx, text)
	if err != nil {
		return "", err
	}

	if favorite == "" {
		return "", nil
	}

	return fmt.Sprintf("わぁい%s あかり%s大好き", favorite, favorite), nil
}

// findFavorite returns a favorite that was found in the text.
// Returns a empty string if the text does not have a favorite.
func (a *Akari) findFavorite(ctx context.Context, text string) (string, error) {
	res, err := a.client.Parse(ctx, text)
	if err != nil {
		return "", fmt.Errorf("fail to parse the text: %w", err)
	}

	if len(res.Chunks) == 0 {
		return "", nil
	}

	var favoChunk da.Chunk
	var favoMorphemIndex int
	favoriteID := -1

	chunkMap := make(map[int][]da.Chunk)
	for _, chunk := range res.Chunks {
		chunkMap[chunk.Head] = append(chunkMap[chunk.Head], chunk)

		if favoriteID < 0 {
			for i, t := range chunk.Tokens {
				if isFavoriteToken(t) {
					favoChunk = chunk
					favoMorphemIndex = i
					favoriteID = chunk.ID
					break
				}
			}
		}
	}

	if favoriteID < 0 {
		return "", nil
	}

	deps, ok := chunkMap[favoriteID]
	if !ok {
		if favoMorphemIndex < 1 {
			return "", nil
		}

		var favorite string
		for i := favoMorphemIndex - 1; i >= 0; i-- {
			t := favoChunk.Tokens[i]
			if t.PartOfSpeech() == "助詞" {
				continue
			}

			favorite = t.Surface()
			break
		}

		return favorite, nil
	}

	var favorite string
loop:
	for _, dep := range deps {
		tlen := len(dep.Tokens)
		if tlen < 2 {
			continue
		}

		lastWord := dep.Tokens[tlen-1]
		if lastWord.PartOfSpeech() != "助詞" {
			continue
		}

		switch lastWord.Surface() {
		case "の":
			if dep.Tokens[tlen-2].PartOfSpeech() == "動詞" {
				favorite = dep.String()
				break loop
			}
		case "が", "も", "を":
			favorite = ""
			for i := 0; i < tlen-1; i++ {
				t := dep.Tokens[i]
				favorite += t.Surface()
			}
			break loop
		}
	}

	return favorite, nil
}

func isFavoriteToken(t da.Token) bool {
	s := t.Surface()
	pos := t.PartOfSpeech()

	return (strings.HasPrefix(s, "大好き") || strings.HasPrefix(s, "好き")) &&
		(pos == "形容詞" || pos == "形容動詞")
}
