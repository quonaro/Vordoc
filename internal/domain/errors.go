package domain

import "errors"

// ErrDocNotFound is returned when a documentation does not exist.
var ErrDocNotFound = errors.New("documentation not found")

// ErrPageNotFound is returned when a page does not exist.
var ErrPageNotFound = errors.New("page not found")

// ErrPasswordRequired is returned when a page requires a password.
var ErrPasswordRequired = errors.New("password required")

// ErrInvalidPassword is returned when a page password is incorrect.
var ErrInvalidPassword = errors.New("invalid password")

// ErrAccessDenied is returned when access to a page is denied.
var ErrAccessDenied = errors.New("access denied")

// ErrAssetNotFound is returned when a static asset does not exist.
var ErrAssetNotFound = errors.New("asset not found")
