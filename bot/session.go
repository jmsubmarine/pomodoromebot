package bot

import (
	"sync"
	"time"
)

type Phase string

const (
	PhaseWork  Phase = "work"
	PhaseBreak Phase = "break"
)

type PomodoroSession struct {
	CurrentRound  int
	TotalRounds   int
	Timer         *time.Timer
	AwaitingInput bool
	Phase         Phase
	StartTime     time.Time
}

type SessionStore struct {
	sync.Mutex
	sessions map[int64]*PomodoroSession
}

var Store = &SessionStore{
	sessions: make(map[int64]*PomodoroSession),
}

func (s *SessionStore) Get(chatID int64) (*PomodoroSession, bool) {
	s.Lock()
	defer s.Unlock()
	session, exists := s.sessions[chatID]
	return session, exists
}

func (s *SessionStore) Set(chatID int64, session *PomodoroSession) {
	s.Lock()
	defer s.Unlock()
	s.sessions[chatID] = session
}

func (s *SessionStore) Delete(chatID int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.sessions, chatID)
}
