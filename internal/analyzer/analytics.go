package analyzer

import (
	"time"

	"github.com/gotd/td/tg"
)

type analytics struct {
	TotalViews              int              `json:"total_views,omitempty"`
	TotalComments           int              `json:"total_comments,omitempty"`
	TotalReactions          int              `json:"total_reactions,omitempty"`
	TotalPosts              int              `json:"total_posts,omitempty"`
	TotalForwarded          int              `json:"total_forwarded,omitempty"`
	MonthlyView             map[string]int   `json:"monthly_view,omitempty"`
	ReactionCounter         map[string]int   `json:"reaction_counter,omitempty"`
	PostCountPerday         map[string][]int `json:"post_count_perday,omitempty"`
	PostCountPerMonth       map[string]int   `json:"post_count_per_month,omitempty"`
	PopularPostID           int              `json:"popular_post_id,omitempty"`
	PopularPostViewCount    int              `json:"popular_post_view_count,omitempty"`
	PopularPostByCommentID  int              `json:"popular_post_by_comment_id,omitempty"`
	PopularPostCommentCount int              `json:"popular_post_comment_count,omitempty"`
	ForwardCount            map[int]int      `json:"forward_count,omitempty"`
}

func getDateTime(date int) time.Time {
	t := time.Unix(int64(date), 0)
	return t
}

func countNumOfReactions(reactions tg.MessageReactions) (map[string]int, int) {
	counter := make(map[string]int)
	// * Important: I am counting the custom reactions too
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
func NewAnalytics() analytics {
	var a analytics
	a.MonthlyView = make(map[string]int)
	a.ReactionCounter = make(map[string]int)
	a.PostCountPerday = make(map[string][]int)
	a.PostCountPerMonth = make(map[string]int)
	a.ForwardCount = make(map[int]int)
	return a
}
func (a *analytics) addDateCount(date time.Time) {
	month := date.Month().String()
	t := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := t.AddDate(0, 0, -1)
	if len(a.PostCountPerday[month]) == 0 {
		a.PostCountPerday[month] = make([]int, lastDay.Day())
	}
	a.PostCountPerday[month][date.Day()-1] += 1
}
