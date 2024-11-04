package sr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Pagination struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	TotalHits  int `json:"totalhits"`
	TotalPages int `json:"totalpages"`
}

type EpisodesPage struct {
	Pagination Pagination `json:"pagination"`
	Episodes   []Episode  `json:"episodes"`
}

type Episode struct {
	ID                int           `json:"id"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	URL               string        `json:"url"`
	Program           Program       `json:"program"`
	AudioPreference   string        `json:"audiopreference"`
	AudioPriority     string        `json:"audiopriority"`
	AudioPresentation string        `json:"audiopresentation"`
	PublishDate       Time          `json:"publishdateutc"`
	ImageURL          string        `json:"imageurl"`
	ImageURLTemplate  string        `json:"imageurltemplate"`
	Photographer      string        `json:"photographer"`
	Broadcast         *Broadcast    `json:"broadcast,omitempty"`
	BroadcastTime     BroadcastTime `json:"broadcasttime"`
	ChannelID         int           `json:"channelid"`
}

type Program struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Broadcast struct {
	AvailableStop Time            `json:"availablestoputc"`
	Files         []BroadcastFile `json:"broadcastfiles"`
}

type BroadcastFile struct {
	Duration    int    `json:"duration"`
	PublishDate Time   `json:"publishdateutc"`
	ID          int    `json:"id"`
	URL         string `json:"url"`
	StatKey     string `json:"statkey"`
}

type BroadcastTime struct {
	StartTime Time `json:"starttimeutc"`
	EndTime   Time `json:"endtimeutc"`
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		*t = Time{
			Time: time.Time{},
		}
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return nil
	}

	// Expected format: "/Date(1728810000000)/"
	if len(s) != 21 {
		return fmt.Errorf("invalid sr time: invalid length")
	}

	ts, err := strconv.ParseInt(s[6:19], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid sr time: %w", err)
	}

	sec := ts / 1000
	nsec := int64((float64)(ts)/1000 - (float64)(sec))
	*t = Time{
		Time: time.Unix(sec, nsec).UTC(),
	}

	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"/Date(%d)/"`, t.UnixMilli())), nil
}
