package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"palario/pkg/llm"
	"strconv"
	"time"

	"github.com/rodaine/table"
)

var (
	courseID, subjectID, moduleID int

	plarioToken string

	totalCorrent, totalWrong int
)

func main() {
	flag.StringVar(&plarioToken, "ptoken", "", "plario access token")
	flag.Parse()

	if plarioToken == "" {
		log.Fatal("gotta provide valid plario access token")
	}

	groq := llm.NewGroq(
		os.Getenv("GROQ_TOKEN"),
		"openai/gpt-oss-120b",
		"Ты решаешь тест по математическому анализу. Ты получишь вопрос и множество ответов в формате latex, тебе нужно вернуть только id ответа",
	)

	client := &http.Client{}
	plario := NewPlario(plarioToken)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	subjects, err := plario.GetAvailable(client)
	if err != nil {
		log.Fatal("GetAvailable err", err)
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

		logger.SetPrefix(Info(fmt.Sprintf("[C: %d, M: %d, Q: %d] ", plario.CourseID, plario.ModuleID, question.Exercise.ActivityID)))

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

		logger.Println("sending question to groq")
		groqResponse, err := groq.SendGroqRequest(client, question.Exercise.ToString())
		if err != nil {
			logger.Fatal(err)
		}

		answer, _ := strconv.Atoi(groqResponse.Choices[0].Message.Content)

		response, err := plario.PostAnswer(client, question.Exercise.ActivityID, []int{answer}, false)
		if err != nil {
			logger.Fatal(err)
		}

		rightAnswer := response.RightAnswerIDs[0]
		logger.Printf(Important("groq -> %d, righ -> %d", answer, rightAnswer))

		if answer != rightAnswer {
			logger.Printf(Warn("wrong answer, submited %d but right one is %d, will submite again and go further", answer, rightAnswer))
			totalWrong++
			_, _ = plario.PostAnswer(client, question.Exercise.ActivityID, []int{rightAnswer}, true)
		} else {
			logger.Println(Success("first try hit"))
			totalCorrent++
		}
		time.Sleep(time.Duration(randomSleep) * time.Second)
	}

	logger.Println(Info("Total correct:", totalCorrent))
	logger.Println(Info("Total wrong:", totalWrong))
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
