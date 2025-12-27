package analyzer

import (
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

type OverallMetrics struct {
	TotalViews     int `json:"total_views"`
	TotalComments  int `json:"total_comments"`
	TotalReactions int `json:"total_reactions"`
	TotalPosts     int `json:"total_posts"`
	TotalForwards  int `json:"total_forwards"`
}

func (m *OverallMetrics) UpdateMetrics(msg *tg.Message) {
	_, totalReactions := countNumOfReactions(msg.Reactions)
	m.TotalViews += msg.Views
	m.TotalReactions += totalReactions
	m.TotalComments += msg.Replies.Replies
	m.TotalPosts += 1
	_, ok := msg.FwdFrom.GetFromID()
	if !ok {
		return
	}
	m.TotalForwards += 1
}

type TimeTrends struct {
	ViewsByMonth         map[string]int   `json:"views_by_month,omitempty"`
	PostsByDay           map[string][]int `json:"posts_by_day,omitempty"`
	PostsByMonth         map[string]int   `json:"posts_by_month,omitempty"`
	PostsByHour          map[int]int      `json:"posts_by_hour,omitempty"`
	LongestPostingStreak int              `json:"longest_posting_streak,omitempty"`
}

func (t *TimeTrends) UpdateTrends(mm *tg.Message) {
	dateTime := getDateTime(mm.Date)
	month := dateTime.Month().String()
	monthKey := fmt.Sprintf("%d-%s", dateTime.Year(), month)
	t.PostsByMonth[monthKey] += 1
	t.ViewsByMonth[monthKey] += mm.Views
	t.PostsByHour[dateTime.Hour()] += 1
	tt := time.Date(dateTime.Year(), dateTime.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	if len(t.PostsByDay[monthKey]) == 0 {
		lastDay := tt.AddDate(0, 0, -1)
		t.PostsByDay[monthKey] = make([]int, lastDay.Day())
	}
	t.PostsByDay[monthKey][dateTime.Day()-1] += 1
}

type TopPosts struct {
	MostViewedID       int            `json:"most_viewed_id"`
	MostViewedCount    int            `json:"most_viewed_count"`
	MostCommentedID    int            `json:"most_commented_id"`
	MostCommentedCount int            `json:"most_commented_count"`
	ForwardsBySource   map[int]int    `json:"forwards_by_post"`
	ReactionsByType    map[string]int `json:"reactions_by_type"`
}

func (tp *TopPosts) UpdateTopPosts(msg *tg.Message) {
	if tp.MostViewedID == 0 || tp.MostViewedCount < msg.Views {
		tp.MostViewedID = msg.ID
		tp.MostViewedCount = msg.Views
	}
	if tp.MostCommentedID == 0 || tp.MostCommentedCount < msg.Replies.Replies {
		tp.MostCommentedID = msg.ID
		tp.MostCommentedCount = msg.Replies.Replies
	}
	reactionCounter, _ := (countNumOfReactions(msg.Reactions))
	tp.ReactionsByType = mergeMaps(tp.ReactionsByType, reactionCounter)
	fromID, ok := msg.FwdFrom.GetFromID()
	if !ok {
		return
	}
	if ch, ok := fromID.(*tg.PeerChannel); ok {
		tp.ForwardsBySource[int(ch.ChannelID)] += 1
	}
}

type Analytics struct {
	ChannelProfile string         `json:"channel_profile"`
	ChannelName    string         `json:"channel_name"`
	Totals         OverallMetrics `json:"totals"`
	Trends         TimeTrends     `json:"trends"`
	Highlights     TopPosts       `json:"highlights"`
}

func NewAnalytics(name string) Analytics {
	var a Analytics
	a.ChannelName = name
	a.Trends.ViewsByMonth = make(map[string]int)
	a.Highlights.ReactionsByType = make(map[string]int)
	a.Trends.PostsByDay = make(map[string][]int)
	a.Trends.PostsByMonth = make(map[string]int)
	a.Highlights.ForwardsBySource = make(map[int]int)
	a.Trends.PostsByHour = make(map[int]int)
	return a
}
func (a *Analytics) updateFromChannelMessages(m *tg.MessagesChannelMessages) int {
	if m == nil {
		return 0
	}
	offSet := 0
	startDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	minDateUnix := int(startDate.Unix())
	for i, msg := range m.Messages {
		mm, ok := msg.(*tg.Message)
		if !ok || i == 0 {
			continue
		}
		offSet = mm.Date
		if mm.Date <= minDateUnix {
			return minDateUnix
		}
		a.Highlights.UpdateTopPosts(mm)
		a.Totals.UpdateMetrics(mm)
		a.Trends.UpdateTrends(mm)
	}
	return offSet
}
func (a *Analytics) GetLongestStreak() {
	array := make([]int, 0)
	for _, m := range a.Trends.PostsByDay {
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
	a.Trends.LongestPostingStreak = current
}
