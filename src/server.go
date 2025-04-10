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
	path = strings.TrimPrefix(path, "/")

	if strings.HasPrefix(path, "opensearch") {
		if path == "opensearch.xml" {
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
					<ShortName>Универсальный поиск</ShortName>
					<Description>Универсальный поиск ищет правильнее других</Description>
					<Url rel="self" type="application/opensearchdescription+xml" template="http://unisearch:8080/opensearch.xml"/>
					<Url type="application/x-suggestions+json" method="GET" template="https://suggest.yandex.com/suggest-ff.cgi?part={searchTerms}"/>
					<Url type="text/html" template="http://unisearch:8080/{searchTerms}"/>
				</OpenSearchDescription>
				`))
		} else {
			w.Write([]byte(`<!DOCTYPE html>
				<html lang="en">
				<head>
					<link rel="search" type="application/opensearchdescription+xml" title="Ultisearch" href="opensearch.xml"/>
				</head>
				<body>Please add the engine with ^click on the URL</body>
				</html>`))
		}
		return
	}

	// defaults
	engine := s.config.Default
	prepend := ""
	var replacements []Replacement = nil

	for _, r := range path {
		for _, conf := range s.config.Runes {
			if r >= conf.FromRune && r <= conf.ToRune {
				engine = conf.Engine
				break
			}
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
				prepend = ""
				replacements = nil
			} else {
				r := data.MustRight()
				engine = r.Engine
				prepend = r.Prepend
				replacements = r.Replace
			}
			path = path[idx+1:]
		}
	}

	s.log.Debugw("Request", "path", path, "engine", engine, "prepend", prepend)

	newPath, ok := s.config.Engines[engine]
	if !ok {
		s.log.Errorf("Engine %s not found", engine)
		return
	}

	if replacements != nil {
		for _, repl := range replacements {
			path = strings.Replace(path, repl.From, repl.To, -1)
		}
	}

	path = url.PathEscape(path)

	if prepend != "" {
		newPath += prepend + " " + path
	} else {
		newPath += path
	}

	if s.config.Debug {
		_, _ = w.Write([]byte(newPath))
	} else {
		w.Header().Add("Location", newPath)
		w.WriteHeader(http.StatusFound)
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf("[::]:%v", s.config.Port), s)
}
