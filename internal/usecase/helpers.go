package service

import "github.com/Enthreeka/tg-bot-quiz/internal/entity"

func updateArgsToModel(args entity.Args) []entity.Answer {
	answer := make([]entity.Answer, len(args.Answers))
	for key, value := range args.Answers {
		answer[key] = entity.Answer{
			Answer:         value.Answer,
			CostOfResponse: value.Cost,
		}
	}

	return answer
}
