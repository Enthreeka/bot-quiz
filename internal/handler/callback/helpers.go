package callback

import (
	"github.com/Enthreeka/tg-bot-quiz/internal/entity"
	"strconv"
	"strings"
)

func GetThirdValue(data string) int {
	parts := strings.Split(data, "_")
	if len(parts) > 3 {
		return 0
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}

	return id
}

func GetThirdValueString(data string) string {
	parts := strings.Split(data, "_")
	if len(parts) > 3 {
		return ""
	}

	return parts[2]
}

func AnswerToArgsModel(answer []entity.Answer) *entity.Args {
	if answer == nil || len(answer) == 0 {
		return &entity.Args{}
	}
	args := new(entity.Args)
	args.Answers = make([]entity.AnswerArgs, len(answer))
	for key, value := range answer {
		args.Answers[key] = entity.AnswerArgs{
			Answer: value.Answer,
			Cost:   value.CostOfResponse,
		}
	}

	return args
}
