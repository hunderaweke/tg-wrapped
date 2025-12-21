package analyzer

import (
	"time"

	"github.com/gotd/td/tg"
)

func getDateTime(date int) time.Time {
	t := time.Unix(int64(date), 0)
	return t
}

func countNumOfReactions(reactions tg.MessageReactions) (map[string]int, int) {
	counter := make(map[string]int)
	// * Important: I am counting the custom reactions to total counter
	totalCount := 0
	for _, r := range reactions.Results {
		totalCount += r.Count
		switch r.Reaction.(type) {
		case *tg.ReactionEmoji:
			emojiReaction := r.Reaction.(*tg.ReactionEmoji)
			counter[emojiReaction.Emoticon] += r.Count
		default:
			continue
		}
	}
	return counter, totalCount
}
func mergeMaps(firstMap, secondMap map[string]int) map[string]int {
	for key, val := range secondMap {
		firstMap[key] += val
	}
	return firstMap
}
