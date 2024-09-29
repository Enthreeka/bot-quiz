package query

import (
	"github.com/google/go-querystring/query"
	"net/url"
)

type Query struct {
	QuestionID int `url:"questionID"`
	ChannelID  int `url:"channelID"`
}

func (q *Query) QueryParam() (url.Values, error) {
	params, err := query.Values(q)
	if err != nil {
		return nil, err
	}

	return params, nil
}
