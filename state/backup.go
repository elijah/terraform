package state

import "github.com/hashicorp/terraform/terraform"

// BackupState wraps a State that backs up the state on the first time that
// a WriteState or PersistState is called.
//
// If Path exists, it will be overwritten.
type BackupState struct {
	Real State
	Path string

	done bool
}

func (s *BackupState) State() *terraform.State {
	return s.Real.State()
}

func (s *BackupState) RefreshState() error {
	return s.Real.RefreshState()
}

func (s *BackupState) WriteState(state *terraform.State) error {
	if !s.done {
		if err := s.backup(); err != nil {
			return err
		}
	}

	return s.Real.WriteState(state)
}

func (s *BackupState) PersistState() error {
	if !s.done {
		if err := s.backup(); err != nil {
			return err
		}
	}

	return s.Real.PersistState()
}

// all states get wrapped by BackupState, so it has to be a Locker
func (s *BackupState) Lock(reason string) error {
	if s, ok := s.Real.(Locker); ok {
		return s.Lock(reason)
	}
	return nil
}

func (s *BackupState) Unlock() error {
	if s, ok := s.Real.(Locker); ok {
		return s.Unlock()
	}
	return nil
}

func (s *BackupState) backup() error {
	state := s.Real.State()
	if state == nil {
		if err := s.Real.RefreshState(); err != nil {
			return err
		}

		state = s.Real.State()
	}

	// LocalState.WriteState ensures that a file always exists for locking
	// purposes, but we don't need a backup or lock if the state is empty, so
	// skip this with a nil state.
	if state != nil {
		ls := &LocalState{Path: s.Path}
		if err := ls.WriteState(state); err != nil {
			return err
		}
	}

	s.done = true
	return nil
}
