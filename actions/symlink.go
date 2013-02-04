package actions

import (
	"os"
)

type Symlink struct {
	// source/destination need better names I think.
	Source      string // This is the original file
	Destination string // This is the link to the original file.
	Result
}

func NewSymlink(source, destination string) *Symlink {
	return &Symlink{Source: source, Destination: destination}
}

func (s *Symlink) Run() {
	symlinkExists, correctSrc := s.CheckSymlink()

	if symlinkExists && correctSrc {
		s.Result.Success = true
		return
	}

	if symlinkExists && correctSrc != true {
		err := os.Remove(s.Destination)
		if err != nil {
			s.Result.Error = err.Error()
			return
		}
		s.Result.Changed = true
	}

	if err := s.CreateSymlink(); err != nil {
		s.Result.Error = err.Error()
		return
	}

	s.Result.Success = true
	return
}

func (s *Symlink) CreateSymlink() error {
	err := os.Symlink(s.Source, s.Destination)
	if err != nil {
		return err
	}
	s.Result.Changed = true
	return nil
}

func (s *Symlink) CheckSymlink() (symlinkExists, correctSrc bool) {
	fi, err := os.Lstat(s.Destination)
	if err != nil {
		return
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		symlinkExists = true
	} else {
		return
	}

	source, err := os.Readlink(s.Destination)
	if err != nil {
		return
	}
	if source == s.Source {
		correctSrc = true
	}
	return
}
