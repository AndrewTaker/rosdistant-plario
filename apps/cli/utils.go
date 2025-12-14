package main

import (
	"log"
	"log/slog"
	"net/http"
	pl "pkg/plario"
)

// get subjects, courses, modules and mastery levels
func QuizMode(plario *pl.Plario, client *http.Client) error {
	subjects, err := plario.GetAvailable(client)
	if err != nil {
		return err
	}

	courses := make(map[pl.Course][]pl.Module)
	for _, s := range subjects {
		for _, c := range s.Courses {
			plario.CourseID = c.ID
			modules, err := plario.GetModules(client)
			if err != nil {
				return err
			}
			courses[c] = modules
		}
	}
	printSubjectsTable(subjects, courses)
	return nil
}

func InitLogger(level string) *slog.Logger {
	var l slog.Level

	switch level {
	case "info":
		l = slog.LevelInfo
	case "debug":
		l = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(log.Writer(), &slog.HandlerOptions{Level: l})
	return slog.New(handler)
}
