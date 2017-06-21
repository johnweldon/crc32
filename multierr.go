package main

import (
	"os"
	"strings"
)

func multiClose(err error, fds ...*os.File) error {
	var merr *multiErr
	if err != nil {
		merr = newErr(err)
	}
	for _, e := range fds {
		if e2 := e.Close(); e2 != nil {
			if merr == nil {
				merr = newErr(e2)
			} else {
				merr.Related = append(merr.Related, e2)
			}
		}
	}
	if merr == nil {
		return err
	}
	return merr
}

func newErr(err error) *multiErr { return &multiErr{Primary: err} }

type multiErr struct {
	Primary  error
	CausedBy []error
	Related  []error
}

func (e *multiErr) Error() string {
	if e == nil || e.Primary == nil {
		return "<nil>"
	}
	msgs := []string{}
	msgs = append(msgs, e.Primary.Error())
	if len(e.CausedBy) > 0 {
		msgs = append(msgs, "Caused By:")
		for _, c := range e.CausedBy {
			msgs = append(msgs, "  "+c.Error())
		}
	}
	if len(e.Related) > 0 {
		msgs = append(msgs, "Related:")
		for _, c := range e.Related {
			msgs = append(msgs, "  "+c.Error())
		}
	}
	return strings.Join(msgs, "\n")
}
