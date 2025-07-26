package markdown

import (
	"errors"

	"github.com/gomarkdown/markdown/ast"
)

var ErrHeadingNotFound = errors.New("heading not found for path")

// FindHeadingByPath finds a heading node by HeadingPath.
func FindHeadingByPath(doc ast.Node, path *HeadingPath) (*ast.Heading, error) {
	res, err := NavigateHeadingPath(doc, path)
	if err != nil {
		return nil, err
	}
	if res.TargetHeading == nil {
		return nil, ErrHeadingNotFound
	}
	return res.TargetHeading, nil
}
