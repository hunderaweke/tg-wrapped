package analyzer

import (
	"time"

	"github.com/gotd/td/tg"
)

type Analytics struct {
	ChannelName             string           `json:"channel_name,omitempty"`
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
	LongestStreak           int              `json:"longest_streak,omitempty"`
	visited                 map[int]struct{}
}

func NewAnalytics(name string) Analytics {
	var a Analytics
	a.ChannelName = name
	a.MonthlyView = make(map[string]int)
	a.ReactionCounter = make(map[string]int)
	a.PostCountPerday = make(map[string][]int)
	a.PostCountPerMonth = make(map[string]int)
	a.ForwardCount = make(map[int]int)
	a.visited = make(map[int]struct{})
	return a
}
func (a *Analytics) addDateCount(date time.Time) {
	month := date.Month().String()
	t := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := t.AddDate(0, 0, -1)
	if len(a.PostCountPerday[month]) == 0 {
		a.PostCountPerday[month] = make([]int, lastDay.Day())
	}
	a.PostCountPerday[month][date.Day()-1] += 1
}
func (a *Analytics) updateFromChannelMessages(m *tg.MessagesChannelMessages) int {
	offSet := 0
	last := &tg.Message{}
	for _, msg := range m.Messages {
		if mm, ok := msg.(*tg.Message); ok {
			if _, ok := a.visited[mm.ID]; ok {
				continue
			}
			if a.PopularPostID == 0 || a.PopularPostViewCount < mm.Views {
				a.PopularPostID = mm.ID
				a.PopularPostViewCount = mm.Views
			}
			if a.PopularPostByCommentID == 0 || a.PopularPostCommentCount < mm.Replies.Replies {
				a.PopularPostByCommentID = mm.ID
				a.PopularPostCommentCount = mm.Replies.Replies
			}
			offSet = mm.Date
			a.TotalViews += mm.Views
			a.TotalComments += mm.Replies.Replies
			t := getDateTime(mm.Date)
			a.MonthlyView[t.Month().String()] += mm.Views
			reactionCounter, totalReactions := (countNumOfReactions(mm.Reactions))
			a.ReactionCounter = mergeMaps(a.ReactionCounter, reactionCounter)
			a.TotalReactions += totalReactions
			a.PostCountPerMonth[t.Month().String()] += 1
			a.addDateCount(t)
			if fromID, ok := mm.FwdFrom.GetFromID(); ok {
				if ch, ok := fromID.(*tg.PeerChannel); ok {
					a.ForwardCount[int(ch.ChannelID)] += 1
				}
				a.TotalForwarded += 1
			}
			last = mm
		}
	}
	a.visited[last.ID] = struct{}{}
	return offSet
}
func (a *Analytics) GetLongestStreak() {
	array := make([]int, 0)
	for _, m := range a.PostCountPerday {
		array = append(array, m...)
	}
	current := 0
	left := 0
	for left < len(array) {
		for left < len(array) && array[left] == 0 {
			left += 1
		}
		right := left
		for right < len(array) && array[right] != 0 {
			right += 1
		}
		current = max(current, right-left)
		left = right
	}
	a.LongestStreak = current
}
