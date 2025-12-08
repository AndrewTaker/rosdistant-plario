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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		randomSleep := RandInRange(r, 5, 10)
		log.Println("random sleep is set to", randomSleep)

		question, err := p.GetQuestion(client)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("question", question)
		attempt, err := p.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			log.Fatal(err)
		}
		p.Attempt = attempt
		if len(question.Exercise.PossibleAnswers) == 0 {
			log.Println("no answers in response, probably a theory, subbmiting")
			err := p.CompleteLesson(client, question.Exercise.ActivityID)
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Duration(randomSleep) * time.Second)
			continue
		}

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
		time.Sleep(time.Duration(randomSleep) * time.Second)
		log.Println("-----------------------------------")
	}
}

func RandInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}
