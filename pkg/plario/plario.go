package plario

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

var (
	ErrNoMoreActivity = fmt.Errorf("NoMoreActivity")
)

type Plario struct {
	BaseURL string

	SubjectID int
	CourseID  int
	ModuleID  int

	Culture string
	Token   string
	Attempt int

	logger *slog.Logger
}

func NewPlario(token string, logger *slog.Logger) *Plario {
	return &Plario{
		BaseURL:   "https://api.plario.ru",
		SubjectID: 0,
		CourseID:  0,
		ModuleID:  0,
		Culture:   "ru",
		Token:     token,
		Attempt:   0,
		logger:    logger,
	}

}

func (p *Plario) GetModules(client *http.Client) ([]Module, error) {
	baseURL, err := url.Parse(p.BaseURL + "/learner/module/availableToLearner")
	if err != nil {
		return nil, err
	}

	q := baseURL.Query()
	q.Set("culture", p.Culture)
	q.Set("dateFrom", "2025-11-30T21:00:00.000Z")
	// q.Set("dateTo", time.Now().Format(time.RFC3339))
	q.Set("dateTo", "2025-12-31T20:59:59.999Z")
	q.Set("teacherCourseId", strconv.Itoa(p.CourseID))
	baseURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var modules []Module
	if err := json.NewDecoder(resp.Body).Decode(&modules); err != nil {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("decode err %d: %s", resp.StatusCode, string(b))
	}

	return modules, nil
}

// attempt is based on current activity id
// activity id is just an id of a question
// start activity -> get activity id -> get attempt based on activity -> post answer
func (p *Plario) GetAvailable(client *http.Client) ([]Subject, error) {
	baseURL, err := url.Parse(p.BaseURL + "/learner/course/available")
	if err != nil {
		return nil, err
	}

	q := baseURL.Query()
	q.Set("culture", p.Culture)
	baseURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var subjects []Subject
	if err := json.NewDecoder(resp.Body).Decode(&subjects); err != nil {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("decode err %d: %s", resp.StatusCode, string(b))
	}

	return subjects, nil
}

func (p *Plario) CompleteLesson(client *http.Client, activityID int) error {
	baseURL, err := url.Parse(p.BaseURL + fmt.Sprintf("/learner/adaptiveLearning/completeLesson/%d/%d", activityID, p.Attempt))
	if err != nil {
		return err
	}

	queryParams := baseURL.Query()
	queryParams.Add("moduleId", strconv.Itoa(p.ModuleID))
	queryParams.Add("teacherCourseId", strconv.Itoa(p.CourseID))
	queryParams.Add("culture", p.Culture)
	baseURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("POST", baseURL.String(), nil)
	if err != nil {
		return err
	}

	p.setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("p.CompleteLesson: bad status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
func (p *Plario) PostAnswer(client *http.Client, activityID int, answers []int, secondAttempt bool) (*PlarioAnswerResponse, error) {
	var baseURL *url.URL
	var bPayload []byte
	var err error
	var payload PlarionAnswerRequest

	if !secondAttempt {
		baseURL, err = url.Parse(p.BaseURL + "/learner/adaptiveLearning/checkAnswer")
		if err != nil {
			return nil, err
		}

		payload = PlarionAnswerRequest{
			ActivityID:      activityID,
			AnswerIDs:       answers,
			AttemptID:       p.Attempt,
			ModuleID:        p.ModuleID,
			TeacherCourseID: p.CourseID,
		}

		queryParams := baseURL.Query()
		queryParams.Add("culture", p.Culture)
		baseURL.RawQuery = queryParams.Encode()

		bPayload, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	} else {
		baseURL, err = url.Parse(p.BaseURL + fmt.Sprintf("/learner/adaptiveLearning/answerAttempt/%d/%d", activityID, p.Attempt))
		if err != nil {
			return nil, err
		}

		payload := PlarionAnswerRequest{
			AnswerIDs: answers,
			ModuleID:  p.ModuleID,
		}

		queryParams := baseURL.Query()
		queryParams.Add("culture", p.Culture)
		baseURL.RawQuery = queryParams.Encode()

		bPayload, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("POST", baseURL.String(), bytes.NewBuffer(bPayload))
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// logger.Debug("post_answer",
	// 	"request_url", req.URL.String(),
	// 	"request_method", req.Method,
	// 	"request_payload", string(bPayload),
	// )
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound && strings.Contains(string(body), "ModuleSessionNotFoundOrExpired") {
			_, err := client.Do(req)
			if err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("p.PostAnswer: bad status code %d: %s", resp.StatusCode, string(body))
	}

	var par PlarioAnswerResponse
	if err := json.NewDecoder(resp.Body).Decode(&par); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return &par, nil
}
func (p *Plario) GetAttempt(client *http.Client, activityID int) (int, error) {
	// modules/10/activities/851/attempts?culture=ru
	// https://api.plario.ru/learner/adaptiveLearning/modules/12/activities/343/attempts%3Fculture=ru
	baseURL, err := url.Parse(p.BaseURL + "/learner/adaptiveLearning")
	if err != nil {
		return 0, err
	}

	urlExtra := fmt.Sprintf("/modules/%d/activities/%d/attempts", p.ModuleID, activityID)
	baseURL.Path = path.Join(baseURL.Path, urlExtra)

	q := baseURL.Query()
	q.Set("culture", p.Culture)
	baseURL.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", baseURL.String(), nil)
	if err != nil {
		return 0, err
	}

	p.setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil
	}

	value, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, nil
	}

	return value, nil
}

func (p *Plario) GetQuestion(client *http.Client) (*PlarioQuestionResponse, error) {
	baseURL, err := url.Parse(p.BaseURL + "/learner/adaptiveLearning")
	if err != nil {
		return nil, err
	}

	queryParams := baseURL.Query()
	queryParams.Add("moduleId", strconv.Itoa(p.ModuleID))
	queryParams.Add("teacherCourseId", strconv.Itoa(p.CourseID))
	queryParams.Add("culture", p.Culture)
	baseURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	p.setHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("p.GetQuestion: bad status code %d: %s", resp.StatusCode, string(body))
	}

	var pqr PlarioQuestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&pqr); err != nil {
		return nil, err
	}

	if pqr.ActivityStatus == "NoMoreActivity" {
		return nil, fmt.Errorf("no more activity, too many mistakes")
	}

	return &pqr, nil
}

func (p *Plario) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:145.0) Gecko/20100101 Firefox/145.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.Token))
	req.Header.Set("Origin", "https://my.plario.ru")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://my.plario.ru/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("TE", "trailers")
}
