package entity

type Args struct {
	Answers []AnswerArgs `json:"варианты_ответы"`
}

type AnswerArgs struct {
	Answer string `json:"ответ"`
	Cost   int    `json:"цена_ответа"`
}
