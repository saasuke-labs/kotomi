package comments

import (
	"sync"
	"time"
)

func NewSitePagesIndex() *SitePagesIndex {
	return &SitePagesIndex{
		data: make(map[string]map[string][]Comment),
	}
}

// Comment represents a comment or a reply.
type Comment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	ParentID  string    `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SitePagesIndex struct {
	sync.RWMutex
	data map[string]map[string][]Comment
}

func (s *SitePagesIndex) AddPageComment(site, page string, comment Comment) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.data[site]; !ok {
		s.data[site] = make(map[string][]Comment)
	}
	if _, ok := s.data[site][page]; !ok {
		s.data[site][page] = []Comment{}
	}

	s.data[site][page] = append(s.data[site][page], comment)
}

func (s *SitePagesIndex) GetPageComments(site, page string) []Comment {
	s.RLock()
	defer s.RUnlock()
	if _, ok := s.data[site]; !ok {
		return []Comment{}
	}
	if _, ok := s.data[site][page]; !ok {
		return []Comment{}
	}

	return s.data[site][page]
}
