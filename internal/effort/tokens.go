package effort

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

const tokenizerName = "cl100k_base"

type cl100k struct{ enc *tiktoken.Tiktoken }

// NewCL100K returns the cl100k_base tokenizer, loaded from the vocabulary
// embedded in the pinned tiktoken-go-loader module, so a count never depends
// on the network and is the same on every machine. It is a fixed, published
// yardstick for comparing apps, not a measure of any particular model's cost.
func NewCL100K() (Tokenizer, error) {
	tiktoken.SetBpeLoader(tiktokenloader.NewOfflineLoader())
	enc, err := tiktoken.GetEncoding(tokenizerName)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", tokenizerName, err)
	}
	return cl100k{enc}, nil
}

func (c cl100k) Count(text string) int { return len(c.enc.Encode(text, nil, nil)) }
func (cl100k) Name() string            { return tokenizerName }
