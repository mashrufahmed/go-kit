package uuid

import guuid "github.com/google/uuid"

func New() guuid.UUID                        { return guuid.New() }
func NewString() string                      { return guuid.NewString() }
func Parse(value string) (guuid.UUID, error) { return guuid.Parse(value) }
func Validate(value string) bool             { _, err := Parse(value); return err == nil }
