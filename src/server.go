package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

type Server struct {
	config *Config
	log    *zap.SugaredLogger
}

func NewServer(config *Config) *Server {
	var logger *zap.Logger
	var err error

	if config.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}

	return &Server{
		config: config,
		log:    logger.Sugar(),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := url.QueryUnescape(r.URL.EscapedPath())
	if err != nil {
		s.log.Errorf("failed to unescape path: %v", err)
		return
	}
	path = strings.Replace(path, "/search/", "", 1)

	// defaults
	engine := s.config.Engines.Fallback
	prepend := ""

	for _, r := range path {
		if (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') {
			engine = s.config.Engines.Russian
			continue
		}

		if r >= 0x3040 && r <= 0x4db0 {
			engine = s.config.Engines.Japanese
			continue
		}
	}

	// try to find a keyword
	idx := strings.IndexRune(path, ' ')
	if idx != -1 {
		kwd := path[0:idx]
		data, ok := s.config.Keywords[kwd]
		if ok {
			if data.IsLeft() {
				engine = data.MustLeft()
			} else {
				r := data.MustRight()
				engine = r.Engine
				prepend = r.Prepend
			}
			path = path[idx+1:]
		}
	}

	s.log.Debugw("Request", "path", path, "engine", engine, "prepend", prepend)

	newPath, ok := s.config.Engines.URLS[engine]
	if !ok {
		s.log.Errorf("Engine %s not found", engine)
		return
	}

	if prepend != "" {
		newPath += prepend + " " + path
	} else {
		newPath += path
	}

	if s.config.Debug {
		w.Write([]byte(newPath))
	} else {
		w.Header().Add("Location", newPath)
		w.WriteHeader(http.StatusFound)
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf("[::]:%v", s.config.Port), s)
}
