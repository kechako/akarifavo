package akarifavo

import (
	"context"
	"fmt"

	da "github.com/kechako/go-yahoo-da"
)

// A Akari represents a generator of Akari favorite statement.
type Akari struct {
	client *da.Client
}

// New returns a new *Akari.
func New(appID string) *Akari {
	return &Akari{
		client: da.NewClient(appID),
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

	if len(res.Results) == 0 || len(res.Results[0].Chunks) == 0 {
		return "", nil
	}

	var favoChunk da.Chunk
	var favoMorphemIndex int
	favoriteID := -1

	chunkMap := make(map[int][]da.Chunk)
	for _, chunk := range res.Results[0].Chunks {
		chunkMap[chunk.Dependency] = append(chunkMap[chunk.Dependency], chunk)

		if favoriteID < 0 {
			for i, m := range chunk.Morphemes {
				if (m.Surface == "大好き" || m.Surface == "好き") && m.POS == "形容動詞" {
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
			m := favoChunk.Morphemes[i]
			if m.POS == "助詞" {
				continue
			}

			favorite = m.Surface
			break
		}

		return favorite, nil
	}

	var favorite string
loop:
	for _, dep := range deps {
		mlen := len(dep.Morphemes)
		if mlen < 2 {
			continue
		}

		lastWord := dep.Morphemes[mlen-1]
		if lastWord.POS != "助詞" {
			continue
		}

		switch lastWord.Surface {
		case "の":
			if dep.Morphemes[mlen-2].POS == "動詞" {
				favorite = dep.String()
				break loop
			}
		case "が", "も", "を":
			favorite = ""
			for i := 0; i < mlen-1; i++ {
				m := dep.Morphemes[i]
				favorite += m.Surface
			}
			break loop
		}
	}

	return favorite, nil
}
