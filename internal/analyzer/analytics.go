package analyzer

import (
	"time"

	"github.com/gotd/td/tg"
)

type Analytics struct {
	ChannelName            string           `json:"channel_name,omitempty"`
	TotalViews             int              `json:"total_views,omitempty"`
	TotalComments          int              `json:"total_comments,omitempty"`
	TotalReactions         int              `json:"total_reactions,omitempty"`
	TotalPosts             int              `json:"total_posts,omitempty"`
	TotalForwards          int              `json:"total_forwards,omitempty"`
	ViewsByMonth           map[string]int   `json:"views_by_month,omitempty"`
	ReactionsByType        map[string]int   `json:"reactions_by_type,omitempty"`
	PostsByHour            map[int]int      `json:"posts_by_hour,omitempty"`
	PostsByDay             map[string][]int `json:"posts_by_day,omitempty"`
	PostsByMonth           map[string]int   `json:"posts_by_month,omitempty"`
	MostViewedPostID       int              `json:"most_viewed_post_id,omitempty"`
	MostViewedPostCount    int              `json:"most_viewed_post_count,omitempty"`
	MostCommentedPostID    int              `json:"most_commented_post_id,omitempty"`
	MostCommentedPostCount int              `json:"most_commented_post_count,omitempty"`
	ForwardsFromCount      map[int]int      `json:"forwards_from_count,omitempty"`
	LongestPostingStreak   int              `json:"longest_posting_streak,omitempty"`
	visited                map[int]struct{}
}

func NewAnalytics(name string) Analytics {
	var a Analytics
	a.ChannelName = name
	a.ViewsByMonth = make(map[string]int)
	a.ReactionsByType = make(map[string]int)
	a.PostsByDay = make(map[string][]int)
	a.PostsByMonth = make(map[string]int)
	a.ForwardsFromCount = make(map[int]int)
	a.visited = make(map[int]struct{})
	a.PostsByHour = make(map[int]int)
	return a
}
func (a *Analytics) addDateCount(date time.Time) {
	month := date.Month().String()
	t := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastDay := t.AddDate(0, 0, -1)
	if len(a.PostsByDay[month]) == 0 {
		a.PostsByDay[month] = make([]int, lastDay.Day())
	}
	a.PostsByDay[month][date.Day()-1] += 1
}
func (a *Analytics) updateFromChannelMessages(m *tg.MessagesChannelMessages) int {
	offSet := 0
	last := &tg.Message{}
	for _, msg := range m.Messages {
		if mm, ok := msg.(*tg.Message); ok {
			if _, ok := a.visited[mm.ID]; ok {
				continue
			}
			if a.MostViewedPostID == 0 || a.MostViewedPostCount < mm.Views {
				a.MostViewedPostID = mm.ID
				a.MostViewedPostCount = mm.Views
			}
			if a.MostCommentedPostID == 0 || a.MostCommentedPostCount < mm.Replies.Replies {
				a.MostCommentedPostID = mm.ID
				a.MostCommentedPostCount = mm.Replies.Replies
			}
			offSet = mm.Date
			a.TotalViews += mm.Views
			a.TotalComments += mm.Replies.Replies
			t := getDateTime(mm.Date)
			a.ViewsByMonth[t.Month().String()] += mm.Views
			reactionCounter, totalReactions := (countNumOfReactions(mm.Reactions))
			a.ReactionsByType = mergeMaps(a.ReactionsByType, reactionCounter)
			a.TotalReactions += totalReactions
			a.PostsByMonth[t.Month().String()] += 1
			a.PostsByHour[t.Hour()] += 1
			a.addDateCount(t)
			if fromID, ok := mm.FwdFrom.GetFromID(); ok {
				if ch, ok := fromID.(*tg.PeerChannel); ok {
					a.ForwardsFromCount[int(ch.ChannelID)] += 1
				}
				a.TotalForwards += 1
			}
			last = mm
		}
	}
	a.visited[last.ID] = struct{}{}
	return offSet
}
func (a *Analytics) GetLongestStreak() {
	array := make([]int, 0)
	for _, m := range a.PostsByDay {
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
	a.LongestPostingStreak = current
}
