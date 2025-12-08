package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const MODULE_ID = 12           // name of the section, for example "Пределы дробнорациональных/иррациональных выражений"
const TEACHER_COURSE_ID = 2274 // name of the subject in uni, for example "Математический анализ"
const CULTURE = "ru"

func main() {
	token := os.Getenv("TOKEN")
	client := &http.Client{}
	p := NewPlario(MODULE_ID, TEACHER_COURSE_ID, token)

	for {
		question, err := p.GetQuestion(client)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("question", question)
		if len(question.Exercise.PossibleAnswers) == 0 {
			log.Println("no answers in response")
			time.Sleep(5 * time.Second)
			continue
		}

		attempt, err := p.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			log.Fatal(err)
		}
		p.Attempt = attempt

		index := rand.Intn(len(question.Exercise.PossibleAnswers))
		requestAnswer := question.Exercise.PossibleAnswers[index].AnswerID

		response, err := p.PostAnswer(client, question.Exercise.ActivityID, []int{requestAnswer}, false)
		if err != nil {
			log.Fatal(err)
		}

		responseAnswer := response.RightAnswerIDs

		if requestAnswer != responseAnswer[0] {
			log.Printf("wrong answer, submited %d but right one is %d", requestAnswer, responseAnswer[0])
			log.Println("submitting again with", responseAnswer[0])
			_, _ = p.PostAnswer(client, question.Exercise.ActivityID, []int{responseAnswer[0]}, true)
		}
		time.Sleep(5 * time.Second)
	}
}
