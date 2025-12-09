package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"palario/pkg/database"
	"time"

	"github.com/rodaine/table"
)

var (
	courseID, subjectID, moduleID int

	plarioToken string
)

func main() {
	flag.StringVar(&plarioToken, "ptoken", "", "plario access token")
	flag.Parse()

	if plarioToken == "" {
		log.Fatal("gotta provide valid plario access token")
	}

	db, err := database.New(context.Background(), "plario.db")
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	plario := NewPlario(plarioToken)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// subjects: highest level like "Высшая математика"
	subjects, err := plario.GetAvailable(client)
	if err != nil {
		log.Fatal("GetAvailable err", err)
	}

	// every subjects has an array of courses like "Математический анализ"
	for _, s := range subjects {
		err := db.CreateSubject(s.ID, s.Name)
		if err != nil {
			log.Fatal("db err", err)
		}
		for _, c := range s.Courses {
			err = db.CreateCourse(c.ID, c.Name, s.ID)
			if err != nil {
				log.Fatal("db err", err)
			}
		}
	}

	printCoursesTable(subjects[0].Courses)
	log.Println("enter course id to start:")
	fmt.Scan(&courseID)
	plario.SubjectID = subjects[0].ID
	plario.CourseID = courseID

	// module belongs to a course
	// 	plario subject 								-> Высшая математика
	//	rosdistant course 							-> Математический анализ
	//	learning module (thematic set as in a book) -> Пределы дробнорациональных/иррациональных выражений
	//	------------------------------------------------------------------------------------------------------------
	//	| subject				|	course					|	module												|
	// 	| Высшая математика		|	Математический анализ	|	Пределы дробнорациональных/иррациональных выражений	|
	//	------------------------------------------------------------------------------------------------------------
	// heirerchy is
	//	-> subject
	//		-> []course
	//			-> []module
	modules, err := plario.GetModules(client)
	if err != nil {
		log.Fatal(err)
	}
	if len(modules) == 0 {
		log.Fatal("no modules")
	}

	for _, m := range modules {
		err := db.CreateModule(m.ID, m.Name, courseID)
		if err != nil {
			log.Fatal(err)
		}
	}

	printModulesTable(modules)
	log.Println("enter module id to start:")
	fmt.Scan(&moduleID)
	plario.ModuleID = moduleID

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmsgprefix)
	for {
		randomSleep := RandInRange(r, 5, 10)

		question, err := plario.GetQuestion(client)
		if err != nil {
			logger.Fatal(err)
		}
		logger.SetPrefix(fmt.Sprintf("[C: %d, M: %d, Q: %d] ", plario.CourseID, plario.ModuleID, question.Exercise.ActivityID))
		logger.Println("start")

		attempt, err := plario.GetAttempt(client, question.Exercise.ActivityID)
		if err != nil {
			logger.Fatal(err)
		}
		plario.Attempt = attempt

		if len(question.Exercise.PossibleAnswers) == 0 {
			logger.Println("no answers in response, probably a theory, submitting")
			err := plario.CompleteLesson(client, question.Exercise.ActivityID)
			if err != nil {
				logger.Fatal(err)
			}
			time.Sleep(time.Duration(randomSleep) * time.Second)
			continue
		}

		logger.Println("trying to get answer from db")
		answer, err := db.GetAnswer(question.Exercise.ActivityID, subjectID, courseID, moduleID)
		if err != nil {
			logger.Fatal(err)
		}

		var passedAnswer int
		if answer != 0 {
			logger.Println("there is an answer in database ->", answer)
			passedAnswer = answer
		} else {
			logger.Println("no answer in database")
			index := rand.Intn(len(question.Exercise.PossibleAnswers))
			passedAnswer = question.Exercise.PossibleAnswers[index].AnswerID
		}

		response, err := plario.PostAnswer(client, question.Exercise.ActivityID, []int{passedAnswer}, false)
		if err != nil {
			logger.Fatal(err)
		}

		responseAnswer := response.RightAnswerIDs

		if passedAnswer != responseAnswer[0] {
			logger.Printf("wrong answer, submited %d but right one is %d", passedAnswer, responseAnswer[0])
			logger.Println("submitting again and saving to database", responseAnswer[0])

			err := db.CreateQuestion(
				question.Exercise.ActivityID,
				question.Exercise.Content,
				responseAnswer[0],
				plario.SubjectID,
				plario.CourseID,
				plario.ModuleID,
			)
			if err != nil {
				logger.Fatal(err)
			}
			_, _ = plario.PostAnswer(client, question.Exercise.ActivityID, []int{responseAnswer[0]}, true)
		} else {
			logger.Println("first try hit, saving to database and going further")
			err := db.CreateQuestion(
				question.Exercise.ActivityID,
				question.Exercise.Content,
				passedAnswer,
				plario.SubjectID,
				plario.CourseID,
				plario.ModuleID,
			)
			if err != nil {
				logger.Fatal(err)
			}
		}
		time.Sleep(time.Duration(randomSleep) * time.Second)
	}
}

func RandInRange(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}

func printModulesTable(objs []Module) {
	t := table.New("ID", "NAME", "MASTERY")
	for _, o := range objs {
		t.AddRow(o.ID, o.Name, fmt.Sprintf("%.2f (%.2f%%)", o.Mastery, o.Mastery*100))
	}
	log.Println()
	t.Print()
}

func printSubjectsTable(objs []Subject) {
	t := table.New("ID", "NAME")
	for _, o := range objs {
		for _, s := range o.Courses {
			t.AddRow(s.ID, s.Name)
		}
	}
	log.Println()
	t.Print()
}

func printCoursesTable(objs []Course) {
	t := table.New("ID", "NAME")
	for _, o := range objs {
		t.AddRow(o.ID, o.Name)
	}
	log.Println()
	t.Print()
}
